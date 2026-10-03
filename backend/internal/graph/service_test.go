package graph_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/graph"
	"github.com/google/uuid"
)

func TestService_PersistEdges(t *testing.T) {
	db, err := sql.Open("fake_graph_driver", "test_persist_edges")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	svc := graph.NewService(db, nil)
	snapID := uuid.New()
	srcID := uuid.New()
	tgtID := uuid.New()

	edges := []domain.DependencyEdge{
		{
			ID:              uuid.New(),
			SnapshotID:      snapID,
			SourceFileID:    srcID,
			TargetFileID:    &tgtID,
			EdgeType:        domain.EdgeTypeImports,
			IsDeterministic: true,
			RawTarget:       "./utils",
		},
	}

	ctx := context.Background()
	if err := svc.PersistEdges(ctx, snapID, edges); err != nil {
		t.Fatalf("failed to persist edges: %v", err)
	}
}

func TestService_GetCandidateImpact(t *testing.T) {
	db, err := sql.Open("fake_graph_driver", "test_candidate_impact")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	snapID := uuid.New()
	targetFileID := uuid.New()
	depFileID := uuid.New()
	upstreamFileID := uuid.New()

	testGraphDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		// 1. Target file lookup
		if strings.Contains(query, "SELECT id FROM repository_files") {
			return newRows([]string{"id"}, [][]driver.Value{{targetFileID.String()}}), nil, nil
		}

		// 2. Downstream impact CTE
		if strings.Contains(query, "impact_graph AS") {
			return newRows(
				[]string{"id", "path", "depth", "edge_type", "is_deterministic"},
				[][]driver.Value{
					{depFileID.String(), "cmd/server/main.go", int64(1), "IMPORTS", true},
				},
			), nil, nil
		}

		// 3. Upstream dependency CTE
		if strings.Contains(query, "dep_graph AS") {
			return newRows(
				[]string{"id", "path", "depth", "edge_type", "is_deterministic"},
				[][]driver.Value{
					{upstreamFileID.String(), "internal/storage/db.go", int64(1), "IMPORTS", true},
				},
			), nil, nil
		}

		// 4. Affected symbols query
		if strings.Contains(query, "FROM code_symbols") {
			return newRows(
				[]string{"id", "snapshot_id", "file_id", "name", "kind", "start_line", "end_line", "signature", "is_exported", "created_at"},
				[][]driver.Value{
					{uuid.New().String(), snapID.String(), targetFileID.String(), "ProcessRequest", "FUNCTION", int64(15), int64(30), "func ProcessRequest() error", true, time.Now()},
				},
			), nil, nil
		}

		return newRows([]string{}, nil), nil, nil
	}

	svc := graph.NewService(db, nil)
	ctx := context.Background()

	report, err := svc.GetCandidateImpact(ctx, snapID, "internal/api/router.go")
	if err != nil {
		t.Fatalf("failed to get candidate impact: %v", err)
	}

	if report.TargetFile != "internal/api/router.go" {
		t.Errorf("expected target file internal/api/router.go, got %s", report.TargetFile)
	}
	if len(report.DownstreamCandidates) != 1 {
		t.Errorf("expected 1 downstream candidate, got %d", len(report.DownstreamCandidates))
	}
	if len(report.UpstreamDependencies) != 1 {
		t.Errorf("expected 1 upstream dependency, got %d", len(report.UpstreamDependencies))
	}
	if len(report.AffectedSymbols) != 1 {
		t.Errorf("expected 1 affected symbol, got %d", len(report.AffectedSymbols))
	}
	if report.AffectedSymbols[0].Name != "ProcessRequest" {
		t.Errorf("expected symbol ProcessRequest, got %s", report.AffectedSymbols[0].Name)
	}
}

func TestService_GetSubsystemConnections(t *testing.T) {
	db, err := sql.Open("fake_graph_driver", "test_subsystem_conn")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	testGraphDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "FROM dependency_edges") {
			return newRows(
				[]string{"source_path", "target_path"},
				[][]driver.Value{
					{"cmd/server/main.go", "internal/api/router.go"},
					{"internal/api/router.go", "internal/storage/db.go"},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	svc := graph.NewService(db, nil)
	ctx := context.Background()
	snapID := uuid.New()

	conns, err := svc.GetSubsystemConnections(ctx, snapID)
	if err != nil {
		t.Fatalf("failed to get subsystem connections: %v", err)
	}

	if len(conns) == 0 {
		t.Fatal("expected subsystem connections, got 0")
	}

	// Verify cmd connects to api
	targets, ok := conns["core-cmd"]
	if !ok || len(targets) == 0 {
		t.Errorf("expected connections from core-cmd, got %+v", conns)
	}
}

func TestFeatureTraceBuilder(t *testing.T) {
	db, err := sql.Open("fake_graph_driver", "test_traces")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	snapID := uuid.New()
	mainID := uuid.New()
	routerID := uuid.New()

	testGraphDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "FROM repository_files") {
			return newRows(
				[]string{"id", "path", "language", "content"},
				[][]driver.Value{
					{mainID.String(), "cmd/server/main.go", "Go", "package main\n\nfunc main() {}\n"},
					{routerID.String(), "internal/api/router.go", "Go", "package api\n\nfunc SetupRouter() {}\n"},
				},
			), nil, nil
		}
		if strings.Contains(query, "FROM code_symbols") {
			return newRows(
				[]string{"file_id", "path", "name", "kind", "start_line", "signature"},
				[][]driver.Value{
					{mainID.String(), "cmd/server/main.go", "main", "FUNCTION", int64(3), "func main()"},
					{routerID.String(), "internal/api/router.go", "SetupRouter", "FUNCTION", int64(3), "func SetupRouter()"},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	builder := graph.NewFeatureTraceBuilder(db)
	ctx := context.Background()

	traces, err := builder.BuildFeatureTraces(ctx, snapID)
	if err != nil {
		t.Fatalf("failed to build traces: %v", err)
	}

	if len(traces) == 0 {
		t.Fatal("expected at least 1 feature trace, got 0")
	}

	trace, err := builder.GetFeatureTraceByID(ctx, snapID, traces[0].ID)
	if err != nil {
		t.Fatalf("failed to get trace by ID: %v", err)
	}
	if len(trace.Steps) < 2 {
		t.Errorf("expected at least 2 steps in lifecycle trace, got %d", len(trace.Steps))
	}
}
