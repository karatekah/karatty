//go:build !windows

package tests

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/creack/pty"
	"github.com/karatekah/karatty/internal/security"
	"golang.org/x/crypto/ssh"
)

type tuiCapture struct {
	mu sync.Mutex
	bytes.Buffer
}

func (c *tuiCapture) Write(data []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Buffer.Write(data)
}
func (c *tuiCapture) plain() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ansi.Strip(c.Buffer.String())
}

type tuiProcess struct {
	terminal *os.File
	output   *tuiCapture
	done     chan error
}

func (p *tuiProcess) waitFor(t *testing.T, match func(string) bool) string {
	t.Helper()
	deadline := time.After(15 * time.Second)
	tick := time.NewTicker(25 * time.Millisecond)
	defer tick.Stop()
	for {
		text := p.output.plain()
		if match(text) {
			return text
		}
		select {
		case <-deadline:
			tail := []rune(text)
			if len(tail) > 1200 {
				tail = tail[len(tail)-1200:]
			}
			t.Fatalf("native encrypted TUI did not reach the expected state; test terminal tail: %q", string(tail))
		case <-tick.C:
		}
	}
}

func (p *tuiProcess) typeLine(t *testing.T, value string) {
	t.Helper()
	if _, err := p.terminal.Write([]byte(value + "\r")); err != nil {
		t.Fatal("could not write to native TUI", err)
	}
}

func TestE2EEClientNativeTUI(t *testing.T) {
	cryptoBin := os.Getenv("KARATTY_CRYPTO_BIN")
	if cryptoBin == "" {
		t.Skip("set KARATTY_CRYPTO_BIN to run the native encrypted TUI")
	}
	f := setup(t)
	f.a.PublicSignup, f.c.DirectOnly, f.c.CiphertextOnly = true, true, true
	f.api.IdentityRequests = security.NewLimiter(1000, 1000)
	root := t.TempDir()
	clientBin := filepath.Join(root, "karatty")
	build := exec.CommandContext(f.ctx, "go", "build", "-o", clientBin, "../cmd/termlink")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build native client: %v: %s", err, output)
	}
	caPath := filepath.Join(root, "ca.pem")
	if err := os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: f.http.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	makeUser := func(name string) (string, string) {
		_, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		signer, err := ssh.NewSignerFromKey(private)
		if err != nil {
			t.Fatal(err)
		}
		keyPath := filepath.Join(root, name+".key")
		block, err := ssh.MarshalPrivateKey(private, "test-only")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(keyPath, pem.EncodeToMemory(block), 0600); err != nil {
			t.Fatal(err)
		}
		http := f.httpClient(t)
		if _, err = http.Login(f.ctx, signer, name, ""); err != nil {
			t.Fatal(err)
		}
		peer := "alice"
		if name == "alice" {
			peer = "bob"
		}
		if _, err = http.Direct(f.ctx, peer); err != nil && name != "alice" {
			t.Fatal(err)
		}
		return keyPath, filepath.Join(root, name+"-state")
	}
	aliceKey, aliceState := makeUser("alice")
	bobKey, bobState := makeUser("bob")
	// Alice's first request preceded Bob's registration; open the reciprocal pair.
	aliceHTTP := f.httpClient(t)
	keyData, err := os.ReadFile(aliceKey)
	if err != nil {
		t.Fatal(err)
	}
	aliceSigner, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = aliceHTTP.Login(f.ctx, aliceSigner, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err = aliceHTTP.Direct(f.ctx, "bob"); err != nil {
		t.Fatal(err)
	}
	start := func(key, state, peer string) *tuiProcess {
		cmd := exec.CommandContext(f.ctx, clientBin, "chat", "--server", f.http.URL, "--ca-cert", caPath, "--key", key, "--state-dir", state, "--crypto-bin", cryptoBin)
		cmd.Env = append(os.Environ(), "TERM=xterm-256color", "NO_COLOR=1")
		terminal, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 30, Cols: 120})
		if err != nil {
			t.Fatal(err)
		}
		p := &tuiProcess{terminal: terminal, output: &tuiCapture{}, done: make(chan error, 1)}
		go func() { _, _ = io.Copy(p.output, terminal) }()
		go func() { p.done <- cmd.Wait(); close(p.done) }()
		t.Cleanup(func() {
			_, _ = terminal.Write([]byte{3})
			select {
			case <-p.done:
			case <-time.After(3 * time.Second):
				_ = cmd.Process.Kill()
				<-p.done
			}
			_ = terminal.Close()
		})
		p.waitFor(t, func(text string) bool { return strings.Contains(text, "archivo local cifrado") })
		p.typeLine(t, "native local passphrase")
		// The initial room selection clears identityView. Its tab must be on
		// screen before /myfingerprint, even if the peer is not online yet.
		p.waitFor(t, func(text string) bool {
			return strings.Contains(text, "["+peer+"]")
		})
		return p
	}
	alice, bob := start(aliceKey, aliceState, "bob"), start(bobKey, bobState, "alice")
	fingerprint := regexp.MustCompile(`Tu huella:\s*([0-9a-f]{64})`)
	own := func(p *tuiProcess) string {
		p.typeLine(t, "/myfingerprint")
		text := p.waitFor(t, func(text string) bool { return fingerprint.MatchString(text) })
		return fingerprint.FindStringSubmatch(text)[1]
	}
	aliceFingerprint, bobFingerprint := own(alice), own(bob)
	alice.typeLine(t, "/verify bob "+bobFingerprint)
	bob.typeLine(t, "/verify alice "+aliceFingerprint)
	for _, p := range []*tuiProcess{alice, bob} {
		// The renderer can reuse the prefix of "Conectado" already on screen.
		// SYNC is a new marker emitted on the first successful subscription.
		p.waitFor(t, func(text string) bool { return strings.Contains(text, "SYNC 0") })
	}
	alice.typeLine(t, "mensaje nativo cifrado")
	bob.waitFor(t, func(text string) bool { return strings.Contains(text, "mensaje nativo cifrado") })
	alice.waitFor(t, func(text string) bool { return strings.Contains(text, "ACK") })
	var body string
	if err = f.db.QueryRow(f.ctx, "SELECT body FROM messages LIMIT 1").Scan(&body); err != nil || !security.Envelope(body) || strings.Contains(body, "mensaje nativo cifrado") {
		t.Fatal("native TUI did not persist an opaque encrypted envelope", err)
	}
}
