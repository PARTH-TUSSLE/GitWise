-- 000005_code_symbols.up.sql
-- GitWise Phase 4 Schema: Code Symbols from Structural AST Analysis

CREATE TABLE IF NOT EXISTS code_symbols (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL REFERENCES repository_snapshots(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES repository_files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(50) NOT NULL, -- FUNCTION, METHOD, STRUCT, INTERFACE, CLASS, TYPE
    start_line INT NOT NULL,
    end_line INT NOT NULL,
    signature TEXT,
    is_exported BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_code_symbols_snapshot ON code_symbols(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_code_symbols_snapshot_name ON code_symbols(snapshot_id, name);
CREATE INDEX IF NOT EXISTS idx_code_symbols_file ON code_symbols(file_id);
CREATE INDEX IF NOT EXISTS idx_code_symbols_snapshot_kind ON code_symbols(snapshot_id, kind);
CREATE UNIQUE INDEX IF NOT EXISTS idx_code_symbols_file_name_kind_line ON code_symbols(file_id, name, kind, start_line);
