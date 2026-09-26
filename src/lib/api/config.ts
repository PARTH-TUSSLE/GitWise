// src/lib/api/config.ts
// Centralized configuration for frontend API access

export type ApiMode = "mock" | "live";

export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";
}

export function getApiMode(): ApiMode {
  const mode = process.env.NEXT_PUBLIC_API_MODE;
  if (mode === "live") {
    return "live";
  }
  return "mock";
}

export function isLiveMode(): boolean {
  return getApiMode() === "live";
}
