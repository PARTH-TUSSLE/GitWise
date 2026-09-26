package postgres_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/storage/migrations"
	"github.com/gitwise/backend/internal/storage/postgres"
)

func TestPostgres_ConnectionFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Invalid port should fail clearly, close sql.DB, and return a nil *DB instance
	db, err := postgres.New(ctx, "postgres://invalid:invalid@localhost:59999/invalid?sslmode=disable", logger)
	if err == nil {
		t.Fatal("expected error connecting to non-existent postgres instance, got nil")
	}
	if db != nil {
		t.Errorf("expected db to be nil on connection failure, got: %v", db)
	}
}

func TestPostgres_Integration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping live postgres test: DATABASE_URL not set")
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db, err := postgres.New(ctx, dbURL, logger)
	if err != nil {
		t.Skipf("Skipping live postgres test (instance not running): %v", err)
	}
	defer db.Close()

	if err := db.CheckHealth(ctx); err != nil {
		t.Fatalf("database health check failed: %v", err)
	}

	if err := db.ApplyMigrations(ctx, migrations.InitSchemaUp); err != nil {
		t.Fatalf("failed to apply Phase 1 migrations: %v", err)
	}
}
