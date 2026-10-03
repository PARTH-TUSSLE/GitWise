package domain

import (
	"time"

	"github.com/google/uuid"
)

// EvidenceRef models a lightweight anchor referencing an exact code slice.
// Storage Axiom: Stores byte/line offsets and content SHA-256 hash without duplicating raw code in evidence_refs.
type EvidenceRef struct {
	ID          string    `json:"id"` // e.g. "ev_01", "ev_a1b2c3"
	SnapshotID  uuid.UUID `json:"snapshotId"`
	FileID      uuid.UUID `json:"fileId"`
	FilePath    string    `json:"filePath"`
	StartLine   int       `json:"startLine"`
	EndLine     int       `json:"endLine"`
	ContentHash string    `json:"contentHash"`
	Provenance  string    `json:"provenance"`
	Snippet     string    `json:"snippet,omitempty"` // Hydrated dynamically on read
	CreatedAt   time.Time `json:"createdAt"`
}

// EvidencePackage is the assembled grounded evidence container passed to the LLM.
type EvidencePackage struct {
	SnapshotID uuid.UUID     `json:"snapshotId"`
	CommitSHA  string        `json:"commitSha"`
	Query      string        `json:"query"`
	Items      []EvidenceRef `json:"items"`
	TotalItems int           `json:"totalItems"`
}
