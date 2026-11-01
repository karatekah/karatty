ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_body_check;
ALTER TABLE messages ADD CONSTRAINT messages_body_check CHECK(octet_length(body) BETWEEN 1 AND 8192);
