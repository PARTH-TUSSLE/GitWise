package domain

import (
	"time"

	"github.com/google/uuid"
)

// CodeChunk models a semantically bounded code slice for vector & full-text search.
type CodeChunk struct {
	ID             uuid.UUID  `json:"id"`
	SnapshotID     uuid.UUID  `json:"snapshotId"`
	FileID         uuid.UUID  `json:"fileId"`
	FilePath       string     `json:"filePath,omitempty"`
	SymbolID       *uuid.UUID `json:"symbolId,omitempty"`
	SymbolName     string     `json:"symbolName,omitempty"`
	StartLine      int        `json:"startLine"`
	EndLine        int        `json:"endLine"`
	Scope          string     `json:"scope,omitempty"`
	Content        string     `json:"content"`
	Embedding      []float32  `json:"embedding,omitempty"`
	EmbeddingModel string     `json:"embeddingModel"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// SearchResult models a hybrid retrieval hit with tiered ranking metadata.
type SearchResult struct {
	Chunk          CodeChunk `json:"chunk"`
	Score          float64   `json:"score"`
	RankTier       string    `json:"rankTier"` // "SYMBOL_EXACT", "SYMBOL_PREFIX", "FTS_TEXT", "VECTOR_SEMANTIC"
	SymbolMatch    bool      `json:"symbolMatch"`
	LexicalRank    float64   `json:"lexicalRank,omitempty"`
	VectorDistance float64   `json:"vectorDistance,omitempty"`
}
