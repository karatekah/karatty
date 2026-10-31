package tests

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
)

func TestAPIBoundsAndSessionExpiry(t *testing.T) {
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "bounded")
	if err != nil {
		t.Fatal(err)
	}
	alice, _, session := f.register(t, "alice", conv)
	for _, body := range []string{strings.Repeat("x", 4097), "\x1b]52;malicious\a", "hidden\u202e"} {
		if _, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: security.ID(), Body: body}); err == nil {
			t.Fatal("unsafe message accepted")
		}
	}
	f.c.Sends = security.NewLimiter(0, 10)
	for i := 0; i < 10; i++ {
		if _, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: security.ID(), Body: "within burst"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: security.ID(), Body: "over burst"}); err == nil {
		t.Fatal("send limit missing")
	}
	if _, err = alice.History(f.ctx, conv, -1, 100); err == nil {
		t.Fatal("negative cursor")
	}
	if _, err = alice.History(f.ctx, conv, 0, 101); err == nil {
		t.Fatal("unbounded page")
	}
	ctx, cancel := context.WithTimeout(f.ctx, 5*time.Second)
	defer cancel()
	endpoint := strings.Replace(f.http.URL, "https://", "wss://", 1) + "/v1/conversations/" + conv + "/events"
	conn, resp, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{HTTPClient: f.http.Client(), HTTPHeader: http.Header{"Authorization": []string{"Bearer " + session.Token}, "Origin": []string{"https://untrusted.invalid"}}})
	if conn != nil {
		conn.CloseNow()
	}
	if err == nil || resp == nil || resp.StatusCode != 403 {
		t.Fatal("untrusted WebSocket origin accepted", err)
	}
	if _, err = f.db.Exec(f.ctx, "UPDATE sessions SET expires_at=now()-interval '1 second' WHERE token_hash=$1", security.Hash(session.Token)); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.History(f.ctx, conv, 0, 100); err == nil {
		t.Fatal("expired session accepted")
	}
}
