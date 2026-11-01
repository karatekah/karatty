package auth

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

// StartRecovery prepares a one-use challenge for a new SSH key. The response
// is identical for unknown usernames, so it does not expose account existence.
func (s *Service) StartRecovery(ctx context.Context, in protocol.RecoveryStart) (protocol.Challenge, error) {
	if !security.Identifier(in.Username) {
		return protocol.Challenge{}, security.ErrInvalid
	}
	_, raw, err := PublicKey(in.PublicKey)
	if err != nil {
		return protocol.Challenge{}, err
	}
	challenge := protocol.Challenge{ID: security.ID(), Payload: "KARATTY-RECOVERY-v2\n" + security.ID()}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.Challenge{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092637)"); err != nil {
		return protocol.Challenge{}, err
	}
	if _, err = tx.Exec(ctx, "DELETE FROM recovery_challenges WHERE expires_at<=now()"); err != nil {
		return protocol.Challenge{}, err
	}
	var count int
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM recovery_challenges").Scan(&count); err != nil {
		return protocol.Challenge{}, err
	}
	if count >= 4096 {
		return protocol.Challenge{}, security.ErrLimited
	}
	var uid string
	err = tx.QueryRow(ctx, "SELECT id FROM users WHERE username=$1 AND active AND recovery_hash IS NOT NULL", in.Username).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return challenge, tx.Commit(ctx)
	}
	if err != nil {
		return protocol.Challenge{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO recovery_challenges(id,user_id,public_key,payload,expires_at)
        VALUES($1,$2,$3,$4,now()+interval '2 minutes')
        ON CONFLICT(user_id) DO UPDATE SET id=EXCLUDED.id, public_key=EXCLUDED.public_key,
        payload=EXCLUDED.payload, expires_at=EXCLUDED.expires_at`, challenge.ID, uid, raw, challenge.Payload)
	if err != nil {
		return protocol.Challenge{}, err
	}
	return challenge, tx.Commit(ctx)
}

// CompleteRecovery consumes the challenge even when the proof or code is wrong.
// The 192-bit saved code is the sole recovery credential when the old key is lost.
func (s *Service) CompleteRecovery(ctx context.Context, in protocol.RecoveryProof) (protocol.RecoveryResult, error) {
	if !security.Identifier(in.ID) || len(in.Signature) > 2048 {
		return protocol.RecoveryResult{}, security.ErrInvalid
	}
	var uid, raw, payload string
	err := s.DB.QueryRow(ctx, `DELETE FROM recovery_challenges WHERE id=$1 AND expires_at>now()
        RETURNING user_id,public_key,payload`, in.ID).Scan(&uid, &raw, &payload)
	if err != nil {
		return protocol.RecoveryResult{}, security.ErrUnauthorized
	}
	key, _, err := PublicKey(raw)
	var signature ssh.Signature
	if err != nil || ssh.Unmarshal(in.Signature, &signature) != nil || key.Verify([]byte(payload), &signature) != nil {
		return protocol.RecoveryResult{}, security.ErrUnauthorized
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.RecoveryResult{}, err
	}
	defer tx.Rollback(ctx)
	var recoveryHash string
	if err = tx.QueryRow(ctx, "SELECT recovery_hash FROM users WHERE id=$1 AND active FOR UPDATE", uid).Scan(&recoveryHash); err != nil {
		return protocol.RecoveryResult{}, security.ErrUnauthorized
	}
	if len(in.RecoveryCode) != 48 || subtle.ConstantTimeCompare([]byte(recoveryHash), []byte(security.Hash(in.RecoveryCode))) != 1 {
		return protocol.RecoveryResult{}, security.ErrUnauthorized
	}
	nextCode := security.ID()
	if _, err = tx.Exec(ctx, "UPDATE sessions SET revoked_at=now() WHERE user_id=$1", uid); err != nil {
		return protocol.RecoveryResult{}, err
	}
	if _, err = tx.Exec(ctx, "UPDATE ssh_keys SET revoked_at=now() WHERE user_id=$1", uid); err != nil {
		return protocol.RecoveryResult{}, err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO ssh_keys(id,user_id,public_key,fingerprint) VALUES($1,$2,$3,$4)", security.ID(), uid, raw, ssh.FingerprintSHA256(key)); err != nil {
		return protocol.RecoveryResult{}, security.ErrInvalid
	}
	if _, err = tx.Exec(ctx, "UPDATE users SET recovery_hash=$1 WHERE id=$2", security.Hash(nextCode), uid); err != nil {
		return protocol.RecoveryResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return protocol.RecoveryResult{}, err
	}
	return protocol.RecoveryResult{RecoveryCode: nextCode}, nil
}
