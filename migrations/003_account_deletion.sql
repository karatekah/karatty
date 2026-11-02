ALTER TABLE conversations ADD COLUMN IF NOT EXISTS closed_at timestamptz;
CREATE INDEX IF NOT EXISTS messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS memberships_user_id ON memberships(user_id);
CREATE INDEX IF NOT EXISTS sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS ssh_keys_user_id ON ssh_keys(user_id);
