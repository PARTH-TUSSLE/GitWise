// src/lib/api/gitstat.ts
// Centralized GitStat telemetry data access layer

import { ContributorProfile, CONTRIBUTORS } from "@/lib/mockData";
import { isLiveMode } from "./config";
import { apiFetch } from "./client";

/**
 * Retrieves contributor telemetry profile.
 * Respects NEXT_PUBLIC_API_MODE (mock vs live) while preserving contract parity.
 */
export async function fetchContributorProfile(username: string): Promise<ContributorProfile> {
  const normUser = username.toLowerCase();

  if (isLiveMode()) {
    try {
      const data = await apiFetch<ContributorProfile>(
        `/api/v1/gitstat/${encodeURIComponent(username)}`,
        { cache: "no-store" }
      );
      return data;
    } catch (err) {
      console.warn(`[GitWise Live API] Failed to fetch gitstat for ${username}, falling back to mock:`, err);
    }
  }

  // Mock mode / fallback
  const mock =
    CONTRIBUTORS[normUser] ||
    CONTRIBUTORS["alexr_dev"] ||
    CONTRIBUTORS["alexR_dev"];

  return mock;
}
