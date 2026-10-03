-- 000006_dependency_edges.up.sql
-- GitWise Phase 5 Schema: Dependency Edges for Code Intelligence Graph & Impact Analysis

CREATE TABLE IF NOT EXISTS dependency_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL REFERENCES repository_snapshots(id) ON DELETE CASCADE,
    source_file_id UUID NOT NULL REFERENCES repository_files(id) ON DELETE CASCADE,
    target_file_id UUID REFERENCES repository_files(id) ON DELETE CASCADE,
    edge_type VARCHAR(50) NOT NULL, -- IMPORTS, CALLS_NAME, IMPLEMENTS
    is_deterministic BOOLEAN NOT NULL DEFAULT true, -- false for name-inferred calls
    line_number INT,
    raw_target VARCHAR(1024),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dep_edges_snapshot ON dependency_edges(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_dep_edges_source ON dependency_edges(source_file_id);
CREATE INDEX IF NOT EXISTS idx_dep_edges_target ON dependency_edges(target_file_id);
CREATE INDEX IF NOT EXISTS idx_dep_edges_source_target ON dependency_edges(source_file_id, target_file_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_dep_edges_unique ON dependency_edges(
    snapshot_id,
    source_file_id,
    COALESCE(target_file_id, '00000000-0000-0000-0000-000000000000'::uuid),
    edge_type,
    COALESCE(line_number, 0),
    COALESCE(raw_target, '')
);
