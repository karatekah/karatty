package tests

import (
	"testing"

	"github.com/karatekah/karatty/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

func TestDeleteAccountPreservesPeerMessages(t *testing.T) {
	f := setup(t)
	channel, err := f.c.Channel(f.ctx, "account-delete")
	if err != nil {
		t.Fatal(err)
	}
	alice, aliceKey, aliceSession := f.register(t, "alice", channel)
	bob, _, _ := f.register(t, "bob", channel)
	direct, err := alice.Direct(f.ctx, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if !direct.Pending || direct.ID != "" {
		t.Fatal("direct request opened a chat without acceptance")
	}
	direct, err = bob.Direct(f.ctx, "alice")
	if err != nil || direct.Pending || direct.ID == "" {
		t.Fatal("recipient could not accept direct request", err)
	}
	for _, conv := range []string{channel, direct.ID} {
		if _, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: "alice-" + conv, Body: "alice content"}); err != nil {
			t.Fatal(err)
		}
		if _, err = bob.Send(f.ctx, conv, protocol.Send{ClientID: "bob-" + conv, Body: "bob content"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = f.a.Challenge(f.ctx, protocol.ChallengeRequest{PublicKey: string(ssh.MarshalAuthorizedKey(aliceKey.PublicKey()))}); err != nil {
		t.Fatal(err)
	}
	if err = alice.DeleteAccount(f.ctx); err != nil {
		t.Fatal(err)
	}
	for _, conv := range []string{channel, direct.ID} {
		page, e := bob.History(f.ctx, conv, 0, 100)
		if e != nil || len(page.Messages) != 1 || page.Messages[0].Body != "bob content" || page.Messages[0].Seq != 2 {
			t.Fatalf("peer history after deletion: %+v %v", page, e)
		}
	}
	rooms, err := bob.List(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	for _, room := range rooms {
		if room.ID == direct.ID {
			closed = room.Closed && room.Name == "Cuenta eliminada"
		}
	}
	if !closed {
		t.Fatal("direct conversation was not archived")
	}
	if _, err = bob.Send(f.ctx, direct.ID, protocol.Send{ClientID: "closed", Body: "unread"}); err == nil {
		t.Fatal("archived direct accepted a message")
	}
	if _, err = bob.Send(f.ctx, channel, protocol.Send{ClientID: "still-open", Body: "channel continues"}); err != nil {
		t.Fatal("shared channel stopped working:", err)
	}
	if _, err = bob.Direct(f.ctx, "alice"); err == nil {
		t.Fatal("deleted username still reachable")
	}
	if _, err = f.httpClient(t).Login(f.ctx, aliceKey, "", ""); err == nil {
		t.Fatal("deleted SSH key logged in")
	}
	for table, query := range map[string]string{
		"users":       "SELECT count(*) FROM users WHERE id=$1",
		"keys":        "SELECT count(*) FROM ssh_keys WHERE user_id=$1",
		"sessions":    "SELECT count(*) FROM sessions WHERE user_id=$1",
		"memberships": "SELECT count(*) FROM memberships WHERE user_id=$1",
		"messages":    "SELECT count(*) FROM messages WHERE sender_id=$1",
	} {
		var count int
		if err = f.db.QueryRow(f.ctx, query, aliceSession.UserID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s retained deleted account data: %d %v", table, count, err)
		}
	}
	var challenges int
	if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM auth_challenges WHERE public_key=$1", string(ssh.MarshalAuthorizedKey(aliceKey.PublicKey()))).Scan(&challenges); err != nil || challenges != 0 {
		t.Fatalf("challenge retained: %d %v", challenges, err)
	}
	if err = alice.DeleteAccount(f.ctx); err == nil {
		t.Fatal("deleted account authorized a second deletion")
	}
}
