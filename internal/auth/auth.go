package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

type Service struct{ DB *pgxpool.Pool }

func PublicKey(raw string) (ssh.PublicKey, string, error) {
	if len(raw) > 8192 {
		return nil, "", security.ErrInvalid
	}
	key, _, options, rest, err := ssh.ParseAuthorizedKey([]byte(raw))
	if err != nil || len(options) != 0 || len(strings.TrimSpace(string(rest))) != 0 {
		return nil, "", security.ErrInvalid
	}
	// Hardware keys and RSA policy can be added after interoperability tests.
	if key.Type() != ssh.KeyAlgoED25519 {
		return nil, "", security.ErrInvalid
	}
	return key, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key))), nil
}

func (s *Service) Challenge(ctx context.Context, r protocol.ChallengeRequest) (protocol.Challenge, error) {
	_, canonical, err := PublicKey(r.PublicKey)
	if err != nil {
		return protocol.Challenge{}, err
	}
	if r.Username != "" && (!security.Identifier(r.Username) || len(r.Invitation) > 256 || r.Invitation == "") {
		return protocol.Challenge{}, security.ErrInvalid
	}
	c := protocol.Challenge{ID: security.ID(), Payload: "TERMLINK-AUTH-v1\n" + security.ID()}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return c, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092632)"); err != nil {
		return c, err
	}
	if _, err = tx.Exec(ctx, "DELETE FROM auth_challenges WHERE expires_at <= now()"); err != nil {
		return c, err
	}
	var count int
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM auth_challenges").Scan(&count); err != nil {
		return c, err
	}
	if count >= 4096 {
		return c, security.ErrLimited
	}
	_, err = tx.Exec(ctx, `INSERT INTO auth_challenges(id,payload,public_key,username,invitation_hash,expires_at) VALUES($1,$2,$3,$4,$5,now()+interval '2 minutes')`, c.ID, c.Payload, canonical, r.Username, security.Hash(r.Invitation))
	if err != nil {
		return c, err
	}
	return c, tx.Commit(ctx)
}

func (s *Service) Finish(ctx context.Context, p protocol.Proof) (protocol.Session, error) {
	if !security.Identifier(p.ID) || len(p.Signature) > 2048 {
		return protocol.Session{}, security.ErrInvalid
	}
	var payload, raw, username, invite string
	// Autocommit the one-shot deletion before checking the proof. Failed signup
	// must not roll back consumption of the authentication challenge.
	err := s.DB.QueryRow(ctx, `DELETE FROM auth_challenges WHERE id=$1 AND expires_at>now() RETURNING payload,public_key,username,invitation_hash`, p.ID).Scan(&payload, &raw, &username, &invite)
	if err != nil {
		return protocol.Session{}, security.ErrUnauthorized
	}
	// Consume even invalid proofs so a challenge permits only one attempt.
	key, _, err := PublicKey(raw)
	var sig ssh.Signature
	if err != nil || ssh.Unmarshal(p.Signature, &sig) != nil || key.Verify([]byte(payload), &sig) != nil {
		return protocol.Session{}, security.ErrUnauthorized
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.Session{}, err
	}
	defer tx.Rollback(ctx)
	var uid, kid string
	if username != "" {
		var conversation *string
		err = tx.QueryRow(ctx, `UPDATE invitations SET remaining=remaining-1 WHERE token_hash=$1 AND remaining>0 AND expires_at>now() RETURNING conversation_id`, invite).Scan(&conversation)
		if err != nil {
			return protocol.Session{}, security.ErrUnauthorized
		}
		uid = security.ID()
		kid = security.ID()
		if _, err = tx.Exec(ctx, "INSERT INTO users(id,username) VALUES($1,$2)", uid, username); err != nil {
			return protocol.Session{}, security.ErrInvalid
		}
		if _, err = tx.Exec(ctx, "INSERT INTO ssh_keys(id,user_id,public_key,fingerprint) VALUES($1,$2,$3,$4)", kid, uid, raw, ssh.FingerprintSHA256(key)); err != nil {
			return protocol.Session{}, security.ErrInvalid
		}
		if conversation != nil {
			if _, err = tx.Exec(ctx, "INSERT INTO memberships(conversation_id,user_id) VALUES($1,$2)", *conversation, uid); err != nil {
				return protocol.Session{}, err
			}
		}
	} else {
		err = tx.QueryRow(ctx, `SELECT k.user_id,k.id FROM ssh_keys k JOIN users u ON u.id=k.user_id WHERE k.public_key=$1 AND k.revoked_at IS NULL AND u.active FOR SHARE OF k,u`, raw).Scan(&uid, &kid)
		if err != nil {
			return protocol.Session{}, security.ErrUnauthorized
		}
	}
	session, err := issue(ctx, tx, uid, kid)
	if err != nil {
		return session, err
	}
	return session, tx.Commit(ctx)
}

func issue(ctx context.Context, tx pgx.Tx, uid, kid string) (protocol.Session, error) {
	session := protocol.Session{Token: security.ID(), UserID: uid, ExpiresAt: time.Now().UTC().Add(8 * time.Hour)}
	_, err := tx.Exec(ctx, "INSERT INTO sessions(token_hash,user_id,key_id,expires_at) VALUES($1,$2,$3,$4)", security.Hash(session.Token), uid, kid, session.ExpiresAt)
	return session, err
}

// SSHSession may only be called after the SSH library has verified possession.
func (s *Service) SSHSession(ctx context.Context, key ssh.PublicKey) (protocol.Session, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.Session{}, err
	}
	defer tx.Rollback(ctx)
	var uid, kid string
	err = tx.QueryRow(ctx, `SELECT k.user_id,k.id FROM ssh_keys k JOIN users u ON u.id=k.user_id WHERE k.public_key=$1 AND k.revoked_at IS NULL AND u.active FOR SHARE OF k,u`, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))).Scan(&uid, &kid)
	if err != nil {
		return protocol.Session{}, security.ErrUnauthorized
	}
	session, err := issue(ctx, tx, uid, kid)
	if err != nil {
		return session, err
	}
	return session, tx.Commit(ctx)
}

func (s *Service) KnownKey(ctx context.Context, key ssh.PublicKey) bool {
	var found bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ssh_keys k JOIN users u ON u.id=k.user_id WHERE k.public_key=$1 AND k.revoked_at IS NULL AND u.active)`, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))).Scan(&found)
	return err == nil && found
}

// Principal locks authorization rows until the operation commits.
func Principal(ctx context.Context, tx pgx.Tx, token string) (string, error) {
	if len(token) != 48 {
		return "", security.ErrUnauthorized
	}
	var uid string
	err := tx.QueryRow(ctx, `SELECT s.user_id FROM sessions s JOIN users u ON u.id=s.user_id JOIN ssh_keys k ON k.id=s.key_id WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at>now() AND k.revoked_at IS NULL AND u.active FOR SHARE OF s,u,k`, security.Hash(token)).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", security.ErrUnauthorized
	}
	return uid, err
}

func (s *Service) Logout(ctx context.Context, token string) error {
	_, err := s.DB.Exec(ctx, "UPDATE sessions SET revoked_at=now() WHERE token_hash=$1", security.Hash(token))
	return err
}
func (s *Service) Invite(ctx context.Context, conversation string, uses int, ttl time.Duration) (string, error) {
	if uses < 1 || uses > 100 || ttl <= 0 || ttl > 30*24*time.Hour {
		return "", security.ErrInvalid
	}
	token := security.ID()
	var conv *string
	if conversation != "" {
		conv = &conversation
	}
	_, err := s.DB.Exec(ctx, "INSERT INTO invitations(token_hash,conversation_id,remaining,expires_at) VALUES($1,$2,$3,$4)", security.Hash(token), conv, uses, time.Now().UTC().Add(ttl))
	return token, err
}
func (s *Service) RevokeKey(ctx context.Context, fingerprint string) error {
	_, err := s.DB.Exec(ctx, "UPDATE ssh_keys SET revoked_at=now() WHERE fingerprint=$1", fingerprint)
	return err
}
