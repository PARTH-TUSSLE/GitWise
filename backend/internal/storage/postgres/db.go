package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
	logger *slog.Logger
}

// New creates a new database connection pool.
func New(ctx context.Context, connStr string, logger *slog.Logger) (*DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	logger.Info("Connected to PostgreSQL successfully")
	return &DB{DB: db, logger: logger}, nil
}

// CheckHealth checks if the database is reachable.
func (d *DB) CheckHealth(ctx context.Context) error {
	if d == nil || d.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return d.PingContext(pingCtx)
}

// ApplyMigrations executes raw SQL migrations on startup if connected.
func (d *DB) ApplyMigrations(ctx context.Context, migrationSQL string) error {
	if d == nil || d.DB == nil {
		return fmt.Errorf("cannot apply migrations: database is nil")
	}
	_, err := d.ExecContext(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	d.logger.Info("Database migrations applied successfully")
	return nil
}
