package github_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/github"
)

func TestGitHubClient_RateLimitTracking(t *testing.T) {
	resetEpoch := time.Now().Add(1 * time.Hour).Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetEpoch, 10))
		w.Header().Set("X-RateLimit-Resource", "core")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"login":"testuser"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := github.NewClient(server.URL, "", logger)

	ctx := context.Background()
	user, err := client.GetUser(ctx, "testuser")
	if err != nil {
		t.Fatalf("expected successful GetUser, got: %v", err)
	}
	if user.Login != "testuser" {
		t.Errorf("expected login 'testuser', got: %s", user.Login)
	}

	rl := client.GetRateLimit()
	if rl.Limit != 60 || rl.Remaining != 42 || rl.Resource != "core" {
		t.Errorf("rate limit info mismatch: %+v", rl)
	}
}

func TestGitHubClient_RateLimitExceeded(t *testing.T) {
	resetEpoch := time.Now().Add(30 * time.Minute).Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetEpoch, 10))
		w.Header().Set("X-RateLimit-Resource", "core")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded for user IP"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := github.NewClient(server.URL, "", logger)

	_, err := client.GetUser(context.Background(), "testuser")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !github.IsRateLimit(err) {
		t.Errorf("expected error to be RateLimitError, got: %v", err)
	}
}

func TestGitHubClient_UserNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := github.NewClient(server.URL, "", logger)

	_, err := client.GetUser(context.Background(), "nonexistent_user")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !github.IsNotFound(err) {
		t.Errorf("expected error to be NotFoundError, got: %v", err)
	}
}

func TestGitHubClient_AuthHeaderInjection(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"login":"authorized_user"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := github.NewClient(server.URL, "ghp_secret_token_123", logger)

	_, err := client.GetUser(context.Background(), "authorized_user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAuth != "Bearer ghp_secret_token_123" {
		t.Errorf("expected Bearer token in Authorization header, got: %s", capturedAuth)
	}
}

func TestGitHubClient_GraphQLGracefulDegradation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Without token -> returns ErrGraphQLUnavailable immediately
	unauthClient := github.NewClient("https://api.github.com", "", logger)
	_, err := unauthClient.GetContributionCalendar(context.Background(), "torvalds")
	if err != github.ErrGraphQLUnavailable {
		t.Errorf("expected ErrGraphQLUnavailable for unauthenticated client, got: %v", err)
	}

	// With token and valid GraphQL payload -> returns parsed calendar
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "graphql") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"data": {
					"user": {
						"contributionsCollection": {
							"contributionCalendar": {
								"totalContributions": 240,
								"weeks": [
									{
										"contributionDays": [
											{"date": "2026-01-01", "contributionCount": 4, "contributionLevel": "FIRST_QUARTILE"}
										]
									}
								]
							}
						}
					}
				}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	authClient := github.NewClient(server.URL, "dummy_token", logger)
	cal, err := authClient.GetContributionCalendar(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cal.TotalContributions != 240 {
		t.Errorf("expected 240 total contributions, got: %d", cal.TotalContributions)
	}
	if len(cal.Weeks) != 1 || len(cal.Weeks[0].ContributionDays) != 1 {
		t.Fatalf("unexpected calendar structure: %+v", cal)
	}
	if cal.Weeks[0].ContributionDays[0].ContributionCount != 4 {
		t.Errorf("expected 4 contributions, got: %d", cal.Weeks[0].ContributionDays[0].ContributionCount)
	}
}
