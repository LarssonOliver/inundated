BEGIN;

-- Owned tags. Some items own a tag of their own (a task owns its task tag):
-- the tag's name and archived state follow its owner, and it can't be
-- edited or deleted through the tag API. Each kind of owner keeps its tag
-- in a tag_id column of its own table, with a real foreign key; this view
-- gathers them so tag reads join one place, whatever the owner. A new kind
-- of owner adds a UNION ALL arm here. Deleted owners are left out.
CREATE OR REPLACE VIEW tag_owners AS
    SELECT tag_id, 'task'::text AS kind, id AS owner_id, closed_at AS archived_at
    FROM tasks
    WHERE deleted_at IS NULL;

COMMIT;
