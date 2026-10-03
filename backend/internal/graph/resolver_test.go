package graph_test

import (
	"testing"

	"github.com/gitwise/backend/internal/analysis"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/graph"
	"github.com/google/uuid"
)

func TestResolver_GoImports(t *testing.T) {
	snapID := uuid.New()
	goModID := uuid.New()
	mainID := uuid.New()
	apiID := uuid.New()
	serviceID := uuid.New()

	goModContent := "module github.com/gitwise/testrepo\n\ngo 1.25\n"

	files := []domain.RepositoryFile{
		{ID: goModID, SnapshotID: snapID, Path: "go.mod", Content: &goModContent},
		{ID: mainID, SnapshotID: snapID, Path: "cmd/server/main.go"},
		{ID: apiID, SnapshotID: snapID, Path: "internal/api/router.go"},
		{ID: serviceID, SnapshotID: snapID, Path: "internal/service/user.go"},
	}

	analysisResults := map[string]*analysis.FileAnalysisResult{
		"cmd/server/main.go": {
			Imports: []string{
				"fmt",
				"net/http",
				"github.com/gitwise/testrepo/internal/api",
			},
		},
		"internal/api/router.go": {
			Imports: []string{
				"github.com/gitwise/testrepo/internal/service",
				"github.com/go-chi/chi/v5",
			},
		},
	}

	resolver := graph.NewResolver()
	edges := resolver.ResolveSnapshotEdges(snapID, files, analysisResults)

	if len(edges) == 0 {
		t.Fatal("expected resolved edges, got 0")
	}

	// Verify main.go -> router.go internal edge
	foundMainToAPI := false
	foundMainToFmt := false
	foundAPIToService := false
	foundAPIToChi := false

	for _, e := range edges {
		if e.SourceFileID == mainID && e.TargetFileID != nil && *e.TargetFileID == apiID {
			foundMainToAPI = true
			if !e.IsDeterministic {
				t.Errorf("expected deterministic edge for internal Go import, got false")
			}
		}
		if e.SourceFileID == mainID && e.TargetFileID == nil && e.RawTarget == "fmt" {
			foundMainToFmt = true
		}
		if e.SourceFileID == apiID && e.TargetFileID != nil && *e.TargetFileID == serviceID {
			foundAPIToService = true
		}
		if e.SourceFileID == apiID && e.TargetFileID == nil && e.RawTarget == "github.com/go-chi/chi/v5" {
			foundAPIToChi = true
		}
	}

	if !foundMainToAPI {
		t.Error("expected internal edge from main.go to internal/api/router.go")
	}
	if !foundMainToFmt {
		t.Error("expected external edge from main.go to fmt")
	}
	if !foundAPIToService {
		t.Error("expected internal edge from router.go to internal/service/user.go")
	}
	if !foundAPIToChi {
		t.Error("expected external edge from router.go to chi")
	}
}

func TestResolver_TypeScriptImports(t *testing.T) {
	snapID := uuid.New()
	pageID := uuid.New()
	compID := uuid.New()
	utilsID := uuid.New()
	barrelID := uuid.New()

	files := []domain.RepositoryFile{
		{ID: pageID, SnapshotID: snapID, Path: "src/app/page.tsx"},
		{ID: compID, SnapshotID: snapID, Path: "src/components/Header.tsx"},
		{ID: utilsID, SnapshotID: snapID, Path: "src/lib/utils.ts"},
		{ID: barrelID, SnapshotID: snapID, Path: "src/lib/api/index.ts"},
	}

	analysisResults := map[string]*analysis.FileAnalysisResult{
		"src/app/page.tsx": {
			Imports: []string{
				"react",
				"@/components/Header",
				"../lib/utils",
				"@/lib/api",
			},
		},
	}

	resolver := graph.NewResolver()
	edges := resolver.ResolveSnapshotEdges(snapID, files, analysisResults)

	foundPageToHeader := false
	foundPageToUtils := false
	foundPageToBarrel := false
	foundPageToReact := false

	for _, e := range edges {
		if e.SourceFileID == pageID {
			if e.TargetFileID != nil && *e.TargetFileID == compID {
				foundPageToHeader = true
			}
			if e.TargetFileID != nil && *e.TargetFileID == utilsID {
				foundPageToUtils = true
			}
			if e.TargetFileID != nil && *e.TargetFileID == barrelID {
				foundPageToBarrel = true
			}
			if e.TargetFileID == nil && e.RawTarget == "react" {
				foundPageToReact = true
			}
		}
	}

	if !foundPageToHeader {
		t.Error("expected alias edge from page.tsx to Header.tsx")
	}
	if !foundPageToUtils {
		t.Error("expected relative edge from page.tsx to utils.ts")
	}
	if !foundPageToBarrel {
		t.Error("expected barrel edge from page.tsx to src/lib/api/index.ts")
	}
	if !foundPageToReact {
		t.Error("expected external edge from page.tsx to react")
	}
}
