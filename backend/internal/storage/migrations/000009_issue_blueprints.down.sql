-- Migration: 000009_issue_blueprints.down.sql
-- Description: Reverts issues and issue_blueprints tables.

DROP TABLE IF EXISTS issue_blueprints CASCADE;
DROP TABLE IF EXISTS issues CASCADE;
