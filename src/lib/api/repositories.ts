// src/lib/api/repositories.ts
// Centralized Repository architecture data access layer

import { RepoModel, REPOSITORIES } from "@/lib/repoData";
import { isLiveMode } from "./config";
import { apiFetch } from "./client";

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
