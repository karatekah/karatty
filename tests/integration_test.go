package tests

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/karatekah/termlink/internal/auth"
	"github.com/karatekah/termlink/internal/chat"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/internal/storage"
	"github.com/karatekah/termlink/internal/transport"
	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

type fixture struct {
	ctx    context.Context
	db     *pgxpool.Pool
	config *pgxpool.Config
	a      *auth.Service
	c      *chat.Service
	api    *transport.API
	http   *httptest.Server
}

func setup(t *testing.T) *fixture {
	t.Helper()
	url := os.Getenv("TERMLINK_TEST_DATABASE_URL")
	if url == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("integration database required in CI")
		}
		t.Skip("set TERMLINK_TEST_DATABASE_URL to run real PostgreSQL tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	root, err := storage.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "test_" + security.ID()
	if _, err = root.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	cfg.MaxConns = 16
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{ctx: ctx, db: db, config: cfg}
	f.a = &auth.Service{DB: db}
	f.c = chat.New(db)
	if err = storage.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	f.api = transport.NewAPI(f.a, f.c)
	f.api.Requests = security.NewLimiter(1000, 1000)
	f.http = httptest.NewTLSServer(f.api.Handler())
	t.Cleanup(func() {
		f.api.Close()
		f.http.Close()
		f.db.Close()
		_, _ = root.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		root.Close()
	})
	return f
}
func signer(t *testing.T) ssh.Signer {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func (f *fixture) httpClient(t *testing.T) *client.HTTP {
	t.Helper()
	c, err := client.New(f.http.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	c.Transport = f.http.Client()
	return c
}
func (f *fixture) register(t *testing.T, name, conv string) (*client.HTTP, ssh.Signer, protocol.Session) {
	t.Helper()
	invite, err := f.a.Invite(f.ctx, conv, 1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	key := signer(t)
	c := f.httpClient(t)
	session, err := c.Login(f.ctx, key, name, invite)
	if err != nil {
		t.Fatal(err)
	}
	return c, key, session
}
func receive(t *testing.T, ctx context.Context, ch <-chan client.Event) client.Event {
	t.Helper()
	select {
	case e, ok := <-ch:
		if !ok {
			t.Fatal("stream closed")
		}
		return e
	case <-ctx.Done():
		t.Fatal("stream timeout")
		return client.Event{}
	}
}
func requireError(t *testing.T, err error, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("got %v; want %v", err, want)
	}
}

func TestInvitationProofExpiryAndRevocation(t *testing.T) {
	f := setup(t)
	key := signer(t)
	inv, err := f.a.Invite(f.ctx, "", 1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	r := protocol.ChallengeRequest{PublicKey: string(ssh.MarshalAuthorizedKey(key.PublicKey())), Username: "alice", Invitation: inv}
	challenge, err := f.a.Challenge(f.ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	wrong := signer(t)
	bad, _ := wrong.Sign(rand.Reader, []byte(challenge.Payload))
	_, err = f.a.Finish(f.ctx, protocol.Proof{ID: challenge.ID, Signature: ssh.Marshal(bad)})
	requireError(t, err, security.ErrUnauthorized)
	sig, _ := key.Sign(rand.Reader, []byte(challenge.Payload))
	_, err = f.a.Finish(f.ctx, protocol.Proof{ID: challenge.ID, Signature: ssh.Marshal(sig)})
	requireError(t, err, security.ErrUnauthorized)
	challenge, err = f.a.Challenge(f.ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	sig, _ = key.Sign(rand.Reader, []byte(challenge.Payload))
	proof := protocol.Proof{ID: challenge.ID, Signature: ssh.Marshal(sig)}
	session, err := f.a.Finish(f.ctx, proof)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.a.Finish(f.ctx, proof)
	requireError(t, err, security.ErrUnauthorized)
	other := f.httpClient(t)
	_, err = other.Login(f.ctx, signer(t), "bob", inv)
	if err == nil {
		t.Fatal("reused invitation")
	}
	expired, err := f.a.Invite(f.ctx, "", 1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.db.Exec(f.ctx, "UPDATE invitations SET expires_at=now()-interval '1 second' WHERE token_hash=$1", security.Hash(expired))
	if err != nil {
		t.Fatal(err)
	}
	_, err = other.Login(f.ctx, signer(t), "expired", expired)
	if err == nil {
		t.Fatal("accepted expired invitation")
	}
	challenge, err = f.a.Challenge(f.ctx, protocol.ChallengeRequest{PublicKey: r.PublicKey})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.db.Exec(f.ctx, "UPDATE auth_challenges SET expires_at=now()-interval '1 second' WHERE id=$1", challenge.ID)
	if err != nil {
		t.Fatal(err)
	}
	sig, _ = key.Sign(rand.Reader, []byte(challenge.Payload))
	_, err = f.a.Finish(f.ctx, protocol.Proof{ID: challenge.ID, Signature: ssh.Marshal(sig)})
	requireError(t, err, security.ErrUnauthorized)
	if err = f.a.RevokeKey(f.ctx, ssh.FingerprintSHA256(key.PublicKey())); err != nil {
		t.Fatal(err)
	}
	_, err = f.c.List(f.ctx, session.Token)
	requireError(t, err, security.ErrUnauthorized)
	_, err = other.Login(f.ctx, key, "", "")
	if err == nil {
		t.Fatal("revoked key logged in")
	}
}

func TestConcurrentInvitationLimitedUses(t *testing.T) {
	f := setup(t)
	invite, err := f.a.Invite(f.ctx, "", 1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		key := signer(t)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := f.httpClient(t)
			_, e := c.Login(f.ctx, key, fmt.Sprintf("user%d", i), invite)
			results <- e
		}(i)
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("invitation had %d successful uses", success)
	}
}

func TestMembershipOrderDedupAndPersistence(t *testing.T) {
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "general")
	if err != nil {
		t.Fatal(err)
	}
	alice, key, _ := f.register(t, "alice", conv)
	bob, _, _ := f.register(t, "bob", conv)
	eve, _, _ := f.register(t, "eve", "")
	if _, err = eve.History(f.ctx, conv, 0, 100); err == nil {
		t.Fatal("non-member read")
	}
	if _, err = eve.Send(f.ctx, conv, protocol.Send{ClientID: "stolen", Body: "no"}); err == nil {
		t.Fatal("non-member sent")
	}
	events := eve.Stream(f.ctx, conv, 0)
	if e := receive(t, f.ctx, events); !e.Fatal {
		t.Fatal("non-member subscribed")
	}
	streamCtx, cancel := context.WithCancel(f.ctx)
	defer cancel()
	events = bob.Stream(streamCtx, conv, 0)
	req := protocol.Send{ClientID: "retry-stable-id", Body: "Hola desde el cliente"}
	first, err := alice.Send(f.ctx, conv, req)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := alice.Send(f.ctx, conv, req)
	if err != nil || first.ID != retry.ID || first.Seq != retry.Seq {
		t.Fatal("retry was not deduplicated", err)
	}
	req.Body = "different"
	_, err = alice.Send(f.ctx, conv, req)
	var apiErr *client.Error
	if !errors.As(err, &apiErr) || apiErr.Status != 409 {
		t.Fatal("conflicting retry accepted", err)
	}
	e := receive(t, f.ctx, events)
	if e.Err != nil || len(e.Page.Messages) != 1 || e.Page.Messages[0].ID != first.ID {
		t.Fatal("live event mismatch", e.Err)
	}
	var wg sync.WaitGroup
	seqs := make(chan int64, 8)
	fail := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sender := alice
			if i%2 == 1 {
				sender = bob
			}
			msg, e := sender.Send(f.ctx, conv, protocol.Send{ClientID: fmt.Sprintf("parallel-%d", i), Body: "parallel"})
			if e != nil {
				fail <- e
				return
			}
			seqs <- msg.Seq
		}(i)
	}
	wg.Wait()
	close(seqs)
	close(fail)
	for e := range fail {
		t.Fatal(e)
	}
	seen := map[int64]bool{}
	for seq := range seqs {
		if seen[seq] {
			t.Fatal("duplicate sequence")
		}
		seen[seq] = true
	}
	for i := int64(2); i <= 9; i++ {
		if !seen[i] {
			t.Fatalf("missing sequence %d", i)
		}
	}
	cancel()
	if err = alice.Logout(f.ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.List(f.ctx); err == nil {
		t.Fatal("logged out token still accepted")
	}
	// Reopen a database pool and fresh domain service, with no in-memory message state.
	db2, err := pgxpool.NewWithConfig(f.ctx, f.config.Copy())
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	restarted := chat.New(db2)
	fresh := f.httpClient(t)
	session, err := fresh.Login(f.ctx, key, "", "")
	if err != nil {
		t.Fatal(err)
	}
	page, err := restarted.History(f.ctx, session.Token, conv, 0, 3)
	if err != nil || len(page.Messages) != 3 || page.Next != 3 {
		t.Fatal("persistence/page one", err)
	}
	page, err = restarted.History(f.ctx, session.Token, conv, page.Next, 100)
	if err != nil || len(page.Messages) != 6 || page.Messages[0].Seq != 4 || page.Next != 9 {
		t.Fatal("persistence/page two", err)
	}
	dm, err := alice.Direct(f.ctx, "bob")
	if err == nil {
		t.Fatal("logged-out user opened direct", dm)
	}
	dm, err = fresh.Direct(f.ctx, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if !dm.Pending || dm.ID != "" {
		t.Fatal("direct request opened a chat without acceptance")
	}
	requests, err := bob.DirectRequests(f.ctx)
	if err != nil || len(requests) != 1 || requests[0].Username != "alice" {
		t.Fatal("direct request not visible to recipient", requests, err)
	}
	dm2, err := bob.Direct(f.ctx, "alice")
	if err != nil || dm2.Pending || dm2.ID == "" {
		t.Fatal("recipient could not accept direct", err)
	}
	dm, err = fresh.Direct(f.ctx, "bob")
	if err != nil || dm.ID != dm2.ID {
		t.Fatal("direct pair not canonical", err)
	}
	if _, err = eve.History(f.ctx, dm.ID, 0, 100); err == nil {
		t.Fatal("direct leaked")
	}
}

func TestLiveMembershipAndKeyRevocation(t *testing.T) {
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "private")
	if err != nil {
		t.Fatal(err)
	}
	alice, key, session := f.register(t, "alice", conv)
	events := alice.Stream(f.ctx, conv, 0)
	_, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: "initial", Body: "before revocation"})
	if err != nil {
		t.Fatal(err)
	}
	if e := receive(t, f.ctx, events); e.Err != nil {
		t.Fatal(e.Err)
	}
	if err = f.c.RevokeMember(f.ctx, conv, session.UserID); err != nil {
		t.Fatal(err)
	}
	if e := receive(t, f.ctx, events); !e.Fatal {
		t.Fatal("subscription survived revocation")
	}
	if _, err = alice.History(f.ctx, conv, 0, 100); err == nil {
		t.Fatal("revoked member read")
	}
	if _, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: "after", Body: "forbidden"}); err == nil {
		t.Fatal("revoked member sent")
	}
	if err = f.a.RevokeKey(f.ctx, ssh.FingerprintSHA256(key.PublicKey())); err != nil {
		t.Fatal(err)
	}
	if _, err = alice.List(f.ctx); err == nil {
		t.Fatal("revoked key session active")
	}
}

func TestWebSocketReconnectFromSequence(t *testing.T) {
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "reconnect")
	if err != nil {
		t.Fatal(err)
	}
	alice, _, _ := f.register(t, "alice", conv)
	// A proxy terminates the first event connection after one delivered page,
	// exactly like a transport failure. Subsequent dials reach the real API.
	var mu sync.Mutex
	connections := 0
	interrupted := make(chan struct{})
	resume := make(chan struct{})
	defer func() {
		select {
		case <-resume:
		default:
			close(resume)
		}
	}()
	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		connections++
		n := connections
		mu.Unlock()
		if n > 1 {
			select {
			case <-resume:
			case <-r.Context().Done():
				return
			}
		}
		upstream, _, e := websocket.Dial(r.Context(), strings.Replace(f.http.URL, "https://", "wss://", 1)+r.URL.RequestURI(), &websocket.DialOptions{HTTPClient: f.http.Client(), HTTPHeader: r.Header})
		if e != nil {
			w.WriteHeader(502)
			return
		}
		defer upstream.CloseNow()
		downstream, e := websocket.Accept(w, r, nil)
		if e != nil {
			return
		}
		defer downstream.CloseNow()
		for {
			var p protocol.Page
			if e = wsjson.Read(r.Context(), upstream, &p); e != nil {
				return
			}
			if e = wsjson.Write(r.Context(), downstream, p); e != nil {
				return
			}
			if n == 1 {
				close(interrupted)
				return
			}
		}
	}))
	defer proxy.Close()
	subscriber, err := client.New(proxy.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	subscriber.Transport = proxy.Client()
	subscriber.Token = alice.Token
	// websocket.Dial uses the trusted client for this test proxy.
	streamCtx, cancel := context.WithCancel(f.ctx)
	defer cancel()
	events := subscriber.Stream(streamCtx, conv, 0)
	_, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: "one", Body: "before outage"})
	if err != nil {
		t.Fatal(err)
	}
	// See client.DialOptions HTTPClient: production uses system CA roots.
	e := receive(t, f.ctx, events)
	if e.Err != nil {
		t.Fatal(e.Err)
	}
	if e.Page.Next != 1 {
		t.Fatal("missing first sequence")
	}
	<-interrupted
	_, err = alice.Send(f.ctx, conv, protocol.Send{ClientID: "two", Body: "during outage"})
	if err != nil {
		t.Fatal(err)
	}
	close(resume)
	for {
		e = receive(t, f.ctx, events)
		if e.Fatal {
			t.Fatal(e.Err)
		}
		if e.Err == nil {
			if len(e.Page.Messages) != 1 || e.Page.Messages[0].Seq != 2 {
				t.Fatal("reconnect skipped or duplicated message")
			}
			break
		}
	}
}

func startSSH(t *testing.T, f *fixture) (string, ssh.HostKeyCallback) {
	t.Helper()
	// MarshalPrivateKey requires the underlying private key rather than a Signer.
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	host, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(private, "test host")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "host")
	if err = os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatal(err)
	}
	srv, err := transport.NewSSH(f.a, f.c, path)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(security.BoundListener(listener, 16))
	t.Cleanup(func() { srv.Close() })
	callback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if !bytes.Equal(key.Marshal(), host.PublicKey().Marshal()) {
			return errors.New("host key mismatch")
		}
		return nil
	}
	return listener.Addr().String(), callback
}
func sshConnect(t *testing.T, addr string, key ssh.Signer, host ssh.HostKeyCallback) *ssh.Client {
	t.Helper()
	c, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{User: "chat", Auth: []ssh.AuthMethod{ssh.PublicKeys(key)}, HostKeyCallback: host, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}
func sshExec(t *testing.T, c *ssh.Client, command, body string) ([]byte, error) {
	t.Helper()
	session, err := c.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	session.Stdin = strings.NewReader(body)
	return session.Output(command)
}

func TestSSHAndNativeAcceptance(t *testing.T) {
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "general")
	if err != nil {
		t.Fatal(err)
	}
	alice, aliceKey, _ := f.register(t, "alice", conv)
	bob, bobKey, _ := f.register(t, "bob", conv)
	_, eveKey, _ := f.register(t, "eve", "")
	addr, host := startSSH(t, f)
	sshBob := sshConnect(t, addr, bobKey, host)
	sshEve := sshConnect(t, addr, eveKey, host)
	// API -> SSH history, then SSH -> native WSS, followed by fresh SSH login.
	first, err := alice.Send(f.ctx, conv, protocol.Send{ClientID: "native-to-ssh", Body: "Hola desde nativo"})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := sshExec(t, sshBob, "history "+conv+" 0", "")
	if err != nil {
		t.Fatal(err)
	}
	var page protocol.Page
	if err = json.Unmarshal(raw, &page); err != nil || len(page.Messages) != 1 || page.Messages[0].ID != first.ID {
		t.Fatal("SSH history mismatch", err)
	}
	streamCtx, cancel := context.WithCancel(f.ctx)
	defer cancel()
	events := alice.Stream(streamCtx, conv, first.Seq)
	raw, err = sshExec(t, sshBob, "send "+conv+" ssh-to-native", "Hola desde SSH")
	if err != nil {
		t.Fatal(err)
	}
	var second protocol.Message
	if err = json.Unmarshal(raw, &second); err != nil {
		t.Fatal(err)
	}
	if e := receive(t, f.ctx, events); e.Err != nil || len(e.Page.Messages) != 1 || e.Page.Messages[0].ID != second.ID {
		t.Fatal("SSH event mismatch", e.Err)
	}
	// Same ID over separate authenticated SSH sessions remains idempotent.
	raw, err = sshExec(t, sshBob, "send "+conv+" ssh-to-native", "Hola desde SSH")
	if err != nil {
		t.Fatal(err)
	}
	var retry protocol.Message
	_ = json.Unmarshal(raw, &retry)
	if retry.ID != second.ID {
		t.Fatal("SSH retry duplicated")
	}
	for _, cmd := range []string{"history " + conv + " 0", "send " + conv + " unauthorized"} {
		if _, e := sshExec(t, sshEve, cmd, "denied"); e == nil {
			t.Fatal("SSH nonmember operation accepted")
		}
	}
	if _, err = sshExec(t, sshBob, "rm -rf /", ""); err == nil {
		t.Fatal("arbitrary command accepted")
	}
	if err = alice.Logout(f.ctx); err != nil {
		t.Fatal(err)
	}
	if err = bob.Logout(f.ctx); err != nil {
		t.Fatal(err)
	}
	sshBob.Close()
	fresh := f.httpClient(t)
	if _, err = fresh.Login(f.ctx, aliceKey, "", ""); err != nil {
		t.Fatal(err)
	}
	page, err = fresh.History(f.ctx, conv, 0, 100)
	if err != nil || len(page.Messages) != 2 {
		t.Fatal("history after native logout", err)
	}
	sshBob = sshConnect(t, addr, bobKey, host)
	raw, err = sshExec(t, sshBob, "history "+conv+" 0", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &page); err != nil || len(page.Messages) != 2 {
		t.Fatal("history after SSH reconnect", err)
	}
	// Real PTY runs the shared Bubble Tea model and renders persisted content.
	session, err := sshBob.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err = session.RequestPty("xterm-256color", 24, 100, ssh.TerminalModes{}); err != nil {
		t.Fatal(err)
	}
	out, err := session.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	input, err := session.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Shell(); err != nil {
		t.Fatal(err)
	}
	rendered := make(chan string, 1)
	go func() {
		b := make([]byte, 8192)
		var all strings.Builder
		for {
			n, e := out.Read(b)
			all.Write(b[:n])
			if strings.Contains(all.String(), "Hola desde SSH") {
				rendered <- all.String()
				return
			}
			if e != nil {
				rendered <- all.String()
				return
			}
		}
	}()
	select {
	case text := <-rendered:
		if !strings.Contains(text, "Karatty") || !strings.Contains(text, "Hola desde SSH") {
			t.Fatal("TUI did not render persisted messages")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("SSH TUI timeout")
	}
	_, _ = input.Write([]byte{3})
	if err = f.a.RevokeKey(f.ctx, ssh.FingerprintSHA256(bobKey.PublicKey())); err != nil {
		t.Fatal(err)
	}
	if _, err = sshExec(t, sshBob, "list", ""); err == nil {
		t.Fatal("revoked SSH key session accepted")
	}
}

// Build native CLI and run actual processes against HTTPS, not only its library.
func TestNativeExecutable(t *testing.T) {
	if testing.Short() {
		t.Skip("native executable build")
	}
	f := setup(t)
	conv, err := f.c.Channel(f.ctx, "cli")
	if err != nil {
		t.Fatal(err)
	}
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	invite, err := f.a.Invite(f.ctx, conv, 1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	c := f.httpClient(t)
	if _, err = c.Login(f.ctx, key, "cli", invite); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Send(f.ctx, conv, protocol.Send{ClientID: "cli-persisted", Body: "from executable test"}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	if err = os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: f.http.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(private, "test")
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(dir, "key")
	if err = os.WriteFile(keyPath, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "termlink.exe")
	build := exec.CommandContext(f.ctx, "go", "build", "-o", binary, "../cmd/termlink")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("CLI build: %v %s", err, out)
	}
	cmd := exec.CommandContext(f.ctx, binary, "history", "--legacy", "--server", f.http.URL, "--ca-cert", caPath, "--key", keyPath, "--conversation", conv)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI history failed: %v %s", err, out)
	}
	var page protocol.Page
	if err = json.Unmarshal(out, &page); err != nil || len(page.Messages) != 1 || page.Messages[0].Body != "from executable test" {
		t.Fatal("CLI persisted history", err)
	}
	cmd = exec.CommandContext(f.ctx, binary, "metrics", "--server", f.http.URL, "--ca-cert", caPath, "--key", keyPath)
	out, err = cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "termlink_requests_total") {
		t.Fatalf("CLI metrics failed: %v %s", err, out)
	}
	// Native executable -> authenticated SSH, then SSH -> a fresh executable.
	_, bobKey, _ := f.register(t, "bob", conv)
	addr, host := startSSH(t, f)
	sshBob := sshConnect(t, addr, bobKey, host)
	cmd = exec.CommandContext(f.ctx, binary, "send", "--legacy", "--server", f.http.URL, "--ca-cert", caPath, "--key", keyPath, "--conversation", conv, "--id", "native-process-send", "--message-stdin")
	cmd.Stdin = strings.NewReader("native process message")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native send: %v %s", err, out)
	}
	raw, err := sshExec(t, sshBob, "history "+conv+" 1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &page); err != nil || len(page.Messages) != 1 || page.Messages[0].Body != "native process message" {
		t.Fatal("native -> SSH", err)
	}
	if _, err = sshExec(t, sshBob, "send "+conv+" ssh-process-reply", "SSH reply to executable"); err != nil {
		t.Fatal(err)
	}
	cmd = exec.CommandContext(f.ctx, binary, "history", "--legacy", "--server", f.http.URL, "--ca-cert", caPath, "--key", keyPath, "--conversation", conv, "--after", "2")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native reconnect: %v %s", err, out)
	}
	if err = json.Unmarshal(out, &page); err != nil || len(page.Messages) != 1 || page.Messages[0].Body != "SSH reply to executable" {
		t.Fatal("SSH -> native", err)
	}
	noArgs := exec.CommandContext(f.ctx, binary)
	if err = noArgs.Run(); err == nil {
		t.Fatal("CLI accepted missing key")
	}
	cmd = exec.CommandContext(f.ctx, binary, "delete-account", "--server", f.http.URL, "--ca-cert", caPath, "--key", keyPath, "--confirm-delete")
	if out, err = cmd.CombinedOutput(); err != nil {
		t.Fatalf("CLI account deletion failed: %v %s", err, out)
	}
	if _, err = c.List(f.ctx); err == nil {
		t.Fatal("CLI deletion left the account session active")
	}
}
