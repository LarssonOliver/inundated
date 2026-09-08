BEGIN;

DROP INDEX IF EXISTS idx_tags_user_id;
ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_user_id_fkey;
ALTER TABLE tags DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS idx_projects_user_id;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_user_id_fkey;
ALTER TABLE projects DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS idx_timespans_user_id;
ALTER TABLE timespans DROP CONSTRAINT IF EXISTS timespans_user_id_fkey;
ALTER TABLE timespans DROP COLUMN IF EXISTS user_id;

COMMIT;
