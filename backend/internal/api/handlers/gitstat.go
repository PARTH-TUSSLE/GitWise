package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gitwise/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

type GitStatHandler struct {
	profiles map[string]domain.ContributorProfile
}

func NewGitStatHandler() *GitStatHandler {
	handler := &GitStatHandler{
		profiles: make(map[string]domain.ContributorProfile),
	}
	handler.initMockProfiles()
	return handler
}

func (h *GitStatHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		http.Error(w, `{"error":"username is required"}`, http.StatusBadRequest)
		return
	}

	normUser := strings.ToLower(username)
	profile, exists := h.profiles[normUser]
	if !exists {
		// Dynamic profile fallback with explicit provenance tracking
		profile = domain.ContributorProfile{
			Username:         username,
			Name:             fmt.Sprintf("@%s", username),
			AvatarURL:        fmt.Sprintf("https://github.com/%s.png", username),
			Title:            "Open Source Contributor",
			Bio:              "Autonomous developer contributing to open-source software.",
			Joined:           "2023",
			Status:           "Active Contributor",
			PrimaryLanguages: []string{"Go", "TypeScript", "Python"},
			Metrics: domain.ContributorMetrics{
				MergedPRs:               12,
				OpenPRs:                 2,
				CodeReviewsGiven:        34,
				ReviewCommentVolume:     88,
				IssuesOpened:            8,
				IssuesParticipatedIn:    21,
				IssuesLinkedToMergedPRs: 10,
				ActiveRepositories:      3,
				TotalCommits:            210,
				LinesAdded:              4500,
				LinesDeleted:            1200,
				FilesChanged:            114,
				ReviewTurnaroundHours:   6.5,
				MergeSuccessRatePct:     85.0,
			},
			Repositories: []domain.ContributorRepository{
				{
					Name:        fmt.Sprintf("%s/workspace", username),
					Description: "Developer workspace and active contributions",
					Stars:       15,
					Forks:       4,
					Language:    "Go",
					Commits:     84,
					PRs:         6,
					Role:        "Maintainer",
					EvidenceURL: fmt.Sprintf("https://github.com/%s", username),
				},
			},
			RecentDiffs: []domain.RecentDiff{
				{
					ID:         "d-init",
					Repo:       fmt.Sprintf("%s/workspace", username),
					CommitHash: "a1b2c3d",
					Message:    "Initial repository scaffold and setup",
					Added:      120,
					Deleted:    5,
					Timestamp:  "Just now",
					Type:       "COMMIT",
				},
			},
			ActivityWeeks:  generate52WeeksActivity(1),
			ProvenanceNote: "Synthesized baseline telemetry profile. Live GitHub API sync available in Phase 2.",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(profile)
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
				Date:    fmt.Sprintf("2026-W%02d-%d", w+1, d),
				Level:   level,
				Commits: level * 3 * factor,
				PRs:     map[bool]int{true: 1, false: 0}[level > 2],
				Reviews: map[bool]int{true: 2, false: 0}[level > 1],
			}
		}
		weeks[w] = domain.ActivityWeek{
			Week: fmt.Sprintf("W%d", w+1),
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
			{
				ID:         "d2",
				Repo:       "kubernetes/kubernetes",
				PRNumber:   intPtr(124580),
				CommitHash: "c381da2",
				Message:    "Refactor kube-scheduler pre-filter plugin node score caching to prevent contention",
				Added:      840,
				Deleted:    612,
				Timestamp:  "13:12:44",
				Type:       "COMMIT",
			},
			{
				ID:         "d3",
				Repo:       "tokio-rs/tokio",
				PRNumber:   intPtr(5891),
				CommitHash: "9a21ef4",
				Message:    "Review comment: Validate poll_ready backpressure handling in mpsc channel",
				Added:      0,
				Deleted:    0,
				Timestamp:  "11:05:19",
				Type:       "REVIEW_COMMENT",
			},
			{
				ID:         "d4",
				Repo:       "vercel/next.js",
				PRNumber:   intPtr(54821),
				CommitHash: "3f88be1",
				Message:    "Merge PR #54821: Fix concurrent server action revalidation race in chunked streaming",
				Added:      620,
				Deleted:    140,
				Timestamp:  "09:41:02",
				Type:       "PR_MERGED",
			},
			{
				ID:         "d5",
				Repo:       "kubernetes/kubernetes",
				CommitHash: "e102f9c",
				Message:    "Fix boundary check in chunked kubelet pod status watcher; closes issue #801",
				Added:      45,
				Deleted:    12,
				Timestamp:  "Yesterday",
				Type:       "ISSUE_CLOSED",
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
			{
				Name:        "torvalds/subsurface-for-dirk",
				Description: "Subsurface dive log program",
				Stars:       1200,
				Forks:       230,
				Language:    "C",
				Commits:     1400,
				PRs:         440,
				Role:        "Maintainer",
				EvidenceURL: "https://github.com/torvalds/subsurface-for-dirk",
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
			{
				ID:         "t2",
				Repo:       "torvalds/linux",
				CommitHash: "7b411d9",
				Message:    "Merge branch 'x86/urgent' of git://git.kernel.org/pub/scm/linux/kernel/git/tip/tip",
				Added:      210,
				Deleted:    85,
				Timestamp:  "8h ago",
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
			{
				Name:        "reduxjs/redux",
				Description: "Predictable state container for JavaScript apps",
				Stars:       60400,
				Forks:       15300,
				Language:    "TypeScript",
				Commits:     1100,
				PRs:         420,
				Role:        "Maintainer",
				EvidenceURL: "https://github.com/reduxjs/redux",
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
