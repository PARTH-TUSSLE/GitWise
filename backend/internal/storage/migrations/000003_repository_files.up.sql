-- 000003_repository_files.up.sql
-- GitWise Phase 3 Schema: Repository Files for Commit-Addressed Snapshots

CREATE TABLE IF NOT EXISTS repository_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL REFERENCES repository_snapshots(id) ON DELETE CASCADE,
    path VARCHAR(1024) NOT NULL,
    extension VARCHAR(50),
    language VARCHAR(100),
    size_bytes INT NOT NULL,
    line_count INT NOT NULL,
    sha256_hash CHAR(64) NOT NULL,
    content TEXT,
    is_binary BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_repo_files_snapshot_path ON repository_files(snapshot_id, path);
CREATE INDEX IF NOT EXISTS idx_repo_files_snapshot ON repository_files(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_repo_files_language ON repository_files(snapshot_id, language);
