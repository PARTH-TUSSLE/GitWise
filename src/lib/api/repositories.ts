// src/lib/api/repositories.ts
// Centralized Repository architecture data access layer

import { RepoModel, REPOSITORIES } from "@/lib/repoData";
import { isLiveMode } from "./config";
import { apiFetch } from "./client";

export interface IngestResponse {
  jobId: string;
  snapshotId: string;
  status: "QUEUED" | "PROCESSING" | "COMPLETED" | "FAILED";
  stage: string;
  commitSha: string;
}

export interface AnalysisJob {
  id: string;
  type: string;
  snapshotId?: string;
  status: "QUEUED" | "PROCESSING" | "COMPLETED" | "FAILED";
  stage: string;
  progressPercent: number;
  errorMessage?: string;
  retryCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface RepositoryFile {
  id: string;
  snapshotId: string;
  path: string;
  extension: string;
  language: string;
  sizeBytes: number;
  lineCount: number;
  sha256Hash: string;
  content?: string;
  isBinary: boolean;
  createdAt: string;
}

/**
 * Triggers repository ingestion and snapshot creation.
 */
export async function ingestRepository(
  owner: string,
  repo: string,
  ref: string = "main"
): Promise<IngestResponse> {
  if (isLiveMode()) {
    return await apiFetch<IngestResponse>("/api/v1/repositories/ingest", {
      method: "POST",
      body: JSON.stringify({ owner, repo, ref }),
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });
  }

  // Mock mode fallback response
  return {
    jobId: "00000000-0000-0000-0000-000000000001",
    snapshotId: "00000000-0000-0000-0000-000000000002",
    status: "COMPLETED",
    stage: "DONE",
    commitSha: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4",
  };
}

/**
 * Retrieves background job status.
 */
export async function fetchJob(jobId: string): Promise<AnalysisJob> {
  if (isLiveMode()) {
    return await apiFetch<AnalysisJob>(`/api/v1/jobs/${encodeURIComponent(jobId)}`, {
      cache: "no-store",
    });
  }

  return {
    id: jobId,
    type: "SNAPSHOT_INGEST",
    status: "COMPLETED",
    stage: "DONE",
    progressPercent: 100.0,
    retryCount: 0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };
}

/**
 * Retrieves all files recorded for a repository snapshot.
 */
export async function fetchSnapshotFiles(
  owner: string,
  repo: string,
  commitSha: string
): Promise<RepositoryFile[]> {
  if (isLiveMode()) {
    return await apiFetch<RepositoryFile[]>(
      `/api/v1/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/snapshots/${encodeURIComponent(commitSha)}/files`,
      { cache: "no-store" }
    );
  }

  return [];
}

/**
 * Retrieves repository intelligence snapshot.
 * - mock mode: returns local mock fixtures
 * - live mode: strictly queries Go backend API and propagates any errors
 */
export async function fetchRepository(owner: string, repo: string): Promise<RepoModel> {
  if (isLiveMode()) {
    return await apiFetch<RepoModel>(
      `/api/v1/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`,
      { cache: "no-store" }
    );
  }

  // Mock mode: local fixtures only
  const repoKey = `${owner}/${repo}`.toLowerCase();
  const mock =
    REPOSITORIES[repoKey] ||
    REPOSITORIES["vercel/next.js"];

  return mock;
}


