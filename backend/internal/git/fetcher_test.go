package git_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gitwise/backend/internal/git"
)

func TestShouldIgnorePath(t *testing.T) {
	tests := []struct {
		path   string
		ignore bool
	}{
		{"src/main.go", false},
		{"internal/service.ts", false},
		{"node_modules/express/index.js", true},
		{".git/config", true},
		{"vendor/github.com/foo/bar.go", true},
		{"package-lock.json", true},
		{"pnpm-lock.yaml", true},
		{"go.sum", true},
		{"public/logo.png", true},
		{"assets/fonts/inter.woff2", true},
		{"dist/bundle.min.js", true},
		{"README.md", false},
		{"configs/app.yaml", false},
	}

	for _, tt := range tests {
		got := git.ShouldIgnorePath(tt.path)
		if got != tt.ignore {
			t.Errorf("ShouldIgnorePath(%q) = %v; want %v", tt.path, got, tt.ignore)
		}
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "Go"},
		{"app.ts", "TypeScript"},
		{"component.tsx", "TypeScript (JSX)"},
		{"script.js", "JavaScript"},
		{"main.py", "Python"},
		{"lib.rs", "Rust"},
		{"schema.sql", "SQL"},
		{"run.sh", "Shell"},
		{"unknown.xyz", "Plain Text"},
	}

	for _, tt := range tests {
		got := git.DetectLanguage(tt.path)
		if got != tt.expected {
			t.Errorf("DetectLanguage(%q) = %q; want %q", tt.path, got, tt.expected)
		}
	}
}

func TestMockFetcher(t *testing.T) {
	expectedSHA := "0123456789abcdef0123456789abcdef01234567"
	content := "package main\n\nfunc main() {\n\tprintln(\"hello world\")\n}\n"
	h := sha256.Sum256([]byte(content))
	hashHex := hex.EncodeToString(h[:])

	files := []git.FileEntry{
		{
			Path:       "main.go",
			Extension:  ".go",
			Language:   "Go",
			SizeBytes:  len(content),
			LineCount:  5,
			SHA256Hash: hashHex,
			Content:    content,
			IsBinary:   false,
		},
	}

	fetcher := git.NewMockFetcher(expectedSHA, files)
	ctx := context.Background()

	sha, err := fetcher.ResolveCommitSHA(ctx, "acme", "widgets", "main")
	if err != nil {
		t.Fatalf("unexpected error resolving SHA: %v", err)
	}
	if sha != expectedSHA {
		t.Errorf("expected SHA %s, got %s", expectedSHA, sha)
	}

	tree, err := fetcher.FetchTree(ctx, "acme", "widgets", sha)
	if err != nil {
		t.Fatalf("unexpected error fetching tree: %v", err)
	}
	if tree.Outcome != git.IngestionOutcomeComplete {
		t.Errorf("expected outcome COMPLETE, got %s", tree.Outcome)
	}
	if len(tree.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(tree.Files))
	}
	if tree.Files[0].Path != "main.go" {
		t.Errorf("expected file main.go, got %s", tree.Files[0].Path)
	}
	if tree.Files[0].SHA256Hash != hashHex {
		t.Errorf("expected hash %s, got %s", hashHex, tree.Files[0].SHA256Hash)
	}
}

func TestGitHubFetcher_TruncatedTreeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/git/trees/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"sha": "1234567890123456789012345678901234567890",
				"truncated": true,
				"tree": [
					{"path": "file1.go", "type": "blob", "size": 100, "url": "http://example.com/blob1"}
				]
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	fetcher := git.NewGitHubFetcher(ts.URL, "dummy-token", nil)
	_, err := fetcher.FetchTree(context.Background(), "owner", "repo", "1234567890123456789012345678901234567890")
	if err == nil {
		t.Fatal("expected error for truncated tree response, got nil")
	}
	if !errors.Is(err, git.ErrTreeTruncated) {
		t.Errorf("expected ErrTreeTruncated, got %v", err)
	}
}

func TestGitHubFetcher_BlobFetchFailure(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/git/trees/") {
			w.Header().Set("Content-Type", "application/json")
			blobURL := ts.URL + "/git/blobs/bad-blob"
			fmt.Fprintf(w, `{
				"sha": "1234567890123456789012345678901234567890",
				"truncated": false,
				"tree": [
					{"path": "main.go", "type": "blob", "size": 100, "url": "%s"}
				]
			}`, blobURL)
			return
		}
		if strings.Contains(r.URL.Path, "/git/blobs/bad-blob") {
			http.Error(w, "upstream rate limit or network error", http.StatusBadGateway)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	fetcher := git.NewGitHubFetcher(ts.URL, "dummy-token", nil)
	_, err := fetcher.FetchTree(context.Background(), "owner", "repo", "1234567890123456789012345678901234567890")
	if err == nil {
		t.Fatal("expected error on failed blob fetch, got nil")
	}
	if !strings.Contains(err.Error(), "failed to fetch blob content") {
		t.Errorf("expected blob fetch error message, got %v", err)
	}
}

func TestGitHubFetcher_SingleFileSizeCap(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/git/trees/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"sha": "1234567890123456789012345678901234567890",
				"truncated": false,
				"tree": [
					{"path": "huge.dat", "type": "blob", "size": 600000, "url": "http://example.com/huge"}
				]
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	fetcher := git.NewGitHubFetcher(ts.URL, "dummy-token", nil)
	res, err := fetcher.FetchTree(context.Background(), "owner", "repo", "1234567890123456789012345678901234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Outcome != git.IngestionOutcomePartial {
		t.Errorf("expected Outcome PARTIAL for capped file, got %v", res.Outcome)
	}
	if len(res.Files) != 0 {
		t.Errorf("expected 0 files after skipping oversized file, got %d", len(res.Files))
	}
}
