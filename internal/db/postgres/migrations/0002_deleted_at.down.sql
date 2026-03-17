BEGIN;

DROP INDEX IF EXISTS idx_tags_active;
ALTER TABLE tags DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_projects_active;
ALTER TABLE projects DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_timespans_active;
ALTER TABLE timespans DROP COLUMN IF EXISTS deleted_at;

COMMIT;
