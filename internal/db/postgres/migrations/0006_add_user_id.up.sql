BEGIN;

ALTER TABLE tags ADD COLUMN IF NOT EXISTS user_id UUID DEFAULT NULL;
DO $$ BEGIN
	ALTER TABLE tags ADD CONSTRAINT tags_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

ALTER TABLE projects ADD COLUMN IF NOT EXISTS user_id UUID DEFAULT NULL;
DO $$ BEGIN
	ALTER TABLE projects ADD CONSTRAINT projects_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);

ALTER TABLE timespans ADD COLUMN IF NOT EXISTS user_id UUID DEFAULT NULL;
DO $$ BEGIN
	ALTER TABLE timespans ADD CONSTRAINT timespans_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
CREATE INDEX IF NOT EXISTS idx_timespans_user_id ON timespans(user_id);

COMMIT;
