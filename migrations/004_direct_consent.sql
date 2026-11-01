CREATE TABLE IF NOT EXISTS direct_requests (
 id text PRIMARY KEY,
 requester_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 recipient_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at timestamptz NOT NULL DEFAULT now(),
 expires_at timestamptz NOT NULL DEFAULT now() + interval '7 days',
 UNIQUE(requester_id,recipient_id), CHECK(requester_id<>recipient_id)
);
CREATE INDEX IF NOT EXISTS direct_requests_recipient ON direct_requests(recipient_id,created_at);
CREATE INDEX IF NOT EXISTS direct_requests_expiry ON direct_requests(expires_at);
CREATE TABLE IF NOT EXISTS direct_blocks (
 blocker_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 blocked_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(blocker_id,blocked_id), CHECK(blocker_id<>blocked_id)
);
