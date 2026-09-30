BEGIN;

DROP VIEW IF EXISTS project_effective_tags;

CREATE OR REPLACE VIEW tag_owners AS
    SELECT tag_id, 'task'::text AS kind, id AS owner_id, closed_at AS archived_at
    FROM tasks
    WHERE deleted_at IS NULL;

-- Project tags go with their projects' column. Their links on timespans,
-- tasks and task_effective_tags cascade.
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_tag_id_key;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_tag_id_fkey;
DELETE FROM tags t USING projects p WHERE t.id = p.tag_id;
ALTER TABLE projects DROP COLUMN IF EXISTS tag_id;

COMMIT;
