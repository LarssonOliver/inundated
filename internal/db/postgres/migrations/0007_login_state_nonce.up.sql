BEGIN;

-- Per-login OIDC nonce. Existing rows (in-flight logins at deploy time) get an
-- empty string; those callbacks will fail nonce verification and the user
-- simply logs in again.
ALTER TABLE login_states ADD COLUMN IF NOT EXISTS nonce TEXT NOT NULL DEFAULT '';

COMMIT;
