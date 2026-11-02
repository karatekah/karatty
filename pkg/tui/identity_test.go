package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
)

type identityFixture struct{ client.Backend }

func (identityFixture) Fingerprint(context.Context, string) (string, error) {
	return strings.Repeat("b", 64), nil
}

type retryFixture struct {
	client.Backend
	got protocol.Send
}

func (*retryFixture) Pending(string) []protocol.Send {
	return []protocol.Send{{ClientID: "stable", Body: "guardado"}}
}
func (f *retryFixture) Send(_ context.Context, _ string, in protocol.Send) (protocol.Message, error) {
	f.got = in
	return protocol.Message{ClientID: in.ClientID, Body: in.Body}, nil
}

func TestTUIReopensStoredPendingSendWithSameID(t *testing.T) {
	f := &retryFixture{}
	m := New(context.Background(), f)
	m.rooms = []protocol.Conversation{{ID: "direct", Kind: "direct"}}
	m.input.SetValue("/pending")
	m.directCommand()
	if !strings.Contains(m.identityView, "stable") || !strings.Contains(m.identityView, "guardado") {
		t.Fatal("stored pending message was not available to the user")
	}
	m.input.SetValue("/retry stable")
	cmd := m.directCommand()
	if cmd == nil {
		t.Fatal("stored pending send was not retried")
	}
	cmd()
	if f.got.ClientID != "stable" || f.got.Body != "guardado" {
		t.Fatal("retry changed the stored message identity or body")
	}
}
func (identityFixture) Verify(context.Context, string, string) error        { return nil }
func (identityFixture) SelfFingerprint() string                             { return strings.Repeat("a", 64) }
func (identityFixture) Reverify(context.Context, string, string) error      { return nil }
func (identityFixture) DiscardPending(context.Context, string) (int, error) { return 0, nil }

func TestIdentityCommandsShowFullFingerprintBeforeVerification(t *testing.T) {
	m := New(context.Background(), identityFixture{})
	m.input.SetValue("/myfingerprint")
	if cmd := m.directCommand(); cmd != nil || !strings.Contains(m.identityView, strings.Repeat("a", 64)) {
		t.Fatal("own fingerprint was not shown in full")
	}
	m.input.SetValue("/fingerprint bob")
	cmd := m.directCommand()
	if cmd == nil {
		t.Fatal("peer fingerprint command did not run")
	}
	result, ok := cmd().(identityResult)
	if !ok || result.err != nil || result.fingerprint != strings.Repeat("b", 64) {
		t.Fatal("peer fingerprint lookup failed")
	}
	m.Update(result)
	if !strings.Contains(m.identityView, "Compárala con esa persona por otro medio") || !strings.Contains(m.identityView, result.fingerprint) {
		t.Fatal("TUI did not explain independent verification")
	}
	m.input.SetValue("/verify bob " + result.fingerprint)
	cmd = m.directCommand()
	if cmd == nil {
		t.Fatal("verification command did not run")
	}
	verified, ok := cmd().(identityResult)
	if !ok || !verified.verified || verified.err != nil {
		t.Fatal("verification command failed")
	}
}
