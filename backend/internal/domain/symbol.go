package domain

import (
	"time"

	"github.com/google/uuid"
)

// SymbolKind represents the syntactic classification of a code symbol.
type SymbolKind string

const (
	SymbolKindFunction  SymbolKind = "FUNCTION"
	SymbolKindMethod    SymbolKind = "METHOD"
	SymbolKindStruct    SymbolKind = "STRUCT"
	SymbolKindInterface SymbolKind = "INTERFACE"
	SymbolKindClass     SymbolKind = "CLASS"
	SymbolKindType      SymbolKind = "TYPE"
)

// CodeSymbol represents an individual symbol declaration extracted from a repository file.
// Invariant: Every symbol is strictly tied to an exact snapshot_id and file_id.
type CodeSymbol struct {
	ID         uuid.UUID  `json:"id"`
	SnapshotID uuid.UUID  `json:"snapshotId"`
	FileID     uuid.UUID  `json:"fileId"`
	FilePath   string     `json:"filePath,omitempty"`
	Name       string     `json:"name"`
	Kind       SymbolKind `json:"kind"`
	StartLine  int        `json:"startLine"`
	EndLine    int        `json:"endLine"`
	Signature  string     `json:"signature"`
	IsExported bool       `json:"isExported"`
	CreatedAt  time.Time  `json:"createdAt"`
}
