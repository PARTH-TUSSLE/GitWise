package handlers_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gitwise/backend/internal/api/handlers"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/github"
	"github.com/gitwise/backend/internal/service/gitstat"
	"github.com/go-chi/chi/v5"
)

func TestHealthHandler(t *testing.T) {
	h := handlers.NewHealthHandler(nil, "2.1.0-test")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp handlers.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Version != "2.1.0-test" {
		t.Errorf("expected version '2.1.0-test', got %q", resp.Version)
	}
	if resp.Database != "disconnected" {
		t.Errorf("expected database 'disconnected' with nil db, got %q", resp.Database)
	}
}

func TestGitStatHandler_MockMode_PredefinedProfile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handlers.NewGitStatHandler(nil, logger)

	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/alexR_dev", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var profile domain.ContributorProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if profile.Username != "alexR_dev" {
		t.Errorf("expected username alexR_dev, got %s", profile.Username)
	}
	if profile.Metrics.MergedPRs != 148 {
		t.Errorf("expected 148 merged PRs, got %d", profile.Metrics.MergedPRs)
	}
	if len(profile.Repositories) != 4 {
		t.Errorf("expected 4 repositories, got %d", len(profile.Repositories))
	}
	if len(profile.ActivityWeeks) != 52 {
		t.Errorf("expected 52 activity weeks, got %d", len(profile.ActivityWeeks))
	}
}

func TestGitStatHandler_MockMode_ArbitraryUser_Returns404(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handlers.NewGitStatHandler(nil, logger)

	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/unknown-user-99", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// In mock mode, arbitrary usernames MUST return 404 instead of generating fake metrics
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for unknown user in mock mode, got %d", rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error json: %v", err)
	}
	if errResp["error"] != "not_found" {
		t.Errorf("expected error 'not_found', got %s", errResp["error"])
	}
}

func TestGitStatHandler_LiveMode_KnownDemoUser_GitHub404Returns404(t *testing.T) {
	// Even for known demo username "alexr_dev", when GitHub API returns 404 in live mode,
	// the handler MUST return 404 and NEVER silently substitute mock data!
	mockGH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer mockGH.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(mockGH.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)
	h := handlers.NewGitStatHandler(svc, logger)

	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/alexr_dev", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 in live mode when GitHub returns 404 for alexr_dev, got %d", rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error json: %v", err)
	}
	if errResp["error"] != "not_found" {
		t.Errorf("expected error 'not_found', got %s", errResp["error"])
	}
}

func TestGitStatHandler_LiveMode_KnownDemoUser_Upstream500DoesNotReturnMock(t *testing.T) {
	// For known demo username "alexr_dev", when GitHub API fails with 500,
	// the handler MUST return 502 with sanitized error and NEVER return mock data!
	mockGH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "Internal GitHub Server Error with internal url https://github.internal/secret"}`))
	}))
	defer mockGH.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(mockGH.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)
	h := handlers.NewGitStatHandler(svc, logger)

	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/alexr_dev", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 Bad Gateway on upstream error, got %d", rec.Code)
	}

	// Verify sanitized response does not contain mock data or internal upstream secrets
	bodyStr := rec.Body.String()
	if strings.Contains(bodyStr, "Alex Rivera") || strings.Contains(bodyStr, "github.internal") {
		t.Errorf("leaked mock profile or internal error details in response: %s", bodyStr)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error json: %v", err)
	}
	if errResp["error"] != "upstream_error" {
		t.Errorf("expected error 'upstream_error', got %s", errResp["error"])
	}
	if errResp["message"] != "Failed to retrieve developer telemetry from upstream service" {
		t.Errorf("expected sanitized error message, got %s", errResp["message"])
	}
}

func TestGitStatHandler_LiveMode_ArbitraryUser_CannotReceiveFakeMetrics(t *testing.T) {
	// In live mode, an arbitrary non-existent username must return 404 and NEVER receive fake metrics.
	mockGH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer mockGH.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(mockGH.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)
	h := handlers.NewGitStatHandler(svc, logger)

	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/arbitrary_random_user_999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for arbitrary user in live service, got %d", rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error json: %v", err)
	}
	if errResp["error"] != "not_found" {
		t.Errorf("expected error 'not_found', got %s", errResp["error"])
	}
}

func TestGitStatHandler_LiveMode_RateLimit(t *testing.T) {
	mockGH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1800000000")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer mockGH.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(mockGH.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)
	h := handlers.NewGitStatHandler(svc, logger)

	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/rate_limited_user", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for rate limit, got %d", rec.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error json: %v", err)
	}
	if errResp["error"] != "rate_limit_exceeded" {
		t.Errorf("expected error 'rate_limit_exceeded', got %s", errResp["error"])
	}
}

func TestRepoHandler_PredefinedRepo(t *testing.T) {
	h := handlers.NewRepoHandler(nil, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/api/v1/repositories/{owner}/{repo}", h.GetRepository)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/vercel/next.js", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var model domain.RepoModel
	if err := json.Unmarshal(rec.Body.Bytes(), &model); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if model.Name != "vercel/next.js" {
		t.Errorf("expected name vercel/next.js, got %s", model.Name)
	}
	if len(model.Subsystems) != 4 {
		t.Errorf("expected 4 subsystems, got %d", len(model.Subsystems))
	}
	if len(model.FeatureTraces) == 0 {
		t.Errorf("expected feature traces to be present")
	}
}

func TestRepoHandler_DynamicFallbackRepo(t *testing.T) {
	h := handlers.NewRepoHandler(nil, nil, nil, nil)

	r := chi.NewRouter()
	r.Get("/api/v1/repositories/{owner}/{repo}", h.GetRepository)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/octocat/hello-world", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var model domain.RepoModel
	if err := json.Unmarshal(rec.Body.Bytes(), &model); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if model.Name != "octocat/hello-world" {
		t.Errorf("expected name octocat/hello-world, got %s", model.Name)
	}
	if len(model.Subsystems) == 0 {
		t.Errorf("expected at least 1 fallback subsystem")
	}
}

func TestRouter_NotFound(t *testing.T) {
	h := handlers.NewHealthHandler(nil, "2.1.0-test")
	r := chi.NewRouter()
	r.Get("/healthz", h.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent/endpoint", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown route, got %d", rec.Code)
	}
}

func TestRouter_MissingParams(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handlers.NewGitStatHandler(nil, logger)
	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	// Test missing username
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for trailing slash without username, got %d", rec.Code)
	}
}

func TestRepoHandler_IngestValidation(t *testing.T) {
	h := handlers.NewRepoHandler(nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/api/v1/repositories/ingest", h.IngestRepository)

	// Missing body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repositories/ingest", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// When service is nil, it should report 503
	if rec.Code != http.StatusServiceUnavailable && rec.Code != http.StatusBadRequest {
		t.Errorf("expected 503 or 400, got %d", rec.Code)
	}
}

func TestRepoHandler_GetJobValidation(t *testing.T) {
	h := handlers.NewRepoHandler(nil, nil, nil, nil)
	r := chi.NewRouter()
	r.Get("/api/v1/jobs/{id}", h.GetJob)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable && rec.Code != http.StatusBadRequest {
		t.Errorf("expected 503 or 400, got %d", rec.Code)
	}
}
