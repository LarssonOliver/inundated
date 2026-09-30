BEGIN;

-- Project tags. Each project owns a project tag (tag_id), shown as "@name":
-- timespans and tasks carry it to count toward the project. Like a task
-- tag, its name and archived state follow the project; its color is kept
-- equal to the project's on every project write.
ALTER TABLE projects ADD COLUMN IF NOT EXISTS tag_id UUID;

-- Give every existing project, deleted ones included, a tag of its own.
UPDATE projects SET tag_id = gen_random_uuid() WHERE tag_id IS NULL;
INSERT INTO tags (id, name, color, user_id, deleted_at)
SELECT tag_id, name, color, user_id, deleted_at FROM projects
ON CONFLICT (id) DO NOTHING;

ALTER TABLE projects ALTER COLUMN tag_id SET NOT NULL;
ALTER TABLE projects ADD CONSTRAINT projects_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(id);
ALTER TABLE projects ADD CONSTRAINT projects_tag_id_key UNIQUE (tag_id);

-- Projects join the owners of tags (see migration 0017).
CREATE OR REPLACE VIEW tag_owners AS
    SELECT tag_id, 'task'::text AS kind, id AS owner_id, closed_at AS archived_at
    FROM tasks
    WHERE deleted_at IS NULL
    UNION ALL
    SELECT tag_id, 'project'::text, id, archived_at
    FROM projects
    WHERE deleted_at IS NULL;

-- Every tag that counts toward a project: the ones linked to it, and its
-- own project tag. A project's own tag is never stored in project_tags; it
-- counts by this rule, the way task_effective_tags always includes a
-- task's own task tag.
CREATE OR REPLACE VIEW project_effective_tags AS
    SELECT project_id, tag_id FROM project_tags
    UNION ALL
    SELECT id, tag_id FROM projects;

COMMIT;
