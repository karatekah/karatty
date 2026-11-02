package tests

import (
	"testing"
	"time"

	"github.com/karatekah/karatty/internal/storage"
	"github.com/karatekah/karatty/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

func TestMessageRetention(t *testing.T) {
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "retention")
	if err != nil {
		t.Fatal(err)
	}
	alice, _, session := f.register(t, "alice", conv)
	old, err := alice.Send(f.ctx, conv, protocol.Send{ClientID: "old", Body: "expired content"})
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := alice.Send(f.ctx, conv, protocol.Send{ClientID: "fresh", Body: "retained content"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec(f.ctx, "UPDATE messages SET created_at=now()-interval '721 hours' WHERE id=$1", old.ID); err != nil {
		t.Fatal(err)
	}
	page, err := alice.History(f.ctx, conv, 0, 100)
	if err != nil || len(page.Messages) != 1 || page.Messages[0].ID != fresh.ID || page.Next != fresh.Seq {
		t.Fatalf("expired message visible or sequence broken: %+v %v", page, err)
	}
	// Reusing an expired ID removes the old content before assigning a new sequence.
	retry, err := alice.Send(f.ctx, conv, protocol.Send{ClientID: "old", Body: "new content"})
	if err != nil || retry.ID == old.ID || retry.Seq != fresh.Seq+1 {
		t.Fatalf("expired ID reuse: %+v %v", retry, err)
	}
	if _, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: "old", Body: "different content"}); err == nil {
		t.Fatal("unexpired ID reused with different content")
	}
	if _, err = f.db.Exec(f.ctx, "UPDATE messages SET created_at=now()-interval '721 hours' WHERE id=$1", fresh.ID); err != nil {
		t.Fatal(err)
	}
	deleted, err := storage.PurgeExpiredMessages(f.ctx, f.db)
	if err != nil || deleted != 1 {
		t.Fatalf("purge result: %d %v", deleted, err)
	}
	deleted, err = storage.PurgeExpiredMessages(f.ctx, f.db)
	if err != nil || deleted != 0 {
		t.Fatalf("purge was not idempotent: %d %v", deleted, err)
	}
	page, err = alice.History(f.ctx, conv, 0, 100)
	if err != nil || len(page.Messages) != 1 || page.Messages[0].ID != retry.ID || page.Next != retry.Seq {
		t.Fatalf("history after purge: %+v %v", page, err)
	}
	if err = storage.Migrate(f.ctx, f.db); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	// A backlog larger than one batch is fully removed without touching live data.
	_, err = f.db.Exec(f.ctx, `INSERT INTO messages(id,client_id,conversation_id,sender_id,seq,body,created_at)
		SELECT 'bulk-'||n,'bulk-'||n,$1,$2,1000+n,'old',now()-interval '721 hours'
		FROM generate_series(1,1001) n`, conv, session.UserID)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err = storage.PurgeExpiredMessages(f.ctx, f.db)
	if err != nil || deleted != 1001 {
		t.Fatalf("batched purge: %d %v", deleted, err)
	}
	page, err = alice.History(f.ctx, conv, 0, 100)
	if err != nil || len(page.Messages) != 1 || page.Messages[0].ID != retry.ID {
		t.Fatalf("live message lost during batched purge: %+v %v", page, err)
	}
}

func TestExpiredAuthRecordsArePurged(t *testing.T) {
	f := setup(t)
	alice, key, _ := f.register(t, "alice", "")
	if err := alice.Logout(f.ctx); err != nil {
		t.Fatal(err)
	}
	challenge, err := f.a.Challenge(f.ctx, protocol.ChallengeRequest{PublicKey: string(ssh.MarshalAuthorizedKey(key.PublicKey()))})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.a.Invite(f.ctx, "", 1, time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec(f.ctx, "UPDATE auth_challenges SET expires_at=now()-interval '1 second' WHERE id=$1", challenge.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec(f.ctx, "UPDATE invitations SET expires_at=now()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	if _, err = storage.PurgeExpiredAuthRecords(f.ctx, f.db); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"sessions", "auth_challenges", "invitations"} {
		var count int
		if err = f.db.QueryRow(f.ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s retained expired authentication data: %d %v", table, count, err)
		}
	}
}
