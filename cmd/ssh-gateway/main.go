package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/karatekah/termlink/internal/auth"
	"github.com/karatekah/termlink/internal/chat"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/internal/storage"
	"github.com/karatekah/termlink/internal/transport"
)

func main() {
	if err := run(); err != nil {
		slog.Error("SSH gateway failed; check configuration and database connectivity")
		os.Exit(1)
	}
}
func run() error {
	addr := flag.String("addr", "127.0.0.1:2222", "listen address")
	key := flag.String("host-key", ".local/ssh_host_ed25519", "persistent host private key")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	db, err := storage.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	if err = storage.Migrate(ctx, db); err != nil {
		return err
	}
	server, err := transport.NewSSH(&auth.Service{DB: db}, chat.New(db), *key)
	if err != nil {
		return err
	}
	defer server.Close()
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	result := make(chan error, 1)
	go func() { result <- server.Serve(security.BoundListener(l, 64)) }()
	slog.Info("SSH gateway started", "address", *addr)
	select {
	case <-ctx.Done():
		return nil
	case err = <-result:
		return err
	}
}
