package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/karatekah/termlink/internal/auth"
	"github.com/karatekah/termlink/internal/chat"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/pkg/protocol"
)

type API struct {
	Auth             *auth.Service
	Chat             *chat.Service
	Requests         *security.Limiter
	IdentityRequests *security.Limiter
	Streams          chan struct{}
	Count            atomic.Uint64
	Failures         atomic.Uint64
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewAPI(a *auth.Service, c *chat.Service) *API {
	ctx, cancel := context.WithCancel(context.Background())
	return &API{Auth: a, Chat: c, Requests: security.NewLimiter(10, 30), IdentityRequests: security.NewLimiter(1.0/60, 3), Streams: make(chan struct{}, 64), ctx: ctx, cancel: cancel}
}
func (a *API) Close() { a.cancel() }
func token(r *http.Request) string {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func failure(w http.ResponseWriter, err error) {
	status := 500
	msg := "internal error"
	switch {
	case errors.Is(err, security.ErrUnauthorized):
		status = 401
		msg = err.Error()
	case errors.Is(err, security.ErrForbidden):
		status = 403
		msg = err.Error()
	case errors.Is(err, security.ErrInvalid):
		status = 400
		msg = err.Error()
	case errors.Is(err, security.ErrConflict):
		status = 409
		msg = err.Error()
	case errors.Is(err, security.ErrLimited):
		status = 429
		msg = err.Error()
	}
	write(w, status, protocol.Failure{Error: msg})
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16384)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return security.ErrInvalid
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return security.ErrInvalid
	}
	return nil
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if a.Chat.DB.Ping(ctx) != nil {
			w.WriteHeader(503)
			return
		}
		write(w, 200, map[string]string{"status": "ok"})
	})
	// Metrics are operator-local; no identities or message content appear in labels.
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			w.WriteHeader(403)
			return
		}
		if err := a.Auth.Verify(r.Context(), token(r)); err != nil {
			failure(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "termlink_requests_total %d\ntermlink_stream_failures_total %d\ntermlink_active_streams %d\n", a.Count.Load(), a.Failures.Load(), len(a.Streams))
	})
	mux.HandleFunc("POST /v1/auth/challenge", func(w http.ResponseWriter, r *http.Request) {
		var in protocol.ChallengeRequest
		if err := decode(w, r, &in); err != nil {
			failure(w, err)
			return
		}
		if in.Username != "" {
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			if !a.IdentityRequests.Allow(host) {
				failure(w, security.ErrLimited)
				return
			}
		}
		out, err := a.Auth.Challenge(r.Context(), in)
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("POST /v1/auth/proof", func(w http.ResponseWriter, r *http.Request) {
		var in protocol.Proof
		if err := decode(w, r, &in); err != nil {
			failure(w, err)
			return
		}
		out, err := a.Auth.Finish(r.Context(), in)
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("POST /v1/recovery/start", func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !a.IdentityRequests.Allow(host) {
			failure(w, security.ErrLimited)
			return
		}
		var in protocol.RecoveryStart
		if err := decode(w, r, &in); err != nil {
			failure(w, err)
			return
		}
		out, err := a.Auth.StartRecovery(r.Context(), in)
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("POST /v1/recovery/complete", func(w http.ResponseWriter, r *http.Request) {
		var in protocol.RecoveryProof
		if err := decode(w, r, &in); err != nil {
			failure(w, err)
			return
		}
		out, err := a.Auth.CompleteRecovery(r.Context(), in)
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("DELETE /v1/session", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Auth.Logout(r.Context(), token(r)); err != nil {
			failure(w, err)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("DELETE /v1/account", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Auth.DeleteAccount(r.Context(), token(r)); err != nil {
			failure(w, err)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /v1/conversations", func(w http.ResponseWriter, r *http.Request) {
		out, err := a.Chat.List(r.Context(), token(r))
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("POST /v1/direct", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Username string `json:"username"`
		}
		if err := decode(w, r, &in); err != nil {
			failure(w, err)
			return
		}
		out, err := a.Chat.Direct(r.Context(), token(r), in.Username)
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("GET /v1/direct/requests", func(w http.ResponseWriter, r *http.Request) {
		out, err := a.Chat.DirectRequests(r.Context(), token(r))
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("DELETE /v1/direct/requests/{username}", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Chat.DeclineDirect(r.Context(), token(r), r.PathValue("username")); err != nil {
			failure(w, err)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("PUT /v1/direct/blocks/{username}", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Chat.BlockDirect(r.Context(), token(r), r.PathValue("username")); err != nil {
			failure(w, err)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("DELETE /v1/direct/blocks/{username}", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Chat.UnblockDirect(r.Context(), token(r), r.PathValue("username")); err != nil {
			failure(w, err)
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /v1/conversations/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		after, limit, err := cursor(r)
		if err != nil {
			failure(w, err)
			return
		}
		out, err := a.Chat.History(r.Context(), token(r), r.PathValue("id"), after, limit)
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("POST /v1/conversations/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		var in protocol.Send
		if err := decode(w, r, &in); err != nil {
			failure(w, err)
			return
		}
		out, err := a.Chat.Send(r.Context(), token(r), r.PathValue("id"), in)
		if err != nil {
			failure(w, err)
			return
		}
		write(w, 200, out)
	})
	mux.HandleFunc("GET /v1/conversations/{id}/events", a.events)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		a.Count.Add(1)
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !a.Requests.Allow(host) {
			failure(w, security.ErrLimited)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if strings.HasSuffix(r.URL.Path, "/events") {
			mux.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}
func cursor(r *http.Request) (int64, int, error) {
	after := int64(0)
	limit := 100
	var err error
	if raw := r.URL.Query().Get("after"); raw != "" {
		after, err = strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, 0, security.ErrInvalid
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			return 0, 0, security.ErrInvalid
		}
	}
	if after < 0 || limit < 1 || limit > 100 {
		return 0, 0, security.ErrInvalid
	}
	return after, limit, nil
}
func (a *API) events(w http.ResponseWriter, r *http.Request) {
	after, _, err := cursor(r)
	if err != nil {
		failure(w, err)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	stop := context.AfterFunc(a.ctx, cancel)
	defer stop()
	check, done := context.WithTimeout(ctx, 5*time.Second)
	_, err = a.Chat.History(check, token(r), r.PathValue("id"), after, 1)
	done()
	if err != nil {
		failure(w, err)
		return
	}
	select {
	case a.Streams <- struct{}{}:
		defer func() { <-a.Streams }()
	default:
		failure(w, security.ErrLimited)
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1024)
	ctx = conn.CloseRead(ctx) // clients send via HTTPS; this also detects disconnects.
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			ping, pcancel := context.WithTimeout(ctx, 5*time.Second)
			err = conn.Ping(ping)
			pcancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			query, qcancel := context.WithTimeout(ctx, 5*time.Second)
			page, e := a.Chat.History(query, token(r), r.PathValue("id"), after, 100)
			qcancel()
			if e != nil {
				a.Failures.Add(1)
				_ = conn.Close(websocket.StatusPolicyViolation, "subscription ended")
				return
			}
			if len(page.Messages) > 0 {
				writeCtx, wcancel := context.WithTimeout(ctx, 5*time.Second)
				e = wsjson.Write(writeCtx, conn, page)
				wcancel()
				if e != nil {
					return
				}
				after = page.Next
			}
		}
	}
}
