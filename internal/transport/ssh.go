package transport

import (
	"context"
	"encoding/json"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	sshserver "charm.land/ssh"
	"charm.land/wish/v2"
	wishtea "charm.land/wish/v2/bubbletea"
	"github.com/karatekah/karatty/internal/auth"
	"github.com/karatekah/karatty/internal/chat"
	"github.com/karatekah/karatty/internal/security"
	"github.com/karatekah/karatty/pkg/protocol"
	"github.com/karatekah/karatty/pkg/tui"
)

func NewSSH(a *auth.Service, c *chat.Service, hostKey string) (*sshserver.Server, error) {
	attempts := security.NewLimiter(1, 10)
	channels := make(chan struct{}, 64)
	srv, err := wish.NewServer(wish.WithHostKeyPath(hostKey), wish.WithIdleTimeout(5*time.Minute), wish.WithMaxTimeout(8*time.Hour), wish.WithPublicKeyAuth(func(ctx sshserver.Context, key sshserver.PublicKey) bool {
		host, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())
		return attempts.Allow(host) && a.KnownKey(ctx, key)
	}))
	if err != nil {
		return nil, err
	}
	srv.HandshakeTimeout = 10 * time.Second
	srv.ChannelHandlers = map[string]sshserver.ChannelHandler{"session": func(server *sshserver.Server, conn *gossh.ServerConn, ch gossh.NewChannel, ctx sshserver.Context) {
		select {
		case channels <- struct{}{}:
			defer func() { <-channels }()
			sshserver.DefaultSessionHandler(server, conn, ch, ctx)
		default:
			_ = ch.Reject(gossh.ResourceShortage, "session limit")
		}
	}}
	srv.Handler = func(sess sshserver.Session) {
		query, qcancel := context.WithTimeout(sess.Context(), 5*time.Second)
		session, e := a.SSHSession(query, sess.PublicKey())
		qcancel()
		if e != nil {
			_ = sess.Exit(1)
			return
		}
		ctx, cancel := context.WithCancel(sess.Context())
		defer cancel()
		defer func() {
			end, done := context.WithTimeout(context.Background(), time.Second)
			defer done()
			_ = a.Logout(end, session.Token)
		}()
		// Recheck even idle sessions. Revoking a key or membership does not wait for input.
		go func() {
			tick := time.NewTicker(250 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-tick.C:
					check, done := context.WithTimeout(ctx, 5*time.Second)
					_, err := c.List(check, session.Token)
					done()
					if err != nil {
						cancel()
						_ = sess.Close()
						return
					}
				}
			}
		}()
		local := Local{Chat: c, Token: session.Token}
		if len(sess.Command()) > 0 {
			sshCommand(ctx, sess, local)
			return
		}
		_, _, pty := sess.Pty()
		if !pty {
			_, _ = io.WriteString(sess, "Karatty requiere una terminal. Usa ssh -t o los comandos list/requests/direct/history/send.\n")
			_ = sess.Exit(1)
			return
		}
		middleware := wishtea.Middleware(func(s sshserver.Session) (tea.Model, []tea.ProgramOption) {
			return tui.New(ctx, local), []tea.ProgramOption{tea.WithContext(ctx)}
		})
		middleware(func(sshserver.Session) {})(sess)
	}
	return srv, nil
}

// Restricted commands support scripts and deterministic acceptance tests; this
// endpoint never executes a shell or accepts arbitrary program names.
func sshCommand(ctx context.Context, s sshserver.Session, b Local) {
	args := s.Command()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var out any
	var err error
	switch {
	case len(args) == 1 && args[0] == "list":
		out, err = b.List(ctx)
	case len(args) == 1 && args[0] == "requests":
		out, err = b.DirectRequests(ctx)
	case len(args) == 2 && args[0] == "direct":
		out, err = b.Direct(ctx, args[1])
	case len(args) == 2 && args[0] == "decline":
		err = b.DeclineDirect(ctx, args[1])
		out = map[string]bool{"ok": err == nil}
	case len(args) == 2 && args[0] == "block":
		err = b.BlockDirect(ctx, args[1])
		out = map[string]bool{"ok": err == nil}
	case len(args) == 2 && args[0] == "unblock":
		err = b.UnblockDirect(ctx, args[1])
		out = map[string]bool{"ok": err == nil}
	case len(args) == 3 && args[0] == "history":
		var after int64
		after, err = strconv.ParseInt(args[2], 10, 64)
		if err == nil {
			out, err = b.History(ctx, args[1], after, 100)
		} else {
			err = security.ErrInvalid
		}
	case len(args) == 3 && args[0] == "send":
		var body []byte
		body, err = io.ReadAll(io.LimitReader(s, 4098))
		if err == nil {
			out, err = b.Send(ctx, args[1], protocol.Send{ClientID: args[2], Body: strings.TrimSuffix(strings.TrimSuffix(string(body), "\n"), "\r")})
		}
	default:
		err = security.ErrInvalid
	}
	if err != nil {
		_ = json.NewEncoder(s).Encode(protocol.Failure{Error: "operation denied or invalid"})
		_ = s.Exit(1)
		return
	}
	_ = json.NewEncoder(s).Encode(out)
	_ = s.Exit(0)
}
