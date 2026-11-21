BEGIN;

ALTER TABLE timespans ADD COLUMN IF NOT EXISTS caldav_name TEXT DEFAULT NULL;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS caldav_name TEXT DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_timespans_caldav_name ON timespans (user_id, caldav_name) NULLS NOT DISTINCT
    WHERE caldav_name IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_caldav_name ON tasks (user_id, caldav_name) NULLS NOT DISTINCT
    WHERE caldav_name IS NOT NULL AND deleted_at IS NULL;

COMMIT;
