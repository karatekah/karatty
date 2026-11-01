CREATE TABLE e2ee_devices (
 user_id text PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 ssh_key_id text NOT NULL REFERENCES ssh_keys(id),
 identity_ed text NOT NULL,
 identity_curve text NOT NULL,
 ssh_signature bytea NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE e2ee_one_time_keys (
 user_id text NOT NULL REFERENCES e2ee_devices(user_id) ON DELETE CASCADE,
 public_key text NOT NULL,
 signature text NOT NULL,
 claimed_at timestamptz,
 PRIMARY KEY(user_id,public_key)
);
