package evidence_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/evidence"
	"github.com/google/uuid"
)

func TestExtractSnippet_And_ComputeHash(t *testing.T) {
	content := "line 1\nline 2\nline 3\nline 4\nline 5\n"

	snippet, err := evidence.ExtractSnippet(content, 2, 4)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}
	expected := "line 2\nline 3\nline 4"
	if snippet != expected {
		t.Errorf("expected %q, got %q", expected, snippet)
	}

	hash := evidence.ComputeHash(snippet)
	if len(hash) != 64 {
		t.Errorf("expected 64 char hash, got %d", len(hash))
	}
	if hash != evidence.ComputeHash(snippet) {
		t.Error("expected deterministic hash")
	}

	// Edge case: start exceeds total lines
	_, err = evidence.ExtractSnippet(content, 10, 15)
	if err == nil {
		t.Error("expected error when start line exceeds total lines")
	}

	// Edge case: end exceeds total lines (clamped to max)
	snippetClamped, err := evidence.ExtractSnippet(content, 4, 20)
	if err != nil {
		t.Fatalf("unexpected clamp error: %v", err)
	}
	expectedClamped := "line 4\nline 5"
	if snippetClamped != expectedClamped {
		t.Errorf("expected %q, got %q", expectedClamped, snippetClamped)
	}
}

func TestStore_PersistRefsBatch(t *testing.T) {
	db, err := sql.Open("fake_evidence_driver", "test_persist_batch")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	store := evidence.NewStore(db, nil)
	snapID := uuid.New()
	fileID := uuid.New()

	refs := []domain.EvidenceRef{
		{
			ID:          "ev_01",
			SnapshotID:  snapID,
			FileID:      fileID,
			StartLine:   10,
			EndLine:     20,
			ContentHash: evidence.ComputeHash("func Test() {}"),
			Provenance:  "hybrid_retrieval",
		},
	}

	ctx := context.Background()
	err = store.PersistRefsBatch(ctx, refs)
	if err != nil {
		t.Fatalf("unexpected persist error: %v", err)
	}
}

func TestStore_HydrateEvidence_And_AssemblePackage(t *testing.T) {
	db, err := sql.Open("fake_evidence_driver", "test_hydrate")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	snapID := uuid.New()
	fileID := uuid.New()
	rawContent := "package api\n\nimport \"net/http\"\n\nfunc HealthCheck(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(http.StatusOK)\n}\n"

	testEvidenceDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "SELECT path, content FROM repository_files") {
			return newRows(
				[]string{"path", "content"},
				[][]driver.Value{
					{"api/health.go", rawContent},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	expectedSnippet, err := evidence.ExtractSnippet(rawContent, 5, 7)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}
	expectedHash := evidence.ComputeHash(expectedSnippet)

	store := evidence.NewStore(db, nil)
	ctx := context.Background()

	refs := []domain.EvidenceRef{
		{
			ID:          "ev_raw_01",
			SnapshotID:  snapID,
			FileID:      fileID,
			StartLine:   5,
			EndLine:     7,
			ContentHash: expectedHash,
			Provenance:  "symbol_exact",
		},
	}

	pkg, err := store.AssembleEvidencePackage(ctx, snapID, "commit123", "HealthCheck", refs)
	if err != nil {
		t.Fatalf("unexpected assemble error: %v", err)
	}
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.SnapshotID != snapID {
		t.Errorf("expected snapshot ID %s, got %s", snapID, pkg.SnapshotID)
	}
	if pkg.CommitSHA != "commit123" {
		t.Errorf("expected commit commit123, got %s", pkg.CommitSHA)
	}
	if pkg.TotalItems != 1 {
		t.Errorf("expected 1 item, got %d", pkg.TotalItems)
	}
	if len(pkg.Items) != 1 {
		t.Fatalf("expected 1 item in slice, got %d", len(pkg.Items))
	}

	item := pkg.Items[0]
	if item.ID != "ev_01" {
		t.Errorf("expected citation ID ev_01, got %s", item.ID)
	}
	if item.FilePath != "api/health.go" {
		t.Errorf("expected file path api/health.go, got %s", item.FilePath)
	}
	if item.Snippet != expectedSnippet {
		t.Errorf("expected snippet %q, got %q", expectedSnippet, item.Snippet)
	}
	if item.ContentHash != expectedHash {
		t.Errorf("expected content hash %s, got %s", expectedHash, item.ContentHash)
	}
}
