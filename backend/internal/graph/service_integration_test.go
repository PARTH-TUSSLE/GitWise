package graph_test

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/graph"
	"github.com/gitwise/backend/internal/storage/migrations"
	"github.com/gitwise/backend/internal/storage/postgres"
	"github.com/google/uuid"
)

func TestService_PostgresIntegration_GraphTraversalAndCyclePrevention(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping live postgres test: DATABASE_URL not set")
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pgDB, err := postgres.New(ctx, dbURL, logger)
	if err != nil {
		t.Skipf("Skipping live postgres test (instance not reachable): %v", err)
	}
	defer pgDB.Close()

	runner := migrations.NewRunner(pgDB.DB, logger)
	if err := runner.Run(ctx); err != nil {
		t.Fatalf("failed to run migrations on postgres: %v", err)
	}

	// Dynamic identifiers for test isolation
	testID := uuid.New().String()
	testOwner := fmt.Sprintf("graph-owner-%s", testID)
	testRepo := fmt.Sprintf("graph-repo-%s", testID)

	shaBytes := make([]byte, 20)
	if _, err := crand.Read(shaBytes); err != nil {
		t.Fatalf("failed to generate random commit SHA: %v", err)
	}
	commitSHA := hex.EncodeToString(shaBytes)

	// Register automated cleanup (cascades down to snapshots, files, symbols, edges)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pgDB.DB.ExecContext(cleanupCtx, "DELETE FROM repositories WHERE owner = $1 AND name = $2", testOwner, testRepo)
	})

	// 1. Insert repository and snapshot
	var repoID, snapID uuid.UUID
	err = pgDB.DB.QueryRowContext(ctx, `
		INSERT INTO repositories (github_id, owner, name, default_branch, is_private, created_at, updated_at)
		VALUES ($1, $2, $3, 'main', false, NOW(), NOW())
		RETURNING id`,
		time.Now().UnixNano(), testOwner, testRepo,
	).Scan(&repoID)
	if err != nil {
		t.Fatalf("failed to insert test repo: %v", err)
	}

	err = pgDB.DB.QueryRowContext(ctx, `
		INSERT INTO repository_snapshots (repository_id, commit_sha, ref_name, status, total_files, total_lines)
		VALUES ($1, $2, 'main', 'READY', 6, 100)
		RETURNING id`,
		repoID, commitSHA,
	).Scan(&snapID)
	if err != nil {
		t.Fatalf("failed to insert test snapshot: %v", err)
	}

	// 2. Insert test files:
	// Cycle: A -> B -> C -> A
	// Chain: C -> D -> E -> F (deep chain)
	fileNames := []string{
		"src/a.ts",
		"src/b.ts",
		"src/c.ts",
		"src/d.ts",
		"src/e.ts",
		"src/f.ts",
	}
	fileIDs := make(map[string]uuid.UUID)

	for _, fn := range fileNames {
		var fid uuid.UUID
		err = pgDB.DB.QueryRowContext(ctx, `
			INSERT INTO repository_files (snapshot_id, path, extension, language, size_bytes, line_count, sha256_hash, content)
			VALUES ($1, $2, '.ts', 'TypeScript', 50, 5, 'hash123', 'export const x = 1;')
			RETURNING id`,
			snapID, fn,
		).Scan(&fid)
		if err != nil {
			t.Fatalf("failed to insert file %s: %v", fn, err)
		}
		fileIDs[fn] = fid
	}

	// 3. Insert dependency edges:
	// B imports A (B -> A)
	// C imports B (C -> B)
	// A imports C (A -> C)  <-- CYCLE!
	// D imports C (D -> C)
	// E imports D (E -> D)
	// F imports E (F -> E)  <-- Deep chain!
	svc := graph.NewService(pgDB.DB, logger)

	idA := fileIDs["src/a.ts"]
	idB := fileIDs["src/b.ts"]
	idC := fileIDs["src/c.ts"]
	idD := fileIDs["src/d.ts"]
	idE := fileIDs["src/e.ts"]
	idF := fileIDs["src/f.ts"]

	edges := []domain.DependencyEdge{
		{ID: uuid.New(), SnapshotID: snapID, SourceFileID: idB, TargetFileID: &idA, EdgeType: domain.EdgeTypeImports, IsDeterministic: true, RawTarget: "./a"},
		{ID: uuid.New(), SnapshotID: snapID, SourceFileID: idC, TargetFileID: &idB, EdgeType: domain.EdgeTypeImports, IsDeterministic: true, RawTarget: "./b"},
		{ID: uuid.New(), SnapshotID: snapID, SourceFileID: idA, TargetFileID: &idC, EdgeType: domain.EdgeTypeImports, IsDeterministic: true, RawTarget: "./c"}, // Cycle!
		{ID: uuid.New(), SnapshotID: snapID, SourceFileID: idD, TargetFileID: &idC, EdgeType: domain.EdgeTypeImports, IsDeterministic: true, RawTarget: "./c"},
		{ID: uuid.New(), SnapshotID: snapID, SourceFileID: idE, TargetFileID: &idD, EdgeType: domain.EdgeTypeImports, IsDeterministic: true, RawTarget: "./d"},
		{ID: uuid.New(), SnapshotID: snapID, SourceFileID: idF, TargetFileID: &idE, EdgeType: domain.EdgeTypeImports, IsDeterministic: true, RawTarget: "./e"},
	}

	if err := svc.PersistEdges(ctx, snapID, edges); err != nil {
		t.Fatalf("failed to persist edges to postgres: %v", err)
	}

	// 4. Assert Idempotency: re-persisting edges must succeed without duplicate key error
	if err := svc.PersistEdges(ctx, snapID, edges); err != nil {
		t.Fatalf("re-persisting edges to postgres failed (idempotency broken): %v", err)
	}

	// 5. Query candidate impact for src/a.ts
	// Downstream dependents of A:
	// Depth 1: B (imports A)
	// Depth 2: C (imports B)
	// Depth 3: A (cycle! stopped by visited array), D (imports C)
	// Depth > 3: E (imports D), F (imports E) must be EXCLUDED by depth <= 3 cap!
	report, err := svc.GetCandidateImpact(ctx, snapID, "src/a.ts")
	if err != nil {
		t.Fatalf("GetCandidateImpact failed on real postgres: %v", err)
	}

	if report.TargetFile != "src/a.ts" {
		t.Errorf("expected targetFile src/a.ts, got %s", report.TargetFile)
	}

	// Check max depth bounds
	for _, n := range report.DownstreamCandidates {
		if n.Depth > 3 {
			t.Errorf("candidate %s exceeded max depth 3: depth=%d", n.Path, n.Depth)
		}
		if n.Path == "src/e.ts" || n.Path == "src/f.ts" {
			t.Errorf("file %s at depth > 3 should not be in downstream candidates", n.Path)
		}
	}

	// Verify cycle did not result in duplicate nodes or infinite loop
	pathCounts := make(map[string]int)
	for _, n := range report.DownstreamCandidates {
		pathCounts[n.Path]++
		if pathCounts[n.Path] > 1 {
			t.Errorf("duplicate node in impact traversal: %s", n.Path)
		}
	}
}
