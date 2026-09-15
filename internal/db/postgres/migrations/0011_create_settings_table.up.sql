BEGIN;

CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY,
    user_id UUID DEFAULT NULL,
    week_start_day TEXT NOT NULL DEFAULT 'monday',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    duration_format TEXT NOT NULL DEFAULT 'long',
    time_format TEXT NOT NULL DEFAULT '24h'
);

DO $$ BEGIN
	ALTER TABLE settings ADD CONSTRAINT settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- One settings row per real user.
CREATE UNIQUE INDEX IF NOT EXISTS idx_settings_user_id ON settings(user_id) WHERE user_id IS NOT NULL;
-- At most one unowned (userless-mode) settings row: the indexed expression is
-- constant `true` for every row matching the predicate, so a second one
-- collides with the first.
CREATE UNIQUE INDEX IF NOT EXISTS idx_settings_unowned_singleton ON settings((user_id IS NULL)) WHERE user_id IS NULL;

COMMIT;
