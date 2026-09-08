BEGIN;

DROP INDEX IF EXISTS idx_sessions_token_hash;
ALTER TABLE sessions DROP COLUMN IF EXISTS token_hash;

COMMIT;
