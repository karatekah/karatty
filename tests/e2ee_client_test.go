package tests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
	"github.com/karatekah/termlink/pkg/secureclient"
)

type rejectMessagePost struct{ next http.RoundTripper }

func (r rejectMessagePost) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/messages") {
		return nil, errors.New("simulated unavailable server")
	}
	return r.next.RoundTrip(request)
}

func TestE2EEClientExchangeAndLocalPassphrase(t *testing.T) {
	bin := os.Getenv("KARATTY_CRYPTO_BIN")
	if bin == "" {
		t.Skip("set KARATTY_CRYPTO_BIN to run the local Olm client")
	}
	f := setup(t)
	f.a.PublicSignup = true
	f.c.DirectOnly = true
	f.c.CiphertextOnly = true
	f.api.IdentityRequests = security.NewLimiter(1000, 1000)
	aliceHTTP, bobHTTP := f.httpClient(t), f.httpClient(t)
	aliceKey, bobKey := signer(t), signer(t)
	aliceLogin, err := aliceHTTP.Login(f.ctx, aliceKey, "alice", "")
	if err != nil {
		t.Fatal(err)
	}
	bobLogin, err := bobHTTP.Login(f.ctx, bobKey, "bob", "")
	if err != nil {
		t.Fatal(err)
	}
	aliceDir, bobDir := t.TempDir(), t.TempDir()
	passphrase := []byte("local passphrase 123")
	alice, err := secureclient.Open(f.ctx, aliceHTTP, aliceKey, aliceLogin.UserID, passphrase, aliceDir, bin)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = alice.Close() })
	bob, err := secureclient.Open(f.ctx, bobHTTP, bobKey, bobLogin.UserID, passphrase, bobDir, bin)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = bob.Close() })
	if _, err = aliceHTTP.Direct(f.ctx, "bob"); err != nil {
		t.Fatal(err)
	}
	direct, err := bobHTTP.Direct(f.ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "unverified", Body: "secret"}); err == nil {
		t.Fatal("sent to a peer without fingerprint verification")
	}
	bobFingerprint, err := alice.Fingerprint(f.ctx, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if bobFingerprint != bob.SelfFingerprint() {
		t.Fatal("Bob's shared fingerprint differs from Alice's view")
	}
	if err = alice.Verify(f.ctx, "bob", strings.Repeat("0", 64)); err == nil {
		t.Fatal("accepted wrong fingerprint")
	}
	aliceFingerprint, err := bob.Fingerprint(f.ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if aliceFingerprint != alice.SelfFingerprint() {
		t.Fatal("Alice's shared fingerprint differs from Bob's view")
	}
	if err = alice.Verify(f.ctx, "bob", bobFingerprint); err != nil {
		t.Fatal(err)
	}
	if err = bob.Verify(f.ctx, "alice", aliceFingerprint); err != nil {
		t.Fatal(err)
	}
	first, err := alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "first", Body: "secreto de Alice"})
	if err != nil || first.Body != "secreto de Alice" {
		t.Fatal("first send failed", err)
	}
	raw, err := bobHTTP.History(f.ctx, direct.ID, 0, 100)
	if err != nil || len(raw.Messages) != 1 || !security.Envelope(raw.Messages[0].Body) || strings.Contains(raw.Messages[0].Body, "secreto de Alice") {
		t.Fatal("server persisted readable text", err)
	}
	received, err := bob.Sync(f.ctx, direct.ID, "alice")
	if err != nil || len(received.Messages) != 1 || received.Messages[0].Body != "secreto de Alice" {
		t.Fatal("Bob could not decrypt", err)
	}
	if _, err = bob.Send(f.ctx, direct.ID, "alice", protocol.Send{ClientID: "reply", Body: "respuesta de Bob"}); err != nil {
		t.Fatal(err)
	}
	if _, err = bob.Sync(f.ctx, direct.ID, "alice"); err != nil {
		t.Fatal(err)
	}
	conversation, err := alice.Sync(f.ctx, direct.ID, "bob")
	if err != nil || len(conversation.Messages) != 2 || conversation.Messages[1].Body != "respuesta de Bob" {
		t.Fatal("Alice could not decrypt the ratchet reply", err)
	}
	repeated, err := alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "first", Body: "secreto de Alice"})
	if err != nil || repeated.ID != first.ID {
		t.Fatal("confirmed retry changed message", err)
	}
	if _, err = alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "first", Body: "different"}); err != security.ErrConflict {
		t.Fatal("accepted a changed retry", err)
	}
	if err = bob.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = secureclient.Open(f.ctx, bobHTTP, bobKey, bobLogin.UserID, []byte("wrong passphrase"), bobDir, bin); err == nil {
		t.Fatal("wrong local passphrase unlocked archive")
	}
	bob, err = secureclient.Open(f.ctx, bobHTTP, bobKey, bobLogin.UserID, passphrase, bobDir, bin)
	if err != nil || len(bob.Archive(direct.ID)) != 2 {
		t.Fatal("encrypted local archive did not survive restart", err)
	}
	backend := &secureclient.Backend{HTTP: bobHTTP, Secure: bob}
	visible, err := backend.History(f.ctx, direct.ID, 0, 100)
	if err != nil || len(visible.Messages) != 2 || visible.Messages[0].Body != "secreto de Alice" {
		t.Fatal("local TUI backend exposed ciphertext or lost history", err)
	}
	streamCtx, stop := context.WithTimeout(f.ctx, 5*time.Second)
	stream := backend.Stream(streamCtx, direct.ID, 0)
	initial := receive(t, streamCtx, stream)
	stop()
	if initial.Err != nil || len(initial.Page.Messages) != 2 || initial.Page.Messages[0].Body != "secreto de Alice" {
		t.Fatal("local TUI stream did not show decrypted archive", initial.Err)
	}
	large := strings.Repeat("x", 4096)
	if _, err = alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "large", Body: large}); err != nil {
		t.Fatal("valid maximum length message failed", err)
	}
	received, err = bob.Sync(f.ctx, direct.ID, "alice")
	if err != nil || len(received.Messages) != 1 || received.Messages[0].Body != large {
		t.Fatal("maximum length message did not round trip", err)
	}
	if _, err = alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "tamper", Body: "must not display"}); err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	var rawText string
	if err = f.db.QueryRow(f.ctx, "SELECT body FROM messages WHERE client_id='tamper'").Scan(&rawText); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal([]byte(rawText), &envelope); err != nil {
		t.Fatal(err)
	}
	message := envelope["message"].(map[string]any)
	message["body"] = base64.RawStdEncoding.EncodeToString(make([]byte, 64))
	damaged, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec(f.ctx, "UPDATE messages SET body=$1 WHERE client_id='tamper'", string(damaged)); err != nil {
		t.Fatal(err)
	}
	if _, err = bob.Sync(f.ctx, direct.ID, "alice"); err == nil {
		t.Fatal("tampered ciphertext was accepted")
	}
	if len(bob.Archive(direct.ID)) != 3 {
		t.Fatal("tampered plaintext entered local archive")
	}
	if _, err = bob.Send(f.ctx, direct.ID, "alice", protocol.Send{ClientID: "old-unread", Body: "mensaje antiguo"}); err != nil {
		t.Fatal(err)
	}
	transport := aliceHTTP.Transport.Transport
	aliceHTTP.Transport.Transport = rejectMessagePost{next: transport}
	_, err = alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "uncertain", Body: "envío pendiente"})
	aliceHTTP.Transport.Transport = transport
	if err == nil {
		t.Fatal("offline send unexpectedly succeeded")
	}
	newBobKey := signer(t)
	challenge, err := bobHTTP.StartRecovery(f.ctx, newBobKey, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = bobHTTP.CompleteRecovery(f.ctx, newBobKey, challenge, bobLogin.RecoveryCode); err != nil {
		t.Fatal(err)
	}
	recoveredHTTP := f.httpClient(t)
	if _, err = recoveredHTTP.Login(f.ctx, newBobKey, "", ""); err != nil {
		t.Fatal(err)
	}
	recoveredBob, err := secureclient.Open(f.ctx, recoveredHTTP, newBobKey, bobLogin.UserID, passphrase, bobDir, bin)
	if err != nil {
		t.Fatal("new SSH key could not create a fresh local device", err)
	}
	t.Cleanup(func() { _ = recoveredBob.Close() })
	newFingerprint, err := alice.Fingerprint(f.ctx, "bob")
	if err != nil || newFingerprint != recoveredBob.SelfFingerprint() || newFingerprint == bobFingerprint {
		t.Fatal("recovered device identity was not rotated", err)
	}
	if _, err = alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "after-recovery-unverified", Body: "blocked"}); err == nil {
		t.Fatal("sent to a changed identity without new verification")
	}
	if err = alice.Reverify(f.ctx, direct.ID, "bob", newFingerprint); err == nil {
		t.Fatal("re-verification silently discarded a pending send")
	}
	if count, discardErr := alice.DiscardPending(direct.ID); discardErr != nil || count != 1 {
		t.Fatal("pending send could not be discarded explicitly", discardErr)
	}
	if err = alice.Reverify(f.ctx, direct.ID, "bob", newFingerprint); err != nil {
		t.Fatal("could not re-verify recovered peer", err)
	}
	if err = recoveredBob.Verify(f.ctx, "alice", aliceFingerprint); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.Send(f.ctx, direct.ID, "bob", protocol.Send{ClientID: "after-recovery", Body: "nueva sesión"}); err != nil {
		t.Fatal("new Olm session failed after recovery", err)
	}
	newMessages, err := recoveredBob.Sync(f.ctx, direct.ID, "alice")
	if err != nil || len(newMessages.Messages) != 1 || newMessages.Messages[0].Body != "nueva sesión" {
		t.Fatal("new device could not decrypt after recovery", err)
	}
	oldAndNew, err := alice.Sync(f.ctx, direct.ID, "bob")
	if err != nil || len(oldAndNew.Messages) != 4 || oldAndNew.Messages[2].Body != "[mensaje anterior al cambio de identidad; no disponible]" {
		t.Fatal("old unread ciphertext was not marked unavailable after rotation", err)
	}
}
