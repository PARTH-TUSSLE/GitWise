package evidence

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

// Store manages persistence and dynamic hydration of grounded evidence references.
// Storage Axiom: Evidence references persist coordinate anchors and content SHA-256 hashes
// without duplicating raw code strings in evidence_refs table.
type Store struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewStore creates a new Evidence Store.
func NewStore(db *sql.DB, logger *slog.Logger) *Store {
	if logger == nil {
		logger = slog.Default()
	}
	return &Store{
		db:     db,
		logger: logger,
	}
}

// ComputeHash calculates deterministic SHA-256 hex string of code snippet content.
func ComputeHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// ExtractSnippet extracts 1-indexed lines [startLine, endLine] from full file content.
func ExtractSnippet(fileContent string, startLine, endLine int) (string, error) {
	if startLine < 1 {
		startLine = 1
	}
	// Normalize carriage returns
	fileContent = strings.ReplaceAll(fileContent, "\r\n", "\n")
	lines := strings.Split(fileContent, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return "", nil
	}
	if startLine > len(lines) {
		return "", fmt.Errorf("start line %d exceeds file line count %d", startLine, len(lines))
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if endLine < startLine {
		endLine = startLine
	}
	selected := lines[startLine-1 : endLine]
	return strings.Join(selected, "\n"), nil
}

// PersistRefsBatch saves evidence reference anchors to PostgreSQL.
func (s *Store) PersistRefsBatch(ctx context.Context, refs []domain.EvidenceRef) error {
	if len(refs) == 0 || s.db == nil {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx for evidence refs: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO evidence_refs (
			id, snapshot_id, file_id, start_line, end_line, content_hash, provenance, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (id) DO UPDATE SET
			content_hash = EXCLUDED.content_hash,
			provenance = EXCLUDED.provenance;`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare evidence refs insert: %w", err)
	}
	defer stmt.Close()

	for _, ref := range refs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := stmt.ExecContext(ctx,
			ref.ID,
			ref.SnapshotID,
			ref.FileID,
			ref.StartLine,
			ref.EndLine,
			ref.ContentHash,
			ref.Provenance,
		)
		if err != nil {
			return fmt.Errorf("failed to insert evidence ref %s: %w", ref.ID, err)
		}
	}

	return tx.Commit()
}

// HydrateEvidence populates FilePath and Snippet for given evidence refs by querying repository_files.
// Validates snippet content hash against stored ContentHash.
func (s *Store) HydrateEvidence(ctx context.Context, refs []domain.EvidenceRef) ([]domain.EvidenceRef, error) {
	if len(refs) == 0 || s.db == nil {
		return refs, nil
	}

	// 1. Gather unique file IDs
	fileIDMap := make(map[uuid.UUID]bool)
	for _, ref := range refs {
		fileIDMap[ref.FileID] = true
	}

	type fileData struct {
		path    string
		content string
	}
	files := make(map[uuid.UUID]fileData)

	for fileID := range fileIDMap {
		var path string
		var content sql.NullString
		err := s.db.QueryRowContext(ctx, `
			SELECT path, content FROM repository_files WHERE id = $1
		`, fileID).Scan(&path, &content)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to query repository file for evidence hydration", "file_id", fileID, "error", err)
			continue
		}
		files[fileID] = fileData{
			path:    path,
			content: content.String,
		}
	}

	// 2. Hydrate each ref
	hydrated := make([]domain.EvidenceRef, len(refs))
	for i, ref := range refs {
		hydrated[i] = ref
		fd, exists := files[ref.FileID]
		if !exists {
			continue
		}
		hydrated[i].FilePath = fd.path

		snippet, err := ExtractSnippet(fd.content, ref.StartLine, ref.EndLine)
		if err != nil {
			s.logger.WarnContext(ctx, "failed to extract snippet for evidence ref", "ref_id", ref.ID, "error", err)
			continue
		}
		hydrated[i].Snippet = snippet

		// Check hash integrity
		actualHash := ComputeHash(snippet)
		if ref.ContentHash != "" && actualHash != ref.ContentHash {
			s.logger.WarnContext(ctx, "evidence snippet hash mismatch",
				"ref_id", ref.ID,
				"expected", ref.ContentHash,
				"actual", actualHash,
			)
		}
	}

	return hydrated, nil
}

// AssembleEvidencePackage hydrates refs and bundles them into an EvidencePackage with sequential citation tags (ev_01, ev_02, ...).
func (s *Store) AssembleEvidencePackage(
	ctx context.Context,
	snapshotID uuid.UUID,
	commitSHA string,
	query string,
	refs []domain.EvidenceRef,
) (*domain.EvidencePackage, error) {
	hydrated, err := s.HydrateEvidence(ctx, refs)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate evidence package: %w", err)
	}

	// Normalize citation tags to sequential anchors for prompt context clarity
	for i := range hydrated {
		hydrated[i].ID = fmt.Sprintf("ev_%02d", i+1)
	}

	return &domain.EvidencePackage{
		SnapshotID: snapshotID,
		CommitSHA:  commitSHA,
		Query:      query,
		Items:      hydrated,
		TotalItems: len(hydrated),
	}, nil
}

// GetRefsBySnapshot retrieves all evidence references recorded for a snapshot.
func (s *Store) GetRefsBySnapshot(ctx context.Context, snapshotID uuid.UUID) ([]domain.EvidenceRef, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT er.id, er.snapshot_id, er.file_id, rf.path, er.start_line, er.end_line, er.content_hash, er.provenance, er.created_at
		FROM evidence_refs er
		JOIN repository_files rf ON er.file_id = rf.id
		WHERE er.snapshot_id = $1
		ORDER BY er.created_at ASC;`

	rows, err := s.db.QueryContext(ctx, query, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to query evidence refs: %w", err)
	}
	defer rows.Close()

	var refs []domain.EvidenceRef
	for rows.Next() {
		var ref domain.EvidenceRef
		if err := rows.Scan(
			&ref.ID,
			&ref.SnapshotID,
			&ref.FileID,
			&ref.FilePath,
			&ref.StartLine,
			&ref.EndLine,
			&ref.ContentHash,
			&ref.Provenance,
			&ref.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan evidence ref: %w", err)
		}
		refs = append(refs, ref)
	}

	return refs, nil
}
