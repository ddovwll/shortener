CREATE TABLE IF NOT EXISTS visits
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    link_id    UUID        NOT NULL REFERENCES short_links (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET
);