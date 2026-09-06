BEGIN;

-- Ownership of a resource. NULL means the resource predates user support
-- (or was created while running in userless mode); the first user to log in
-- adopts every such row. See CreateUserAdoptingOrphans in the postgres repo.

ALTER TABLE tags ADD COLUMN IF NOT EXISTS user_id UUID DEFAULT NULL;
ALTER TABLE tags ADD CONSTRAINT tags_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

ALTER TABLE projects ADD COLUMN IF NOT EXISTS user_id UUID DEFAULT NULL;
ALTER TABLE projects ADD CONSTRAINT projects_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);

ALTER TABLE timespans ADD COLUMN IF NOT EXISTS user_id UUID DEFAULT NULL;
ALTER TABLE timespans ADD CONSTRAINT timespans_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_timespans_user_id ON timespans(user_id);

COMMIT;
