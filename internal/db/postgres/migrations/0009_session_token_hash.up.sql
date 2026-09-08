BEGIN;

ALTER TABLE sessions ADD COLUMN IF NOT EXISTS token_hash BYTEA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);

COMMIT;
