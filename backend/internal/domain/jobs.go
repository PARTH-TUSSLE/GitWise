package domain

import (
	"time"

	"github.com/google/uuid"
)

type JobType string

const (
	JobTypeSnapshotIngest JobType = "SNAPSHOT_INGEST"
	JobTypeStaticAnalysis JobType = "STATIC_ANALYSIS"
	JobTypeVectorIndex    JobType = "VECTOR_INDEX"
	JobTypeTelemetrySync  JobType = "TELEMETRY_SYNC"
)

type JobStatus string

const (
	JobStatusQueued     JobStatus = "QUEUED"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

type JobStage string

const (
	JobStageInitializing   JobStage = "INITIALIZING"
	JobStageFetchingTree   JobStage = "FETCHING_TREE"
	JobStageFilteringFiles JobStage = "FILTERING_FILES"
	JobStageAnalyzingAST   JobStage = "ANALYZING_AST"
	JobStageParsingAST     JobStage = "PARSING_AST"
	JobStageBuildingGraph  JobStage = "BUILDING_GRAPH"
	JobStageChunking       JobStage = "CHUNKING"
	JobStageEmbedding      JobStage = "EMBEDDING"
	JobStageFinalizing     JobStage = "FINALIZING"
	JobStageDone           JobStage = "DONE"
)

type AnalysisJob struct {
	ID              uuid.UUID  `json:"id"`
	Type            JobType    `json:"type"`
	SnapshotID      *uuid.UUID `json:"snapshotId,omitempty"`
	Status          JobStatus  `json:"status"`
	Stage           JobStage   `json:"stage"`
	ProgressPercent float64    `json:"progressPercent"`
	ErrorMessage    *string    `json:"errorMessage,omitempty"`
	RetryCount      int        `json:"retryCount"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}
