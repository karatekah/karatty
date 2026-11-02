package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const retentionBatch = 1000

// PurgeExpiredMessages removes message content and identifiers after 720 hours.
// Each batch commits separately so a large backlog does not hold one long lock.
func PurgeExpiredMessages(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	var total int64
	for {
		result, err := db.Exec(ctx, `WITH expired AS (
			SELECT id FROM messages
			WHERE created_at < now() - interval '720 hours'
			ORDER BY created_at, id LIMIT $1 FOR UPDATE
		)
		DELETE FROM messages m USING expired WHERE m.id = expired.id`, retentionBatch)
		if err != nil {
			return total, err
		}
		count := result.RowsAffected()
		total += count
		if count < retentionBatch {
			return total, nil
		}
	}
}

func PurgeExpiredDirectRequests(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	result, err := db.Exec(ctx, "DELETE FROM direct_requests WHERE expires_at<=now()")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func PurgeExpiredRecoveryChallenges(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	result, err := db.Exec(ctx, "DELETE FROM recovery_challenges WHERE expires_at<=now()")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func PurgeExpiredAuthRecords(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var total int64
	for _, query := range []string{
		"DELETE FROM sessions WHERE expires_at<=now() OR revoked_at IS NOT NULL",
		"DELETE FROM auth_challenges WHERE expires_at<=now()",
		"DELETE FROM invitations WHERE expires_at<=now() OR remaining=0",
	} {
		result, e := tx.Exec(ctx, query)
		if e != nil {
			return total, e
		}
		total += result.RowsAffected()
	}
	return total, tx.Commit(ctx)
}
