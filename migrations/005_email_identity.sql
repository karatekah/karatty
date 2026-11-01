ALTER TABLE users ADD COLUMN IF NOT EXISTS email text UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at timestamptz;
ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_hash text;

CREATE TABLE IF NOT EXISTS email_challenges (
 id text PRIMARY KEY,
 purpose text NOT NULL CHECK(purpose IN ('signup','recovery')),
 username text NOT NULL DEFAULT '',
 email text NOT NULL,
 public_key text NOT NULL,
 payload text NOT NULL,
 code_hash text NOT NULL,
 user_id text REFERENCES users(id) ON DELETE CASCADE,
 expires_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(purpose,email)
);
CREATE INDEX IF NOT EXISTS email_challenges_expiry ON email_challenges(expires_at);

CREATE TABLE IF NOT EXISTS email_daily (
 day date PRIMARY KEY,
 sent integer NOT NULL CHECK(sent>=0)
);
