package domain

import (
	"time"

	"github.com/google/uuid"
)

type SnapshotStatus string

const (
	SnapshotStatusQueued     SnapshotStatus = "QUEUED"
	SnapshotStatusProcessing SnapshotStatus = "PROCESSING"
	SnapshotStatusReady      SnapshotStatus = "READY"
	SnapshotStatusPartial    SnapshotStatus = "PARTIAL"
	SnapshotStatusFailed     SnapshotStatus = "FAILED"
)

type Repository struct {
	ID            uuid.UUID `json:"id"`
	GitHubID      int64     `json:"githubId"`
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	DefaultBranch string    `json:"defaultBranch"`
	IsPrivate     bool      `json:"isPrivate"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type RepositorySnapshot struct {
	ID              uuid.UUID      `json:"id"`
	RepositoryID    uuid.UUID      `json:"repositoryId"`
	CommitSHA       string         `json:"commitSha"`
	RefName         string         `json:"refName"`
	Status          SnapshotStatus `json:"status"`
	TotalFiles      int            `json:"totalFiles"`
	TotalLines      int            `json:"totalLines"`
	PrimaryLanguage string         `json:"primaryLanguage"`
	AnalyzedAt      *time.Time     `json:"analyzedAt,omitempty"`
	ExpiresAt       *time.Time     `json:"expiresAt,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
}
