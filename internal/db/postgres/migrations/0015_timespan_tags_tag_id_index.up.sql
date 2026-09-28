BEGIN;

-- Time totals look up timespans by tag (directly, or through a task's
-- task tag), which the (timespan_id, tag_id) primary key can't serve.
CREATE INDEX IF NOT EXISTS idx_timespan_tags_tag_id ON timespan_tags(tag_id);

COMMIT;
