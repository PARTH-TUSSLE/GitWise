-- 000001_init_schema.down.sql
-- GitWise Phase 1 Schema Rollback

DROP TABLE IF EXISTS analysis_jobs;
DROP TABLE IF EXISTS repository_snapshots;
DROP TABLE IF EXISTS repositories;
DROP EXTENSION IF EXISTS vector;
