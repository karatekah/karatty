package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/karatekah/termlink/internal/auth"
	"github.com/karatekah/termlink/internal/chat"
	"github.com/karatekah/termlink/internal/security"
	"github.com/karatekah/termlink/internal/storage"
	"github.com/karatekah/termlink/internal/transport"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server operation failed", "reason", safe(err))
		os.Exit(1)
	}
}
func safe(err error) string {
	if err == security.ErrInvalid {
		return err.Error()
	}
	return "check configuration, database connectivity and arguments"
}
func run() error {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	dev := flag.Bool("dev", false, "allow HTTP on loopback")
	cert := flag.String("tls-cert", "", "TLS certificate")
	key := flag.String("tls-key", "", "TLS private key")
	hostKey := flag.String("ssh-host-key", "", "optional SSH host key (combined process)")
	sshAddr := flag.String("ssh-addr", "127.0.0.1:2222", "combined SSH address")
	channel := flag.String("channel", "", "admin: create channel")
	invite := flag.Bool("invite", false, "admin: create invitation")
	conversation := flag.String("conversation", "", "admin: conversation ID")
	uses := flag.Int("uses", 1, "invitation uses")
	ttl := flag.Duration("ttl", 24*time.Hour, "invitation validity")
	revokeKey := flag.String("revoke-key", "", "admin: key fingerprint")
	revokeMember := flag.String("revoke-member", "", "admin: user ID")
	retention30d := flag.Bool("retention-30d", false, "purge messages older than 30 days at startup and hourly")
	publicSignup := flag.Bool("public-signup", false, "allow pseudonymous SSH-key registration without email")
	publicNetwork := flag.Bool("public-network", false, "allow the E2EE alpha API beyond loopback over TLS (LAN or Internet)")
	purgeExpired := flag.Bool("purge-expired-messages", false, "admin: purge messages older than 30 days once")
	flag.Parse()
	if err := validatePublicMode(*addr, *cert, *key, *hostKey, *publicSignup, *publicNetwork, *dev); err != nil {
		return err
	}
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
	a := &auth.Service{DB: db, PublicSignup: *publicSignup}
	c := chat.New(db)
	c.DirectOnly = *publicSignup
	c.CiphertextOnly = *publicSignup
	switch {
	case *channel != "":
		if *publicSignup {
			return security.ErrInvalid
		}
		id, e := c.Channel(ctx, *channel)
		if e == nil {
			fmt.Println(id)
		}
		return e
	case *invite:
		token, e := a.Invite(ctx, *conversation, *uses, *ttl)
		if e == nil {
			fmt.Println(token)
		}
		return e
	case *revokeKey != "":
		return a.RevokeKey(ctx, *revokeKey)
	case *revokeMember != "":
		return c.RevokeMember(ctx, *conversation, *revokeMember)
	case *purgeExpired:
		count, e := storage.PurgeExpiredMessages(ctx, db)
		if e == nil {
			fmt.Println(count)
		}
		return e
	}
	if err = transport.ValidateHTTP(*addr, *cert, *key, *dev); err != nil {
		return err
	}
	api := transport.NewAPI(a, c)
	defer api.Close()
	server := &http.Server{Addr: *addr, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: time.Minute, MaxHeaderBytes: 8192}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	listener = security.BoundListener(listener, 128)
	defer listener.Close()
	result := make(chan error, 2)
	if *hostKey != "" {
		gateway, e := transport.NewSSH(a, c, *hostKey)
		if e != nil {
			return e
		}
		sshListener, e := net.Listen("tcp", *sshAddr)
		if e != nil {
			return e
		}
		defer gateway.Close()
		go func() { result <- gateway.Serve(security.BoundListener(sshListener, 64)) }()
	}
	go func() {
		if *cert != "" {
			result <- server.ServeTLS(listener, *cert, *key)
		} else {
			result <- server.Serve(listener)
		}
	}()
	if *retention30d || *publicSignup {
		go runRetention(ctx, db)
	}
	slog.Info("karatty started", "address", *addr, "ssh_enabled", *hostKey != "")
	select {
	case <-ctx.Done():
		shutdown, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		api.Close()
		return server.Shutdown(shutdown)
	case err = <-result:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func runRetention(ctx context.Context, db *pgxpool.Pool) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		count, err := storage.PurgeExpiredMessages(ctx, db)
		if err != nil && ctx.Err() == nil {
			slog.Error("message retention purge failed")
		} else if err == nil {
			slog.Info("message retention purge finished", "deleted", count)
		}
		requests, requestErr := storage.PurgeExpiredDirectRequests(ctx, db)
		if requestErr != nil && ctx.Err() == nil {
			slog.Error("direct request purge failed")
		} else if requestErr == nil {
			slog.Info("direct request purge finished", "deleted", requests)
		}
		challenges, challengeErr := storage.PurgeExpiredRecoveryChallenges(ctx, db)
		if challengeErr != nil && ctx.Err() == nil {
			slog.Error("recovery challenge purge failed")
		} else if challengeErr == nil {
			slog.Info("recovery challenge purge finished", "deleted", challenges)
		}
		authRecords, authErr := storage.PurgeExpiredAuthRecords(ctx, db)
		if authErr != nil && ctx.Err() == nil {
			slog.Error("auth record purge failed")
		} else if authErr == nil {
			slog.Info("auth record purge finished", "deleted", authRecords)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
