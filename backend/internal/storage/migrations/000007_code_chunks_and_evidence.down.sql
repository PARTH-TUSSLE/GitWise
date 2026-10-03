-- 000007_code_chunks_and_evidence.down.sql
DROP INDEX IF EXISTS idx_evidence_file;
DROP INDEX IF EXISTS idx_evidence_snapshot;
DROP TABLE IF EXISTS evidence_refs CASCADE;

DROP INDEX IF EXISTS idx_chunks_embedding;
DROP INDEX IF EXISTS idx_chunks_tsv;
DROP INDEX IF EXISTS idx_chunks_file;
DROP INDEX IF EXISTS idx_chunks_snapshot;
DROP TABLE IF EXISTS code_chunks CASCADE;
