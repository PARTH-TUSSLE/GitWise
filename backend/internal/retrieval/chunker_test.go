package retrieval_test

import (
	"strings"
	"testing"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/retrieval"
	"github.com/google/uuid"
)

func TestChunker_ASTDeclarationBoundaries(t *testing.T) {
	chunker := retrieval.NewChunker()

	code := `package math

// Add adds two integers
func Add(a, b int) int {
	return a + b
}

// Multiply multiplies two integers
func Multiply(a, b int) int {
	return a * b
}
`
	fileID := uuid.New()
	snapshotID := uuid.New()
	sym1ID := uuid.New()
	sym2ID := uuid.New()

	file := domain.RepositoryFile{
		ID:         fileID,
		SnapshotID: snapshotID,
		Path:       "math.go",
		Content:    &code,
	}

	symbols := []domain.CodeSymbol{
		{
			ID:        sym1ID,
			FileID:    fileID,
			Name:      "Add",
			Kind:      "FUNCTION",
			StartLine: 4,
			EndLine:   6,
		},
		{
			ID:        sym2ID,
			FileID:    fileID,
			Name:      "Multiply",
			Kind:      "FUNCTION",
			StartLine: 9,
			EndLine:   11,
		},
	}

	chunks := chunker.ChunkFile(file, symbols)
	if len(chunks) == 0 {
		t.Fatal("expected chunks, got 0")
	}

	// Check that we have chunks for preamble, Add, middle comment, Multiply
	var foundAdd, foundMultiply bool
	for _, c := range chunks {
		if c.SnapshotID != snapshotID {
			t.Errorf("expected snapshot ID %s, got %s", snapshotID, c.SnapshotID)
		}
		if c.FileID != fileID {
			t.Errorf("expected file ID %s, got %s", fileID, c.FileID)
		}
		if c.FilePath != "math.go" {
			t.Errorf("expected file path math.go, got %s", c.FilePath)
		}

		if c.SymbolID != nil && *c.SymbolID == sym1ID {
			foundAdd = true
			if c.SymbolName != "Add" {
				t.Errorf("expected symbol name Add, got %s", c.SymbolName)
			}
			if c.StartLine != 4 || c.EndLine != 6 {
				t.Errorf("expected lines 4-6, got %d-%d", c.StartLine, c.EndLine)
			}
			if !strings.Contains(c.Content, "func Add(a, b int) int") {
				t.Errorf("expected Add func content, got %s", c.Content)
			}
		}
		if c.SymbolID != nil && *c.SymbolID == sym2ID {
			foundMultiply = true
			if c.SymbolName != "Multiply" {
				t.Errorf("expected symbol name Multiply, got %s", c.SymbolName)
			}
			if c.StartLine != 9 || c.EndLine != 11 {
				t.Errorf("expected lines 9-11, got %d-%d", c.StartLine, c.EndLine)
			}
			if !strings.Contains(c.Content, "func Multiply(a, b int) int") {
				t.Errorf("expected Multiply func content, got %s", c.Content)
			}
		}
	}

	if !foundAdd {
		t.Error("expected Add function chunk")
	}
	if !foundMultiply {
		t.Error("expected Multiply function chunk")
	}
}

func TestChunker_FallbackWindowed(t *testing.T) {
	chunker := retrieval.NewChunker()

	content := "# Configuration File\nkey1=value1\nkey2=value2\nkey3=value3\n"
	fileID := uuid.New()
	snapshotID := uuid.New()

	file := domain.RepositoryFile{
		ID:         fileID,
		SnapshotID: snapshotID,
		Path:       "config.env",
		Content:    &content,
	}

	chunks := chunker.ChunkFile(file, nil)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Scope != "config.env" {
		t.Errorf("expected scope config.env, got %s", chunks[0].Scope)
	}
	if chunks[0].StartLine != 1 || chunks[0].EndLine != 4 {
		t.Errorf("expected lines 1-4, got %d-%d", chunks[0].StartLine, chunks[0].EndLine)
	}
	if !strings.Contains(chunks[0].Content, "key1=value1") {
		t.Errorf("expected content to contain key1=value1, got %s", chunks[0].Content)
	}
}

func TestChunker_EmptyFile(t *testing.T) {
	chunker := retrieval.NewChunker()

	empty := ""
	file := domain.RepositoryFile{
		ID:         uuid.New(),
		SnapshotID: uuid.New(),
		Path:       "empty.go",
		Content:    &empty,
	}

	chunks := chunker.ChunkFile(file, nil)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty string, got %d", len(chunks))
	}

	file.Content = nil
	chunks = chunker.ChunkFile(file, nil)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for nil content, got %d", len(chunks))
	}
}
