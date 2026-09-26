-- 000002_github_gitstat.up.sql
-- GitWise Phase 2 Schema: GitHub Users & Calculated Metrics Cache

CREATE TABLE IF NOT EXISTS github_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id BIGINT UNIQUE,
    username VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    bio TEXT,
    company VARCHAR(255),
    location VARCHAR(255),
    blog TEXT,
    twitter_username VARCHAR(255),
    public_repos INT NOT NULL DEFAULT 0,
    public_gists INT NOT NULL DEFAULT 0,
    followers INT NOT NULL DEFAULT 0,
    following INT NOT NULL DEFAULT 0,
    github_created_at TIMESTAMPTZ,
    profile_data JSONB NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_github_users_username_lower ON github_users(LOWER(username));

CREATE TABLE IF NOT EXISTS github_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES github_users(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    metric_key VARCHAR(100) NOT NULL,
    metric_value NUMERIC NOT NULL DEFAULT 0,
    provenance VARCHAR(255) NOT NULL,
    sample_size INT NOT NULL DEFAULT 0,
    coverage_limit TEXT NOT NULL,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_github_metrics_user ON github_metrics(user_id);
CREATE INDEX IF NOT EXISTS idx_github_metrics_username_key ON github_metrics(username, metric_key);
