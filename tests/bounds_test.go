package tests

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/karatekah/karatty/internal/security"
	"github.com/karatekah/karatty/pkg/protocol"
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

func TestHTTPAuthAndNoStore(t *testing.T) {
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "http-contract")
	if err != nil {
		t.Fatal(err)
	}
	alice, _, session := f.register(t, "alice", conv)
	check := func(method, path, authorization string, want int) {
		t.Helper()
		req, err := http.NewRequest(method, f.http.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if authorization != "" {
			req.Header.Set("Authorization", authorization)
		}
		resp, err := f.http.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != want {
			t.Fatalf("%s %s (authorization present: %t): status %d, want %d", method, path, authorization != "", resp.StatusCode, want)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Fatalf("%s %s: Cache-Control = %q, want no-store", method, path, got)
		}
	}

	check("GET", "/healthz", "", 200)
	check("GET", "/metrics", "", 401)
	check("GET", "/metrics", session.Token, 401)
	check("GET", "/metrics", "Bearer "+session.Token, 200)
	metrics, err := alice.Metrics(f.ctx)
	if err != nil || !strings.Contains(metrics, "termlink_requests_total") {
		t.Fatalf("authenticated metrics: %v %q", err, metrics)
	}
	check("GET", "/v1/conversations", session.Token, 401)
	check("GET", "/v1/conversations", "Bearer "+session.Token, 200)
	check("DELETE", "/v1/session", "", 401)
	check("DELETE", "/v1/session", session.Token, 401)
	check("DELETE", "/v1/session", "Bearer "+session.Token, 204)
	check("GET", "/metrics", "Bearer "+session.Token, 401)
	check("GET", "/missing", "", 404)
}
