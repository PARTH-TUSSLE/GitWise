// src/lib/api/gitstat.ts
// Centralized GitStat telemetry data access layer

import { ContributorProfile, CONTRIBUTORS } from "@/lib/mockData";
import { isLiveMode } from "./config";
import { apiFetch } from "./client";

/**
 * Retrieves contributor telemetry profile.
 * - mock mode: returns local mock fixtures
 * - live mode: strictly queries Go backend API and propagates any errors
 */
export async function fetchContributorProfile(username: string): Promise<ContributorProfile> {
  if (isLiveMode()) {
    return await apiFetch<ContributorProfile>(
      `/api/v1/gitstat/${encodeURIComponent(username)}`,
      { cache: "no-store" }
    );
  }

  // Mock mode: local fixtures only
  const normUser = username.toLowerCase();
  const mock =
    CONTRIBUTORS[normUser] ||
    CONTRIBUTORS["alexr_dev"] ||
    CONTRIBUTORS["alexR_dev"];

  return mock;
}
