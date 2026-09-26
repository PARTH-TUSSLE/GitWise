package analysis_test

import (
	"testing"

	"github.com/gitwise/backend/internal/analysis"
	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

func TestClassifySubsystems(t *testing.T) {
	files := []domain.RepositoryFile{
		{Path: "cmd/server/main.go", Language: "Go", SizeBytes: 1200},
		{Path: "internal/service/repo.go", Language: "Go", SizeBytes: 3400},
		{Path: "internal/service/user.go", Language: "Go", SizeBytes: 2500},
		{Path: "src/app/page.tsx", Language: "TypeScript (JSX)", SizeBytes: 800},
		{Path: "src/components/Header.tsx", Language: "TypeScript (JSX)", SizeBytes: 1500},
		{Path: "README.md", Language: "Plain Text", SizeBytes: 500},
	}

	snapID := uuid.New()
	symbols := []domain.CodeSymbol{
		{SnapshotID: snapID, FilePath: "cmd/server/main.go", Name: "main", Kind: domain.SymbolKindFunction},
		{SnapshotID: snapID, FilePath: "internal/service/repo.go", Name: "Service", Kind: domain.SymbolKindStruct},
		{SnapshotID: snapID, FilePath: "internal/service/repo.go", Name: "NewService", Kind: domain.SymbolKindFunction},
		{SnapshotID: snapID, FilePath: "src/app/page.tsx", Name: "HomePage", Kind: domain.SymbolKindFunction},
	}

	subsystems := analysis.ClassifySubsystems(files, symbols)

	if len(subsystems) < 3 {
		t.Fatalf("expected at least 3 subsystems, got %d", len(subsystems))
	}

	subMap := make(map[string]domain.SubsystemNode)
	for _, s := range subsystems {
		subMap[s.ID] = s
	}

	// Verify cmd -> cli
	if cli, ok := subMap["cli"]; !ok {
		t.Error("expected cli subsystem to be classified")
	} else {
		if cli.FileCount != 1 {
			t.Errorf("expected cli file count 1, got %d", cli.FileCount)
		}
		if cli.EntryPoint != "cmd/server/main.go" {
			t.Errorf("expected entrypoint cmd/server/main.go, got %s", cli.EntryPoint)
		}
	}

	// Verify internal -> core-service
	if core, ok := subMap["core-service"]; !ok {
		t.Error("expected core-service subsystem to be classified")
	} else {
		if core.FileCount != 2 {
			t.Errorf("expected core-service file count 2, got %d", core.FileCount)
		}
		if core.SymbolCount != 2 {
			t.Errorf("expected core-service symbol count 2, got %d", core.SymbolCount)
		}
	}
}

func TestBuildRepoTree(t *testing.T) {
	files := []domain.RepositoryFile{
		{Path: "cmd/server/main.go", Language: "Go", SizeBytes: 1024},
		{Path: "internal/app.go", Language: "Go", SizeBytes: 2048},
		{Path: "README.md", Language: "Plain Text", SizeBytes: 512},
	}

	snapID := uuid.New()
	symbols := []domain.CodeSymbol{
		{SnapshotID: snapID, FilePath: "cmd/server/main.go", Name: "main", Kind: domain.SymbolKindFunction},
		{SnapshotID: snapID, FilePath: "internal/app.go", Name: "App", Kind: domain.SymbolKindStruct},
	}

	tree := analysis.BuildRepoTree(files, symbols)

	if len(tree) == 0 {
		t.Fatal("expected non-empty tree")
	}

	// Check that directories sort before files
	var firstIsDir bool
	if len(tree) > 0 && tree[0].Type == "directory" {
		firstIsDir = true
	}
	if !firstIsDir {
		t.Error("expected top-level directory to sort before files")
	}
}

func TestClassifySubsystems_BackendPrefixes(t *testing.T) {
	files := []domain.RepositoryFile{
		{Path: "backend/internal/foo.go", Language: "Go", SizeBytes: 100},
		{Path: "backend/pkg/foo.go", Language: "Go", SizeBytes: 200},
		{Path: "backend/api/foo.go", Language: "Go", SizeBytes: 300},
	}

	subsystems := analysis.ClassifySubsystems(files, nil)

	subMap := make(map[string]domain.SubsystemNode)
	for _, s := range subsystems {
		subMap[s.ID] = s
	}

	// Verify backend/internal/foo.go -> core-service
	if core, ok := subMap["core-service"]; !ok {
		t.Error("expected backend/internal/... to be classified under core-service")
	} else if core.FileCount != 1 {
		t.Errorf("expected 1 file in core-service, got %d", core.FileCount)
	}

	// Verify backend/pkg/foo.go -> public-pkg
	if pkg, ok := subMap["public-pkg"]; !ok {
		t.Error("expected backend/pkg/... to be classified under public-pkg")
	} else if pkg.FileCount != 1 {
		t.Errorf("expected 1 file in public-pkg, got %d", pkg.FileCount)
	}

	// Verify backend/api/foo.go -> api-server
	if api, ok := subMap["api-server"]; !ok {
		t.Error("expected backend/api/... to be classified under api-server")
	} else if api.FileCount != 1 {
		t.Errorf("expected 1 file in api-server, got %d", api.FileCount)
	}
}

func TestDetectDominantLanguage_DeterministicTieBreak(t *testing.T) {
	// Equal counts: 1 Go file and 1 TypeScript file
	// Tie-break rule: "Go" < "TypeScript" alphabetically -> "Go" wins deterministically every time
	files := []domain.RepositoryFile{
		{Path: "backend/internal/foo.go", Language: "TypeScript", SizeBytes: 100},
		{Path: "backend/internal/bar.go", Language: "Go", SizeBytes: 100},
	}

	// Run multiple iterations to verify stability against randomized map iteration order
	for i := 0; i < 20; i++ {
		subsystems := analysis.ClassifySubsystems(files, nil)
		if len(subsystems) == 0 {
			t.Fatal("expected at least 1 subsystem")
		}
		if subsystems[0].Language != "Go" {
			t.Fatalf("iteration %d: expected deterministic tie-break 'Go', got %s", i, subsystems[0].Language)
		}
	}
}
