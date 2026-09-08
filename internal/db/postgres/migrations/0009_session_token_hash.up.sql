BEGIN;

-- The session cookie now carries an opaque token whose SHA-256 is stored here;
-- the primary key stays an internal identifier that is never sent to clients.
-- Existing rows get a NULL hash and can no longer be resolved, so any active
-- session must re-authenticate - the correct posture when changing the token
-- scheme.
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS token_hash BYTEA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);

COMMIT;
