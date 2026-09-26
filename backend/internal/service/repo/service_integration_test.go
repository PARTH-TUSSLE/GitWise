package repo_test

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

	"github.com/gitwise/backend/internal/api/sse"
	"github.com/gitwise/backend/internal/git"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/gitwise/backend/internal/service/repo"
	"github.com/gitwise/backend/internal/storage/migrations"
	"github.com/gitwise/backend/internal/storage/postgres"
	"github.com/google/uuid"
)

func TestService_PostgresIntegration_LifecycleAndSymbolPersistence(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping live postgres test: DATABASE_URL not set")
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	// Generate isolated, unique identifiers per test run to prevent cross-run collisions
	testID := uuid.New().String()
	testOwner := fmt.Sprintf("integ-owner-%s", testID)
	testRepo := fmt.Sprintf("integ-repo-%s", testID)

	shaBytes := make([]byte, 20)
	if _, err := crand.Read(shaBytes); err != nil {
		t.Fatalf("failed to generate random commit SHA: %v", err)
	}
	commitSHA := hex.EncodeToString(shaBytes)

	// Clean up all test data on completion (cascades to snapshots, files, symbols, and jobs)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := pgDB.DB.ExecContext(cleanupCtx, "DELETE FROM repositories WHERE owner = $1 AND name = $2", testOwner, testRepo); err != nil {
			t.Logf("cleanup failed for repository %s/%s: %v", testOwner, testRepo, err)
		}
	})

	// Test real end-to-end ingestion and symbol persistence with live PostgreSQL
	goCode := `package main

type ServerConfig struct {
	Port int
	Host string
}

func StartServer() error {
	return nil
}
`
	files := []git.FileEntry{
		{Path: "server.go", Extension: ".go", Language: "Go", SizeBytes: len(goCode), LineCount: 11, SHA256Hash: "hash-real", Content: goCode},
	}
	fetcher := git.NewMockFetcher(commitSHA, files)
	broker := sse.NewBroker(nil)
	jm := jobs.NewJobManager(pgDB.DB, 1, 5, logger, broker)
	svc := repo.NewService(pgDB.DB, fetcher, jm, logger)

	job, snap, err := svc.Ingest(ctx, testOwner, testRepo, "main")
	if err != nil {
		t.Fatalf("ingest failed on real postgres: %v", err)
	}
	if job == nil || snap == nil {
		t.Fatal("expected non-nil job and snapshot")
	}

	// Run ProcessIngestion against real PostgreSQL
	if err := svc.ProcessIngestion(ctx, job.ID); err != nil {
		t.Fatalf("ProcessIngestion failed on real postgres: %v", err)
	}

	// Verify symbols query against real PostgreSQL tables and constraints
	symbols, err := svc.GetSnapshotSymbols(ctx, testOwner, testRepo, commitSHA, "")
	if err != nil {
		t.Fatalf("failed to query symbols from postgres: %v", err)
	}

	if len(symbols) < 2 {
		t.Errorf("expected at least 2 symbols persisted in postgres, got %d", len(symbols))
	}

	// Verify idempotency on real PostgreSQL upsert (idx_code_symbols_file_name_kind_line)
	if err := svc.ProcessIngestion(ctx, job.ID); err != nil {
		t.Fatalf("re-running ProcessIngestion against real postgres failed: %v", err)
	}
}
