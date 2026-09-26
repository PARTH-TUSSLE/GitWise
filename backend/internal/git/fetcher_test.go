package git_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	if len(tree) != 1 {
		t.Fatalf("expected 1 file, got %d", len(tree))
	}
	if tree[0].Path != "main.go" {
		t.Errorf("expected file main.go, got %s", tree[0].Path)
	}
	if tree[0].SHA256Hash != hashHex {
		t.Errorf("expected hash %s, got %s", hashHex, tree[0].SHA256Hash)
	}
}
