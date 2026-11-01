package main

import (
	"bytes"
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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
	"github.com/karatekah/termlink/pkg/secureclient"
	"github.com/karatekah/termlink/pkg/tui"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var version = "dev"

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
	public := flags.Bool("public", false, "register a pseudonym without an invitation")
	dev := flags.Bool("dev", false, "allow loopback HTTP")
	caPath := flags.String("ca-cert", "", "additional trusted PEM CA certificate")
	conv := flags.String("conversation", "", "conversation ID for history/send")
	after := flags.Int64("after", 0, "history cursor")
	id := flags.String("id", "", "stable ID for a send retry")
	messageStdin := flags.Bool("message-stdin", false, "send: read one message from standard input")
	peer := flags.String("peer", "", "username for direct chat")
	fingerprint := flags.String("fingerprint", "", "out-of-band peer fingerprint to verify")
	legacy := flags.Bool("legacy", false, "use the old invitation-mode plaintext client")
	cryptoBin := flags.String("crypto-bin", "", "local karatty-crypto binary path")
	stateDir := flags.String("state-dir", "", "encrypted local state directory")
	confirmDelete := flags.Bool("confirm-delete", false, "delete this account, keys, sessions and own messages")
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	switch command {
	case "chat", "register", "recover", "login", "list", "history", "send", "direct", "requests", "decline", "block", "unblock", "metrics", "delete-account", "my-fingerprint", "fingerprint", "verify":
	case "version":
		fmt.Println("Karatty " + version)
		return nil
	default:
		return errors.New("commands: chat, register, recover, login, list, history, send, direct, requests, decline, block, unblock, my-fingerprint, fingerprint, verify, metrics, delete-account, version")
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
	if command == "recover" {
		if *username == "" {
			return errors.New("recovery requires --username and --key with a new SSH key")
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		challenge, e := c.StartRecovery(ctx, signer, *username)
		if e != nil {
			return e
		}
		recoveryCode, e := readSecret("Código de recuperación guardado fuera del equipo: ")
		if e != nil {
			return e
		}
		result, e := c.CompleteRecovery(ctx, signer, challenge, string(recoveryCode))
		for i := range recoveryCode {
			recoveryCode[i] = 0
		}
		if e != nil {
			return e
		}
		fmt.Printf("Clave SSH reemplazada. Guarda el nuevo código de recuperación fuera del equipo; se muestra una sola vez: %s\n", result.RecoveryCode)
		return nil
	}
	invite := ""
	name := ""
	if command == "register" {
		name = *username
		if name == "" {
			return errors.New("registration requires --username")
		}
		if *public && *invitationStdin {
			return errors.New("choose --public or --invitation-stdin")
		}
		if *public {
			// The server decides whether public registration is enabled.
		} else if *invitationStdin {
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
		if c.Token == "" {
			return
		}
		end, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		_ = c.Logout(end)
	}()
	var secure *secureclient.Backend
	if (command == "register" && *public) || (!*legacy && (command == "chat" || command == "history" || command == "send" || command == "my-fingerprint" || command == "fingerprint" || command == "verify")) {
		if command == "register" {
			fmt.Printf("Identidad creada: %s. Guarda este código de recuperación fuera del equipo; se muestra una sola vez: %s\n", session.UserID, session.RecoveryCode)
		}
		passphrase, e := readSecret("Frase secreta del archivo local cifrado (mínimo 12 caracteres): ")
		if e != nil {
			return errors.New("cannot read local passphrase; use an interactive terminal")
		}
		if command == "register" {
			confirm, confirmErr := readSecret("Repite la frase secreta local: ")
			if confirmErr != nil || !bytes.Equal(passphrase, confirm) {
				clear(passphrase)
				clear(confirm)
				return errors.New("local passphrase confirmation does not match")
			}
			clear(confirm)
		}
		root := *stateDir
		if root == "" {
			config, configErr := os.UserConfigDir()
			if configErr != nil {
				clear(passphrase)
				return configErr
			}
			root = filepath.Join(config, "karatty")
		}
		binary := *cryptoBin
		if binary == "" {
			exe, exeErr := os.Executable()
			if exeErr != nil {
				clear(passphrase)
				return exeErr
			}
			binary = filepath.Join(filepath.Dir(exe), "karatty-crypto")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
		}
		device, openErr := secureclient.Open(ctx, c, signer, session.UserID, passphrase, root, binary)
		clear(passphrase)
		if openErr != nil {
			return openErr
		}
		defer device.Close()
		secure = &secureclient.Backend{HTTP: c, Secure: device}
	}
	switch command {
	case "register":
		if !*public {
			fmt.Printf("Identidad creada: %s. Guarda este código de recuperación fuera del equipo; se muestra una sola vez: %s\n", session.UserID, session.RecoveryCode)
		}
		return nil
	case "login":
		fmt.Printf("Identidad verificada: %s. La sesión temporal se cierra al salir.\n", session.UserID)
		return nil
	case "chat":
		var backend client.Backend = c
		if secure != nil {
			backend = secure
		}
		_, err = tea.NewProgram(tui.New(ctx, backend), tea.WithContext(ctx)).Run()
		return err
	case "list":
		out, e := c.List(ctx)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	case "history":
		var backend client.Backend = c
		if secure != nil {
			backend = secure
		}
		out, e := backend.History(ctx, *conv, *after, 100)
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
	case "requests":
		out, e := c.DirectRequests(ctx)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	case "fingerprint":
		if *peer == "" || secure == nil {
			return errors.New("fingerprint requires --peer and the local encrypted client")
		}
		value, e := secure.Fingerprint(ctx, *peer)
		if e != nil {
			return e
		}
		fmt.Printf("Huella de %s: %s\nCompárala con esa persona por otro medio antes de verificar.\n", *peer, value)
		return nil
	case "my-fingerprint":
		if secure == nil {
			return errors.New("my-fingerprint requires the local encrypted client")
		}
		fmt.Println(secure.SelfFingerprint())
		return nil
	case "verify":
		if *peer == "" || *fingerprint == "" || secure == nil {
			return errors.New("verify requires --peer and --fingerprint")
		}
		if e := secure.Verify(ctx, *peer, *fingerprint); e != nil {
			return e
		}
		fmt.Println("Identidad verificada para", *peer)
		return nil
	case "decline", "block", "unblock":
		if *peer == "" {
			return errors.New("provide --peer with the other username")
		}
		switch command {
		case "decline":
			return c.DeclineDirect(ctx, *peer)
		case "block":
			return c.BlockDirect(ctx, *peer)
		default:
			return c.UnblockDirect(ctx, *peer)
		}
	case "metrics":
		out, e := c.Metrics(ctx)
		if e != nil {
			return e
		}
		_, e = fmt.Print(out)
		return e
	case "delete-account":
		if !*confirmDelete {
			return errors.New("delete-account requires --confirm-delete")
		}
		return c.DeleteAccount(ctx)
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
		var backend client.Backend = c
		if secure != nil {
			backend = secure
		}
		out, e := backend.Send(ctx, *conv, protocol.Send{ClientID: *id, Body: body})
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	}
	return nil
}

func clear(b []byte) {
	for i := range b {
		b[i] = 0
	}
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
