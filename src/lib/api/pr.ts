// src/lib/api/pr.ts
// PR Intelligence and Semantic Review API client

import { PullRequestModel, PR_DATABASE } from "@/lib/prData";
import { isLiveMode } from "./config";
import { apiFetch } from "./client";

/**
 * Fetches all tracked pull requests for a repository.
 */
export async function fetchPullRequests(owner: string, repo: string): Promise<PullRequestModel[]> {
  const repoKey = `${owner}/${repo}`;
  if (isLiveMode()) {
    try {
      const data = await apiFetch<PullRequestModel[]>(`/api/v1/pr/${owner}/${repo}`, {
        cache: "no-store",
      });
      if (Array.isArray(data) && data.length > 0) {
        return data;
      }
    } catch (err) {
      console.warn(`[pr] Live API fetch failed for ${repoKey}, using fallback:`, err);
    }
  }

  return PR_DATABASE[repoKey] || PR_DATABASE["vercel/next.js"] || [];
}

/**
 * Fetches a single pull request and its semantic analysis findings.
 */
export async function fetchPullRequest(
  owner: string,
  repo: string,
  number: number
): Promise<PullRequestModel | null> {
  const repoKey = `${owner}/${repo}`;
  if (isLiveMode()) {
    try {
      return await apiFetch<PullRequestModel>(`/api/v1/pr/${owner}/${repo}/${number}`, {
        cache: "no-store",
      });
    } catch (err) {
      console.warn(`[pr] Live API fetch failed for #${number}, falling back:`, err);
    }
  }

  const list = PR_DATABASE[repoKey] || PR_DATABASE["vercel/next.js"] || [];
  return list.find((p) => p.number === number) || null;
}

/**
 * Requests an on-demand semantic diff review of a pull request.
 */
export async function reviewPullRequest(
  owner: string,
  repo: string,
  number: number,
  options?: { useAi?: boolean; model?: string; provider?: string }
): Promise<PullRequestModel> {
  if (isLiveMode()) {
    return await apiFetch<PullRequestModel>(`/api/v1/pr/${owner}/${repo}/${number}/review`, {
      method: "POST",
      body: JSON.stringify(options || {}),
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });
  }

  const list = PR_DATABASE[`${owner}/${repo}`] || PR_DATABASE["vercel/next.js"] || [];
  const found = list.find((p) => p.number === number);
  if (!found) {
    throw new Error(`Pull Request #${number} not found`);
  }
  return found;
}
