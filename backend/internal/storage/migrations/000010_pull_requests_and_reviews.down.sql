-- Migration: 000010_pull_requests_and_reviews.down.sql
-- Description: Reverts pull_requests and pr_reviews tables.

DROP TABLE IF EXISTS pr_reviews CASCADE;
DROP TABLE IF EXISTS pull_requests CASCADE;
