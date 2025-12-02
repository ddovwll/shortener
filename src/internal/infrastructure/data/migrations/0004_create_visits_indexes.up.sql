CREATE INDEX IF NOT EXISTS idx_visits_link_id_created_at
    ON public.visits (link_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_visits_link_id_user_agent
    ON public.visits (link_id, user_agent);