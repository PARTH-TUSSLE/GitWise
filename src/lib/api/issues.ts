// src/lib/api/issues.ts
// Issue Intelligence and Implementation Blueprint API client

import { IssueModel, ISSUES_DATABASE } from "@/lib/issueData";
import { isLiveMode } from "./config";
import { apiFetch } from "./client";

/**
 * Fetches all tracked issues for a repository.
 */
export async function fetchIssues(owner: string, repo: string): Promise<IssueModel[]> {
  const repoKey = `${owner}/${repo}`;
  if (isLiveMode()) {
    try {
      const data = await apiFetch<IssueModel[]>(`/api/v1/issues/${owner}/${repo}`, {
        cache: "no-store",
      });
      if (Array.isArray(data) && data.length > 0) {
        return data;
      }
    } catch (err) {
      console.warn(`[issues] Live API fetch failed for ${repoKey}, using fallback:`, err);
    }
  }

  return ISSUES_DATABASE[repoKey] || ISSUES_DATABASE["vercel/next.js"] || [];
}

/**
 * Fetches a single issue and its current blueprint.
 */
export async function fetchIssue(owner: string, repo: string, number: number): Promise<IssueModel | null> {
  const repoKey = `${owner}/${repo}`;
  if (isLiveMode()) {
    try {
      return await apiFetch<IssueModel>(`/api/v1/issues/${owner}/${repo}/${number}`, {
        cache: "no-store",
      });
    } catch (err) {
      console.warn(`[issues] Live API fetch failed for #${number}, falling back:`, err);
    }
  }

  const list = ISSUES_DATABASE[repoKey] || ISSUES_DATABASE["vercel/next.js"] || [];
  return list.find((i) => i.number === number) || null;
}

/**
 * Requests generation of a sequenced implementation blueprint.
 */
export async function generateBlueprint(
  owner: string,
  repo: string,
  number: number,
  options?: { snapshotId?: string; ref?: string; useAi?: boolean; model?: string; provider?: string }
): Promise<IssueModel> {
  if (isLiveMode()) {
    return await apiFetch<IssueModel>(`/api/v1/issues/${owner}/${repo}/${number}/blueprint`, {
      method: "POST",
      body: JSON.stringify(options || {}),
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });
  }

  const list = ISSUES_DATABASE[`${owner}/${repo}`] || ISSUES_DATABASE["vercel/next.js"] || [];
  const found = list.find((i) => i.number === number);
  if (!found) {
    throw new Error(`Issue #${number} not found`);
  }
  return found;
}
