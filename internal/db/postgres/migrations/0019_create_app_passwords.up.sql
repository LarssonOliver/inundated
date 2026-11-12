BEGIN;

CREATE TABLE IF NOT EXISTS app_passwords (
    id UUID PRIMARY KEY,
    user_id UUID DEFAULT NULL,
    name TEXT NOT NULL,
    token_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ DEFAULT NULL
);

DO $$ BEGIN
	ALTER TABLE app_passwords ADD CONSTRAINT app_passwords_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_app_passwords_token_hash ON app_passwords(token_hash);
CREATE INDEX IF NOT EXISTS idx_app_passwords_user_id ON app_passwords(user_id);

COMMIT;
