package tests

import (
	"errors"
	"testing"

	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
)

func TestDirectConsentDeclineAndBlock(t *testing.T) {
	f := setup(t)
	alice, _, _ := f.register(t, "alice", "")
	bob, _, _ := f.register(t, "bob", "")

	pending, err := alice.Direct(f.ctx, "bob")
	if err != nil || !pending.Pending || pending.ID != "" {
		t.Fatalf("request should not open a conversation: %+v %v", pending, err)
	}
	requests, err := bob.DirectRequests(f.ctx)
	if err != nil || len(requests) != 1 || requests[0].Username != "alice" {
		t.Fatalf("incoming request missing: %+v %v", requests, err)
	}
	if err = bob.DeclineDirect(f.ctx, "alice"); err != nil {
		t.Fatal(err)
	}
	requests, err = bob.DirectRequests(f.ctx)
	if err != nil || len(requests) != 0 {
		t.Fatalf("declined request still visible: %+v %v", requests, err)
	}
	if pending, err = alice.Direct(f.ctx, "bob"); err != nil || !pending.Pending {
		t.Fatalf("new request after decline failed: %+v %v", pending, err)
	}
	if err = bob.BlockDirect(f.ctx, "alice"); err != nil {
		t.Fatal(err)
	}
	requests, err = bob.DirectRequests(f.ctx)
	if err != nil || len(requests) != 0 {
		t.Fatalf("blocked request still visible: %+v %v", requests, err)
	}
	if _, err = alice.Direct(f.ctx, "bob"); !status(err, 403) {
		t.Fatal("blocked sender could request again", err)
	}
	if _, err = bob.Direct(f.ctx, "alice"); !status(err, 403) {
		t.Fatal("blocked pair opened a chat", err)
	}
	if err = bob.UnblockDirect(f.ctx, "alice"); err != nil {
		t.Fatal(err)
	}
	if pending, err = alice.Direct(f.ctx, "bob"); err != nil || !pending.Pending {
		t.Fatalf("request after unblock failed: %+v %v", pending, err)
	}
	direct, err := bob.Direct(f.ctx, "alice")
	if err != nil || direct.ID == "" || direct.Pending || direct.Name != "alice" {
		t.Fatalf("recipient could not accept: %+v %v", direct, err)
	}
	openRooms, err := alice.List(f.ctx)
	if err != nil || len(openRooms) != 1 || openRooms[0].Name != "bob" {
		t.Fatalf("sender cannot identify direct chat: %+v %v", openRooms, err)
	}
	if _, err = alice.Send(f.ctx, direct.ID, protocol.Send{ClientID: "before-block", Body: "message"}); err != nil {
		t.Fatal(err)
	}
	if err = bob.BlockDirect(f.ctx, "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.Send(f.ctx, direct.ID, protocol.Send{ClientID: "after-block", Body: "blocked"}); !status(err, 403) {
		t.Fatal("closed direct accepted message", err)
	}
	page, err := bob.History(f.ctx, direct.ID, 0, 100)
	if err != nil || len(page.Messages) != 1 {
		t.Fatalf("past messages disappeared: %+v %v", page, err)
	}
	rooms, err := alice.List(f.ctx)
	if err != nil || len(rooms) != 1 || !rooms[0].Closed {
		t.Fatalf("blocked chat was not archived: %+v %v", rooms, err)
	}
}

func TestDirectRequestExpiry(t *testing.T) {
	f := setup(t)
	alice, _, _ := f.register(t, "alice", "")
	bob, _, _ := f.register(t, "bob", "")
	if _, err := alice.Direct(f.ctx, "bob"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(f.ctx, "UPDATE direct_requests SET expires_at=now()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	requests, err := bob.DirectRequests(f.ctx)
	if err != nil || len(requests) != 0 {
		t.Fatalf("expired request remained visible: %+v %v", requests, err)
	}
	pending, err := bob.Direct(f.ctx, "alice")
	if err != nil || !pending.Pending || pending.ID != "" {
		t.Fatalf("expired request was accepted: %+v %v", pending, err)
	}
}

func status(err error, want int) bool {
	var apiErr *client.Error
	return errors.As(err, &apiErr) && apiErr.Status == want
}
