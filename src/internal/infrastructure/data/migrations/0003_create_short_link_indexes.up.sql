CREATE UNIQUE INDEX IF NOT EXISTS idx_short_links_short_code
    ON public.short_links (short_code);