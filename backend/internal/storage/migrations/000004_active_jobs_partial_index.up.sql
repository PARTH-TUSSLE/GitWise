-- 000004_active_jobs_partial_index.up.sql
-- GitWise Phase 3 Migration: Enforce at most one active job (QUEUED or PROCESSING) per snapshot

CREATE UNIQUE INDEX IF NOT EXISTS idx_active_jobs_snapshot
ON analysis_jobs (snapshot_id)
WHERE status IN ('QUEUED', 'PROCESSING') AND snapshot_id IS NOT NULL;
