package graph

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

// FeatureTraceBuilder constructs grounded end-to-end execution flows across subsystems.
type FeatureTraceBuilder struct {
	db *sql.DB
}

// NewFeatureTraceBuilder creates a new trace builder.
func NewFeatureTraceBuilder(db *sql.DB) *FeatureTraceBuilder {
	return &FeatureTraceBuilder{db: db}
}

// BuildFeatureTraces derives verifiable feature traces from real snapshot files and symbols.
func (b *FeatureTraceBuilder) BuildFeatureTraces(ctx context.Context, snapshotID uuid.UUID) ([]domain.FeatureTrace, error) {
	if b.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// 1. Query available key files in the snapshot
	query := `
		SELECT rf.id, rf.path, rf.language, COALESCE(rf.content, '')
		FROM repository_files rf
		WHERE rf.snapshot_id = $1
		ORDER BY rf.path ASC`

	rows, err := b.db.QueryContext(ctx, query, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshot files for traces: %w", err)
	}
	defer rows.Close()

	type fileInfo struct {
		id       uuid.UUID
		path     string
		language string
		content  string
	}

	fileMap := make(map[string]fileInfo)
	for rows.Next() {
		var fi fileInfo
		if err := rows.Scan(&fi.id, &fi.path, &fi.language, &fi.content); err != nil {
			return nil, err
		}
		fileMap[fi.path] = fi
	}

	// 2. Query symbols for line numbers and signatures
	symQuery := `
		SELECT cs.file_id, rf.path, cs.name, cs.kind, cs.start_line, COALESCE(cs.signature, '')
		FROM code_symbols cs
		JOIN repository_files rf ON cs.file_id = rf.id
		WHERE cs.snapshot_id = $1
		ORDER BY cs.start_line ASC`

	symRows, err := b.db.QueryContext(ctx, symQuery, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to query symbols for traces: %w", err)
	}
	defer symRows.Close()

	type symInfo struct {
		name      string
		kind      string
		startLine int
		signature string
	}
	fileSymbols := make(map[string][]symInfo)
	for symRows.Next() {
		var fileID uuid.UUID
		var filePath, name, kind, sig string
		var startLine int
		if err := symRows.Scan(&fileID, &filePath, &name, &kind, &startLine, &sig); err != nil {
			return nil, err
		}
		fileSymbols[filePath] = append(fileSymbols[filePath], symInfo{
			name:      name,
			kind:      kind,
			startLine: startLine,
			signature: sig,
		})
	}

	var traces []domain.FeatureTrace

	// Trace 1: HTTP Request Lifecycle / API Ingestion Path
	var lifecycleSteps []domain.FeatureTraceStep
	stepNum := 1

	// Look for entrypoint (cmd/server/main.go, main.go, or src/index.ts)
	var entryFile *fileInfo
	for p, fi := range fileMap {
		if strings.HasSuffix(p, "main.go") || p == "src/index.ts" || p == "index.ts" || p == "src/main.ts" {
			entryFile = &fi
			break
		}
	}

	if entryFile != nil {
		line := 1
		code := ""
		if syms, ok := fileSymbols[entryFile.path]; ok && len(syms) > 0 {
			line = syms[0].startLine
			code = syms[0].signature
		}
		if code == "" && entryFile.content != "" {
			code = getFirstNonEmptyLine(entryFile.content)
		}
		lifecycleSteps = append(lifecycleSteps, domain.FeatureTraceStep{
			Step:        stepNum,
			Title:       "Application Entrypoint & Signal Handling",
			Subsystem:   classifySubsystemPath(entryFile.path),
			File:        entryFile.path,
			Line:        line,
			Description: "Root process initialization, environment configuration, and service assembly.",
			CodeSnippet: code,
		})
		stepNum++
	}

	// Look for routing / middleware layer
	var routerFile *fileInfo
	for p, fi := range fileMap {
		if strings.Contains(p, "router.go") || strings.Contains(p, "routes.go") || strings.Contains(p, "app/page.tsx") {
			routerFile = &fi
			break
		}
	}

	if routerFile != nil {
		line := 1
		code := ""
		if syms, ok := fileSymbols[routerFile.path]; ok && len(syms) > 0 {
			line = syms[0].startLine
			code = syms[0].signature
		}
		if code == "" && routerFile.content != "" {
			code = getFirstNonEmptyLine(routerFile.content)
		}
		lifecycleSteps = append(lifecycleSteps, domain.FeatureTraceStep{
			Step:        stepNum,
			Title:       "Request Routing & Middleware Execution",
			Subsystem:   classifySubsystemPath(routerFile.path),
			File:        routerFile.path,
			Line:        line,
			Description: "Dispatch of inbound requests through middleware chain to target controller.",
			CodeSnippet: code,
		})
		stepNum++
	}

	// Look for domain service / core logic layer
	var serviceFile *fileInfo
	for p, fi := range fileMap {
		if strings.Contains(p, "service.go") || strings.Contains(p, "controller") || strings.Contains(p, "handler") {
			serviceFile = &fi
			break
		}
	}

	if serviceFile != nil && (routerFile == nil || serviceFile.path != routerFile.path) {
		line := 1
		code := ""
		if syms, ok := fileSymbols[serviceFile.path]; ok && len(syms) > 0 {
			line = syms[0].startLine
			code = syms[0].signature
		}
		if code == "" && serviceFile.content != "" {
			code = getFirstNonEmptyLine(serviceFile.content)
		}
		lifecycleSteps = append(lifecycleSteps, domain.FeatureTraceStep{
			Step:        stepNum,
			Title:       "Domain Logic & Pipeline Execution",
			Subsystem:   classifySubsystemPath(serviceFile.path),
			File:        serviceFile.path,
			Line:        line,
			Description: "Business validation, state transformation, and dependency invocation.",
			CodeSnippet: code,
		})
		stepNum++
	}

	// Look for storage / persistence layer
	var storageFile *fileInfo
	for p, fi := range fileMap {
		if strings.Contains(p, "postgres") || strings.Contains(p, "storage") || strings.Contains(p, "db.go") || strings.Contains(p, "repository") {
			storageFile = &fi
			break
		}
	}

	if storageFile != nil && (serviceFile == nil || storageFile.path != serviceFile.path) {
		line := 1
		code := ""
		if syms, ok := fileSymbols[storageFile.path]; ok && len(syms) > 0 {
			line = syms[0].startLine
			code = syms[0].signature
		}
		if code == "" && storageFile.content != "" {
			code = getFirstNonEmptyLine(storageFile.content)
		}
		lifecycleSteps = append(lifecycleSteps, domain.FeatureTraceStep{
			Step:        stepNum,
			Title:       "Data Persistence & Query Execution",
			Subsystem:   classifySubsystemPath(storageFile.path),
			File:        storageFile.path,
			Line:        line,
			Description: "ACID transactional write or indexed relational query execution.",
			CodeSnippet: code,
		})
		stepNum++
	}

	if len(lifecycleSteps) > 0 {
		traces = append(traces, domain.FeatureTrace{
			ID:          "trace-core-flow",
			Name:        "Core Execution Lifecycle",
			Description: "End-to-end execution flow from entrypoint through routing, domain services, and persistence.",
			Steps:       lifecycleSteps,
		})
	}

	// If no specialized files match, produce a fallback trace from the first available files
	if len(traces) == 0 && len(fileMap) > 0 {
		var fallbackSteps []domain.FeatureTraceStep
		i := 1
		for _, fi := range fileMap {
			if i > 3 {
				break
			}
			fallbackSteps = append(fallbackSteps, domain.FeatureTraceStep{
				Step:        i,
				Title:       fmt.Sprintf("Module Execution Step %d", i),
				Subsystem:   classifySubsystemPath(fi.path),
				File:        fi.path,
				Line:        1,
				Description: fmt.Sprintf("Source file module execution in %s.", fi.path),
				CodeSnippet: getFirstNonEmptyLine(fi.content),
			})
			i++
		}
		traces = append(traces, domain.FeatureTrace{
			ID:          "trace-module-flow",
			Name:        "Repository Module Flow",
			Description: "Sequential flow across primary repository modules.",
			Steps:       fallbackSteps,
		})
	}

	return traces, nil
}

// GetFeatureTraceByID returns a single feature trace matching the given ID.
func (b *FeatureTraceBuilder) GetFeatureTraceByID(ctx context.Context, snapshotID uuid.UUID, traceID string) (*domain.FeatureTrace, error) {
	traces, err := b.BuildFeatureTraces(ctx, snapshotID)
	if err != nil {
		return nil, err
	}

	for _, t := range traces {
		if t.ID == traceID {
			return &t, nil
		}
	}

	// If trace ID is generic or default, return first available trace
	if len(traces) > 0 && (traceID == "default" || traceID == "main" || traceID == "core") {
		return &traces[0], nil
	}

	return nil, fmt.Errorf("feature trace %s not found for snapshot", traceID)
}

func getFirstNonEmptyLine(content string) string {
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
			return trimmed
		}
	}
	return ""
}
