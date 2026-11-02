package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/karatekah/termlink/pkg/client"
)

type identityFixture struct{ client.Backend }

func (identityFixture) Fingerprint(context.Context, string) (string, error) {
	return strings.Repeat("b", 64), nil
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
