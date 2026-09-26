package repo_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/gitwise/backend/internal/git"
	"github.com/gitwise/backend/internal/service/repo"
)

func TestService_IngestValidation(t *testing.T) {
	svc := repo.NewService(nil, nil, nil, nil)
	ctx := context.Background()

	// Missing owner
	_, _, err := svc.Ingest(ctx, "", "repo", "main")
	if err == nil {
		t.Fatal("expected error for empty owner, got nil")
	}

	// Missing repo
	_, _, err = svc.Ingest(ctx, "owner", "", "main")
	if err == nil {
		t.Fatal("expected error for empty repo, got nil")
	}
}

func TestService_WithMockFetcher_Validation(t *testing.T) {
	expectedSHA := "0123456789abcdef0123456789abcdef01234567"
	content := "package main\n\nfunc main() {}\n"
	h := sha256.Sum256([]byte(content))
	hashHex := hex.EncodeToString(h[:])

	files := []git.FileEntry{
		{
			Path:       "main.go",
			Extension:  ".go",
			Language:   "Go",
			SizeBytes:  len(content),
			LineCount:  3,
			SHA256Hash: hashHex,
			Content:    content,
			IsBinary:   false,
		},
	}

	fetcher := git.NewMockFetcher(expectedSHA, files)
	svc := repo.NewService(nil, fetcher, nil, nil)

	ctx := context.Background()
	sha, err := fetcher.ResolveCommitSHA(ctx, "testowner", "testrepo", "main")
	if err != nil {
		t.Fatalf("unexpected error resolving commit: %v", err)
	}
	if sha != expectedSHA {
		t.Errorf("expected sha %s, got %s", expectedSHA, sha)
	}

	// With nil DB, Ingest should return a database error safely
	_, _, err = svc.Ingest(ctx, "testowner", "testrepo", "main")
	if err == nil {
		t.Fatal("expected error with nil database, got nil")
	}
}
