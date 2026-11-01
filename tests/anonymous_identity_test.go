package tests

import (
	"crypto/rand"
	"testing"

	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/internal/storage"
	"github.com/karatekah/termlink/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

func TestAnonymousSignupAndOfflineRecovery(t *testing.T) {
	f := setup(t)
	legacyChannel, err := f.c.Channel(f.ctx, "legacy")
	if err != nil {
		t.Fatal(err)
	}
	f.a.PublicSignup = true
	f.c.DirectOnly = true
	f.c.CiphertextOnly = true
	f.api.IdentityRequests = security.NewLimiter(1000, 1000)
	aliceKey := signer(t)
	alice := f.httpClient(t)
	registered, err := alice.Login(f.ctx, aliceKey, "alice", "")
	if err != nil || registered.Token == "" || len(registered.RecoveryCode) != 48 {
		t.Fatalf("anonymous signup failed: %+v %v", registered, err)
	}
	var hash string
	if err = f.db.QueryRow(f.ctx, "SELECT recovery_hash FROM users WHERE username='alice'").Scan(&hash); err != nil || hash != security.Hash(registered.RecoveryCode) {
		t.Fatal("recovery code was not stored as a hash", err)
	}
	if _, err = f.db.Exec(f.ctx, "INSERT INTO memberships(conversation_id,user_id) VALUES($1,$2)", legacyChannel, registered.UserID); err != nil {
		t.Fatal(err)
	}
	listed, err := alice.List(f.ctx)
	if err != nil || len(listed) != 0 {
		t.Fatal("public mode exposed a channel", listed, err)
	}
	if _, err = alice.History(f.ctx, legacyChannel, 0, 100); !status(err, 403) {
		t.Fatal("public mode allowed channel history", err)
	}
	if _, err = alice.Send(f.ctx, legacyChannel, protocol.Send{ClientID: security.ID(), Body: "plaintext"}); !status(err, 403) {
		t.Fatal("public mode allowed channel send", err)
	}
	if _, err = alice.Login(f.ctx, signer(t), "alice", ""); !status(err, 400) {
		t.Fatal("duplicate pseudonym accepted", err)
	}
	if _, err = alice.Login(f.ctx, signer(t), "bob", "not-an-invitation"); !status(err, 400) {
		t.Fatal("invitation accepted in public mode", err)
	}

	// A failed proof consumes a registration challenge.
	thirdKey := signer(t)
	challenge, err := f.a.Challenge(f.ctx, protocol.ChallengeRequest{Username: "third", PublicKey: string(ssh.MarshalAuthorizedKey(thirdKey.PublicKey()))})
	if err != nil {
		t.Fatal(err)
	}
	wrongSignature, _ := signer(t).Sign(rand.Reader, []byte(challenge.Payload))
	if _, err = f.a.Finish(f.ctx, protocol.Proof{ID: challenge.ID, Signature: ssh.Marshal(wrongSignature)}); err != security.ErrUnauthorized {
		t.Fatal("wrong registration signature accepted", err)
	}
	rightSignature, _ := thirdKey.Sign(rand.Reader, []byte(challenge.Payload))
	if _, err = f.a.Finish(f.ctx, protocol.Proof{ID: challenge.ID, Signature: ssh.Marshal(rightSignature)}); err != security.ErrUnauthorized {
		t.Fatal("consumed registration challenge reused", err)
	}

	newKey := signer(t)
	recovering := f.httpClient(t)
	unknown, err := recovering.StartRecovery(f.ctx, newKey, "unknown")
	if err != nil || unknown.ID == "" {
		t.Fatal("unknown pseudonym leaked during recovery start", err)
	}
	if _, err = recovering.CompleteRecovery(f.ctx, newKey, unknown, registered.RecoveryCode); !status(err, 401) {
		t.Fatal("unknown account recovered", err)
	}
	recovery, err := recovering.StartRecovery(f.ctx, newKey, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = recovering.CompleteRecovery(f.ctx, newKey, recovery, security.ID()); !status(err, 401) {
		t.Fatal("wrong recovery code accepted", err)
	}
	if _, err = recovering.CompleteRecovery(f.ctx, newKey, recovery, registered.RecoveryCode); !status(err, 401) {
		t.Fatal("consumed recovery challenge reused", err)
	}
	recovery, err = recovering.StartRecovery(f.ctx, newKey, "alice")
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := recovering.CompleteRecovery(f.ctx, newKey, recovery, registered.RecoveryCode)
	if err != nil || len(rotated.RecoveryCode) != 48 || rotated.RecoveryCode == registered.RecoveryCode {
		t.Fatalf("offline recovery failed: %+v %v", rotated, err)
	}
	if _, err = alice.List(f.ctx); !status(err, 401) {
		t.Fatal("old session survived recovery", err)
	}
	if _, err = f.httpClient(t).Login(f.ctx, aliceKey, "", ""); !status(err, 401) {
		t.Fatal("old SSH key survived recovery", err)
	}
	if _, err = recovering.Login(f.ctx, newKey, "", ""); err != nil {
		t.Fatal("new SSH key cannot log in", err)
	}
	if _, err = recovering.StartRecovery(f.ctx, signer(t), "alice"); err != nil {
		t.Fatal(err)
	}
	if err = recovering.DeleteAccount(f.ctx); err != nil {
		t.Fatal(err)
	}
	var users, challenges int
	if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM users").Scan(&users); err != nil || users != 0 {
		t.Fatal("account not removed", users, err)
	}
	if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM recovery_challenges").Scan(&challenges); err != nil || challenges != 0 {
		t.Fatal("recovery challenge retained", challenges, err)
	}

	// Migrations do not recreate email storage after the transition.
	if err = storage.Migrate(f.ctx, f.db); err != nil {
		t.Fatal(err)
	}
	var emailColumns int
	if err = f.db.QueryRow(f.ctx, `SELECT count(*) FROM information_schema.columns
        WHERE table_schema=current_schema() AND table_name='users' AND column_name='email'`).Scan(&emailColumns); err != nil || emailColumns != 0 {
		t.Fatal("email column remains", emailColumns, err)
	}
	var versions int
	if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM schema_migrations").Scan(&versions); err != nil || versions != 8 {
		t.Fatal("migration history is incomplete", versions, err)
	}
}
