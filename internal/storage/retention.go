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
