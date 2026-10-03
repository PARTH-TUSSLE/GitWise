package graph

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/gitwise/backend/internal/analysis"
	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

// Resolver maps extracted import strings to concrete repository files within a snapshot.
type Resolver struct{}

// NewResolver creates a new import-to-file dependency resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// ResolveSnapshotEdges analyzes import statements across all files in a snapshot
// and produces directed dependency edges.
func (r *Resolver) ResolveSnapshotEdges(
	snapshotID uuid.UUID,
	files []domain.RepositoryFile,
	analysisResults map[string]*analysis.FileAnalysisResult,
) []domain.DependencyEdge {
	if len(files) == 0 || len(analysisResults) == 0 {
		return nil
	}

	// 1. Build lookup indexes
	pathToID := make(map[string]uuid.UUID, len(files))
	dirToFiles := make(map[string][]domain.RepositoryFile)
	var goModuleName string

	for _, f := range files {
		normPath := normalizePath(f.Path)
		pathToID[normPath] = f.ID

		dir := normalizePath(path.Dir(normPath))
		if dir == "." {
			dir = ""
		}
		dirToFiles[dir] = append(dirToFiles[dir], f)

		// Discover Go module declaration from go.mod
		if path.Base(normPath) == "go.mod" && f.Content != nil {
			goModuleName = extractGoModuleName(*f.Content)
		}
	}

	var edges []domain.DependencyEdge
	seenEdge := make(map[string]bool)

	// 2. Resolve imports for each file
	for _, sourceFile := range files {
		normSourcePath := normalizePath(sourceFile.Path)
		res, hasResult := analysisResults[normSourcePath]
		if !hasResult || res == nil || len(res.Imports) == 0 {
			continue
		}

		isGo := strings.HasSuffix(normSourcePath, ".go")
		sourceDir := normalizePath(path.Dir(normSourcePath))
		if sourceDir == "." {
			sourceDir = ""
		}

		for _, rawImport := range res.Imports {
			rawImport = strings.TrimSpace(rawImport)
			if rawImport == "" {
				continue
			}

			if isGo {
				r.resolveGoImport(
					snapshotID, sourceFile.ID, normSourcePath, rawImport,
					goModuleName, dirToFiles, pathToID, seenEdge, &edges,
				)
			} else {
				r.resolveTSImport(
					snapshotID, sourceFile.ID, normSourcePath, sourceDir,
					rawImport, pathToID, seenEdge, &edges,
				)
			}
		}
	}

	return edges
}

func (r *Resolver) resolveGoImport(
	snapshotID, sourceFileID uuid.UUID,
	sourcePath, rawImport, goModuleName string,
	dirToFiles map[string][]domain.RepositoryFile,
	pathToID map[string]uuid.UUID,
	seen map[string]bool,
	edges *[]domain.DependencyEdge,
) {
	// 1. Check internal module import (e.g. "github.com/org/repo/internal/api")
	var targetDir string
	isInternal := false

	if goModuleName != "" && strings.HasPrefix(rawImport, goModuleName) {
		sub := strings.TrimPrefix(rawImport, goModuleName)
		targetDir = strings.TrimPrefix(normalizePath(sub), "/")
		isInternal = true
	} else if strings.HasPrefix(rawImport, "./") || strings.HasPrefix(rawImport, "../") {
		sourceDir := normalizePath(path.Dir(sourcePath))
		if sourceDir == "." {
			sourceDir = ""
		}
		targetDir = normalizePath(path.Join(sourceDir, rawImport))
		isInternal = true
	}

	if isInternal {
		targetFiles, exists := dirToFiles[targetDir]
		if exists && len(targetFiles) > 0 {
			for _, tf := range targetFiles {
				if tf.ID == sourceFileID {
					continue
				}
				key := makeEdgeKey(sourceFileID, tf.ID, string(domain.EdgeTypeImports), rawImport)
				if !seen[key] {
					seen[key] = true
					targetID := tf.ID
					*edges = append(*edges, domain.DependencyEdge{
						ID:              uuid.New(),
						SnapshotID:      snapshotID,
						SourceFileID:    sourceFileID,
						SourceFilePath:  sourcePath,
						TargetFileID:    &targetID,
						TargetFilePath:  &tf.Path,
						EdgeType:        domain.EdgeTypeImports,
						IsDeterministic: true,
						RawTarget:       rawImport,
					})
				}
			}
			return
		}
	}

	// 2. Directory suffix fallback (e.g. import "backend/internal/api")
	cleanImport := strings.TrimPrefix(rawImport, "/")
	for dir, dirFiles := range dirToFiles {
		if dir == cleanImport || strings.HasSuffix(dir, "/"+cleanImport) {
			for _, tf := range dirFiles {
				if tf.ID == sourceFileID {
					continue
				}
				key := makeEdgeKey(sourceFileID, tf.ID, string(domain.EdgeTypeImports), rawImport)
				if !seen[key] {
					seen[key] = true
					targetID := tf.ID
					*edges = append(*edges, domain.DependencyEdge{
						ID:              uuid.New(),
						SnapshotID:      snapshotID,
						SourceFileID:    sourceFileID,
						SourceFilePath:  sourcePath,
						TargetFileID:    &targetID,
						TargetFilePath:  &tf.Path,
						EdgeType:        domain.EdgeTypeImports,
						IsDeterministic: true,
						RawTarget:       rawImport,
					})
				}
			}
			return
		}
	}

	// 3. External standard library or 3rd party package (target_file_id is nil)
	key := makeEdgeKey(sourceFileID, uuid.Nil, string(domain.EdgeTypeImports), rawImport)
	if !seen[key] {
		seen[key] = true
		*edges = append(*edges, domain.DependencyEdge{
			ID:              uuid.New(),
			SnapshotID:      snapshotID,
			SourceFileID:    sourceFileID,
			SourceFilePath:  sourcePath,
			TargetFileID:    nil,
			TargetFilePath:  nil,
			EdgeType:        domain.EdgeTypeImports,
			IsDeterministic: true,
			RawTarget:       rawImport,
		})
	}
}

func (r *Resolver) resolveTSImport(
	snapshotID, sourceFileID uuid.UUID,
	sourcePath, sourceDir, rawImport string,
	pathToID map[string]uuid.UUID,
	seen map[string]bool,
	edges *[]domain.DependencyEdge,
) {
	// Candidate resolution paths
	var candidates []string

	if strings.HasPrefix(rawImport, "./") || strings.HasPrefix(rawImport, "../") {
		base := normalizePath(path.Join(sourceDir, rawImport))
		candidates = generateTSCandidates(base)
	} else if strings.HasPrefix(rawImport, "@/") {
		trimmed := strings.TrimPrefix(rawImport, "@/")
		candidates = append(candidates, generateTSCandidates(trimmed)...)
		candidates = append(candidates, generateTSCandidates(path.Join("src", trimmed))...)
	} else if strings.HasPrefix(rawImport, "~/") {
		trimmed := strings.TrimPrefix(rawImport, "~/")
		candidates = append(candidates, generateTSCandidates(trimmed)...)
		candidates = append(candidates, generateTSCandidates(path.Join("src", trimmed))...)
	}

	for _, cand := range candidates {
		if targetID, ok := pathToID[cand]; ok {
			if targetID == sourceFileID {
				continue
			}
			key := makeEdgeKey(sourceFileID, targetID, string(domain.EdgeTypeImports), rawImport)
			if !seen[key] {
				seen[key] = true
				targetPath := cand
				*edges = append(*edges, domain.DependencyEdge{
					ID:              uuid.New(),
					SnapshotID:      snapshotID,
					SourceFileID:    sourceFileID,
					SourceFilePath:  sourcePath,
					TargetFileID:    &targetID,
					TargetFilePath:  &targetPath,
					EdgeType:        domain.EdgeTypeImports,
					IsDeterministic: true,
					RawTarget:       rawImport,
				})
			}
			return
		}
	}

	// External npm package (e.g. "react", "next/navigation")
	key := makeEdgeKey(sourceFileID, uuid.Nil, string(domain.EdgeTypeImports), rawImport)
	if !seen[key] {
		seen[key] = true
		*edges = append(*edges, domain.DependencyEdge{
			ID:              uuid.New(),
			SnapshotID:      snapshotID,
			SourceFileID:    sourceFileID,
			SourceFilePath:  sourcePath,
			TargetFileID:    nil,
			TargetFilePath:  nil,
			EdgeType:        domain.EdgeTypeImports,
			IsDeterministic: true,
			RawTarget:       rawImport,
		})
	}
}

func generateTSCandidates(base string) []string {
	exts := []string{
		".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs",
		"/index.ts", "/index.tsx", "/index.js", "/index.jsx",
	}

	candidates := []string{base}
	for _, ext := range exts {
		candidates = append(candidates, base+ext)
	}
	return candidates
}

func normalizePath(p string) string {
	clean := filepath.ToSlash(filepath.Clean(p))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	return clean
}

func extractGoModuleName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
		}
	}
	return ""
}

func makeEdgeKey(source, target uuid.UUID, edgeType, raw string) string {
	return source.String() + "->" + target.String() + ":" + edgeType + ":" + raw
}
