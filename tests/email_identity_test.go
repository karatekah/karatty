package tests

import (
	"context"
	"testing"

	"github.com/karatekah/termlink/internal/security"
)

type mailSink struct {
	to, code, purpose string
	count             int
}

func (m *mailSink) SendCode(_ context.Context, to, code, purpose string) error {
	m.to, m.code, m.purpose = to, code, purpose
	m.count++
	return nil
}

func TestVerifiedSignupAndRecovery(t *testing.T) {
	f := setup(t)
	mail := &mailSink{}
	f.a.Mail = mail
	f.a.PublicSignup = true
	f.api.IdentityRequests = security.NewLimiter(1000, 1000)
	aliceKey := signer(t)
	alice := f.httpClient(t)
	challenge, err := alice.StartSignup(f.ctx, aliceKey, "alice", "Alice@Example.com")
	if err != nil || mail.to != "alice@example.com" || mail.purpose != "signup" || len(mail.code) != 24 {
		t.Fatalf("signup email was not sent: %+v %+v %v", challenge, mail, err)
	}
	var users int
	if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM users").Scan(&users); err != nil || users != 0 {
		t.Fatal("account created before email verification", users, err)
	}
	if _, err = alice.CompleteSignup(f.ctx, aliceKey, challenge, "wrong"); !status(err, 401) {
		t.Fatal("invalid email code accepted", err)
	}
	if _, err = alice.CompleteSignup(f.ctx, aliceKey, challenge, mail.code); !status(err, 401) {
		t.Fatal("consumed email challenge reused", err)
	}
	challenge, err = alice.StartSignup(f.ctx, aliceKey, "alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	created, err := alice.CompleteSignup(f.ctx, aliceKey, challenge, mail.code)
	if err != nil || created.Session.Token == "" || len(created.RecoveryCode) != 48 {
		t.Fatalf("verified signup failed: %+v %v", created, err)
	}
	if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM users WHERE email=$1 AND email_verified_at IS NOT NULL", "alice@example.com").Scan(&users); err != nil || users != 1 {
		t.Fatal("email not verified in database", users, err)
	}
	if _, err = alice.StartSignup(f.ctx, signer(t), "another", "alice@example.com"); err != nil || mail.count != 2 {
		t.Fatal("existing email leaked or received another code", err)
	}
	newKey := signer(t)
	recoverClient := f.httpClient(t)
	recovery, err := recoverClient.StartRecovery(f.ctx, newKey, "alice@example.com")
	if err != nil || mail.purpose != "recovery" || mail.count != 3 {
		t.Fatal("recovery email was not sent", err)
	}
	if _, err = recoverClient.CompleteRecovery(f.ctx, newKey, recovery, mail.code, "wrong"); !status(err, 400) {
		t.Fatal("malformed recovery code accepted", err)
	}
	// A malformed code is rejected before consumption. A well formed wrong code consumes the challenge.
	if _, err = recoverClient.CompleteRecovery(f.ctx, newKey, recovery, mail.code, security.ID()); !status(err, 401) {
		t.Fatal("incorrect recovery code accepted", err)
	}
	if _, err = recoverClient.CompleteRecovery(f.ctx, newKey, recovery, mail.code, created.RecoveryCode); !status(err, 401) {
		t.Fatal("consumed recovery challenge reused", err)
	}
	recovery, err = recoverClient.StartRecovery(f.ctx, newKey, "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := recoverClient.CompleteRecovery(f.ctx, newKey, recovery, mail.code, created.RecoveryCode)
	if err != nil || len(rotated.RecoveryCode) != 48 || rotated.RecoveryCode == created.RecoveryCode {
		t.Fatalf("recovery failed: %+v %v", rotated, err)
	}
	if _, err = alice.List(f.ctx); !status(err, 401) {
		t.Fatal("old session survived recovery", err)
	}
	if _, err = f.httpClient(t).Login(f.ctx, aliceKey, "", ""); !status(err, 401) {
		t.Fatal("old SSH key survived recovery", err)
	}
	if _, err = recoverClient.Login(f.ctx, newKey, "", ""); err != nil {
		t.Fatal("new SSH key cannot log in", err)
	}
	if err = recoverClient.DeleteAccount(f.ctx); err != nil {
		t.Fatal(err)
	}
	if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM users WHERE email=$1", "alice@example.com").Scan(&users); err != nil || users != 0 {
		t.Fatal("account email retained after deletion", users, err)
	}
	if _, err = f.db.Exec(f.ctx, "INSERT INTO email_daily(day,sent) VALUES((now() AT TIME ZONE 'UTC')::date,80) ON CONFLICT(day) DO UPDATE SET sent=80"); err != nil {
		t.Fatal(err)
	}
	before := mail.count
	if _, err = alice.StartSignup(f.ctx, signer(t), "next", "next@example.com"); !status(err, 429) || mail.count != before {
		t.Fatal("daily email budget was exceeded", err)
	}
}
