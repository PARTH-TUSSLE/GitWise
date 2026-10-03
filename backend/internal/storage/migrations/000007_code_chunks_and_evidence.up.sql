-- 000007_code_chunks_and_evidence.up.sql
-- GitWise Phase 6 Schema: Semantic Code Chunks, pgvector Embeddings, and Evidence References

CREATE TABLE IF NOT EXISTS code_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL REFERENCES repository_snapshots(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES repository_files(id) ON DELETE CASCADE,
    symbol_id UUID REFERENCES code_symbols(id) ON DELETE SET NULL,
    start_line INT NOT NULL,
    end_line INT NOT NULL,
    scope VARCHAR(255),
    content TEXT NOT NULL,
    content_tsv TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', content)) STORED,
    embedding VECTOR(768) NOT NULL,
    embedding_model VARCHAR(100) NOT NULL DEFAULT 'text-embedding-004',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chunks_snapshot ON code_chunks(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_chunks_file ON code_chunks(file_id);
CREATE INDEX IF NOT EXISTS idx_chunks_tsv ON code_chunks USING GIN(content_tsv);
CREATE INDEX IF NOT EXISTS idx_chunks_embedding ON code_chunks USING hnsw (embedding vector_cosine_ops);

CREATE TABLE IF NOT EXISTS evidence_refs (
    id VARCHAR(50) PRIMARY KEY, -- e.g. 'ev_01', 'ev_a1b2c3'
    snapshot_id UUID NOT NULL REFERENCES repository_snapshots(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES repository_files(id) ON DELETE CASCADE,
    start_line INT NOT NULL,
    end_line INT NOT NULL,
    content_hash CHAR(64) NOT NULL,
    provenance VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_evidence_snapshot ON evidence_refs(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_evidence_file ON evidence_refs(file_id);
