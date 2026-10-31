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
	DB    *pgxpool.Pool
	Sends *security.Limiter
}

func New(db *pgxpool.Pool) *Service { return &Service{DB: db, Sends: security.NewLimiter(2, 10)} }
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
	rows, err := tx.Query(ctx, "SELECT c.id,c.name,c.kind FROM conversations c JOIN memberships m ON m.conversation_id=c.id WHERE m.user_id=$1 ORDER BY c.name,c.id LIMIT 100 FOR SHARE OF m", uid)
	if err != nil {
		return nil, err
	}
	out := []protocol.Conversation{}
	for rows.Next() {
		var c protocol.Conversation
		if err = rows.Scan(&c.ID, &c.Name, &c.Kind); err != nil {
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
	rows, err := tx.Query(ctx, `SELECT m.id,m.client_id,m.conversation_id,m.sender_id,u.username,m.seq,m.body,m.created_at FROM messages m JOIN users u ON u.id=m.sender_id WHERE conversation_id=$1 AND seq>$2 ORDER BY seq LIMIT $3`, conv, after, limit)
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
	err = tx.QueryRow(ctx, "SELECT next_seq FROM conversations WHERE id=$1 FOR UPDATE", conv).Scan(&seq)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, security.ErrForbidden
	}
	if err != nil {
		return m, err
	}
	if err = member(ctx, tx, uid, conv); err != nil {
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

// Direct reuses a canonical pair. All operations still use membership authorization.
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
	c = protocol.Conversation{ID: security.ID(), Name: username, Kind: "direct"}
	err = tx.QueryRow(ctx, `INSERT INTO conversations(id,name,kind,direct_pair) VALUES($1,$2,'direct',$3) ON CONFLICT(direct_pair) DO UPDATE SET direct_pair=excluded.direct_pair RETURNING id,name,kind`, c.ID, "dm-"+strings.Join([]string{a[:6], b[:6]}, "-"), pair).Scan(&c.ID, &c.Name, &c.Kind)
	if err != nil {
		return c, err
	}
	_, err = tx.Exec(ctx, "INSERT INTO memberships(conversation_id,user_id) VALUES($1,$2),($1,$3) ON CONFLICT DO NOTHING", c.ID, uid, peer)
	if err != nil {
		return c, err
	}
	return c, tx.Commit(ctx)
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
