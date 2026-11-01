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

type Service struct {
	DB           *pgxpool.Pool
	PublicSignup bool
}

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
	if r.Username != "" {
		if !security.Identifier(r.Username) || len(r.Invitation) > 256 || (r.Invitation == "" && !s.PublicSignup) || (r.Invitation != "" && s.PublicSignup) {
			return protocol.Challenge{}, security.ErrInvalid
		}
	} else if r.Invitation != "" {
		return protocol.Challenge{}, security.ErrInvalid
	}
	prefix := "TERMLINK-AUTH-v1\n"
	if r.Username != "" && s.PublicSignup {
		prefix = "KARATTY-PUBLIC-REGISTER-v1\n"
	}
	c := protocol.Challenge{ID: security.ID(), Payload: prefix + security.ID()}
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
	expectedPrefix := "TERMLINK-AUTH-v1\n"
	if username != "" && s.PublicSignup {
		expectedPrefix = "KARATTY-PUBLIC-REGISTER-v1\n"
	}
	if !strings.HasPrefix(payload, expectedPrefix) {
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
	var uid, kid, recoveryCode string
	if username != "" {
		var conversation *string
		if !s.PublicSignup {
			err = tx.QueryRow(ctx, `UPDATE invitations SET remaining=remaining-1 WHERE token_hash=$1 AND remaining>0 AND expires_at>now() RETURNING conversation_id`, invite).Scan(&conversation)
			if err != nil {
				return protocol.Session{}, security.ErrUnauthorized
			}
		}
		uid = security.ID()
		kid = security.ID()
		recoveryCode = security.ID()
		if _, err = tx.Exec(ctx, "INSERT INTO users(id,username,recovery_hash) VALUES($1,$2,$3)", uid, username, security.Hash(recoveryCode)); err != nil {
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
	session.RecoveryCode = recoveryCode
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

// PrincipalKey returns the active SSH key bound to the current session.
func PrincipalKey(ctx context.Context, tx pgx.Tx, token string) (string, string, string, error) {
	if len(token) != 48 {
		return "", "", "", security.ErrUnauthorized
	}
	var uid, kid, publicKey string
	err := tx.QueryRow(ctx, `SELECT s.user_id,s.key_id,k.public_key FROM sessions s
		JOIN users u ON u.id=s.user_id JOIN ssh_keys k ON k.id=s.key_id
		WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at>now()
		AND k.revoked_at IS NULL AND u.active FOR SHARE OF s,u,k`, security.Hash(token)).Scan(&uid, &kid, &publicKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", security.ErrUnauthorized
	}
	return uid, kid, publicKey, err
}

// Verify checks an active session for endpoints without a conversation operation.
func (s *Service) Verify(ctx context.Context, token string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = Principal(ctx, tx, token); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if len(token) != 48 {
		return security.ErrUnauthorized
	}
	var hash string
	err := s.DB.QueryRow(ctx, `UPDATE sessions s SET revoked_at=now()
		FROM ssh_keys k, users u
		WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at>now()
		AND k.id=s.key_id AND k.revoked_at IS NULL
		AND u.id=s.user_id AND u.active
		RETURNING s.token_hash`, security.Hash(token)).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return security.ErrUnauthorized
	}
	return err
}

// DeleteAccount removes this user's data while preserving messages written by
// other members. Their direct conversations become read-only archives.
func (s *Service) DeleteAccount(ctx context.Context, token string) error {
	if len(token) != 48 {
		return security.ErrUnauthorized
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var uid string
	if err = tx.QueryRow(ctx, "SELECT user_id FROM sessions WHERE token_hash=$1", security.Hash(token)).Scan(&uid); err != nil {
		return security.ErrUnauthorized
	}
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092634,hashtext($1))", uid); err != nil {
		return err
	}
	if _, err = Principal(ctx, tx, token); err != nil {
		return err
	}
	var username string
	if err = tx.QueryRow(ctx, "SELECT username FROM users WHERE id=$1 FOR UPDATE", uid).Scan(&username); err != nil {
		return err
	}
	queries := []struct {
		sql  string
		args []any
	}{
		{`UPDATE conversations SET direct_pair=NULL,name='Cuenta eliminada',closed_at=now()
			WHERE kind='direct' AND (split_part(direct_pair,':',1)=$1 OR split_part(direct_pair,':',2)=$1)`, []any{uid}},
		{`DELETE FROM auth_challenges WHERE username=$1 OR public_key IN (SELECT public_key FROM ssh_keys WHERE user_id=$2)`, []any{username, uid}},
		{`DELETE FROM recovery_challenges WHERE user_id=$1`, []any{uid}},
		{`DELETE FROM messages WHERE sender_id=$1`, []any{uid}},
		{`DELETE FROM memberships WHERE user_id=$1`, []any{uid}},
		{`DELETE FROM sessions WHERE user_id=$1`, []any{uid}},
		{`DELETE FROM e2ee_devices WHERE user_id=$1`, []any{uid}},
		{`DELETE FROM ssh_keys WHERE user_id=$1`, []any{uid}},
		{`DELETE FROM users WHERE id=$1`, []any{uid}},
	}
	for _, query := range queries {
		if _, err = tx.Exec(ctx, query.sql, query.args...); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
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
