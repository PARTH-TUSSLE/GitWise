-- Migration: 000009_issue_blueprints.up.sql
-- Description: Creates issues and issue_blueprints tables for Phase 8 Issue Intelligence and Implementation Blueprints.

CREATE TABLE IF NOT EXISTS issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    number INT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT 'ghost',
    state TEXT NOT NULL DEFAULT 'open',
    subsystem TEXT NOT NULL DEFAULT 'Core',
    reported_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_repository_issue_number UNIQUE (repository_id, number)
);

CREATE INDEX IF NOT EXISTS idx_issues_repo_number ON issues(repository_id, number);
CREATE INDEX IF NOT EXISTS idx_issues_repo_state ON issues(repository_id, state);

CREATE TABLE IF NOT EXISTS issue_blueprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    snapshot_id UUID REFERENCES repository_snapshots(id) ON DELETE SET NULL,
    blast_radius JSONB NOT NULL DEFAULT '{}'::jsonb,
    summary TEXT NOT NULL DEFAULT '',
    prerequisites JSONB NOT NULL DEFAULT '[]'::jsonb,
    affected_files JSONB NOT NULL DEFAULT '[]'::jsonb,
    stages JSONB NOT NULL DEFAULT '{}'::jsonb,
    test_strategy JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_issue_blueprint UNIQUE (issue_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_blueprints_issue_id ON issue_blueprints(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_blueprints_snapshot_id ON issue_blueprints(snapshot_id);
