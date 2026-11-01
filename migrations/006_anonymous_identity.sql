DROP TABLE IF EXISTS email_challenges;
DROP TABLE IF EXISTS email_daily;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email;

CREATE TABLE IF NOT EXISTS recovery_challenges (
 id text PRIMARY KEY,
 user_id text NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
 public_key text NOT NULL,
 payload text NOT NULL,
 expires_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS recovery_challenges_expiry ON recovery_challenges(expires_at);
