package tests

import "testing"

func TestDirectCannotRestoreRevokedMembership(t *testing.T) {
	f := setup(t)
	alice, _, _ := f.register(t, "alice", "")
	bob, _, session := f.register(t, "bob", "")
	dm, err := alice.Direct(f.ctx, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if err = f.c.RevokeMember(f.ctx, dm.ID, session.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err = bob.Direct(f.ctx, "alice"); err == nil {
		t.Fatal("revoked participant restored own membership")
	}
	if _, err = alice.Direct(f.ctx, "bob"); err == nil {
		t.Fatal("other participant restored revoked membership")
	}
	if _, err = bob.History(f.ctx, dm.ID, 0, 100); err == nil {
		t.Fatal("revoked direct history leaked")
	}
}
