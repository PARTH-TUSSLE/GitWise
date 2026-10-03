package domain

import (
	"time"

	"github.com/google/uuid"
)

// EdgeType defines the dependency edge relationship classification.
type EdgeType string

const (
	// EdgeTypeImports represents an explicit module or file import statement (deterministic).
	EdgeTypeImports EdgeType = "IMPORTS"
	// EdgeTypeCallsName represents a name-inferred function or method call target (heuristic).
	EdgeTypeCallsName EdgeType = "CALLS_NAME"
	// EdgeTypeImplements represents an interface implementation relationship.
	EdgeTypeImplements EdgeType = "IMPLEMENTS"
)

// DependencyEdge represents a directed edge in the code intelligence graph.
type DependencyEdge struct {
	ID              uuid.UUID  `json:"id"`
	SnapshotID      uuid.UUID  `json:"snapshotId"`
	SourceFileID    uuid.UUID  `json:"sourceFileId"`
	SourceFilePath  string     `json:"sourceFilePath,omitempty"`
	TargetFileID    *uuid.UUID `json:"targetFileId,omitempty"`
	TargetFilePath  *string    `json:"targetFilePath,omitempty"`
	EdgeType        EdgeType   `json:"edgeType"`
	IsDeterministic bool       `json:"isDeterministic"`
	LineNumber      *int       `json:"lineNumber,omitempty"`
	RawTarget       string     `json:"rawTarget,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// CandidateNode represents a node encountered during candidate impact graph traversal.
type CandidateNode struct {
	FileID          uuid.UUID `json:"fileId"`
	Path            string    `json:"path"`
	Depth           int       `json:"depth"`
	EdgeType        EdgeType  `json:"edgeType"`
	IsDeterministic bool      `json:"isDeterministic"`
	Subsystem       string    `json:"subsystem,omitempty"`
}

// CandidateImpactReport models the candidate impact blast radius for a file.
// Architecture Invariant: Downstream candidates are strictly candidate paths (heuristic Level C),
// and never presented as confirmed runtime guarantees.
type CandidateImpactReport struct {
	TargetFile           string          `json:"targetFile"`
	TargetFileID         uuid.UUID       `json:"targetFileId"`
	DirectImports        []string        `json:"directImports"`
	DirectDependents     []string        `json:"directDependents"`
	DownstreamCandidates []CandidateNode `json:"downstreamCandidates"`
	UpstreamDependencies []CandidateNode `json:"upstreamDependencies"`
	AffectedSymbols      []CodeSymbol    `json:"affectedSymbols"`
	Subsystem            string          `json:"subsystem,omitempty"`
	DepthCapped          bool            `json:"depthCapped"`
	MaxDepth             int             `json:"maxDepth"`
	TotalCandidateCount  int             `json:"totalCandidateCount"`
}
