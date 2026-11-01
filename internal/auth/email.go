package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

func canonicalEmail(raw string) (string, error) {
	if len(raw) < 3 || len(raw) > 254 || strings.TrimSpace(raw) != raw || strings.ContainsAny(raw, "\r\n") {
		return "", security.ErrInvalid
	}
	address, err := mail.ParseAddress(raw)
	if err != nil || address.Address != raw || strings.Count(raw, "@") != 1 {
		return "", security.ErrInvalid
	}
	parts := strings.Split(raw, "@")
	if len(parts[0]) > 64 || !strings.Contains(parts[1], ".") {
		return "", security.ErrInvalid
	}
	return strings.ToLower(raw), nil
}

func verifyEmailProof(payload, raw, expectedHash string, proof protocol.EmailProof) error {
	if !security.Identifier(proof.ID) || len(proof.Code) != 24 || len(proof.Signature) > 2048 ||
		subtle.ConstantTimeCompare([]byte(expectedHash), []byte(security.Hash(proof.Code))) != 1 {
		return security.ErrUnauthorized
	}
	key, _, err := PublicKey(raw)
	var sig ssh.Signature
	if err != nil || ssh.Unmarshal(proof.Signature, &sig) != nil || key.Verify([]byte(payload), &sig) != nil {
		return security.ErrUnauthorized
	}
	return nil
}

func (s *Service) StartSignup(ctx context.Context, in protocol.SignupStart) (protocol.Challenge, error) {
	if !s.PublicSignup || s.Mail == nil {
		return protocol.Challenge{}, security.ErrForbidden
	}
	if !security.Identifier(in.Username) {
		return protocol.Challenge{}, security.ErrInvalid
	}
	email, err := canonicalEmail(in.Email)
	if err != nil {
		return protocol.Challenge{}, err
	}
	_, key, err := PublicKey(in.PublicKey)
	if err != nil {
		return protocol.Challenge{}, err
	}
	challenge := protocol.Challenge{ID: security.ID(), Payload: "TERMLINK-SIGNUP-v1\n" + security.ID()}
	code := security.ID()[:24]
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.Challenge{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092637)"); err != nil {
		return protocol.Challenge{}, err
	}
	if _, err = tx.Exec(ctx, "DELETE FROM email_challenges WHERE expires_at<=now()"); err != nil {
		return protocol.Challenge{}, err
	}
	var count int
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM email_challenges").Scan(&count); err != nil {
		return protocol.Challenge{}, err
	}
	if count >= 4096 {
		return protocol.Challenge{}, security.ErrLimited
	}
	var taken bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1 OR email=$2)", in.Username, email).Scan(&taken); err != nil {
		return protocol.Challenge{}, err
	}
	if taken {
		return challenge, tx.Commit(ctx)
	}
	result, err := tx.Exec(ctx, `INSERT INTO email_challenges(id,purpose,username,email,public_key,payload,code_hash,expires_at)
		VALUES($1,'signup',$2,$3,$4,$5,$6,now()+interval '10 minutes')
		ON CONFLICT(purpose,email) DO NOTHING`, challenge.ID, in.Username, email, key, challenge.Payload, security.Hash(code))
	if err != nil {
		return protocol.Challenge{}, err
	}
	if result.RowsAffected() > 0 {
		if err = reserveEmail(ctx, tx); err != nil {
			return protocol.Challenge{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return protocol.Challenge{}, err
	}
	if result.RowsAffected() == 0 {
		return challenge, nil
	}
	if err = s.Mail.SendCode(ctx, email, code, "signup"); err != nil {
		_, _ = s.DB.Exec(context.Background(), "DELETE FROM email_challenges WHERE id=$1", challenge.ID)
		return protocol.Challenge{}, err
	}
	return challenge, nil
}

func (s *Service) CompleteSignup(ctx context.Context, proof protocol.EmailProof) (protocol.SignupResult, error) {
	if !s.PublicSignup || s.Mail == nil {
		return protocol.SignupResult{}, security.ErrForbidden
	}
	if !security.Identifier(proof.ID) {
		return protocol.SignupResult{}, security.ErrInvalid
	}
	var username, email, raw, payload, codeHash string
	err := s.DB.QueryRow(ctx, `DELETE FROM email_challenges WHERE id=$1 AND purpose='signup' AND expires_at>now()
		RETURNING username,email,public_key,payload,code_hash`, proof.ID).Scan(&username, &email, &raw, &payload, &codeHash)
	if err != nil {
		return protocol.SignupResult{}, security.ErrUnauthorized
	}
	if err = verifyEmailProof(payload, raw, codeHash, proof); err != nil {
		return protocol.SignupResult{}, err
	}
	key, _, err := PublicKey(raw)
	if err != nil {
		return protocol.SignupResult{}, err
	}
	uid, kid := security.ID(), security.ID()
	recoveryCode := security.ID()
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.SignupResult{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO users(id,username,email,email_verified_at,recovery_hash)
		VALUES($1,$2,$3,now(),$4)`, uid, username, email, security.Hash(recoveryCode)); err != nil {
		return protocol.SignupResult{}, security.ErrInvalid
	}
	if _, err = tx.Exec(ctx, "INSERT INTO ssh_keys(id,user_id,public_key,fingerprint) VALUES($1,$2,$3,$4)", kid, uid, raw, ssh.FingerprintSHA256(key)); err != nil {
		return protocol.SignupResult{}, security.ErrInvalid
	}
	session, err := issue(ctx, tx, uid, kid)
	if err != nil {
		return protocol.SignupResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return protocol.SignupResult{}, err
	}
	return protocol.SignupResult{Session: session, RecoveryCode: recoveryCode}, nil
}

func (s *Service) StartRecovery(ctx context.Context, in protocol.RecoveryStart) (protocol.Challenge, error) {
	if s.Mail == nil {
		return protocol.Challenge{}, security.ErrForbidden
	}
	email, err := canonicalEmail(in.Email)
	if err != nil {
		return protocol.Challenge{}, err
	}
	_, key, err := PublicKey(in.PublicKey)
	if err != nil {
		return protocol.Challenge{}, err
	}
	challenge := protocol.Challenge{ID: security.ID(), Payload: "TERMLINK-RECOVERY-v1\n" + security.ID()}
	code := security.ID()[:24]
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.Challenge{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092637)"); err != nil {
		return protocol.Challenge{}, err
	}
	if _, err = tx.Exec(ctx, "DELETE FROM email_challenges WHERE expires_at<=now()"); err != nil {
		return protocol.Challenge{}, err
	}
	var uid string
	err = tx.QueryRow(ctx, "SELECT id FROM users WHERE email=$1 AND email_verified_at IS NOT NULL AND recovery_hash IS NOT NULL AND active", email).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return challenge, tx.Commit(ctx)
	}
	if err != nil {
		return protocol.Challenge{}, err
	}
	result, err := tx.Exec(ctx, `INSERT INTO email_challenges(id,purpose,email,public_key,payload,code_hash,user_id,expires_at)
		VALUES($1,'recovery',$2,$3,$4,$5,$6,now()+interval '10 minutes')
		ON CONFLICT(purpose,email) DO NOTHING`, challenge.ID, email, key, challenge.Payload, security.Hash(code), uid)
	if err != nil {
		return protocol.Challenge{}, err
	}
	if result.RowsAffected() > 0 {
		if err = reserveEmail(ctx, tx); err != nil {
			return protocol.Challenge{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return protocol.Challenge{}, err
	}
	if result.RowsAffected() == 0 {
		return challenge, nil
	}
	if err = s.Mail.SendCode(ctx, email, code, "recovery"); err != nil {
		_, _ = s.DB.Exec(context.Background(), "DELETE FROM email_challenges WHERE id=$1", challenge.ID)
		return protocol.Challenge{}, err
	}
	return challenge, nil
}

func (s *Service) CompleteRecovery(ctx context.Context, proof protocol.RecoveryProof) (protocol.RecoveryResult, error) {
	if s.Mail == nil {
		return protocol.RecoveryResult{}, security.ErrForbidden
	}
	if !security.Identifier(proof.ID) || len(proof.RecoveryCode) != 48 {
		return protocol.RecoveryResult{}, security.ErrInvalid
	}
	var uid, email, raw, payload, codeHash string
	err := s.DB.QueryRow(ctx, `DELETE FROM email_challenges WHERE id=$1 AND purpose='recovery' AND expires_at>now()
		RETURNING user_id,email,public_key,payload,code_hash`, proof.ID).Scan(&uid, &email, &raw, &payload, &codeHash)
	if err != nil {
		return protocol.RecoveryResult{}, security.ErrUnauthorized
	}
	if err = verifyEmailProof(payload, raw, codeHash, proof.EmailProof); err != nil {
		return protocol.RecoveryResult{}, err
	}
	key, _, err := PublicKey(raw)
	if err != nil {
		return protocol.RecoveryResult{}, err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return protocol.RecoveryResult{}, err
	}
	defer tx.Rollback(ctx)
	var recoveryHash string
	if err = tx.QueryRow(ctx, "SELECT recovery_hash FROM users WHERE id=$1 AND email=$2 AND active FOR UPDATE", uid, email).Scan(&recoveryHash); err != nil {
		return protocol.RecoveryResult{}, security.ErrUnauthorized
	}
	if subtle.ConstantTimeCompare([]byte(recoveryHash), []byte(security.Hash(proof.RecoveryCode))) != 1 {
		return protocol.RecoveryResult{}, security.ErrUnauthorized
	}
	newRecoveryCode := security.ID()
	if _, err = tx.Exec(ctx, "UPDATE sessions SET revoked_at=now() WHERE user_id=$1", uid); err != nil {
		return protocol.RecoveryResult{}, err
	}
	if _, err = tx.Exec(ctx, "UPDATE ssh_keys SET revoked_at=now() WHERE user_id=$1", uid); err != nil {
		return protocol.RecoveryResult{}, err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO ssh_keys(id,user_id,public_key,fingerprint) VALUES($1,$2,$3,$4)", security.ID(), uid, raw, ssh.FingerprintSHA256(key)); err != nil {
		return protocol.RecoveryResult{}, security.ErrInvalid
	}
	if _, err = tx.Exec(ctx, "UPDATE users SET recovery_hash=$1 WHERE id=$2", security.Hash(newRecoveryCode), uid); err != nil {
		return protocol.RecoveryResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return protocol.RecoveryResult{}, err
	}
	return protocol.RecoveryResult{RecoveryCode: newRecoveryCode}, nil
}

func reserveEmail(ctx context.Context, tx pgx.Tx) error {
	var count int
	err := tx.QueryRow(ctx, `INSERT INTO email_daily(day,sent) VALUES((now() AT TIME ZONE 'UTC')::date,1)
		ON CONFLICT(day) DO UPDATE SET sent=email_daily.sent+1 WHERE email_daily.sent<80
		RETURNING sent`).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return security.ErrLimited
	}
	return err
}
