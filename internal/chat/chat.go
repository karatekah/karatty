package chat

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/karatekah/termlink/internal/auth"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
)

type Service struct {
	DB       *pgxpool.Pool
	Sends    *security.Limiter
	Requests *security.Limiter
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db, Sends: security.NewLimiter(2, 10), Requests: security.NewLimiter(1.0/60, 3)}
}
func member(ctx context.Context, tx pgx.Tx, uid, conv string) error {
	var found string
	err := tx.QueryRow(ctx, "SELECT user_id FROM memberships WHERE conversation_id=$1 AND user_id=$2 FOR SHARE", conv, uid).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return security.ErrForbidden
	}
	return err
}
func (s *Service) List(ctx context.Context, token string) ([]protocol.Conversation, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	uid, err := auth.Principal(ctx, tx, token)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT c.id,
		CASE WHEN c.kind='direct' AND c.direct_pair IS NOT NULL THEN
			(SELECT username FROM users WHERE id=CASE WHEN split_part(c.direct_pair,':',1)=$1
				THEN split_part(c.direct_pair,':',2) ELSE split_part(c.direct_pair,':',1) END)
		ELSE c.name END,
		c.kind,c.closed_at IS NOT NULL
		FROM conversations c JOIN memberships m ON m.conversation_id=c.id
		WHERE m.user_id=$1 ORDER BY c.name,c.id LIMIT 100 FOR SHARE OF m`, uid)
	if err != nil {
		return nil, err
	}
	out := []protocol.Conversation{}
	for rows.Next() {
		var c protocol.Conversation
		if err = rows.Scan(&c.ID, &c.Name, &c.Kind, &c.Closed); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, c)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return out, tx.Commit(ctx)
}
func (s *Service) History(ctx context.Context, token, conv string, after int64, limit int) (protocol.Page, error) {
	page := protocol.Page{Messages: []protocol.Message{}, Next: after}
	if !security.Identifier(conv) || after < 0 || limit < 1 || limit > 100 {
		return page, security.ErrInvalid
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return page, err
	}
	defer tx.Rollback(ctx)
	uid, err := auth.Principal(ctx, tx, token)
	if err != nil {
		return page, err
	}
	if err = member(ctx, tx, uid, conv); err != nil {
		return page, err
	}
	rows, err := tx.Query(ctx, `SELECT m.id,m.client_id,m.conversation_id,m.sender_id,u.username,m.seq,m.body,m.created_at FROM messages m JOIN users u ON u.id=m.sender_id WHERE conversation_id=$1 AND seq>$2 AND m.created_at >= now() - interval '720 hours' ORDER BY seq LIMIT $3`, conv, after, limit)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var m protocol.Message
		if err = rows.Scan(&m.ID, &m.ClientID, &m.ConversationID, &m.SenderID, &m.Username, &m.Seq, &m.Body, &m.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Messages = append(page.Messages, m)
		page.Next = m.Seq
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return page, err
	}
	return page, tx.Commit(ctx)
}
func (s *Service) Send(ctx context.Context, token, conv string, r protocol.Send) (protocol.Message, error) {
	var m protocol.Message
	if !security.Identifier(conv) || !security.Identifier(r.ClientID) || !security.Body(r.Body) {
		return m, security.ErrInvalid
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return m, err
	}
	defer tx.Rollback(ctx)
	uid, err := auth.Principal(ctx, tx, token)
	if err != nil {
		return m, err
	}
	// Shared lock order with membership revocation: conversation, then membership.
	var seq int64
	var closed bool
	err = tx.QueryRow(ctx, "SELECT next_seq,closed_at IS NOT NULL FROM conversations WHERE id=$1 FOR UPDATE", conv).Scan(&seq, &closed)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, security.ErrForbidden
	}
	if err != nil {
		return m, err
	}
	if err = member(ctx, tx, uid, conv); err != nil {
		return m, err
	}
	if closed {
		return m, security.ErrForbidden
	}
	// An expired ID can be reused only after removing its old message in this
	// transaction. No expired body is returned through the retry ACK.
	if _, err = tx.Exec(ctx, `DELETE FROM messages WHERE conversation_id=$1 AND sender_id=$2 AND client_id=$3 AND created_at < now() - interval '720 hours'`, conv, uid, r.ClientID); err != nil {
		return m, err
	}
	err = tx.QueryRow(ctx, `SELECT m.id,m.client_id,m.conversation_id,m.sender_id,u.username,m.seq,m.body,m.created_at FROM messages m JOIN users u ON u.id=m.sender_id WHERE conversation_id=$1 AND sender_id=$2 AND client_id=$3`, conv, uid, r.ClientID).Scan(&m.ID, &m.ClientID, &m.ConversationID, &m.SenderID, &m.Username, &m.Seq, &m.Body, &m.CreatedAt)
	if err == nil {
		if m.Body != r.Body {
			return m, security.ErrConflict
		}
		return m, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return m, err
	}
	if !s.Sends.Allow(uid) {
		return m, security.ErrLimited
	}
	if _, err = tx.Exec(ctx, "UPDATE conversations SET next_seq=next_seq+1 WHERE id=$1", conv); err != nil {
		return m, err
	}
	m = protocol.Message{ID: security.ID(), ClientID: r.ClientID, ConversationID: conv, SenderID: uid, Seq: seq + 1, Body: r.Body}
	err = tx.QueryRow(ctx, `INSERT INTO messages(id,client_id,conversation_id,sender_id,seq,body) VALUES($1,$2,$3,$4,$5,$6) RETURNING created_at`, m.ID, m.ClientID, conv, uid, m.Seq, m.Body).Scan(&m.CreatedAt)
	if err != nil {
		return m, err
	}
	if err = tx.QueryRow(ctx, "SELECT username FROM users WHERE id=$1", uid).Scan(&m.Username); err != nil {
		return m, err
	}
	return m, tx.Commit(ctx)
}

// Direct creates a request; a reciprocal call accepts it and opens the pair.
func (s *Service) Direct(ctx context.Context, token, username string) (protocol.Conversation, error) {
	var c protocol.Conversation
	if !security.Identifier(username) {
		return c, security.ErrInvalid
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return c, err
	}
	defer tx.Rollback(ctx)
	uid, err := auth.Principal(ctx, tx, token)
	if err != nil {
		return c, err
	}
	var peer string
	if err = tx.QueryRow(ctx, "SELECT id FROM users WHERE username=$1 AND active", username).Scan(&peer); err != nil {
		return c, security.ErrInvalid
	}
	if uid == peer {
		return c, security.ErrInvalid
	}
	a, b := uid, peer
	if a > b {
		a, b = b, a
	}
	pair := a + ":" + b
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092636,hashtext($1))", pair); err != nil {
		return c, err
	}
	var blocked bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM direct_blocks
		WHERE (blocker_id=$1 AND blocked_id=$2) OR (blocker_id=$2 AND blocked_id=$1))`, uid, peer).Scan(&blocked); err != nil {
		return c, err
	}
	if blocked {
		return c, security.ErrForbidden
	}
	err = tx.QueryRow(ctx, "SELECT id,name,kind,closed_at IS NOT NULL FROM conversations WHERE direct_pair=$1 FOR UPDATE", pair).Scan(&c.ID, &c.Name, &c.Kind, &c.Closed)
	if err == nil {
		// Reopening an existing direct must never restore revoked membership.
		if err = member(ctx, tx, uid, c.ID); err != nil {
			return c, err
		}
		if err = member(ctx, tx, peer, c.ID); err != nil {
			return c, err
		}
		c.Name = username
		return c, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return c, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM direct_requests WHERE expires_at<=now()
		AND ((requester_id=$1 AND recipient_id=$2) OR (requester_id=$2 AND recipient_id=$1))`, uid, peer); err != nil {
		return c, err
	}
	var requestID string
	err = tx.QueryRow(ctx, "SELECT id FROM direct_requests WHERE requester_id=$1 AND recipient_id=$2 FOR UPDATE", peer, uid).Scan(&requestID)
	if err == nil {
		c = protocol.Conversation{ID: security.ID(), Name: username, Kind: "direct"}
		if err = tx.QueryRow(ctx, `INSERT INTO conversations(id,name,kind,direct_pair) VALUES($1,$2,'direct',$3) RETURNING id,name,kind,closed_at IS NOT NULL`, c.ID, "dm-"+strings.Join([]string{a[:6], b[:6]}, "-"), pair).Scan(&c.ID, &c.Name, &c.Kind, &c.Closed); err != nil {
			return c, err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO memberships(conversation_id,user_id) VALUES($1,$2),($1,$3)", c.ID, uid, peer); err != nil {
			return c, err
		}
		if _, err = tx.Exec(ctx, "DELETE FROM direct_requests WHERE (requester_id=$1 AND recipient_id=$2) OR (requester_id=$2 AND recipient_id=$1)", uid, peer); err != nil {
			return c, err
		}
		c.Name = username
		return c, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return c, err
	}
	var existing string
	err = tx.QueryRow(ctx, "SELECT id FROM direct_requests WHERE requester_id=$1 AND recipient_id=$2", uid, peer).Scan(&existing)
	if err == nil {
		c = protocol.Conversation{Name: username, Kind: "direct", Pending: true}
		return c, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return c, err
	}
	if !s.Requests.Allow(uid) {
		return c, security.ErrLimited
	}
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092635,hashtext($1))", uid); err != nil {
		return c, err
	}
	var count int
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM direct_requests WHERE requester_id=$1 AND expires_at>now()", uid).Scan(&count); err != nil {
		return c, err
	}
	if count >= 10 {
		return c, security.ErrLimited
	}
	if _, err = tx.Exec(ctx, "INSERT INTO direct_requests(id,requester_id,recipient_id) VALUES($1,$2,$3)", security.ID(), uid, peer); err != nil {
		return c, err
	}
	c = protocol.Conversation{Name: username, Kind: "direct", Pending: true}
	return c, tx.Commit(ctx)
}

func (s *Service) DirectRequests(ctx context.Context, token string) ([]protocol.DirectRequest, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	uid, err := auth.Principal(ctx, tx, token)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT u.username,r.created_at FROM direct_requests r
		JOIN users u ON u.id=r.requester_id
		WHERE r.recipient_id=$1 AND r.expires_at>now() AND u.active
		ORDER BY r.created_at,r.id LIMIT 100`, uid)
	if err != nil {
		return nil, err
	}
	requests := []protocol.DirectRequest{}
	for rows.Next() {
		var request protocol.DirectRequest
		if err = rows.Scan(&request.Username, &request.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		requests = append(requests, request)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return requests, tx.Commit(ctx)
}

// DeclineDirect discards an incoming request without opening a conversation.
func (s *Service) DeclineDirect(ctx context.Context, token, username string) error {
	return s.changeDirectAccess(ctx, token, username, "decline")
}

// BlockDirect removes both requests and closes an existing direct conversation.
func (s *Service) BlockDirect(ctx context.Context, token, username string) error {
	return s.changeDirectAccess(ctx, token, username, "block")
}

func (s *Service) UnblockDirect(ctx context.Context, token, username string) error {
	return s.changeDirectAccess(ctx, token, username, "unblock")
}

func (s *Service) changeDirectAccess(ctx context.Context, token, username, action string) error {
	if !security.Identifier(username) {
		return security.ErrInvalid
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	uid, err := auth.Principal(ctx, tx, token)
	if err != nil {
		return err
	}
	var peer string
	if err = tx.QueryRow(ctx, "SELECT id FROM users WHERE username=$1 AND active", username).Scan(&peer); err != nil {
		return security.ErrInvalid
	}
	if uid == peer {
		return security.ErrInvalid
	}
	a, b := uid, peer
	if a > b {
		a, b = b, a
	}
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092636,hashtext($1))", a+":"+b); err != nil {
		return err
	}
	switch action {
	case "decline":
		_, err = tx.Exec(ctx, "DELETE FROM direct_requests WHERE requester_id=$1 AND recipient_id=$2", peer, uid)
	case "block":
		if _, err = tx.Exec(ctx, "INSERT INTO direct_blocks(blocker_id,blocked_id) VALUES($1,$2) ON CONFLICT DO NOTHING", uid, peer); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "DELETE FROM direct_requests WHERE (requester_id=$1 AND recipient_id=$2) OR (requester_id=$2 AND recipient_id=$1)", uid, peer); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "UPDATE conversations SET direct_pair=NULL,closed_at=now(),name='Conversación bloqueada' WHERE direct_pair=$1", a+":"+b)
	case "unblock":
		_, err = tx.Exec(ctx, "DELETE FROM direct_blocks WHERE blocker_id=$1 AND blocked_id=$2", uid, peer)
	default:
		return security.ErrInvalid
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) Channel(ctx context.Context, name string) (string, error) {
	if !security.Identifier(name) {
		return "", security.ErrInvalid
	}
	id := security.ID()
	_, err := s.DB.Exec(ctx, "INSERT INTO conversations(id,name,kind) VALUES($1,$2,'channel')", id, name)
	return id, err
}
func (s *Service) RevokeMember(ctx context.Context, conv, uid string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT id FROM conversations WHERE id=$1 FOR UPDATE", conv); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "DELETE FROM memberships WHERE conversation_id=$1 AND user_id=$2", conv, uid); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
