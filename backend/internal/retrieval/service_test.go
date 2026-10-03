package retrieval_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/retrieval"
	"github.com/google/uuid"
)

func TestService_PersistChunks(t *testing.T) {
	db, err := sql.Open("fake_retrieval_driver", "test_persist_chunks")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	svc := retrieval.NewService(db, nil, nil)
	snapID := uuid.New()
	fileID := uuid.New()

	chunks := []domain.CodeChunk{
		{
			ID:         uuid.New(),
			SnapshotID: snapID,
			FileID:     fileID,
			StartLine:  1,
			EndLine:    10,
			Scope:      "package main",
			Content:    "package main\n\nfunc main() {}",
		},
	}

	ctx := context.Background()
	err = svc.PersistChunks(ctx, snapID, chunks)
	if err != nil {
		t.Fatalf("unexpected persist error: %v", err)
	}
}

func TestService_HybridSearch_SymbolMatchPrecedence(t *testing.T) {
	db, err := sql.Open("fake_retrieval_driver", "test_hybrid_search")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	snapID := uuid.New()
	fileID := uuid.New()
	symChunkID := uuid.New()
	symID := uuid.New()
	ftsChunkID := uuid.New()

	testRetrievalDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		// 1. Tier 1: Symbol Match query
		if strings.Contains(query, "WHERE c.snapshot_id = $1") && strings.Contains(query, "LOWER(cs.name)") {
			return newRows(
				[]string{
					"id", "snapshot_id", "file_id", "path", "symbol_id", "name",
					"start_line", "end_line", "scope", "content", "embedding_model", "created_at",
				},
				[][]driver.Value{
					{
						symChunkID.String(), snapID.String(), fileID.String(), "auth/service.go",
						symID.String(), "ValidateToken", int64(10), int64(25), "FUNCTION ValidateToken",
						"func ValidateToken(tok string) bool { return true }", "mock-768", time.Now(),
					},
				},
			), nil, nil
		}

		// 2. Tier 2: FTS query
		if strings.Contains(query, "content_tsv @@ plainto_tsquery") {
			return newRows(
				[]string{
					"id", "snapshot_id", "file_id", "path", "symbol_id", "name",
					"start_line", "end_line", "scope", "content", "embedding_model", "created_at",
					"lexical_score",
				},
				[][]driver.Value{
					{
						ftsChunkID.String(), snapID.String(), fileID.String(), "auth/helpers.go",
						"", "", int64(30), int64(45), "helpers",
						"// Helper for token parsing", "mock-768", time.Now(),
						float64(0.85),
					},
				},
			), nil, nil
		}

		// 3. Tier 3: Vector query
		if strings.Contains(query, "c.embedding <=> $2::vector") {
			return newRows(
				[]string{
					"id", "snapshot_id", "file_id", "path", "symbol_id", "name",
					"start_line", "end_line", "scope", "content", "embedding_model", "created_at",
					"cosine_distance",
				},
				[][]driver.Value{
					{
						ftsChunkID.String(), snapID.String(), fileID.String(), "auth/helpers.go",
						"", "", int64(30), int64(45), "helpers",
						"// Helper for token parsing", "mock-768", time.Now(),
						float64(0.25),
					},
				},
			), nil, nil
		}

		return newRows([]string{}, nil), nil, nil
	}

	svc := retrieval.NewService(db, nil, nil)
	ctx := context.Background()

	results, err := svc.HybridSearch(ctx, snapID, "ValidateToken", 5)
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results, got 0")
	}

	// Symbol exact match must be rank 1 with score 1.0 and tier SYMBOL_EXACT
	if results[0].Chunk.ID != symChunkID {
		t.Errorf("expected chunk ID %s, got %s", symChunkID, results[0].Chunk.ID)
	}
	if results[0].Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", results[0].Score)
	}
	if results[0].RankTier != "SYMBOL_EXACT" {
		t.Errorf("expected rank tier SYMBOL_EXACT, got %s", results[0].RankTier)
	}
	if !results[0].SymbolMatch {
		t.Error("expected SymbolMatch true")
	}
	if results[0].Chunk.SymbolName != "ValidateToken" {
		t.Errorf("expected symbol ValidateToken, got %s", results[0].Chunk.SymbolName)
	}

	// Second result should be the fused FTS/Vector hit
	if len(results) > 1 {
		if results[1].Chunk.ID != ftsChunkID {
			t.Errorf("expected fused chunk ID %s, got %s", ftsChunkID, results[1].Chunk.ID)
		}
		if results[1].SymbolMatch {
			t.Error("expected SymbolMatch false for fused hit")
		}
		if results[1].Score >= 1.0 {
			t.Errorf("expected score < 1.0 for fused hit, got %f", results[1].Score)
		}
	}
}

func TestService_HybridSearch_EmptyQuery(t *testing.T) {
	db, err := sql.Open("fake_retrieval_driver", "test_empty_query")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	svc := retrieval.NewService(db, nil, nil)
	ctx := context.Background()

	results, err := svc.HybridSearch(ctx, uuid.New(), "   ", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty query, got %v", results)
	}
}

func BenchmarkHybridRetrieval(b *testing.B) {
	db, err := sql.Open("fake_retrieval_driver", "bench_retrieval")
	if err != nil {
		b.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	snapID := uuid.New()
	fileID := uuid.New()
	symChunkID := uuid.New()
	ftsChunkID := uuid.New()

	testRetrievalDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "WHERE c.snapshot_id = $1") && strings.Contains(query, "LOWER(cs.name)") {
			return newRows(
				[]string{
					"id", "snapshot_id", "file_id", "path", "symbol_id", "name",
					"start_line", "end_line", "scope", "content", "embedding_model", "created_at",
				},
				[][]driver.Value{
					{
						symChunkID.String(), snapID.String(), fileID.String(), "auth/service.go",
						uuid.New().String(), "ValidateToken", int64(10), int64(25), "FUNCTION ValidateToken",
						"func ValidateToken(tok string) bool { return true }", "mock-768", time.Now(),
					},
				},
			), nil, nil
		}
		if strings.Contains(query, "content_tsv @@ plainto_tsquery") {
			return newRows(
				[]string{
					"id", "snapshot_id", "file_id", "path", "symbol_id", "name",
					"start_line", "end_line", "scope", "content", "embedding_model", "created_at",
					"lexical_score",
				},
				[][]driver.Value{
					{
						ftsChunkID.String(), snapID.String(), fileID.String(), "auth/helpers.go",
						"", "", int64(30), int64(45), "helpers",
						"// Helper for token parsing", "mock-768", time.Now(),
						float64(0.85),
					},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	svc := retrieval.NewService(db, nil, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := svc.HybridSearch(ctx, snapID, "ValidateToken", 5)
		if err != nil {
			b.Fatalf("search error: %v", err)
		}
		if len(results) == 0 || results[0].Chunk.ID != symChunkID {
			b.Fatalf("benchmark failed: exact symbol did not outrank other chunks")
		}
	}
}
