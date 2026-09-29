BEGIN;

DROP TRIGGER IF EXISTS task_effective_tags_task_insert ON tasks;
DROP TRIGGER IF EXISTS task_effective_tags_task_update ON tasks;
DROP TRIGGER IF EXISTS task_effective_tags_task_tags_insert ON task_tags;
DROP TRIGGER IF EXISTS task_effective_tags_task_tags_delete ON task_tags;
DROP TRIGGER IF EXISTS task_effective_tags_task_tags_update ON task_tags;
DROP TRIGGER IF EXISTS task_effective_tags_tag_update ON tags;

DROP FUNCTION IF EXISTS task_effective_tags_on_task_insert();
DROP FUNCTION IF EXISTS task_effective_tags_on_task_update();
DROP FUNCTION IF EXISTS task_effective_tags_on_task_tags_insert();
DROP FUNCTION IF EXISTS task_effective_tags_on_task_tags_delete();
DROP FUNCTION IF EXISTS task_effective_tags_on_task_tags_update();
DROP FUNCTION IF EXISTS task_effective_tags_on_tag_update();
DROP FUNCTION IF EXISTS refresh_task_effective_tags(UUID[]);

COMMIT;
