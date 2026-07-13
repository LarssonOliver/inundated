BEGIN;

CREATE TABLE IF NOT EXISTS login_states (
    id UUID PRIMARY KEY,
    redirect_uri TEXT NOT NULL,
    code_verifier TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_login_states_expires_at ON public.login_states(expires_at);

COMMIT;
