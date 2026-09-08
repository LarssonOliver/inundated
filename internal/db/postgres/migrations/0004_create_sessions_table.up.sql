BEGIN;

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    sub TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

/* Intentionally not creating a foreign key constraint on user_id */

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

COMMIT;
