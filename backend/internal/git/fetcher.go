package git

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Ingestion safety thresholds
const (
	MaxSingleFileSizeBytes = 500 * 1024        // 500 KB
	MaxTotalRepoSizeBytes  = 150 * 1024 * 1024 // 150 MB
	MaxTotalFiles          = 5000
)

// ErrTreeTruncated is returned when the GitHub Git Data API truncates the tree response.
var ErrTreeTruncated = errors.New("github git tree response truncated: repository exceeds tree limit")

// IngestionOutcome classifies the completeness of the retrieved repository snapshot.
type IngestionOutcome string

const (
	IngestionOutcomeComplete IngestionOutcome = "COMPLETE"
	IngestionOutcomePartial  IngestionOutcome = "PARTIAL"
)

// FileEntry represents a processed source code file from a commit snapshot.
type FileEntry struct {
	Path       string
	Extension  string
	Language   string
	SizeBytes  int
	LineCount  int
	SHA256Hash string
	Content    string
	IsBinary   bool
}

// FetchResult packages the ingested files alongside completeness classification.
type FetchResult struct {
	Files        []FileEntry
	Outcome      IngestionOutcome
	CappedReason string
}

// Fetcher defines the interface for repository tree resolution and ingestion.
type Fetcher interface {
	ResolveCommitSHA(ctx context.Context, owner, repo, ref string) (string, error)
	FetchTree(ctx context.Context, owner, repo, commitSHA string) (*FetchResult, error)
}

// GitHubFetcher implements Fetcher against GitHub REST & Git Data APIs.
type GitHubFetcher struct {
	baseURL    string
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewGitHubFetcher creates a GitHub tree and blob fetcher.
func NewGitHubFetcher(baseURL, token string, logger *slog.Logger) *GitHubFetcher {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &GitHubFetcher{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
		logger: logger,
	}
}

// ResolveCommitSHA resolves a ref (branch name, tag, or partial SHA) to a 40-character commit SHA.
func (f *GitHubFetcher) ResolveCommitSHA(ctx context.Context, owner, repo, ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}
	// If ref is already a 40-character commit SHA, return directly
	if len(ref) == 40 && isHex(ref) {
		return ref, nil
	}

	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s", f.baseURL, owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	f.setHeaders(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query GitHub commit ref: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub commit resolution failed with status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to decode GitHub commit payload: %w", err)
	}
	if payload.SHA == "" {
		return "", errors.New("empty commit sha returned from GitHub")
	}

	return payload.SHA, nil
}

// gitTreeResponse represents the GitHub Git Trees API response.
type gitTreeResponse struct {
	SHA       string        `json:"sha"`
	Tree      []gitTreeItem `json:"tree"`
	Truncated bool          `json:"truncated"`
}

type gitTreeItem struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"` // "blob" or "tree"
	SHA  string `json:"sha"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

// FetchTree queries GitHub Trees API recursively, filters safety bounds, and fetches file contents.
func (f *GitHubFetcher) FetchTree(ctx context.Context, owner, repo, commitSHA string) (*FetchResult, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/trees/%s?recursive=1", f.baseURL, owner, repo, commitSHA)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	f.setHeaders(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch git tree from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub git tree query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var treeResp gitTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		return nil, fmt.Errorf("failed to decode git tree: %w", err)
	}

	if treeResp.Truncated {
		f.logger.Error("GitHub git tree response was truncated",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.String("commit_sha", commitSHA),
		)
		return nil, ErrTreeTruncated
	}

	var results []FileEntry
	var totalSizeBytes int
	outcome := IngestionOutcomeComplete
	var cappedReason string

	for _, item := range treeResp.Tree {
		if item.Type != "blob" {
			continue
		}

		if ShouldIgnorePath(item.Path) {
			continue
		}

		if item.Size > MaxSingleFileSizeBytes {
			f.logger.Debug("Skipping file exceeding size limit",
				slog.String("path", item.Path),
				slog.Int("size", item.Size),
			)
			outcome = IngestionOutcomePartial
			if cappedReason == "" {
				cappedReason = fmt.Sprintf("file %s exceeded single file limit of %d bytes", item.Path, MaxSingleFileSizeBytes)
			}
			continue
		}

		if totalSizeBytes+item.Size > MaxTotalRepoSizeBytes {
			f.logger.Warn("Repository exceeded total ingestion size cap, capping snapshot as partial",
				slog.Int("total_bytes", totalSizeBytes),
				slog.Int("max_bytes", MaxTotalRepoSizeBytes),
			)
			outcome = IngestionOutcomePartial
			cappedReason = fmt.Sprintf("repository exceeded total size cap of %d MB", MaxTotalRepoSizeBytes/(1024*1024))
			break
		}

		if len(results) >= MaxTotalFiles {
			f.logger.Warn("Repository exceeded total file count limit, capping snapshot as partial",
				slog.Int("total_files", len(results)),
				slog.Int("max_files", MaxTotalFiles),
			)
			outcome = IngestionOutcomePartial
			cappedReason = fmt.Sprintf("repository exceeded total file count limit of %d", MaxTotalFiles)
			break
		}

		// Fetch content for eligible file
		content, isBinary, err := f.fetchBlobContent(ctx, item.URL)
		if err != nil {
			f.logger.Error("Failed to fetch blob content, aborting snapshot ingestion",
				slog.String("path", item.Path),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to fetch blob content for %s: %w", item.Path, err)
		}

		totalSizeBytes += item.Size

		ext := strings.ToLower(path.Ext(item.Path))
		lang := DetectLanguage(item.Path)
		h := sha256.Sum256([]byte(content))
		sha256Hex := hex.EncodeToString(h[:])

		lineCount := 0
		if !isBinary {
			lineCount = countLines(content)
		}

		results = append(results, FileEntry{
			Path:       item.Path,
			Extension:  ext,
			Language:   lang,
			SizeBytes:  len(content),
			LineCount:  lineCount,
			SHA256Hash: sha256Hex,
			Content:    content,
			IsBinary:   isBinary,
		})
	}

	return &FetchResult{
		Files:        results,
		Outcome:      outcome,
		CappedReason: cappedReason,
	}, nil
}

func (f *GitHubFetcher) fetchBlobContent(ctx context.Context, blobURL string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, blobURL, nil)
	if err != nil {
		return "", false, err
	}
	f.setHeaders(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("blob fetch status %d", resp.StatusCode)
	}

	var blobResp struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
		Size     int    `json:"size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&blobResp); err != nil {
		return "", false, err
	}

	var rawBytes []byte
	switch blobResp.Encoding {
	case "base64":
		cleaned := strings.ReplaceAll(blobResp.Content, "\n", "")
		cleaned = strings.ReplaceAll(cleaned, "\r", "")
		decoded, err := base64.StdEncoding.DecodeString(cleaned)
		if err != nil {
			return "", false, fmt.Errorf("failed to decode base64 blob: %w", err)
		}
		rawBytes = decoded
	default:
		rawBytes = []byte(blobResp.Content)
	}

	isBinary := isBinaryContent(rawBytes)
	if isBinary {
		return "", true, nil
	}

	return string(rawBytes), false, nil
}

func (f *GitHubFetcher) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "GitWise-Ingestion-Engine/2.1")
	if f.token != "" {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}
}

// ShouldIgnorePath checks if a file or directory path should be skipped during repository ingestion.
func ShouldIgnorePath(filePath string) bool {
	clean := filepath.ToSlash(filePath)
	parts := strings.Split(clean, "/")

	// Check ignored directories
	for _, part := range parts {
		switch part {
		case ".git", "node_modules", "vendor", "dist", "build", ".next",
			"target", "bin", "obj", ".idea", ".vscode", "coverage", ".turbo":
			return true
		}
	}

	// Check ignored lockfiles & generated files
	base := path.Base(clean)
	switch base {
	case "package-lock.json", "pnpm-lock.yaml", "yarn.lock", "Cargo.lock",
		"go.sum", "composer.lock", "Pipfile.lock", "poetry.lock":
		return true
	}

	// Check binary or asset file extensions
	ext := strings.ToLower(path.Ext(clean))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".webp", ".svg",
		".pdf", ".zip", ".tar", ".gz", ".7z", ".rar", ".wasm",
		".exe", ".so", ".dylib", ".dll", ".bin", ".iso", ".dmg",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".mp4", ".mp3", ".mov", ".avi", ".wav", ".flac",
		".min.js", ".min.css", ".map":
		return true
	}

	return false
}

// DetectLanguage maps file extensions to canonical language names.
func DetectLanguage(filePath string) string {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".go":
		return "Go"
	case ".ts":
		return "TypeScript"
	case ".tsx":
		return "TypeScript (JSX)"
	case ".js":
		return "JavaScript"
	case ".jsx":
		return "JavaScript (JSX)"
	case ".py":
		return "Python"
	case ".rs":
		return "Rust"
	case ".java":
		return "Java"
	case ".c", ".h":
		return "C"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "C++"
	case ".cs":
		return "C#"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".swift":
		return "Swift"
	case ".kt", ".kts":
		return "Kotlin"
	case ".scala":
		return "Scala"
	case ".sql":
		return "SQL"
	case ".sh", ".bash":
		return "Shell"
	case ".html", ".htm":
		return "HTML"
	case ".css", ".scss", ".sass", ".less":
		return "CSS"
	case ".json":
		return "JSON"
	case ".yaml", ".yml":
		return "YAML"
	case ".toml":
		return "TOML"
	case ".md", ".markdown":
		return "Markdown"
	default:
		return "Plain Text"
	}
}

// isBinaryContent inspects initial bytes for null characters indicating binary data.
func isBinaryContent(data []byte) bool {
	checkLen := len(data)
	if checkLen > 8000 {
		checkLen = 8000
	}
	return bytes.IndexByte(data[:checkLen], 0) != -1
}

func countLines(s string) int {
	if len(s) == 0 {
		return 0
	}
	lines := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		lines++
	}
	return lines
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// MockFetcher provides in-memory mock repository file trees for tests.
type MockFetcher struct {
	CommitSHA    string
	Files        []FileEntry
	Outcome      IngestionOutcome
	CappedReason string
	Err          error
}

func NewMockFetcher(commitSHA string, files []FileEntry) *MockFetcher {
	if commitSHA == "" {
		commitSHA = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4"
	}
	return &MockFetcher{
		CommitSHA: commitSHA,
		Files:     files,
		Outcome:   IngestionOutcomeComplete,
	}
}

func (m *MockFetcher) ResolveCommitSHA(ctx context.Context, owner, repo, ref string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return m.CommitSHA, nil
}

func (m *MockFetcher) FetchTree(ctx context.Context, owner, repo, commitSHA string) (*FetchResult, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	outcome := m.Outcome
	if outcome == "" {
		outcome = IngestionOutcomeComplete
	}
	return &FetchResult{
		Files:        m.Files,
		Outcome:      outcome,
		CappedReason: m.CappedReason,
	}, nil
}
