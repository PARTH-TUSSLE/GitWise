-- 000001_init_schema.up.sql
-- GitWise Phase 1 Schema: Repositories, Snapshots, Analysis Jobs + Vector Extension

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- 1. Repositories
CREATE TABLE IF NOT EXISTS repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id BIGINT UNIQUE NOT NULL,
    owner VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    default_branch VARCHAR(255) NOT NULL DEFAULT 'main',
    is_private BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_repos_owner_name ON repositories(owner, name);

-- 2. Snapshots (Commit-Addressed)
CREATE TABLE IF NOT EXISTS repository_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    commit_sha CHAR(40) NOT NULL,
    ref_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED', -- QUEUED, PROCESSING, READY, PARTIAL, FAILED
    total_files INT NOT NULL DEFAULT 0,
    total_lines INT NOT NULL DEFAULT 0,
    primary_language VARCHAR(100),
    analyzed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_snapshots_repo_commit ON repository_snapshots(repository_id, commit_sha);

-- 3. Background Analysis Jobs
CREATE TABLE IF NOT EXISTS analysis_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    snapshot_id UUID REFERENCES repository_snapshots(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED', -- QUEUED, PROCESSING, COMPLETED, FAILED
    stage VARCHAR(100) NOT NULL DEFAULT 'INITIALIZING',
    progress_percent NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    error_message TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON analysis_jobs(status);
