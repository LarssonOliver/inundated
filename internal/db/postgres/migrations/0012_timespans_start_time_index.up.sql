BEGIN;

CREATE INDEX IF NOT EXISTS idx_timespans_user_id_start_time
	ON public.timespans (user_id, start_time);

COMMIT;
