// src/lib/api/repositories.ts
// Centralized Repository architecture data access layer

import { RepoModel, REPOSITORIES } from "@/lib/repoData";
import { isLiveMode } from "./config";
import { apiFetch } from "./client";

/**
 * Retrieves repository intelligence snapshot.
 * Respects NEXT_PUBLIC_API_MODE (mock vs live) while preserving contract parity.
 */
export async function fetchRepository(owner: string, repo: string): Promise<RepoModel> {
  const repoKey = `${owner}/${repo}`.toLowerCase();

  if (isLiveMode()) {
    try {
      const data = await apiFetch<RepoModel>(
        `/api/v1/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`,
        { cache: "no-store" }
      );
      return data;
    } catch (err) {
      console.warn(`[GitWise Live API] Failed to fetch repository for ${owner}/${repo}, falling back to mock:`, err);
    }
  }

  // Mock mode / fallback
  const mock =
    REPOSITORIES[repoKey] ||
    REPOSITORIES["vercel/next.js"];

  return mock;
}
