package e2ee

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/karatekah/karatty/internal/auth"
	"github.com/karatekah/karatty/internal/security"
	"github.com/karatekah/karatty/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

type Service struct {
	DB     *pgxpool.Pool
	Claims *security.Limiter
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db, Claims: security.NewLimiter(1.0/60, 3)}
}

func public32(s string) ([]byte, bool) {
	b, err := base64.RawStdEncoding.DecodeString(s)
	return b, err == nil && len(b) == 32 && base64.RawStdEncoding.EncodeToString(b) == s
}

func signature64(s string) ([]byte, bool) {
	b, err := base64.RawStdEncoding.DecodeString(s)
	return b, err == nil && len(b) == ed25519.SignatureSize && base64.RawStdEncoding.EncodeToString(b) == s
}

func identityPayload(uid string, in protocol.DeviceUpload) string {
	return "KARATTY-E2EE-IDENTITY-v1\n" + uid + "\n" + in.IdentityEd + "\n" + in.IdentityCurve
}

func validateKeys(in protocol.DeviceUpload) error {
	ed, ok := public32(in.IdentityEd)
	if !ok {
		return security.ErrInvalid
	}
	if _, ok = public32(in.IdentityCurve); !ok || len(in.SSHSignature) == 0 || len(in.SSHSignature) > 2048 || len(in.OneTimeKeys) == 0 || len(in.OneTimeKeys) > 20 {
		return security.ErrInvalid
	}
	seen := map[string]bool{}
	for _, otk := range in.OneTimeKeys {
		if _, ok = public32(otk.Key); !ok || seen[otk.Key] {
			return security.ErrInvalid
		}
		seen[otk.Key] = true
		sig, valid := signature64(otk.Signature)
		if !valid || !ed25519.Verify(ed, []byte("KARATTY-OTK-v1\n"+in.IdentityCurve+"\n"+otk.Key), sig) {
			return security.ErrInvalid
		}
	}
	return nil
}

// Upload binds one Olm identity and its signed one-time keys to this SSH session.
func (s *Service) Upload(ctx context.Context, token string, in protocol.DeviceUpload) error {
	if err := validateKeys(in); err != nil {
		return err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	uid, kid, raw, err := auth.PrincipalKey(ctx, tx, token)
	if err != nil {
		return err
	}
	key, _, err := auth.PublicKey(raw)
	var sig ssh.Signature
	if err != nil || ssh.Unmarshal(in.SSHSignature, &sig) != nil || key.Verify([]byte(identityPayload(uid, in)), &sig) != nil {
		return security.ErrInvalid
	}
	var existingKid, existingEd, existingCurve string
	err = tx.QueryRow(ctx, "SELECT ssh_key_id,identity_ed,identity_curve FROM e2ee_devices WHERE user_id=$1 FOR UPDATE", uid).Scan(&existingKid, &existingEd, &existingCurve)
	if errors.Is(err, pgx.ErrNoRows) {
		_, err = tx.Exec(ctx, "INSERT INTO e2ee_devices(user_id,ssh_key_id,identity_ed,identity_curve,ssh_signature) VALUES($1,$2,$3,$4,$5)", uid, kid, in.IdentityEd, in.IdentityCurve, in.SSHSignature)
	} else if err == nil && (existingKid != kid || existingEd != in.IdentityEd || existingCurve != in.IdentityCurve) {
		return security.ErrConflict
	}
	if err != nil {
		return err
	}
	for _, otk := range in.OneTimeKeys {
		if _, err = tx.Exec(ctx, "INSERT INTO e2ee_one_time_keys(user_id,public_key,signature) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", uid, otk.Key, otk.Signature); err != nil {
			return err
		}
	}
	var count int
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM e2ee_one_time_keys WHERE user_id=$1 AND claimed_at IS NULL", uid).Scan(&count); err != nil {
		return err
	}
	if count > 100 {
		return security.ErrLimited
	}
	return tx.Commit(ctx)
}

func allowedPeer(ctx context.Context, tx pgx.Tx, uid, username string) (string, error) {
	if !security.Identifier(username) {
		return "", security.ErrInvalid
	}
	var peer string
	err := tx.QueryRow(ctx, "SELECT id FROM users WHERE username=$1 AND active FOR SHARE", username).Scan(&peer)
	if errors.Is(err, pgx.ErrNoRows) || peer == uid {
		return "", security.ErrForbidden
	}
	if err != nil {
		return "", err
	}
	a, b := uid, peer
	if strings.Compare(a, b) > 0 {
		a, b = b, a
	}
	var conversation string
	err = tx.QueryRow(ctx, `SELECT c.id FROM conversations c
		JOIN memberships m1 ON m1.conversation_id=c.id AND m1.user_id=$2
		JOIN memberships m2 ON m2.conversation_id=c.id AND m2.user_id=$3
		WHERE c.kind='direct' AND c.direct_pair=$1 AND c.closed_at IS NULL
		FOR SHARE OF c,m1,m2`, a+":"+b, uid, peer).Scan(&conversation)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", security.ErrForbidden
	}
	if err != nil {
		return "", err
	}
	return peer, nil
}

// Peer returns the signed identity only after both users accepted a direct.
func (s *Service) Peer(ctx context.Context, token, username string, claim bool) (protocol.DeviceBundle, error) {
	var out protocol.DeviceBundle
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	uid, err := auth.Principal(ctx, tx, token)
	if err != nil {
		return out, err
	}
	peer, err := allowedPeer(ctx, tx, uid, username)
	if err != nil {
		return out, err
	}
	if claim && !s.Claims.Allow(uid) {
		return out, security.ErrLimited
	}
	err = tx.QueryRow(ctx, `SELECT d.user_id,u.username,k.public_key,d.identity_ed,d.identity_curve,d.ssh_signature
		FROM e2ee_devices d JOIN users u ON u.id=d.user_id JOIN ssh_keys k ON k.id=d.ssh_key_id
		WHERE d.user_id=$1 AND k.revoked_at IS NULL FOR SHARE OF d,u,k`, peer).
		Scan(&out.UserID, &out.Username, &out.SSHKey, &out.IdentityEd, &out.IdentityCurve, &out.SSHSignature)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, security.ErrForbidden
	}
	if err != nil {
		return out, err
	}
	if claim {
		var otk protocol.OneTimeKey
		err = tx.QueryRow(ctx, "SELECT public_key,signature FROM e2ee_one_time_keys WHERE user_id=$1 AND claimed_at IS NULL ORDER BY public_key LIMIT 1 FOR UPDATE SKIP LOCKED", peer).Scan(&otk.Key, &otk.Signature)
		if errors.Is(err, pgx.ErrNoRows) {
			return out, security.ErrLimited
		}
		if err != nil {
			return out, err
		}
		if _, err = tx.Exec(ctx, "UPDATE e2ee_one_time_keys SET claimed_at=now(),signature='' WHERE user_id=$1 AND public_key=$2", peer, otk.Key); err != nil {
			return out, err
		}
		out.OneTimeKey = &otk
	}
	return out, tx.Commit(ctx)
}
