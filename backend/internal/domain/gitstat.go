package domain

import "time"

// Metric wraps any calculated metric with explicit provenance and sample bounds.
// In accordance with GitWise principles: never imply lifetime completeness
// unless the underlying data actually supports it.
type Metric[T any] struct {
	Value         T         `json:"value"`
	Provenance    string    `json:"provenance"`    // e.g. "github.rest.pull_requests.merged"
	SampleSize    int       `json:"sampleSize"`    // e.g. 100
	CoverageLimit string    `json:"coverageLimit"` // e.g. "Most recent 100 merged PRs across accessible public repos"
	ComputedAt    time.Time `json:"computedAt"`
}

type ContributorMetrics struct {
	MergedPRs               int     `json:"mergedPRs"`
	OpenPRs                 int     `json:"openPRs"`
	CodeReviewsGiven        int     `json:"codeReviewsGiven"`
	ReviewCommentVolume     int     `json:"reviewCommentVolume"`
	IssuesOpened            int     `json:"issuesOpened"`
	IssuesParticipatedIn    int     `json:"issuesParticipatedIn"`
	IssuesLinkedToMergedPRs int     `json:"issuesLinkedToMergedPRs"`
	ActiveRepositories      int     `json:"activeRepositories"`
	TotalCommits            int     `json:"totalCommits"`
	LinesAdded              int     `json:"linesAdded"`
	LinesDeleted            int     `json:"linesDeleted"`
	FilesChanged            int     `json:"filesChanged"`
	ReviewTurnaroundHours   float64 `json:"reviewTurnaroundHours"`
	MergeSuccessRatePct     float64 `json:"mergeSuccessRatePct"`
}

type ContributorRepository struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
	Forks       int    `json:"forks"`
	Language    string `json:"language"`
	Commits     int    `json:"commits"`
	PRs         int    `json:"prs"`
	Role        string `json:"role"` // Maintainer | Core Contributor | External Contributor
	EvidenceURL string `json:"evidenceUrl"`
}

type RecentDiff struct {
	ID         string `json:"id"`
	Repo       string `json:"repo"`
	PRNumber   *int   `json:"prNumber,omitempty"`
	CommitHash string `json:"commitHash"`
	Message    string `json:"message"`
	Added      int    `json:"added"`
	Deleted    int    `json:"deleted"`
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"` // PR_MERGED | COMMIT | REVIEW_COMMENT | ISSUE_CLOSED
}

type ActivityDay struct {
	Date    string `json:"date"`
	Level   int    `json:"level"` // 0 | 1 | 2 | 3 | 4
	Commits int    `json:"commits"`
	PRs     int    `json:"prs"`
	Reviews int    `json:"reviews"`
}

type ActivityWeek struct {
	Week string        `json:"week"`
	Days []ActivityDay `json:"days"`
}

type ContributorProfile struct {
	Username         string                  `json:"username"`
	Name             string                  `json:"name"`
	AvatarURL        string                  `json:"avatarUrl"`
	Title            string                  `json:"title"`
	Bio              string                  `json:"bio"`
	Joined           string                  `json:"joined"`
	Status           string                  `json:"status"`
	PrimaryLanguages []string                `json:"primaryLanguages"`
	Metrics          ContributorMetrics      `json:"metrics"`
	Repositories     []ContributorRepository `json:"repositories"`
	RecentDiffs      []RecentDiff            `json:"recentDiffs"`
	ActivityWeeks    []ActivityWeek          `json:"activityWeeks"`
	ProvenanceNote   string                  `json:"provenanceNote,omitempty"`
}
