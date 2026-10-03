package domain

import (
	"time"

	"github.com/google/uuid"
)

// AffectedFile represents a file impacted by an issue or blueprint.
type AffectedFile struct {
	Path         string `json:"path"`
	LinesChanged int    `json:"linesChanged"`
	Role         string `json:"role"` // "entry", "core_logic", "type_contract", "test_spec"
	Description  string `json:"description"`
	Snippet      string `json:"snippet,omitempty"`
}

// StepDiff contains contextual code diff changes for an implementation step.
type StepDiff struct {
	Target string `json:"target"`
	Before string `json:"before,omitempty"`
	After  string `json:"after"`
}

// ImplementationStep represents an actionable, sequenced task in the implementation blueprint.
type ImplementationStep struct {
	StepNumber  int      `json:"stepNumber"`
	Title       string   `json:"title"`
	File        string   `json:"file"`
	Action      string   `json:"action"` // "modify", "create", "test"
	Explanation string   `json:"explanation"`
	Diff        StepDiff `json:"diff"`
}

// GroundedTestSuite outlines verification requirements and execution commands.
type GroundedTestSuite struct {
	Name           string `json:"name"`
	Type           string `json:"type"` // "unit", "integration", "e2e"
	TestFile       string `json:"testFile"`
	Command        string `json:"command"`
	ExpectedOutput string `json:"expectedOutput"`
}

// BlastRadius models estimated blast radius and risk assessment for an issue.
type BlastRadius struct {
	Score              float64  `json:"score"` // Scale 1.0 - 5.0
	FileCount          int      `json:"fileCount"`
	SubsystemsAffected []string `json:"subsystemsAffected"`
	RiskAssessment     string   `json:"riskAssessment"`
}

// Stage01Triage captures initial triage diagnostics.
type Stage01Triage struct {
	RootCauseAnalysis string   `json:"rootCauseAnalysis"`
	ReproductionSteps []string `json:"reproductionSteps"`
	ScopeBoundary     string   `json:"scopeBoundary"`
}

// Stage02ImpactedPaths details traversed call graphs and critical symbols.
type Stage02ImpactedPaths struct {
	CallChain       []string `json:"callChain"`
	CriticalSymbols []string `json:"criticalSymbols"`
	StateMutations  string   `json:"stateMutations"`
}

// Stage03Blueprint holds the ordered implementation steps.
type Stage03Blueprint struct {
	Steps []ImplementationStep `json:"steps"`
}

// IssueStages groups the 3-stage intelligence breakdown.
type IssueStages struct {
	Stage01Triage        Stage01Triage        `json:"stage01_triage"`
	Stage02ImpactedPaths Stage02ImpactedPaths `json:"stage02_impacted_paths"`
	Stage03Blueprint     Stage03Blueprint     `json:"stage03_blueprint"`
}

// TestStrategy details test suites and verification checklists.
type TestStrategy struct {
	Suites                []GroundedTestSuite `json:"suites"`
	EdgeCases             []string            `json:"edgeCases"`
	VerificationChecklist []string            `json:"verificationChecklist"`
}

// IssueModel is the complete representation of an issue and its implementation blueprint.
type IssueModel struct {
	ID            string         `json:"id"`
	Number        int            `json:"number"`
	Repo          string         `json:"repo"`
	Title         string         `json:"title"`
	Status        string         `json:"status"` // "open", "in_triage", "investigating"
	Subsystem     string         `json:"subsystem"`
	ReportedDate  string         `json:"reportedDate"`
	Author        string         `json:"author"`
	BlastRadius   BlastRadius    `json:"blastRadius"`
	Summary       string         `json:"summary"`
	Prerequisites []string       `json:"prerequisites"`
	AffectedFiles []AffectedFile `json:"affectedFiles"`
	Stages        IssueStages    `json:"stages"`
	TestStrategy  TestStrategy   `json:"testStrategy"`
}

// IssueRecord represents the database row in the issues table.
type IssueRecord struct {
	ID           uuid.UUID `json:"id"`
	RepositoryID uuid.UUID `json:"repositoryId"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Author       string    `json:"author"`
	State        string    `json:"state"`
	Subsystem    string    `json:"subsystem"`
	ReportedDate time.Time `json:"reportedDate"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CreateBlueprintRequest models the payload for generating or re-generating a blueprint.
type CreateBlueprintRequest struct {
	SnapshotID *uuid.UUID `json:"snapshotId,omitempty"`
	Ref        string     `json:"ref,omitempty"`
	UseAI      bool       `json:"useAi,omitempty"`
	Provider   string     `json:"provider,omitempty"`
	Model      string     `json:"model,omitempty"`
}
