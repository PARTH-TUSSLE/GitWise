// src/lib/api/client.ts
// Base HTTP client with timeout, error envelope handling, and non-2xx detection

import { getApiBaseUrl } from "./config";

export interface ApiErrorPayload {
  error: {
    code: string;
    message: string;
    requestId?: string;
  };
}

export class ApiError extends Error {
  public status: number;
  public code: string;
  public requestId?: string;

  constructor(status: number, message: string, code = "API_ERROR", requestId?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.requestId = requestId;
  }
}

export interface HealthResponse {
  status: string;
  version: string;
  uptime: string;
  database: string;
  timestamp: string;
}

/**
 * Base fetch wrapper with configurable timeout and error handling.
 */
export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const baseUrl = getApiBaseUrl();
  const url = `${baseUrl}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`;

  const timeoutMs = 8000;
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const res = await fetch(url, {
      ...options,
      signal: controller.signal,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    clearTimeout(timeoutId);

    if (!res.ok) {
      let errorMessage = `HTTP ${res.status}: ${res.statusText}`;
      let errorCode = "HTTP_ERROR";
      let reqId = res.headers.get("X-Request-ID") || undefined;

      try {
        const errJson = await res.json();
        if (errJson?.error) {
          if (typeof errJson.error === "string") {
            errorMessage = errJson.error;
          } else if (errJson.error.message) {
            errorMessage = errJson.error.message;
            errorCode = errJson.error.code || errorCode;
          }
        }
        if (errJson?.requestId) {
          reqId = errJson.requestId;
        }
      } catch {
        // Response body was not JSON, retain default status message
      }

      throw new ApiError(res.status, errorMessage, errorCode, reqId);
    }

    return (await res.json()) as T;
  } catch (err: unknown) {
    clearTimeout(timeoutId);
    if (err instanceof ApiError) {
      throw err;
    }
    if (err instanceof Error && err.name === "AbortError") {
      throw new ApiError(408, "Request timed out", "TIMEOUT");
    }
    throw new ApiError(0, err instanceof Error ? err.message : "Network failure", "NETWORK_ERROR");
  }
}

/**
 * Health check querying GET /healthz
 */
export async function checkBackendHealth(): Promise<HealthResponse> {
  return apiFetch<HealthResponse>("/healthz");
}
