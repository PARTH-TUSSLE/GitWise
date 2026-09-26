package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gitwise/backend/internal/api/handlers"
	"github.com/gitwise/backend/internal/domain"
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

func TestGitStatHandler_PredefinedProfile(t *testing.T) {
	h := handlers.NewGitStatHandler()

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

func TestGitStatHandler_DynamicFallbackProfile(t *testing.T) {
	h := handlers.NewGitStatHandler()

	r := chi.NewRouter()
	r.Get("/api/v1/gitstat/{username}", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/unknown-user-99", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var profile domain.ContributorProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if profile.Username != "unknown-user-99" {
		t.Errorf("expected username unknown-user-99, got %s", profile.Username)
	}
	if len(profile.ActivityWeeks) != 52 {
		t.Errorf("expected 52 activity weeks, got %d", len(profile.ActivityWeeks))
	}
	if profile.ProvenanceNote == "" {
		t.Errorf("expected provenance note explaining synthesized baseline")
	}
}

func TestRepoHandler_PredefinedRepo(t *testing.T) {
	h := handlers.NewRepoHandler()

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
	h := handlers.NewRepoHandler()

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
	h := handlers.NewGitStatHandler()
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
