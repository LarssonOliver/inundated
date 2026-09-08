BEGIN;

-- Session creation time, so sliding renewal can be bounded by an absolute cap.
-- Existing rows are backfilled to now(); their absolute expiry is measured from
-- the deploy rather than their true start, which is acceptable for a one-off.
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now();

COMMIT;
