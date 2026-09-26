package analysis

import (
	"context"
	"path"
	"strings"
	"sync"

	"github.com/gitwise/backend/internal/domain"
)

// CapabilityLevel defines the structural analysis capability depth.
type CapabilityLevel int

const (
	// Level0Metadata represents file metadata only (size, extension, binary flag).
	Level0Metadata CapabilityLevel = 0
	// Level1Structure represents basic tokenization / top-level structural blocks.
	Level1Structure CapabilityLevel = 1
	// Level2Symbols represents function, method, struct, interface, class, and type declarations + imports.
	Level2Symbols CapabilityLevel = 2
	// Level3References represents call-graph relationships and intra-package references (Phase 5).
	Level3References CapabilityLevel = 3
	// Level4Semantic represents compiler/type-checker exact semantics (Phase 5+).
	Level4Semantic CapabilityLevel = 4
)

// CapabilityReport communicates the exact capabilities and known boundaries of an analyzer.
type CapabilityReport struct {
	Language        string          `json:"language"`
	Level           CapabilityLevel `json:"level"`
	IsDeterministic bool            `json:"isDeterministic"`
	Features        []string        `json:"features"`
	Limitations     []string        `json:"limitations"`
}

// RawSymbol represents an extracted symbol declaration before database persistence.
type RawSymbol struct {
	Name       string
	Kind       domain.SymbolKind
	StartLine  int
	EndLine    int
	Signature  string
	IsExported bool
}

// FileAnalysisResult contains the symbols and structural metadata extracted from a file.
type FileAnalysisResult struct {
	Symbols     []RawSymbol
	Imports     []string
	ParseErrors []string
	Level       CapabilityLevel
}

// LanguageAnalyzer is the core interface for structural code analysis.
type LanguageAnalyzer interface {
	Language() string
	SupportsExtension(ext string) bool
	Capability() CapabilityReport
	AnalyzeFile(ctx context.Context, filePath, content string) (*FileAnalysisResult, error)
}

// Registry manages language analyzers and routes files by extension.
type Registry struct {
	mu        sync.RWMutex
	analyzers []LanguageAnalyzer
}

// NewRegistry initializes an analysis registry with default analyzers.
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make([]LanguageAnalyzer, 0),
	}
}

// Register adds an analyzer to the registry.
func (r *Registry) Register(analyzer LanguageAnalyzer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.analyzers = append(r.analyzers, analyzer)
}

// FindAnalyzer returns the analyzer supporting the given file extension, or nil.
func (r *Registry) FindAnalyzer(filePath string) LanguageAnalyzer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ext := strings.ToLower(path.Ext(filePath))
	for _, a := range r.analyzers {
		if a.SupportsExtension(ext) {
			return a
		}
	}
	return nil
}

// Analyze dispatches file analysis to the registered analyzer, falling back to Level 0 metadata.
func (r *Registry) Analyze(ctx context.Context, filePath, content string) *FileAnalysisResult {
	analyzer := r.FindAnalyzer(filePath)
	if analyzer == nil {
		return &FileAnalysisResult{
			Symbols: nil,
			Imports: nil,
			Level:   Level0Metadata,
		}
	}

	result, err := analyzer.AnalyzeFile(ctx, filePath, content)
	if err != nil {
		return &FileAnalysisResult{
			Symbols:     nil,
			Imports:     nil,
			ParseErrors: []string{err.Error()},
			Level:       Level0Metadata,
		}
	}

	return result
}
