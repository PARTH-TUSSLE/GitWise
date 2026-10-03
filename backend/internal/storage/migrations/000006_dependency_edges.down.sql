-- 000006_dependency_edges.down.sql
DROP INDEX IF EXISTS idx_dep_edges_unique;
DROP INDEX IF EXISTS idx_dep_edges_source_target;
DROP INDEX IF EXISTS idx_dep_edges_target;
DROP INDEX IF EXISTS idx_dep_edges_source;
DROP INDEX IF EXISTS idx_dep_edges_snapshot;
DROP TABLE IF EXISTS dependency_edges CASCADE;
