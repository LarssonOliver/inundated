BEGIN;

-- Task tags are removed with their tasks, along with any associations
-- other rows still hold to them.
CREATE TEMPORARY TABLE task_tag_ids ON COMMIT DROP AS SELECT tag_id FROM tasks;

DROP TABLE IF EXISTS task_tags;
DROP TABLE IF EXISTS tasks;

DELETE FROM tags WHERE id IN (SELECT tag_id FROM task_tag_ids);

COMMIT;
