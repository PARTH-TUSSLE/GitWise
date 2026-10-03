package domain

import "time"

// DiffSnippet models the focused code before/after hunk for a file.
type DiffSnippet struct {
	Target string `json:"target"`
	Before string `json:"before,omitempty"`
	After  string `json:"after"`
}

// PRFileDiff captures file-level diff and behavioral analysis.
type PRFileDiff struct {
	Path              string      `json:"path"`
	Additions         int         `json:"additions"`
	Deletions         int         `json:"deletions"`
	Status            string      `json:"status"` // "modified", "added", "deleted"
	Subsystem         string      `json:"subsystem"`
	BehavioralSummary string      `json:"behavioralSummary"`
	DiffSnippet       DiffSnippet `json:"diffSnippet"`
}

// SubsystemCoupling captures downstream ripple effects on another subsystem.
type SubsystemCoupling struct {
	Name     string `json:"name"`
	Coupling string `json:"coupling"` // "tight", "loose"
	Impact   string `json:"impact"`
}

// ArchitecturalShift models high-level architectural and contract mutations.
type ArchitecturalShift struct {
	PublicAPIContract    string              `json:"publicApiContract"` // "unchanged", "extended", "breaking"
	PublicAPIExplanation string              `json:"publicApiExplanation"`
	DownstreamSubsystems []SubsystemCoupling `json:"downstreamSubsystems"`
	ExecutionFlowDelta   string              `json:"executionFlowDelta"`
	StateMutationRisk    string              `json:"stateMutationRisk"` // "none", "low", "moderate", "high"
	RiskExplanation      string              `json:"riskExplanation"`
}

// ReviewFinding represents automated linter, regression, convention, or security advice.
type ReviewFinding struct {
	RuleID     string `json:"ruleId"`
	Severity   string `json:"severity"` // "clean", "advisory", "warning"
	Category   string `json:"category"` // "convention", "regression", "performance", "coverage"
	File       string `json:"file"`
	Line       int    `json:"line"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// PRStats summarizes quantitative patch metrics.
type PRStats struct {
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
	FilesChanged int `json:"filesChanged"`
	CommitsCount int `json:"commitsCount"`
}

// PRTestVerification describes regression verification commands and expectations.
type PRTestVerification struct {
	Command        string `json:"command"`
	TargetSuite    string `json:"targetSuite"`
	ExpectedResult string `json:"expectedResult"`
	CoverageDelta  string `json:"coverageDelta"`
}

// PullRequestModel is the complete representation of a PR review.
type PullRequestModel struct {
	ID                 string             `json:"id"`
	Number             int                `json:"number"`
	Repo               string             `json:"repo"`
	Title              string             `json:"title"`
	Author             string             `json:"author"`
	Status             string             `json:"status"` // "ready_for_review", "in_review", "approved", "changes_requested"
	CreatedAt          string             `json:"createdAt"`
	Stats              PRStats            `json:"stats"`
	Summary            string             `json:"summary"`
	ArchitecturalShift ArchitecturalShift `json:"architecturalShift"`
	Files              []PRFileDiff       `json:"files"`
	ReviewFindings     []ReviewFinding    `json:"reviewFindings"`
	TestVerification   PRTestVerification `json:"testVerification"`
}

// ReviewPullRequestRequest models parameters for reviewing a PR.
type ReviewPullRequestRequest struct {
	Patch    string `json:"patch,omitempty"`
	UseAI    bool   `json:"useAi,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

// PullRequestRecord models a row in the pull_requests database table.
type PullRequestRecord struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repositoryId"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Author       string    `json:"author"`
	Status       string    `json:"status"`
	BaseRef      string    `json:"baseRef"`
	HeadRef      string    `json:"headRef"`
	HeadSHA      string    `json:"headSha"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
