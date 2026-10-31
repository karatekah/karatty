CREATE TABLE IF NOT EXISTS users (
 id text PRIMARY KEY, username text UNIQUE NOT NULL, active boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS ssh_keys (
 id text PRIMARY KEY, user_id text NOT NULL REFERENCES users(id), public_key text UNIQUE NOT NULL,
 fingerprint text UNIQUE NOT NULL, revoked_at timestamptz, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS sessions (
 token_hash text PRIMARY KEY, user_id text NOT NULL REFERENCES users(id), key_id text NOT NULL REFERENCES ssh_keys(id),
 expires_at timestamptz NOT NULL, revoked_at timestamptz
);
CREATE INDEX IF NOT EXISTS sessions_key ON sessions(key_id);
CREATE TABLE IF NOT EXISTS conversations (
 id text PRIMARY KEY, name text NOT NULL, kind text NOT NULL CHECK (kind IN ('channel','direct')),
 next_seq bigint NOT NULL DEFAULT 0, direct_pair text UNIQUE
);
CREATE TABLE IF NOT EXISTS memberships (
 conversation_id text NOT NULL REFERENCES conversations(id), user_id text NOT NULL REFERENCES users(id),
 PRIMARY KEY(conversation_id,user_id)
);
CREATE TABLE IF NOT EXISTS invitations (
 token_hash text PRIMARY KEY, expires_at timestamptz NOT NULL, remaining integer NOT NULL CHECK(remaining>=0),
 conversation_id text REFERENCES conversations(id)
);
CREATE TABLE IF NOT EXISTS auth_challenges (
 id text PRIMARY KEY, payload text NOT NULL, public_key text NOT NULL, username text NOT NULL,
 invitation_hash text NOT NULL, expires_at timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS messages (
 id text PRIMARY KEY, client_id text NOT NULL, conversation_id text NOT NULL REFERENCES conversations(id),
 sender_id text NOT NULL REFERENCES users(id), seq bigint NOT NULL CHECK(seq>0), body text NOT NULL CHECK(octet_length(body) BETWEEN 1 AND 4096),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(conversation_id,seq), UNIQUE(conversation_id,sender_id,client_id)
);
