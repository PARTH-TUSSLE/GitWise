package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/github"
	"github.com/gitwise/backend/internal/service/gitstat"
	"github.com/go-chi/chi/v5"
)

type GitStatHandler struct {
	service  *gitstat.Service
	logger   *slog.Logger
	profiles map[string]domain.ContributorProfile
}

func NewGitStatHandler(service *gitstat.Service, logger *slog.Logger) *GitStatHandler {
	if logger == nil {
		logger = slog.Default()
	}
	handler := &GitStatHandler{
		service:  service,
		logger:   logger,
		profiles: make(map[string]domain.ContributorProfile),
	}
	handler.initMockProfiles()
	return handler
}

func (h *GitStatHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(chi.URLParam(r, "username"))
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "bad_request",
			"message": "Username parameter is required",
		})
		return
	}

	normUser := strings.ToLower(username)
	refresh := r.URL.Query().Get("refresh") == "true"

	// 1. LIVE SERVICE PATH:
	// Strictly queries live GitHub client & cache. Never substitutes fabricated or mock data on error.
	if h.service != nil {
		profile, err := h.service.GetProfile(r.Context(), username, refresh)
		if err == nil && profile != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(profile)
			return
		}

		// GitHub user missing -> 404
		if github.IsNotFound(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "GitHub user not found",
			})
			return
		}

		// GitHub rate limited -> 429
		if github.IsRateLimit(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "GitHub API rate limit reached. Please retry later or configure authentication.",
			})
			return
		}

		// Other upstream error: log full detail internally with context, return sanitized message to client
		h.logger.ErrorContext(r.Context(), "failed to retrieve gitstat profile from upstream",
			slog.String("username", username),
			slog.String("error", err.Error()),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "upstream_error",
			"message": "Failed to retrieve developer telemetry from upstream service",
		})
		return
	}

	// 2. MOCK MODE / DEMO PATH:
	// Only returns explicitly predefined mock fixtures (alexR_dev, torvalds, gaearon).
	// Arbitrary usernames NEVER receive generated fake metrics.
	mock, exists := h.profiles[normUser]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "not_found",
			"message": "Contributor profile not found in mock fixtures",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mock)
}

func generate52WeeksActivity(factor int) []domain.ActivityWeek {
	weeks := make([]domain.ActivityWeek, 52)
	for w := 0; w < 52; w++ {
		days := make([]domain.ActivityDay, 7)
		for d := 0; d < 7; d++ {
			seed := (w*7 + d*13) % 100
			level := 0
			if d == 0 || d == 6 {
				if seed > 80 {
					level = 1
				}
			} else if seed > 85 {
				level = 4
			} else if seed > 65 {
				level = 3
			} else if seed > 40 {
				level = 2
			} else if seed > 20 {
				level = 1
			}

			days[d] = domain.ActivityDay{
				Date:    "2026-01-01",
				Level:   level,
				Commits: level * 3 * factor,
				PRs:     map[bool]int{true: 1, false: 0}[level > 2],
				Reviews: map[bool]int{true: 2, false: 0}[level > 1],
			}
		}
		weeks[w] = domain.ActivityWeek{
			Week: "W",
			Days: days,
		}
	}
	return weeks
}

func (h *GitStatHandler) initMockProfiles() {
	// 1. alexr_dev
	h.profiles["alexr_dev"] = domain.ContributorProfile{
		Username:         "alexR_dev",
		Name:             "Alex Rivera",
		AvatarURL:        "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=256&q=80",
		Title:            "Senior Infrastructure & Systems Engineer",
		Bio:              "Building developer tooling, distributed consensus, and platform infrastructure.",
		Joined:           "2018",
		Status:           "Active Contributor",
		PrimaryLanguages: []string{"Rust", "TypeScript", "Go", "C++"},
		Metrics: domain.ContributorMetrics{
			MergedPRs:               148,
			OpenPRs:                 6,
			CodeReviewsGiven:        291,
			ReviewCommentVolume:     842,
			IssuesOpened:            63,
			IssuesParticipatedIn:    117,
			IssuesLinkedToMergedPRs: 102,
			ActiveRepositories:      18,
			TotalCommits:            4112,
			LinesAdded:              54200,
			LinesDeleted:            31700,
			FilesChanged:            1847,
			ReviewTurnaroundHours:   4.2,
			MergeSuccessRatePct:     89.5,
		},
		Repositories: []domain.ContributorRepository{
			{
				Name:        "vercel/next.js",
				Description: "The React Framework for the Web with App Router and Turbopack compiler",
				Stars:       122000,
				Forks:       26400,
				Language:    "TypeScript",
				Commits:     542,
				PRs:         38,
				Role:        "Core Contributor",
				EvidenceURL: "https://github.com/vercel/next.js",
			},
			{
				Name:        "kubernetes/kubernetes",
				Description: "Production-Grade Container Scheduling and Automated Workload Orchestration",
				Stars:       108000,
				Forks:       39100,
				Language:    "Go",
				Commits:     418,
				PRs:         44,
				Role:        "External Contributor",
				EvidenceURL: "https://github.com/kubernetes/kubernetes",
			},
			{
				Name:        "tokio-rs/tokio",
				Description: "A runtime for writing reliable, asynchronous, and slim applications with Rust",
				Stars:       25400,
				Forks:       2400,
				Language:    "Rust",
				Commits:     34,
				PRs:         12,
				Role:        "External Contributor",
				EvidenceURL: "https://github.com/tokio-rs/tokio",
			},
			{
				Name:        "facebook/react",
				Description: "The library for web and native user interfaces",
				Stars:       228000,
				Forks:       46200,
				Language:    "TypeScript",
				Commits:     18,
				PRs:         6,
				Role:        "External Contributor",
				EvidenceURL: "https://github.com/facebook/react",
			},
		},
		RecentDiffs: []domain.RecentDiff{
			{
				ID:         "d1",
				Repo:       "vercel/next.js",
				PRNumber:   intPtr(62145),
				CommitHash: "7b89f0a",
				Message:    "Merge PR #62145: Add Flight stream backpressure drain controller for server actions",
				Added:      1420,
				Deleted:    380,
				Timestamp:  "14:38:05",
				Type:       "PR_MERGED",
			},
		},
		ActivityWeeks:  generate52WeeksActivity(1),
		ProvenanceNote: "Curated developer profile with complete GitWise mock parity.",
	}

	// 2. torvalds
	h.profiles["torvalds"] = domain.ContributorProfile{
		Username:         "torvalds",
		Name:             "Linus Torvalds",
		AvatarURL:        "https://avatars.githubusercontent.com/u/1024025?v=4",
		Title:            "Creator of Linux and Git",
		Bio:              "Software developer. Creator of Linux and Git. Linux kernel maintainer.",
		Joined:           "2011",
		Status:           "Kernel Maintainer",
		PrimaryLanguages: []string{"C", "Shell", "Makefile"},
		Metrics: domain.ContributorMetrics{
			MergedPRs:               2840,
			OpenPRs:                 14,
			CodeReviewsGiven:        14200,
			ReviewCommentVolume:     38900,
			IssuesOpened:            18,
			IssuesParticipatedIn:    4890,
			IssuesLinkedToMergedPRs: 2100,
			ActiveRepositories:      4,
			TotalCommits:            31200,
			LinesAdded:              890400,
			LinesDeleted:            640200,
			FilesChanged:            28400,
			ReviewTurnaroundHours:   1.8,
			MergeSuccessRatePct:     98.2,
		},
		Repositories: []domain.ContributorRepository{
			{
				Name:        "torvalds/linux",
				Description: "Linux kernel source tree",
				Stars:       182000,
				Forks:       54000,
				Language:    "C",
				Commits:     29800,
				PRs:         2400,
				Role:        "Maintainer",
				EvidenceURL: "https://github.com/torvalds/linux",
			},
		},
		RecentDiffs: []domain.RecentDiff{
			{
				ID:         "t1",
				Repo:       "torvalds/linux",
				CommitHash: "9a01f42",
				Message:    "Linux 6.12-rc7 release tag and merge window synchronization",
				Added:      8420,
				Deleted:    5120,
				Timestamp:  "3h ago",
				Type:       "PR_MERGED",
			},
		},
		ActivityWeeks:  generate52WeeksActivity(4),
		ProvenanceNote: "Curated developer profile with complete GitWise mock parity.",
	}

	// 3. gaearon
	h.profiles["gaearon"] = domain.ContributorProfile{
		Username:         "gaearon",
		Name:             "Dan Abramov",
		AvatarURL:        "https://avatars.githubusercontent.com/u/810438?v=4",
		Title:            "Software Engineer & Co-author of Redux",
		Bio:              "Working on React, Redux, and open-source JavaScript architecture.",
		Joined:           "2011",
		Status:           "Core Maintainer",
		PrimaryLanguages: []string{"JavaScript", "TypeScript", "CSS"},
		Metrics: domain.ContributorMetrics{
			MergedPRs:               1840,
			OpenPRs:                 22,
			CodeReviewsGiven:        3910,
			ReviewCommentVolume:     12400,
			IssuesOpened:            340,
			IssuesParticipatedIn:    2180,
			IssuesLinkedToMergedPRs: 1420,
			ActiveRepositories:      42,
			TotalCommits:            8490,
			LinesAdded:              210400,
			LinesDeleted:            148900,
			FilesChanged:            7200,
			ReviewTurnaroundHours:   3.4,
			MergeSuccessRatePct:     94.1,
		},
		Repositories: []domain.ContributorRepository{
			{
				Name:        "facebook/react",
				Description: "The library for web and native user interfaces",
				Stars:       228000,
				Forks:       46200,
				Language:    "JavaScript",
				Commits:     2410,
				PRs:         840,
				Role:        "Core Contributor",
				EvidenceURL: "https://github.com/facebook/react",
			},
		},
		RecentDiffs: []domain.RecentDiff{
			{
				ID:         "g1",
				Repo:       "facebook/react",
				CommitHash: "e441da0",
				Message:    "Refactor Suspense error boundary reconciliation priority",
				Added:      412,
				Deleted:    198,
				Timestamp:  "5h ago",
				Type:       "PR_MERGED",
			},
		},
		ActivityWeeks:  generate52WeeksActivity(2),
		ProvenanceNote: "Curated developer profile with complete GitWise mock parity.",
	}
}

func intPtr(i int) *int {
	return &i
}
