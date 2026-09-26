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
