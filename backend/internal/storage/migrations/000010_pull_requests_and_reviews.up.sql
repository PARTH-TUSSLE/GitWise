-- Migration: 000010_pull_requests_and_reviews.up.sql
-- Description: Creates pull_requests and pr_reviews tables for Phase 9 PR Intelligence & Semantic Diff Analysis.

CREATE TABLE IF NOT EXISTS pull_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    number INT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT 'ghost',
    status TEXT NOT NULL DEFAULT 'ready_for_review',
    base_ref TEXT NOT NULL DEFAULT 'main',
    head_ref TEXT NOT NULL DEFAULT '',
    head_sha TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_repository_pr_number UNIQUE (repository_id, number)
);

CREATE INDEX IF NOT EXISTS idx_pull_requests_repo_number ON pull_requests(repository_id, number);
CREATE INDEX IF NOT EXISTS idx_pull_requests_repo_status ON pull_requests(repository_id, status);

CREATE TABLE IF NOT EXISTS pr_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pull_request_id UUID NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    snapshot_id UUID REFERENCES repository_snapshots(id) ON DELETE SET NULL,
    stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    summary TEXT NOT NULL DEFAULT '',
    architectural_shift JSONB NOT NULL DEFAULT '{}'::jsonb,
    files JSONB NOT NULL DEFAULT '[]'::jsonb,
    review_findings JSONB NOT NULL DEFAULT '[]'::jsonb,
    test_verification JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_pr_review UNIQUE (pull_request_id)
);

CREATE INDEX IF NOT EXISTS idx_pr_reviews_pr_id ON pr_reviews(pull_request_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_snapshot_id ON pr_reviews(snapshot_id);
