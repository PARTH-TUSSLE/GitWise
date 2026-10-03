package handlers_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/api"
	"github.com/gitwise/backend/internal/api/middleware"
	"github.com/gitwise/backend/internal/config"
	"github.com/gitwise/backend/internal/domain"
	"strings"
)

func TestE2E_FullPlatformWorkflow(t *testing.T) {
	mockGH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.ToLower(r.URL.Path)
		if path == "/users/alexr_dev" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"login": "alexR_dev",
				"id": 12345,
				"name": "Alex Rivera",
				"public_repos": 18,
				"followers": 150,
				"created_at": "2018-05-10T12:00:00Z"
			}`))
			return
		}
		if path == "/users/alexr_dev/repos" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"name": "next.js",
					"full_name": "vercel/next.js",
					"stargazers_count": 120000,
					"forks_count": 25000,
					"language": "TypeScript",
					"owner": {"login": "vercel"}
				}
			]`))
			return
		}
		if path == "/search/issues" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"total_count": 148}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer mockGH.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		AppEnv:           "test",
		HTTPHost:         "127.0.0.1",
		HTTPPort:         "8080",
		FrontendOrigin:   "http://localhost:3000",
		GitHubAPIBaseURL: mockGH.URL,
		AIProvider:       "mock",
	}

	router := api.NewRouter(cfg, nil, logger, "2.1.0-production")

	// Step 1: Verify Health Check endpoint
	reqHealth := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recHealth := httptest.NewRecorder()
	router.ServeHTTP(recHealth, reqHealth)

	if recHealth.Code != http.StatusOK {
		t.Fatalf("Step 1 Healthz: expected 200, got %d", recHealth.Code)
	}

	var healthResp struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(recHealth.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("Step 1: failed to decode health response: %v", err)
	}
	if healthResp.Status != "ok" || healthResp.Version != "2.1.0-production" {
		t.Errorf("Step 1: unexpected health response: %+v", healthResp)
	}

	// Step 2: Verify Telemetry GitStat Profile (Mock Mode predefined profile)
	reqProfile := httptest.NewRequest(http.MethodGet, "/api/v1/gitstat/alexR_dev", nil)
	recProfile := httptest.NewRecorder()
	router.ServeHTTP(recProfile, reqProfile)

	if recProfile.Code != http.StatusOK {
		t.Fatalf("Step 2 GitStat: expected 200, got %d", recProfile.Code)
	}

	var profile domain.ContributorProfile
	if err := json.Unmarshal(recProfile.Body.Bytes(), &profile); err != nil {
		t.Fatalf("Step 2: failed to decode profile: %v", err)
	}
	if profile.Username != "alexR_dev" {
		t.Errorf("Step 2: expected username alexR_dev, got %s", profile.Username)
	}
	if profile.Metrics.MergedPRs <= 0 {
		t.Errorf("Step 2: expected positive merged PRs, got %d", profile.Metrics.MergedPRs)
	}

	// Step 3: Verify Repository Architecture Model (Predefined Repository)
	reqRepo := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/vercel/next.js", nil)
	recRepo := httptest.NewRecorder()
	router.ServeHTTP(recRepo, reqRepo)

	if recRepo.Code != http.StatusOK {
		t.Fatalf("Step 3 Repository: expected 200, got %d", recRepo.Code)
	}

	var repoModel domain.RepoModel
	if err := json.Unmarshal(recRepo.Body.Bytes(), &repoModel); err != nil {
		t.Fatalf("Step 3: failed to decode repo model: %v", err)
	}
	if repoModel.Name != "vercel/next.js" {
		t.Errorf("Step 3: expected repo name vercel/next.js, got %s", repoModel.Name)
	}
	if len(repoModel.Subsystems) == 0 {
		t.Errorf("Step 3: expected subsystems in repo model, got 0")
	}

	// Step 4: Verify Non-existent route returns 404
	req404 := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent-route", nil)
	rec404 := httptest.NewRecorder()
	router.ServeHTTP(rec404, req404)

	if rec404.Code != http.StatusNotFound {
		t.Errorf("Step 4: expected 404 for nonexistent route, got %d", rec404.Code)
	}
}

func TestE2E_RateLimitingUnderLoad(t *testing.T) {
	rl := middleware.NewRateLimiter(3, 3, time.Minute)
	defer rl.Close()

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	clientIP := "203.0.113.42:8000"

	// Fire 3 requests within burst limit -> all 200 OK
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/test/repo", nil)
		req.RemoteAddr = clientIP
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 4th request must be rejected with 429
	req4 := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/test/repo", nil)
	req4.RemoteAddr = clientIP
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)

	if rec4.Code != http.StatusTooManyRequests {
		t.Fatalf("request 4: expected 429, got %d", rec4.Code)
	}

	var errResp map[string]interface{}
	if err := json.Unmarshal(rec4.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode 429 response: %v", err)
	}
	if errResp["error"] != "rate_limit_exceeded" {
		t.Errorf("expected error rate_limit_exceeded, got %v", errResp["error"])
	}

	// /healthz bypass must succeed with 200 even after rate limit exhaustion
	reqHealth := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	reqHealth.RemoteAddr = clientIP
	recHealth := httptest.NewRecorder()
	handler.ServeHTTP(recHealth, reqHealth)

	if recHealth.Code != http.StatusOK {
		t.Fatalf("healthcheck bypass: expected 200, got %d", recHealth.Code)
	}
}
