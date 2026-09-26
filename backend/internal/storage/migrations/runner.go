package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"
)

//go:embed *.sql
var MigrationFS embed.FS

// Migration represents a single versioned migration script.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Runner handles discovering, sorting, and executing pending migrations.
type Runner struct {
	db     *sql.DB
	logger *slog.Logger
	fsys   fs.FS
}

// NewRunner creates a new migration runner using embedded migrations.
func NewRunner(db *sql.DB, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runner{
		db:     db,
		logger: logger,
		fsys:   MigrationFS,
	}
}

// SetFS overrides the filesystem for testing.
func (r *Runner) SetFS(fsys fs.FS) {
	r.fsys = fsys
}

// LoadMigrations reads and parses all .up.sql files from the specified filesystem.
// Migrations are strictly sorted by version ascending.
func LoadMigrations(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	var migrationList []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		name := entry.Name()
		// Parse version from prefix (e.g. "000001_init_schema.up.sql" -> 1)
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid migration filename format: %s", name)
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in filename %s: %w", name, err)
		}

		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		migrationList = append(migrationList, Migration{
			Version: version,
			Name:    strings.TrimSuffix(name, ".up.sql"),
			SQL:     string(content),
		})
	}

	// Strictly sort by numeric version ascending
	sort.Slice(migrationList, func(i, j int) bool {
		return migrationList[i].Version < migrationList[j].Version
	})

	return migrationList, nil
}

// Run applies all pending migrations in transactional, deterministic order.
func (r *Runner) Run(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 1. Ensure schema_migrations table exists
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	if _, err := r.db.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to initialize schema_migrations table: %w", err)
	}

	// 2. Discover already applied migrations
	rows, err := r.db.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version ASC;")
	if err != nil {
		return fmt.Errorf("failed to query applied schema migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading applied migrations: %w", err)
	}

	// 3. Load all embedded migrations sorted by version
	available, err := LoadMigrations(r.fsys)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// 4. Apply only pending migrations inside transactions
	appliedCount := 0
	for _, m := range available {
		if applied[m.Version] {
			r.logger.Debug("Migration already applied, skipping",
				slog.Int("version", m.Version),
				slog.String("name", m.Name),
			)
			continue
		}

		r.logger.Info("Applying pending migration",
			slog.Int("version", m.Version),
			slog.String("name", m.Name),
		)

		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", m.Name, err)
		}

		if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s (version %d): %w", m.Name, m.Version, err)
		}

		recordSQL := "INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, NOW());"
		if _, err := tx.ExecContext(ctx, recordSQL, m.Version, m.Name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s in schema_migrations: %w", m.Name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for migration %s: %w", m.Name, err)
		}

		applied[m.Version] = true
		appliedCount++
		r.logger.Info("Applied migration successfully",
			slog.Int("version", m.Version),
			slog.String("name", m.Name),
		)
	}

	if appliedCount == 0 {
		r.logger.Info("Database schema is up to date, 0 pending migrations")
	} else {
		r.logger.Info("Database migrations completed", slog.Int("applied", appliedCount))
	}

	return nil
}

// Run executes pending migrations using the default runner.
func Run(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	return NewRunner(db, logger).Run(ctx)
}
