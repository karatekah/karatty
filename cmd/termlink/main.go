package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
	"github.com/karatekah/termlink/pkg/tui"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	command := "chat"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	base := flags.String("server", "https://127.0.0.1:8443", "API URL")
	keyPath := flags.String("key", "", "Ed25519 private key path")
	username := flags.String("username", "", "username for registration")
	invitationStdin := flags.Bool("invitation-stdin", false, "read invitation from standard input (registration)")
	dev := flags.Bool("dev", false, "allow loopback HTTP")
	caPath := flags.String("ca-cert", "", "additional trusted PEM CA certificate")
	conv := flags.String("conversation", "", "conversation ID for history/send")
	after := flags.Int64("after", 0, "history cursor")
	id := flags.String("id", "", "stable ID for a send retry")
	messageStdin := flags.Bool("message-stdin", false, "send: read one message from standard input")
	peer := flags.String("peer", "", "username for direct chat")
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	switch command {
	case "chat", "register", "login", "list", "history", "send", "direct", "metrics":
	default:
		return errors.New("commands: chat, register, login, list, history, send, direct, metrics")
	}
	if *keyPath == "" {
		return errors.New("provide --key with an Ed25519 private key path")
	}
	raw, err := os.ReadFile(*keyPath)
	if err != nil {
		return errors.New("cannot read private key file")
	}
	signer, err := ssh.ParsePrivateKey(raw)
	var missing *ssh.PassphraseMissingError
	if errors.As(err, &missing) {
		pass, e := readSecret("Frase de la clave SSH: ")
		if e != nil {
			return errors.New("cannot read key passphrase; use an interactive terminal")
		}
		signer, err = ssh.ParsePrivateKeyWithPassphrase(raw, pass)
		for i := range pass {
			pass[i] = 0
		}
	}
	for i := range raw {
		raw[i] = 0
	}
	if err != nil {
		return errors.New("cannot unlock SSH key")
	}
	c, err := client.New(*base, *dev)
	if err != nil {
		return err
	}
	if *caPath != "" {
		ca, e := os.ReadFile(*caPath)
		if e != nil {
			return errors.New("cannot read CA certificate")
		}
		roots, e := x509.SystemCertPool()
		if e != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(ca) {
			return errors.New("invalid CA certificate")
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
		c.Transport = &http.Client{Transport: transport, Timeout: 10 * time.Second}
	}
	invite := ""
	name := ""
	if command == "register" {
		name = *username
		if name == "" {
			return errors.New("registration requires --username")
		}
		if *invitationStdin {
			if _, err = fmt.Fscan(os.Stdin, &invite); err != nil {
				return errors.New("cannot read invitation")
			}
		} else {
			b, e := readSecret("Invitación (entrada oculta): ")
			if e != nil {
				return errors.New("use a terminal or --invitation-stdin")
			}
			invite = string(b)
			for i := range b {
				b[i] = 0
			}
		}
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	loginCtx, done := context.WithTimeout(ctx, 10*time.Second)
	session, err := c.Login(loginCtx, signer, name, invite)
	done()
	if err != nil {
		return err
	}
	defer func() {
		end, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		_ = c.Logout(end)
	}()
	switch command {
	case "register", "login":
		fmt.Printf("Identidad verificada: %s. La sesión temporal se cierra al salir.\n", session.UserID)
		return nil
	case "chat":
		_, err = tea.NewProgram(tui.New(ctx, c), tea.WithContext(ctx)).Run()
		return err
	case "list":
		out, e := c.List(ctx)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	case "history":
		out, e := c.History(ctx, *conv, *after, 100)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	case "direct":
		out, e := c.Direct(ctx, *peer)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	case "metrics":
		out, e := c.Metrics(ctx)
		if e != nil {
			return e
		}
		_, e = fmt.Print(out)
		return e
	case "send":
		if *id == "" {
			return errors.New("send requires --id (reuse it when retrying)")
		}
		var b []byte
		var e error
		if *messageStdin {
			b, e = io.ReadAll(io.LimitReader(os.Stdin, 4098))
			b = []byte(strings.TrimSuffix(strings.TrimSuffix(string(b), "\n"), "\r"))
		} else {
			b, e = readSecret("Mensaje: ")
		}
		if e != nil {
			return e
		}
		body := string(b)
		out, e := c.Send(ctx, *conv, protocol.Send{ClientID: *id, Body: body})
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	}
	return nil
}

// readSecret uses the controlling terminal when stdin carries an invitation or
// a message from a pipe. The secret is never read from that pipe by accident.
func readSecret(prompt string) ([]byte, error) {
	input := os.Stdin
	if !term.IsTerminal(int(input.Fd())) {
		name := "/dev/tty"
		if runtime.GOOS == "windows" {
			name = "CONIN$"
		}
		console, err := os.OpenFile(name, os.O_RDWR, 0)
		if err != nil {
			return nil, err
		}
		defer console.Close()
		input = console
	}
	fmt.Fprint(os.Stderr, prompt)
	secret, err := term.ReadPassword(int(input.Fd()))
	fmt.Fprintln(os.Stderr)
	return secret, err
}
