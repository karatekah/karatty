package tests

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

func deviceUpload(t *testing.T, uid string, signer ssh.Signer) protocol.DeviceUpload {
	t.Helper()
	ed, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	curve := make([]byte, 32)
	otk := make([]byte, 32)
	if _, err = rand.Read(curve); err != nil {
		t.Fatal(err)
	}
	if _, err = rand.Read(otk); err != nil {
		t.Fatal(err)
	}
	enc := base64.RawStdEncoding.EncodeToString
	in := protocol.DeviceUpload{IdentityEd: enc(ed), IdentityCurve: enc(curve)}
	key := enc(otk)
	in.OneTimeKeys = []protocol.OneTimeKey{{Key: key, Signature: enc(ed25519.Sign(private, []byte("KARATTY-OTK-v1\n"+in.IdentityCurve+"\n"+key)))}}
	sshSig, err := signer.Sign(rand.Reader, []byte("KARATTY-E2EE-IDENTITY-v1\n"+uid+"\n"+in.IdentityEd+"\n"+in.IdentityCurve))
	if err != nil {
		t.Fatal(err)
	}
	in.SSHSignature = ssh.Marshal(sshSig)
	return in
}

func TestE2EEKeyDistributionRequiresAcceptedDirect(t *testing.T) {
	f := setup(t)
	f.a.PublicSignup = true
	f.c.DirectOnly = true
	f.api.IdentityRequests = security.NewLimiter(1000, 1000)
	alice := f.httpClient(t)
	bob := f.httpClient(t)
	eve := f.httpClient(t)
	aliceKey, bobKey := signer(t), signer(t)
	if _, err := alice.Login(f.ctx, aliceKey, "alice", ""); err != nil {
		t.Fatal(err)
	}
	bobSession, err := bob.Login(f.ctx, bobKey, "bob", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = eve.Login(f.ctx, signer(t), "eve", ""); err != nil {
		t.Fatal(err)
	}
	bundle := deviceUpload(t, bobSession.UserID, bobKey)
	wrong := deviceUpload(t, bobSession.UserID, signer(t))
	if err = bob.UploadDevice(f.ctx, wrong); !status(err, 400) {
		t.Fatal("wrong SSH signature accepted", err)
	}
	if err = bob.UploadDevice(f.ctx, bundle); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.Peer(f.ctx, "bob"); !status(err, 403) {
		t.Fatal("unaccepted direct exposed identity", err)
	}
	if _, err = alice.Direct(f.ctx, "bob"); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.Peer(f.ctx, "bob"); !status(err, 403) {
		t.Fatal("one-sided request exposed identity", err)
	}
	if _, err = bob.Direct(f.ctx, "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err = eve.ClaimPeer(f.ctx, "bob"); !status(err, 403) {
		t.Fatal("stranger claimed key", err)
	}
	peer, err := alice.Peer(f.ctx, "bob")
	if err != nil || peer.UserID != bobSession.UserID || peer.OneTimeKey != nil || peer.IdentityEd != bundle.IdentityEd || peer.IdentityCurve != bundle.IdentityCurve {
		t.Fatal("invalid identity bundle", peer, err)
	}
	claimed, err := alice.ClaimPeer(f.ctx, "bob")
	if err != nil || claimed.OneTimeKey == nil || claimed.OneTimeKey.Key != bundle.OneTimeKeys[0].Key {
		t.Fatal("one-time key not claimed", claimed, err)
	}
	if _, err = alice.ClaimPeer(f.ctx, "bob"); !status(err, 429) {
		t.Fatal("one-time key reused", err)
	}
	if err = bob.UploadDevice(f.ctx, bundle); err != nil {
		t.Fatal("idempotent key upload failed", err)
	}
	if _, err = alice.ClaimPeer(f.ctx, "bob"); !status(err, 429) {
		t.Fatal("consumed key was reactivated", err)
	}
	changed := deviceUpload(t, bobSession.UserID, bobKey)
	if err = bob.UploadDevice(f.ctx, changed); !status(err, 409) {
		t.Fatal("identity changed silently", err)
	}
	newSSHKey := signer(t)
	recovery, err := bob.StartRecovery(f.ctx, newSSHKey, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = bob.CompleteRecovery(f.ctx, newSSHKey, recovery, bobSession.RecoveryCode); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.Peer(f.ctx, "bob"); !status(err, 403) {
		t.Fatal("recovery retained old E2EE identity", err)
	}
	recovered := f.httpClient(t)
	if _, err = recovered.Login(f.ctx, newSSHKey, "", ""); err != nil {
		t.Fatal(err)
	}
	if err = recovered.DeleteAccount(f.ctx); err != nil {
		t.Fatal("account deletion left E2EE key rows", err)
	}
}
