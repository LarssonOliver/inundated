BEGIN;

DROP INDEX IF EXISTS idx_tasks_caldav_name;
DROP INDEX IF EXISTS idx_timespans_caldav_name;
ALTER TABLE tasks DROP COLUMN IF EXISTS caldav_name;
ALTER TABLE timespans DROP COLUMN IF EXISTS caldav_name;

COMMIT;
