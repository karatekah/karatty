package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/karatekah/termlink/pkg/protocol"
)

func TestMaximumEnvelopePageOverHTTPAndWebSocket(t *testing.T) {
	page := protocol.Page{Next: 100}
	for seq := int64(1); seq <= 100; seq++ {
		page.Messages = append(page.Messages, protocol.Message{Seq: seq, Body: strings.Repeat(`"`, 12288)})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/events") {
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer conn.CloseNow()
			_ = wsjson.Write(r.Context(), conn, page)
			return
		}
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()
	c, err := New(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got, err := c.History(ctx, "direct", 0, 100)
	if err != nil || len(got.Messages) != 100 {
		t.Fatal("maximum HTTP page was truncated", err)
	}
	select {
	case event := <-c.Stream(ctx, "direct", 0):
		if event.Err != nil || len(event.Page.Messages) != 100 {
			t.Fatal("maximum WebSocket page was truncated", event.Err)
		}
	case <-ctx.Done():
		t.Fatal("maximum WebSocket page timed out")
	}
}
