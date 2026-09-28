BEGIN;

-- Closure table of every live task's effective tags: its own task tag, its
-- regular tags, and the effective tags of its ancestors. A timespan counts
-- toward a tag when it carries the tag itself, or carries the task tag of a
-- task whose effective tags include it. Rows are rewritten on task writes,
-- in the same transaction, and removed when a task is deleted.
CREATE TABLE IF NOT EXISTS task_effective_tags (
    task_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    PRIMARY KEY (task_id, tag_id),
    CONSTRAINT task_effective_tags_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CONSTRAINT task_effective_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_effective_tags_tag_id ON task_effective_tags(tag_id);

-- Backfill from the tasks that already exist.
WITH RECURSIVE ancestry(task_id, ancestor_id) AS (
    SELECT id, id FROM tasks WHERE deleted_at IS NULL
    UNION
    SELECT a.task_id, p.parent_id
    FROM ancestry a
    JOIN tasks p ON p.id = a.ancestor_id
    WHERE p.parent_id IS NOT NULL
)
INSERT INTO task_effective_tags (task_id, tag_id)
SELECT a.task_id, k.tag_id
FROM ancestry a
JOIN tasks k ON k.id = a.ancestor_id
UNION
SELECT a.task_id, tt.tag_id
FROM ancestry a
JOIN task_tags tt ON tt.task_id = a.ancestor_id
ON CONFLICT DO NOTHING;

COMMIT;
