package migrations_test

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"testing"
	"testing/fstest"

	"github.com/gitwise/backend/internal/storage/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestLoadMigrations_Embedded(t *testing.T) {
	list, err := migrations.LoadMigrations(migrations.MigrationFS)
	if err != nil {
		t.Fatalf("failed to load embedded migrations: %v", err)
	}

	if len(list) < 9 {
		t.Fatalf("expected at least 9 migrations, got %d", len(list))
	}

	// Verify versions strictly sorted in ascending order
	for i := 0; i < len(list)-1; i++ {
		if list[i].Version >= list[i+1].Version {
			t.Errorf("migrations not sorted in strict ascending order: index %d (%d) >= index %d (%d)",
				i, list[i].Version, i+1, list[i+1].Version)
		}
	}

	// Verify versions 1-9
	if list[0].Version != 1 || list[0].Name != "000001_init_schema" {
		t.Errorf("expected version 1 to be 000001_init_schema, got: %+v", list[0])
	}
	if list[1].Version != 2 || list[1].Name != "000002_github_gitstat" {
		t.Errorf("expected version 2 to be 000002_github_gitstat, got: %+v", list[1])
	}
	if list[2].Version != 3 || list[2].Name != "000003_repository_files" {
		t.Errorf("expected version 3 to be 000003_repository_files, got: %+v", list[2])
	}
	if list[3].Version != 4 || list[3].Name != "000004_active_jobs_partial_index" {
		t.Errorf("expected version 4 to be 000004_active_jobs_partial_index, got: %+v", list[3])
	}
	if list[4].Version != 5 || list[4].Name != "000005_code_symbols" {
		t.Errorf("expected version 5 to be 000005_code_symbols, got: %+v", list[4])
	}
	if list[5].Version != 6 || list[5].Name != "000006_dependency_edges" {
		t.Errorf("expected version 6 to be 000006_dependency_edges, got: %+v", list[5])
	}
	if list[6].Version != 7 || list[6].Name != "000007_code_chunks_and_evidence" {
		t.Errorf("expected version 7 to be 000007_code_chunks_and_evidence, got: %+v", list[6])
	}
	if list[7].Version != 8 || list[7].Name != "000008_mentor_sessions_and_chat" {
		t.Errorf("expected version 8 to be 000008_mentor_sessions_and_chat, got: %+v", list[7])
	}
	if list[8].Version != 9 || list[8].Name != "000009_issue_blueprints" {
		t.Errorf("expected version 9 to be 000009_issue_blueprints, got: %+v", list[8])
	}

	// Verify SQL contents are loaded
	if len(list[0].SQL) == 0 {
		t.Errorf("migration 1 SQL content is empty")
	}
	if len(list[1].SQL) == 0 {
		t.Errorf("migration 2 SQL content is empty")
	}
}

func TestLoadMigrations_DeterministicOrdering(t *testing.T) {
	// Create out-of-order test filesystem
	testFS := fstest.MapFS{
		"000004_four.up.sql":  &fstest.MapFile{Data: []byte("SELECT 4;")},
		"000002_two.up.sql":   &fstest.MapFile{Data: []byte("SELECT 2;")},
		"000001_one.up.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
		"000003_three.up.sql": &fstest.MapFile{Data: []byte("SELECT 3;")},
		"000001_one.down.sql": &fstest.MapFile{Data: []byte("DROP 1;")},
		"non_migration.txt":   &fstest.MapFile{Data: []byte("IGNORED")},
	}

	list, err := migrations.LoadMigrations(testFS)
	if err != nil {
		t.Fatalf("failed to load custom migrations: %v", err)
	}

	if len(list) != 4 {
		t.Fatalf("expected 4 up migrations, got %d", len(list))
	}

	expectedVersions := []int{1, 2, 3, 4}
	for i, exp := range expectedVersions {
		if list[i].Version != exp {
			t.Errorf("expected position %d to have version %d, got %d", i, exp, list[i].Version)
		}
	}
}

func TestLoadMigrations_InvalidFormat(t *testing.T) {
	invalidFS := fstest.MapFS{
		"invalid_no_version.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}

	_, err := migrations.LoadMigrations(invalidFS)
	if err == nil {
		t.Errorf("expected error loading migration with non-numeric prefix, got nil")
	}
}

func TestRunner_IdempotencyAndStateTracking(t *testing.T) {
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("skipping runner integration test: TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("skipping runner integration test: database unreachable: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := migrations.NewRunner(db, logger)

	// First run: should apply all pending migrations
	if err := runner.Run(ctx); err != nil {
		t.Fatalf("first migration run failed: %v", err)
	}

	// Verify schema_migrations table records both versions
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version IN (1, 2);").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count < 2 {
		t.Errorf("expected at least 2 migrations recorded in schema_migrations, got %d", count)
	}

	// Second run: must be idempotent and apply 0 migrations without error
	if err := runner.Run(ctx); err != nil {
		t.Fatalf("second migration run failed: %v", err)
	}
}
