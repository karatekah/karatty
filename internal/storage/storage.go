package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/karatekah/termlink/migrations"
)

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 16
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// Migrate serializes schema installation, including separate gateway/API starts.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(872092631)"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.Initial); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.Retention); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
