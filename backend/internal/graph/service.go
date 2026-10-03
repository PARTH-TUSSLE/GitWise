package graph

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

// Service provides code intelligence graph persistence and depth-bounded candidate impact traversal.
type Service struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewService creates a new graph service.
func NewService(db *sql.DB, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		db:     db,
		logger: logger,
	}
}

// PersistEdges inserts dependency edges into the database with idempotency conflict handling.
func (s *Service) PersistEdges(ctx context.Context, snapshotID uuid.UUID, edges []domain.DependencyEdge) error {
	if len(edges) == 0 || s.db == nil {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction for dependency edges: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO dependency_edges (
			snapshot_id, source_file_id, target_file_id, edge_type, is_deterministic, line_number, raw_target, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (
			snapshot_id, source_file_id, COALESCE(target_file_id, '00000000-0000-0000-0000-000000000000'::uuid),
			edge_type, COALESCE(line_number, 0), COALESCE(raw_target, '')
		) DO NOTHING`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare edge insert statement: %w", err)
	}
	defer stmt.Close()

	for _, e := range edges {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var lineNumVal sql.NullInt32
		if e.LineNumber != nil {
			lineNumVal = sql.NullInt32{Int32: int32(*e.LineNumber), Valid: true}
		}

		_, err := stmt.ExecContext(ctx,
			snapshotID,
			e.SourceFileID,
			e.TargetFileID,
			string(e.EdgeType),
			e.IsDeterministic,
			lineNumVal,
			e.RawTarget,
		)
		if err != nil {
			return fmt.Errorf("failed to insert dependency edge: %w", err)
		}
	}

	return tx.Commit()
}

// GetCandidateImpact performs depth-bounded recursive Common Table Expression (CTE)
// traversal in PostgreSQL to calculate the candidate impact blast radius for a file.
// Hard boundaries: Max depth <= 3, cycle detection array, and max 50 candidate nodes.
func (s *Service) GetCandidateImpact(ctx context.Context, snapshotID uuid.UUID, filePath string) (*domain.CandidateImpactReport, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	cleanPath := filepath.ToSlash(filepath.Clean(filePath))
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	// 1. Resolve target file
	var targetFileID uuid.UUID
	fileQuery := `SELECT id FROM repository_files WHERE snapshot_id = $1 AND path = $2`
	err := s.db.QueryRowContext(ctx, fileQuery, snapshotID, cleanPath).Scan(&targetFileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file %s not found in snapshot", cleanPath)
		}
		return nil, fmt.Errorf("failed to query target file: %w", err)
	}

	// 2. Transitive Downstream Candidates (Files that depend on targetFileID, depth <= 3)
	downstreamQuery := `
		WITH RECURSIVE impact_graph AS (
			-- Base case: direct dependents of the target file
			SELECT
				e.source_file_id AS file_id,
				e.edge_type,
				e.is_deterministic,
				1 AS depth,
				ARRAY[e.target_file_id, e.source_file_id] AS visited
			FROM dependency_edges e
			WHERE e.snapshot_id = $1
			  AND e.target_file_id = $2
			
			UNION ALL
			
			-- Recursive step: files that import the dependent files
			SELECT
				e.source_file_id AS file_id,
				e.edge_type,
				e.is_deterministic,
				ig.depth + 1 AS depth,
				ig.visited || e.source_file_id AS visited
			FROM dependency_edges e
			JOIN impact_graph ig ON e.target_file_id = ig.file_id
			WHERE e.snapshot_id = $1
			  AND ig.depth < 3
			  AND NOT (e.source_file_id = ANY(ig.visited))
		)
		SELECT DISTINCT ON (rf.id)
			rf.id,
			rf.path,
			ig.depth,
			ig.edge_type,
			ig.is_deterministic
		FROM impact_graph ig
		JOIN repository_files rf ON ig.file_id = rf.id
		ORDER BY rf.id, ig.depth ASC
		LIMIT 50;`

	downRows, err := s.db.QueryContext(ctx, downstreamQuery, snapshotID, targetFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute downstream impact traversal: %w", err)
	}
	defer downRows.Close()

	var downstream []domain.CandidateNode
	directDependents := make([]string, 0)

	for downRows.Next() {
		var n domain.CandidateNode
		var edgeTypeStr string
		if err := downRows.Scan(&n.FileID, &n.Path, &n.Depth, &edgeTypeStr, &n.IsDeterministic); err != nil {
			return nil, err
		}
		n.EdgeType = domain.EdgeType(edgeTypeStr)
		n.Subsystem = classifySubsystemPath(n.Path)
		downstream = append(downstream, n)
		if n.Depth == 1 {
			directDependents = append(directDependents, n.Path)
		}
	}

	// 3. Transitive Upstream Dependencies (Files that targetFileID imports, depth <= 3)
	upstreamQuery := `
		WITH RECURSIVE dep_graph AS (
			-- Base case: direct dependencies of the target file
			SELECT
				e.target_file_id AS file_id,
				e.edge_type,
				e.is_deterministic,
				1 AS depth,
				ARRAY[e.source_file_id, e.target_file_id] AS visited
			FROM dependency_edges e
			WHERE e.snapshot_id = $1
			  AND e.source_file_id = $2
			  AND e.target_file_id IS NOT NULL
			
			UNION ALL
			
			-- Recursive step: files imported by the dependencies
			SELECT
				e.target_file_id AS file_id,
				e.edge_type,
				e.is_deterministic,
				dg.depth + 1 AS depth,
				dg.visited || e.target_file_id AS visited
			FROM dependency_edges e
			JOIN dep_graph dg ON e.source_file_id = dg.file_id
			WHERE e.snapshot_id = $1
			  AND dg.depth < 3
			  AND e.target_file_id IS NOT NULL
			  AND NOT (e.target_file_id = ANY(dg.visited))
		)
		SELECT DISTINCT ON (rf.id)
			rf.id,
			rf.path,
			dg.depth,
			dg.edge_type,
			dg.is_deterministic
		FROM dep_graph dg
		JOIN repository_files rf ON dg.file_id = rf.id
		ORDER BY rf.id, dg.depth ASC
		LIMIT 50;`

	upRows, err := s.db.QueryContext(ctx, upstreamQuery, snapshotID, targetFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute upstream dependency traversal: %w", err)
	}
	defer upRows.Close()

	var upstream []domain.CandidateNode
	directImports := make([]string, 0)

	for upRows.Next() {
		var n domain.CandidateNode
		var edgeTypeStr string
		if err := upRows.Scan(&n.FileID, &n.Path, &n.Depth, &edgeTypeStr, &n.IsDeterministic); err != nil {
			return nil, err
		}
		n.EdgeType = domain.EdgeType(edgeTypeStr)
		n.Subsystem = classifySubsystemPath(n.Path)
		upstream = append(upstream, n)
		if n.Depth == 1 {
			directImports = append(directImports, n.Path)
		}
	}

	// 4. Query Exported Code Symbols for the target file
	symbolsQuery := `
		SELECT id, snapshot_id, file_id, name, kind, start_line, end_line, signature, is_exported, created_at
		FROM code_symbols
		WHERE file_id = $1 AND is_exported = true
		ORDER BY start_line ASC`

	symRows, err := s.db.QueryContext(ctx, symbolsQuery, targetFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query affected symbols: %w", err)
	}
	defer symRows.Close()

	var symbols []domain.CodeSymbol
	for symRows.Next() {
		var sym domain.CodeSymbol
		var kindStr string
		var sig sql.NullString
		if err := symRows.Scan(
			&sym.ID, &sym.SnapshotID, &sym.FileID,
			&sym.Name, &kindStr, &sym.StartLine, &sym.EndLine,
			&sig, &sym.IsExported, &sym.CreatedAt,
		); err != nil {
			return nil, err
		}
		sym.FilePath = cleanPath
		sym.Kind = domain.SymbolKind(kindStr)
		if sig.Valid {
			sym.Signature = sig.String
		}
		symbols = append(symbols, sym)
	}

	depthCapped := len(downstream) >= 50
	for _, n := range downstream {
		if n.Depth >= 3 {
			depthCapped = true
			break
		}
	}

	if directImports == nil {
		directImports = []string{}
	}
	if directDependents == nil {
		directDependents = []string{}
	}
	if downstream == nil {
		downstream = []domain.CandidateNode{}
	}
	if upstream == nil {
		upstream = []domain.CandidateNode{}
	}
	if symbols == nil {
		symbols = []domain.CodeSymbol{}
	}

	return &domain.CandidateImpactReport{
		TargetFile:           cleanPath,
		TargetFileID:         targetFileID,
		DirectImports:        directImports,
		DirectDependents:     directDependents,
		DownstreamCandidates: downstream,
		UpstreamDependencies: upstream,
		AffectedSymbols:      symbols,
		Subsystem:            classifySubsystemPath(cleanPath),
		DepthCapped:          depthCapped,
		MaxDepth:             3,
		TotalCandidateCount:  len(downstream),
	}, nil
}

// GetSubsystemConnections calculates inter-subsystem directed dependencies based on real import edges.
func (s *Service) GetSubsystemConnections(ctx context.Context, snapshotID uuid.UUID) (map[string][]string, error) {
	query := `
		SELECT rf1.path AS source_path, rf2.path AS target_path
		FROM dependency_edges e
		JOIN repository_files rf1 ON e.source_file_id = rf1.id
		JOIN repository_files rf2 ON e.target_file_id = rf2.id
		WHERE e.snapshot_id = $1
		  AND e.target_file_id IS NOT NULL;`

	rows, err := s.db.QueryContext(ctx, query, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to query edge subsystem connections: %w", err)
	}
	defer rows.Close()

	connectionSet := make(map[string]map[string]bool)
	for rows.Next() {
		var srcPath, tgtPath string
		if err := rows.Scan(&srcPath, &tgtPath); err != nil {
			return nil, err
		}

		srcSub := classifySubsystemID(srcPath)
		tgtSub := classifySubsystemID(tgtPath)

		if srcSub != "" && tgtSub != "" && srcSub != tgtSub {
			if connectionSet[srcSub] == nil {
				connectionSet[srcSub] = make(map[string]bool)
			}
			connectionSet[srcSub][tgtSub] = true
		}
	}

	result := make(map[string][]string, len(connectionSet))
	for src, targets := range connectionSet {
		list := make([]string, 0, len(targets))
		for tgt := range targets {
			list = append(list, tgt)
		}
		result[src] = list
	}

	return result, nil
}

func classifySubsystemPath(p string) string {
	clean := filepath.ToSlash(filepath.Clean(p))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	parts := strings.Split(clean, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "Root / Config"
	}
	first := parts[0]
	switch first {
	case "cmd":
		return "CLI / Entrypoint"
	case "internal", "pkg", "backend":
		if first == "backend" && len(parts) > 1 {
			switch parts[1] {
			case "cmd":
				return "CLI / Entrypoint"
			case "internal", "pkg":
				if len(parts) > 2 {
					return fmt.Sprintf("Internal / %s", strings.Title(parts[2]))
				}
				return "Internal Architecture"
			}
		}
		if len(parts) > 1 {
			return fmt.Sprintf("Internal / %s", strings.Title(parts[1]))
		}
		return "Internal Architecture"
	case "api":
		return "API Layer"
	case "src":
		if len(parts) > 1 {
			switch parts[1] {
			case "components":
				return "Frontend UI Components"
			case "app", "pages":
				return "Frontend Application Routes"
			case "lib", "utils":
				return "Core Utilities & Libraries"
			}
		}
		return "Frontend Core"
	case "components":
		return "Frontend UI Components"
	case "lib", "utils":
		return "Core Utilities"
	case "test", "tests", "testdata":
		return "Testing & Quality Assurance"
	case "docs":
		return "Documentation"
	default:
		return "Root / Config"
	}
}

func classifySubsystemID(p string) string {
	clean := filepath.ToSlash(filepath.Clean(p))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	parts := strings.Split(clean, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "core-root"
	}
	first := parts[0]
	switch first {
	case "cmd":
		return "core-cmd"
	case "internal", "pkg":
		if len(parts) > 1 {
			return fmt.Sprintf("internal-%s", strings.ToLower(parts[1]))
		}
		return "internal-core"
	case "backend":
		if len(parts) > 1 {
			if parts[1] == "cmd" {
				return "core-cmd"
			}
			if len(parts) > 2 {
				return fmt.Sprintf("internal-%s", strings.ToLower(parts[2]))
			}
			return "internal-backend"
		}
		return "internal-backend"
	case "src":
		if len(parts) > 1 {
			return fmt.Sprintf("src-%s", strings.ToLower(parts[1]))
		}
		return "src-core"
	default:
		return fmt.Sprintf("subsystem-%s", strings.ToLower(first))
	}
}
