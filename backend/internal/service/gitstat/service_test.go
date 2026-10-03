package gitstat_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gitwise/backend/internal/github"
	"github.com/gitwise/backend/internal/service/gitstat"
)

func TestGitStatService_CalculationAndAggregation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		w.Header().Set("Content-Type", "application/json")

		if path == "/users/octocat" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"login": "octocat",
				"id": 583231,
				"name": "The Octocat",
				"avatar_url": "https://avatars.githubusercontent.com/u/583231?v=4",
				"bio": "GitHub mascot",
				"public_repos": 8,
				"followers": 12000,
				"created_at": "2011-01-25T18:44:36Z"
			}`))
			return
		}

		if path == "/users/octocat/repos" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"name": "Hello-World",
					"full_name": "octocat/Hello-World",
					"description": "My first repository",
					"stargazers_count": 2500,
					"forks_count": 1800,
					"language": "Go",
					"html_url": "https://github.com/octocat/Hello-World",
					"owner": {"login": "octocat"}
				},
				{
					"name": "Spoon-Knife",
					"full_name": "octocat/Spoon-Knife",
					"description": "Repositroy for testing forks",
					"stargazers_count": 12000,
					"forks_count": 8000,
					"language": "TypeScript",
					"html_url": "https://github.com/octocat/Spoon-Knife",
					"owner": {"login": "octocat"}
				}
			]`))
			return
		}

		if path == "/users/octocat/events/public" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"id": "1",
					"type": "PushEvent",
					"created_at": "2026-03-01T10:00:00Z",
					"repo": {"name": "octocat/Hello-World"},
					"payload": {
						"size": 3,
						"commits": [{"sha": "a1b2c3d4e5", "message": "Add test cases"}]
					}
				},
				{
					"id": "2",
					"type": "PullRequestReviewEvent",
					"created_at": "2026-03-01T12:00:00Z",
					"repo": {"name": "octocat/Spoon-Knife"},
					"payload": {}
				}
			]`))
			return
		}

		if path == "/search/issues" {
			w.WriteHeader(http.StatusOK)
			q := r.URL.Query().Get("q")
			if strings.Contains(q, "is:merged") {
				_, _ = w.Write([]byte(`{"total_count": 42}`))
			} else if strings.Contains(q, "is:open") {
				_, _ = w.Write([]byte(`{"total_count": 8}`))
			} else {
				_, _ = w.Write([]byte(`{"total_count": 15}`))
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(server.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)

	profile, err := svc.GetProfile(context.Background(), "octocat", false)
	if err != nil {
		t.Fatalf("unexpected error fetching profile: %v", err)
	}

	if profile.Username != "octocat" {
		t.Errorf("expected username octocat, got %s", profile.Username)
	}
	if profile.Name != "The Octocat" {
		t.Errorf("expected name The Octocat, got %s", profile.Name)
	}
	if profile.Metrics.MergedPRs != 42 {
		t.Errorf("expected 42 merged PRs, got %d", profile.Metrics.MergedPRs)
	}
	if profile.Metrics.OpenPRs != 8 {
		t.Errorf("expected 8 open PRs, got %d", profile.Metrics.OpenPRs)
	}
	expectedMergeRate := 84.0 // 42 / (42 + 8) * 100 = 84.0%
	if profile.Metrics.MergeSuccessRatePct != expectedMergeRate {
		t.Errorf("expected merge rate %.1f%%, got %.1f%%", expectedMergeRate, profile.Metrics.MergeSuccessRatePct)
	}
	if profile.Metrics.CodeReviewsGiven < 1 {
		t.Errorf("expected at least 1 code review given, got %d", profile.Metrics.CodeReviewsGiven)
	}
	if len(profile.Repositories) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(profile.Repositories))
	}
	if profile.Repositories[0].Role != "Maintainer" {
		t.Errorf("expected Maintainer role for owned repo, got %s", profile.Repositories[0].Role)
	}
	if len(profile.ActivityWeeks) != 52 {
		t.Errorf("expected 52 activity weeks, got %d", len(profile.ActivityWeeks))
	}
}

func TestGitStatService_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(server.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)

	_, err := svc.GetProfile(context.Background(), "unknown_ghost_user_404", false)
	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}

	if !github.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %v", err)
	}
}

func TestGitStatService_Caching(t *testing.T) {
	requestsCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/cacheduser" {
			requestsCount++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"login": "cacheduser", "name": "Cached User"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(server.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)

	ctx := context.Background()

	// First call -> hits mock server
	_, err := svc.GetProfile(ctx, "cacheduser", false)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if requestsCount != 1 {
		t.Errorf("expected 1 request, got %d", requestsCount)
	}

	// Second call without forceRefresh -> served from in-memory cache
	_, err = svc.GetProfile(ctx, "cacheduser", false)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if requestsCount != 1 {
		t.Errorf("expected cached response without new request, got count %d", requestsCount)
	}

	// Third call with forceRefresh=true -> hits server again
	_, err = svc.GetProfile(ctx, "cacheduser", true)
	if err != nil {
		t.Fatalf("third call failed: %v", err)
	}
	if requestsCount != 2 {
		t.Errorf("expected 2 requests with forceRefresh, got %d", requestsCount)
	}
}

func TestGitStatService_TurnaroundAndLinesCalculation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/users/metricuser" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"login": "metricuser", "name": "Metric User", "created_at": "2020-01-01T00:00:00Z"}`))
			return
		}
		if r.URL.Path == "/users/metricuser/repos" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"name": "repo1", "full_name": "metricuser/repo1", "stargazers_count": 10, "owner": {"login": "metricuser"}}]`))
			return
		}
		if r.URL.Path == "/users/metricuser/events/public" {
			w.WriteHeader(http.StatusOK)
			// PR took 6 hours from creation to merge, with 120 additions and 30 deletions
			_, _ = w.Write([]byte(`[
				{
					"id": "100",
					"type": "PullRequestEvent",
					"created_at": "2026-03-01T16:00:00Z",
					"repo": {"name": "metricuser/repo1"},
					"payload": {
						"action": "closed",
						"pull_request": {
							"created_at": "2026-03-01T10:00:00Z",
							"merged_at": "2026-03-01T16:00:00Z",
							"additions": 120,
							"deletions": 30,
							"changed_files": 3
						}
					}
				}
			]`))
			return
		}
		if r.URL.Path == "/search/issues" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"total_count": 5}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ghClient := github.NewClient(server.URL, "", logger)
	svc := gitstat.NewService(ghClient, nil, logger)

	profile, err := svc.GetProfile(context.Background(), "metricuser", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.Metrics.LinesAdded != 120 {
		t.Errorf("expected 120 lines added, got %d", profile.Metrics.LinesAdded)
	}
	if profile.Metrics.LinesDeleted != 30 {
		t.Errorf("expected 30 lines deleted, got %d", profile.Metrics.LinesDeleted)
	}
	if profile.Metrics.ReviewTurnaroundHours != 6.0 {
		t.Errorf("expected 6.0 review turnaround hours, got %.1f", profile.Metrics.ReviewTurnaroundHours)
	}
}
