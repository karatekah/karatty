package secureclient

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/karatekah/karatty/pkg/localstate"
	"github.com/karatekah/karatty/pkg/protocol"
)

func TestLocalArchivePurgesExpiredMessagesAndPendingSends(t *testing.T) {
	root := t.TempDir()
	passphrase := []byte("local phrase for test")
	store, _, err := localstate.Open(root, "https://example.dev", "user-id", passphrase)
	if err != nil {
		t.Fatal(err)
	}
	c := &Client{store: store, state: newDisk("user-id", "ssh-key")}
	old := time.Now().Add(-721 * time.Hour)
	recent := time.Now().Add(-time.Hour)
	c.state.Archive["direct"] = []protocol.Message{
		{ID: "old", ConversationID: "direct", Body: "expired", CreatedAt: old},
		{ID: "recent", ConversationID: "direct", Body: "keep", CreatedAt: recent},
	}
	c.state.Outbox["old"] = outgoing{ClientID: "old", CreatedAt: old}
	c.state.Outbox["recent"] = outgoing{ClientID: "recent", CreatedAt: recent}
	c.state.Uncertain = []outgoing{{ClientID: "old", CreatedAt: old}, {ClientID: "recent", CreatedAt: recent}}
	if err = c.save(c.state); err != nil {
		t.Fatal(err)
	}
	if err = c.Close(); err != nil {
		t.Fatal(err)
	}
	store, data, err := localstate.Open(root, "https://example.dev", "user-id", passphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var saved disk
	if err = json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.Archive["direct"]) != 1 || saved.Archive["direct"][0].ID != "recent" || len(saved.Outbox) != 1 || len(saved.Uncertain) != 1 || saved.Uncertain[0].ClientID != "recent" {
		t.Fatal("expired plaintext or pending send survived local save")
	}
}

func TestArchiveOrdersSendAckBeforeBacklogAndHidesExpired(t *testing.T) {
	c := &Client{state: newDisk("user-id", "ssh-key")}
	for _, seq := range []int64{3, 1, 2} {
		appendArchive(&c.state, protocol.Message{ID: string(rune('a' + seq)), ConversationID: "direct", Seq: seq, CreatedAt: time.Now()})
	}
	appendArchive(&c.state, protocol.Message{ID: "expired", ConversationID: "direct", Seq: 4, CreatedAt: time.Now().Add(-721 * time.Hour)})
	got := c.Archive("direct")
	if len(got) != 3 || got[0].Seq != 1 || got[1].Seq != 2 || got[2].Seq != 3 {
		t.Fatal("archive skipped or reordered backlog after a newer send acknowledgement")
	}
}
