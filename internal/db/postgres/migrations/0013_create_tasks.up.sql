BEGIN;

-- Tasks. Each task owns a task tag (tag_id): timespans and projects carry
-- that tag to log time on the task or to be assigned it. The task tag's
-- name follows the task, and its archived state and color are derived from
-- the task when read, so neither is kept on the tag row.
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    user_id UUID DEFAULT NULL,
    tag_id UUID NOT NULL,
    parent_id UUID DEFAULT NULL,
    name TEXT NOT NULL,
    due_date DATE DEFAULT NULL,
    estimate INTERVAL DEFAULT NULL,
    rank TEXT COLLATE "C" NOT NULL,
    closed_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    close_reason TEXT DEFAULT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    CONSTRAINT tasks_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT tasks_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(id),
    CONSTRAINT tasks_tag_id_key UNIQUE (tag_id),
    CONSTRAINT tasks_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES tasks(id),
    CONSTRAINT tasks_close_reason_check CHECK (close_reason IN ('done', 'ignored')),
    CONSTRAINT tasks_closed_check CHECK ((closed_at IS NULL) = (close_reason IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_tasks_siblings ON tasks(user_id, parent_id, rank) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(user_id, due_date) WHERE deleted_at IS NULL;

-- Task-Tag join table. Holds a task's regular tags only, never task tags.
CREATE TABLE IF NOT EXISTS task_tags (
    task_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    PRIMARY KEY (task_id, tag_id),
    CONSTRAINT task_tags_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    CONSTRAINT task_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_tags_tag_id ON task_tags(tag_id);

COMMIT;
