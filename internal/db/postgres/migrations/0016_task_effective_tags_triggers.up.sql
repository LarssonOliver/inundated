BEGIN;

-- Keep task_effective_tags up to date in the database instead of in each
-- write path: any change to a task's place in the tree, its regular tags,
-- or a tag's deleted state rewrites the affected rows in the same
-- statement, whichever code (or hand-written SQL) made it. Rows never
-- name a deleted tag, so readers need no live-tag join of their own.

-- refresh_task_effective_tags rewrites the rows of every task in roots and
-- of every live task below them. Deleted tasks keep no rows.
CREATE OR REPLACE FUNCTION refresh_task_effective_tags(roots UUID[]) RETURNS void
LANGUAGE plpgsql AS $$
DECLARE
    subtree UUID[];
BEGIN
    WITH RECURSIVE sub AS (
        SELECT id FROM tasks WHERE id = ANY(roots)
        UNION
        SELECT c.id FROM tasks c JOIN sub ON c.parent_id = sub.id
        WHERE c.deleted_at IS NULL
    )
    SELECT array_agg(id) INTO subtree FROM sub;
    IF subtree IS NULL THEN
        RETURN;
    END IF;

    DELETE FROM task_effective_tags WHERE task_id = ANY(subtree);

    -- Share-lock the regular tags about to be written, so a concurrent tag
    -- delete either waits for this transaction (and then removes its rows)
    -- or has already committed (and the insert below skips the tag). Locking
    -- in id order matches tagsInScope and DeleteTask.
    PERFORM 1 FROM tags t
    WHERE t.deleted_at IS NULL AND t.id IN (
        WITH RECURSIVE ancestry(task_id, ancestor_id) AS (
            SELECT id, id FROM tasks WHERE id = ANY(subtree) AND deleted_at IS NULL
            UNION
            SELECT a.task_id, p.parent_id
            FROM ancestry a
            JOIN tasks p ON p.id = a.ancestor_id
            WHERE p.parent_id IS NOT NULL
        )
        SELECT tt.tag_id FROM ancestry a JOIN task_tags tt ON tt.task_id = a.ancestor_id
    )
    ORDER BY t.id
    FOR SHARE;

    WITH RECURSIVE ancestry(task_id, ancestor_id) AS (
        SELECT id, id FROM tasks WHERE id = ANY(subtree) AND deleted_at IS NULL
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
    JOIN tags lt ON lt.id = tt.tag_id AND lt.deleted_at IS NULL;
END;
$$;

-- New tasks.
CREATE OR REPLACE FUNCTION task_effective_tags_on_task_insert() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM refresh_task_effective_tags(ARRAY(SELECT id FROM new_rows));
    RETURN NULL;
END;
$$;

-- Moved, deleted or restored tasks.
CREATE OR REPLACE FUNCTION task_effective_tags_on_task_update() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM refresh_task_effective_tags(ARRAY(
        SELECT n.id FROM new_rows n JOIN old_rows o ON o.id = n.id
        WHERE n.parent_id IS DISTINCT FROM o.parent_id
            OR n.tag_id IS DISTINCT FROM o.tag_id
            OR n.deleted_at IS DISTINCT FROM o.deleted_at
    ));
    RETURN NULL;
END;
$$;

-- Tasks whose regular tags changed, from whichever transition tables the
-- trigger has.
CREATE OR REPLACE FUNCTION task_effective_tags_on_task_tags_insert() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM refresh_task_effective_tags(ARRAY(SELECT DISTINCT task_id FROM new_rows));
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION task_effective_tags_on_task_tags_delete() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM refresh_task_effective_tags(ARRAY(SELECT DISTINCT task_id FROM old_rows));
    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION task_effective_tags_on_task_tags_update() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM refresh_task_effective_tags(ARRAY(
        SELECT task_id FROM new_rows UNION SELECT task_id FROM old_rows
    ));
    RETURN NULL;
END;
$$;

-- Deleted tags leave every task; restored ones come back to the tasks that
-- still link them.
CREATE OR REPLACE FUNCTION task_effective_tags_on_tag_update() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM task_effective_tags
    WHERE tag_id IN (
        SELECT n.id FROM new_rows n JOIN old_rows o ON o.id = n.id
        WHERE o.deleted_at IS NULL AND n.deleted_at IS NOT NULL
    );
    PERFORM refresh_task_effective_tags(ARRAY(
        SELECT DISTINCT tt.task_id
        FROM new_rows n
        JOIN old_rows o ON o.id = n.id
        JOIN task_tags tt ON tt.tag_id = n.id
        WHERE o.deleted_at IS NOT NULL AND n.deleted_at IS NULL
    ));
    RETURN NULL;
END;
$$;

-- Transition tables allow one event per trigger.
CREATE TRIGGER task_effective_tags_task_insert
    AFTER INSERT ON tasks REFERENCING NEW TABLE AS new_rows
    FOR EACH STATEMENT EXECUTE FUNCTION task_effective_tags_on_task_insert();
CREATE TRIGGER task_effective_tags_task_update
    AFTER UPDATE ON tasks REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
    FOR EACH STATEMENT EXECUTE FUNCTION task_effective_tags_on_task_update();
CREATE TRIGGER task_effective_tags_task_tags_insert
    AFTER INSERT ON task_tags REFERENCING NEW TABLE AS new_rows
    FOR EACH STATEMENT EXECUTE FUNCTION task_effective_tags_on_task_tags_insert();
CREATE TRIGGER task_effective_tags_task_tags_delete
    AFTER DELETE ON task_tags REFERENCING OLD TABLE AS old_rows
    FOR EACH STATEMENT EXECUTE FUNCTION task_effective_tags_on_task_tags_delete();
CREATE TRIGGER task_effective_tags_task_tags_update
    AFTER UPDATE ON task_tags REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
    FOR EACH STATEMENT EXECUTE FUNCTION task_effective_tags_on_task_tags_update();
CREATE TRIGGER task_effective_tags_tag_update
    AFTER UPDATE ON tags REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
    FOR EACH STATEMENT EXECUTE FUNCTION task_effective_tags_on_tag_update();

-- Drop the rows migration 0014's backfill and earlier tag deletes left for
-- deleted tags and deleted tasks.
DELETE FROM task_effective_tags te
USING tags t
WHERE t.id = te.tag_id AND t.deleted_at IS NOT NULL;
DELETE FROM task_effective_tags te
USING tasks k
WHERE k.id = te.task_id AND k.deleted_at IS NOT NULL;

COMMIT;
