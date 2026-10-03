package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"strings"
	"time"

	"github.com/gitwise/backend/internal/ai"
	"github.com/gitwise/backend/internal/analysis"
	"github.com/gitwise/backend/internal/analysis/golang"
	"github.com/gitwise/backend/internal/analysis/typescript"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/evidence"
	"github.com/gitwise/backend/internal/git"
	"github.com/gitwise/backend/internal/graph"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/gitwise/backend/internal/retrieval"
	"github.com/gitwise/backend/internal/service/mentor"
	"github.com/google/uuid"
)

// Service coordinates repository ingestion, commit snapshots, and file persistence.
type Service struct {
	db               *sql.DB
	fetcher          git.Fetcher
	jobManager       *jobs.JobManager
	analyzerRegistry *analysis.Registry
	graphSvc         *graph.Service
	graphResolver    *graph.Resolver
	traceBuilder     *graph.FeatureTraceBuilder
	retrievalSvc     *retrieval.Service
	chunker          *retrieval.Chunker
	evidenceStore    *evidence.Store
	mentorSvc        *mentor.Service
	logger           *slog.Logger
}

// NewService creates a new repository service instance.
func NewService(
	db *sql.DB,
	fetcher git.Fetcher,
	jobManager *jobs.JobManager,
	logger *slog.Logger,
	optionalAI ...ai.Client,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	registry := analysis.NewRegistry()
	registry.Register(golang.NewGoAnalyzer())
	registry.Register(typescript.NewTypeScriptAnalyzer())

	mockEmbedder := retrieval.NewMockEmbedder()
	retrievalSvc := retrieval.NewService(db, mockEmbedder, logger)
	chunker := retrieval.NewChunker()
	evidenceStore := evidence.NewStore(db, logger)
	var aiClient ai.Client = ai.NewMockClient()
	if len(optionalAI) > 0 && optionalAI[0] != nil {
		aiClient = optionalAI[0]
	}
	mentorSvc := mentor.NewService(db, retrievalSvc, evidenceStore, aiClient, logger)

	svc := &Service{
		db:               db,
		fetcher:          fetcher,
		jobManager:       jobManager,
		analyzerRegistry: registry,
		graphSvc:         graph.NewService(db, logger),
		graphResolver:    graph.NewResolver(),
		traceBuilder:     graph.NewFeatureTraceBuilder(db),
		retrievalSvc:     retrievalSvc,
		chunker:          chunker,
		evidenceStore:    evidenceStore,
		mentorSvc:        mentorSvc,
		logger:           logger,
	}

	// Register the task handler for SNAPSHOT_INGEST
	if jobManager != nil {
		jobManager.RegisterHandler(domain.JobTypeSnapshotIngest, svc.ProcessIngestion)
	}

	return svc
}

// SetAIClient updates the active AI client at runtime.
func (s *Service) SetAIClient(client ai.Client) {
	if client != nil && s.mentorSvc != nil {
		s.mentorSvc.SetAIClient(client)
	}
}

// AIClient returns the active AI client.
func (s *Service) AIClient() ai.Client {
	if s.mentorSvc != nil {
		return s.mentorSvc.AIClient()
	}
	return nil
}

// SetAnalyzerRegistry allows setting a custom analyzer registry (for testing or plugins).
func (s *Service) SetAnalyzerRegistry(r *analysis.Registry) {
	s.analyzerRegistry = r
}

// SetGraphService allows setting a custom graph service (for testing).
func (s *Service) SetGraphService(g *graph.Service) {
	s.graphSvc = g
}

// SetTraceBuilder allows setting a custom feature trace builder (for testing).
func (s *Service) SetTraceBuilder(tb *graph.FeatureTraceBuilder) {
	s.traceBuilder = tb
}

// SetRetrievalService allows setting a custom retrieval service (for testing).
func (s *Service) SetRetrievalService(r *retrieval.Service) {
	s.retrievalSvc = r
}

// SetChunker allows setting a custom code chunker (for testing).
func (s *Service) SetChunker(c *retrieval.Chunker) {
	s.chunker = c
}

// SetEvidenceStore allows setting a custom evidence store (for testing).
func (s *Service) SetEvidenceStore(e *evidence.Store) {
	s.evidenceStore = e
}

// SetMentorService allows setting a custom mentor service (for testing).
func (s *Service) SetMentorService(m *mentor.Service) {
	s.mentorSvc = m
}

// Ingest triggers or returns an existing snapshot ingestion job for a repository.
func (s *Service) Ingest(ctx context.Context, owner, repoName, ref string) (*domain.AnalysisJob, *domain.RepositorySnapshot, error) {
	if owner == "" || repoName == "" {
		return nil, nil, errors.New("owner and repo name are required")
	}
	if ref == "" {
		ref = "main"
	}
	if s.db == nil {
		return nil, nil, errors.New("database is not configured")
	}
	if s.fetcher == nil {
		return nil, nil, errors.New("git fetcher is not configured")
	}

	// 1. Resolve or create repository record
	repo, err := s.getOrCreateRepo(ctx, owner, repoName, ref)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register repository: %w", err)
	}

	// 2. Resolve commit ref to immutable 40-character SHA
	commitSHA, err := s.fetcher.ResolveCommitSHA(ctx, owner, repoName, ref)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve commit SHA for ref %s: %w", ref, err)
	}

	// 3. Check for existing snapshot
	existingSnap, err := s.getSnapshot(ctx, repo.ID, commitSHA)
	if err == nil && existingSnap != nil {
		if existingSnap.Status == domain.SnapshotStatusReady || existingSnap.Status == domain.SnapshotStatusPartial {
			// Idempotent: Snapshot already ingested and ready or partial
			dummyJob := &domain.AnalysisJob{
				ID:              uuid.New(),
				Type:            domain.JobTypeSnapshotIngest,
				SnapshotID:      &existingSnap.ID,
				Status:          domain.JobStatusCompleted,
				Stage:           domain.JobStageDone,
				ProgressPercent: 100.0,
				CreatedAt:       existingSnap.CreatedAt,
				UpdatedAt:       time.Now().UTC(),
			}
			return dummyJob, existingSnap, nil
		}

		// Check if an active job already exists for this snapshot
		activeJob, err := s.getActiveJobForSnapshot(ctx, existingSnap.ID)
		if err == nil && activeJob != nil {
			s.logger.Info("Reusing existing active job for snapshot",
				slog.String("snapshot_id", existingSnap.ID.String()),
				slog.String("job_id", activeJob.ID.String()),
				slog.String("status", string(activeJob.Status)),
			)
			return activeJob, existingSnap, nil
		}
	}

	// 4. Create or reset snapshot record in QUEUED state
	snapshot, err := s.createOrResetSnapshot(ctx, repo.ID, commitSHA, ref)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create snapshot record: %w", err)
	}

	// Double check if an active job was created concurrently
	if activeJob, err := s.getActiveJobForSnapshot(ctx, snapshot.ID); err == nil && activeJob != nil {
		return activeJob, snapshot, nil
	}

	// 5. Create background analysis job
	job, err := s.createJob(ctx, snapshot.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create analysis job: %w", err)
	}

	// 6. Enqueue job for background processing
	if s.jobManager != nil && job.Status == domain.JobStatusQueued {
		if err := s.jobManager.Enqueue(ctx, job.ID); err != nil {
			s.logger.Error("Failed to enqueue ingestion job", slog.String("job_id", job.ID.String()), slog.String("error", err.Error()))
			_ = s.failJobDirect(ctx, job.ID, "Queue is full, please retry shortly")
			return nil, nil, fmt.Errorf("failed to enqueue job: %w", err)
		}
	}

	return job, snapshot, nil
}

// ProcessIngestion executes the background repository ingestion pipeline.
func (s *Service) ProcessIngestion(ctx context.Context, jobID uuid.UUID) error {
	startTime := time.Now()

	// 1. Load job
	if s.jobManager == nil {
		return errors.New("job manager is nil")
	}
	job, err := s.jobManager.GetJob(ctx, jobID)
	if err != nil {
		s.failJobAndSnapshot(jobID, nil, err.Error())
		return fmt.Errorf("failed to load job: %w", err)
	}
	if job.SnapshotID == nil {
		s.failJobAndSnapshot(jobID, nil, "job has no associated snapshot ID")
		return errors.New("job has no associated snapshot ID")
	}

	// 2. Load snapshot and repo
	snapshot, repo, err := s.getSnapshotAndRepoByID(ctx, *job.SnapshotID)
	if err != nil {
		s.failJobAndSnapshot(jobID, job.SnapshotID, err.Error())
		return fmt.Errorf("failed to load snapshot or repository: %w", err)
	}

	// 3. Mark snapshot and job as PROCESSING
	_ = s.updateSnapshotStatus(ctx, snapshot.ID, domain.SnapshotStatusProcessing)
	_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageFetchingTree, 15.0, 0, 0, "Fetching git tree from repository")

	// 4. Fetch tree
	fetchRes, err := s.fetcher.FetchTree(ctx, repo.Owner, repo.Name, snapshot.CommitSHA)
	if err != nil {
		s.failJobAndSnapshot(jobID, &snapshot.ID, err.Error())
		return fmt.Errorf("failed to fetch git tree: %w", err)
	}

	files := fetchRes.Files

	_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageFilteringFiles, 50.0, len(files), len(files), "Filtering safety bounds and analyzing files")

	// 5. Calculate statistics and primary language
	totalLines := 0
	langCounts := make(map[string]int)
	for _, f := range files {
		totalLines += f.LineCount
		if f.Language != "" && f.Language != "Plain Text" {
			langCounts[f.Language] += f.LineCount
		}
	}

	primaryLang := "Plain Text"
	maxLines := -1
	for lang, count := range langCounts {
		// Deterministic tie-break rule: highest line count wins; on equal counts,
		// the lexicographically smaller language name (alphabetical order) wins.
		if count > maxLines || (count == maxLines && (primaryLang == "Plain Text" || lang < primaryLang)) {
			maxLines = count
			primaryLang = lang
		}
	}

	// 6. Persist files in database transaction
	_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageFinalizing, 60.0, len(files), len(files), "Persisting repository files into snapshot")

	pathToFileID, err := s.persistFiles(ctx, snapshot.ID, files)
	if err != nil {
		s.failJobAndSnapshot(jobID, &snapshot.ID, err.Error())
		return fmt.Errorf("failed to persist files: %w", err)
	}

	// 7. Structural AST Analysis for Go and TypeScript source files
	_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageAnalyzingAST, 75.0, 0, len(files), "Analyzing Go and TypeScript source structures")

	var allSymbols []domain.CodeSymbol
	analysisMap := make(map[string]*analysis.FileAnalysisResult)
	if s.analyzerRegistry != nil {
		for _, f := range files {
			if f.IsBinary || f.Content == "" {
				continue
			}
			fileID, ok := pathToFileID[f.Path]
			if !ok {
				continue
			}

			res, err := s.analyzerRegistry.Analyze(ctx, f.Path, f.Content)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
					s.failJobAndSnapshot(jobID, &snapshot.ID, fmt.Sprintf("analysis canceled: %v", err))
					return fmt.Errorf("analysis canceled: %w", err)
				}
			}
			if res != nil {
				analysisMap[f.Path] = res
				if len(res.ParseErrors) > 0 {
					s.logger.Warn("Structural analysis warning",
						slog.String("path", f.Path),
						slog.Any("warnings", res.ParseErrors),
					)
				}
				for _, raw := range res.Symbols {
					allSymbols = append(allSymbols, domain.CodeSymbol{
						ID:         uuid.New(),
						SnapshotID: snapshot.ID,
						FileID:     fileID,
						FilePath:   f.Path,
						Name:       raw.Name,
						Kind:       raw.Kind,
						StartLine:  raw.StartLine,
						EndLine:    raw.EndLine,
						Signature:  raw.Signature,
						IsExported: raw.IsExported,
					})
				}
			}
		}

		if err := s.persistSymbols(ctx, snapshot.ID, allSymbols); err != nil {
			s.failJobAndSnapshot(jobID, &snapshot.ID, fmt.Sprintf("failed to persist code symbols: %v", err))
			return fmt.Errorf("failed to persist code symbols: %w", err)
		}
	}

	// Construct snapshot repository files for graph resolution and semantic chunking
	repoFiles := make([]domain.RepositoryFile, len(files))
	for i, f := range files {
		rfID := uuid.Nil
		if id, ok := pathToFileID[f.Path]; ok {
			rfID = id
		}
		contentStr := f.Content
		repoFiles[i] = domain.RepositoryFile{
			ID:         rfID,
			SnapshotID: snapshot.ID,
			Path:       f.Path,
			Extension:  f.Extension,
			Language:   f.Language,
			SizeBytes:  f.SizeBytes,
			LineCount:  f.LineCount,
			SHA256Hash: f.SHA256Hash,
			Content:    &contentStr,
			IsBinary:   f.IsBinary,
		}
	}

	// 7b. Resolve and persist Code Intelligence Graph Dependency Edges
	if s.graphResolver != nil && s.graphSvc != nil {
		edges := s.graphResolver.ResolveSnapshotEdges(snapshot.ID, repoFiles, analysisMap)
		if err := s.graphSvc.PersistEdges(ctx, snapshot.ID, edges); err != nil {
			s.failJobAndSnapshot(jobID, &snapshot.ID, fmt.Sprintf("failed to persist dependency edges: %v", err))
			return fmt.Errorf("failed to persist dependency edges: %w", err)
		}
	}

	// 7c. Semantic Code Chunking, Dense Vector Embeddings, and Chunk Persistence
	if s.retrievalSvc != nil && s.chunker != nil {
		_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageAnalyzingAST, 88.0, len(repoFiles), len(repoFiles), "Generating semantic code chunks and embeddings")
		var allChunks []domain.CodeChunk
		for _, rf := range repoFiles {
			if rf.IsBinary || rf.Content == nil {
				continue
			}
			fileChunks := s.chunker.ChunkFile(rf, allSymbols)
			allChunks = append(allChunks, fileChunks...)
		}

		if err := s.retrievalSvc.PersistChunks(ctx, snapshot.ID, allChunks); err != nil {
			s.failJobAndSnapshot(jobID, &snapshot.ID, fmt.Sprintf("failed to persist code chunks: %v", err))
			return fmt.Errorf("failed to persist code chunks: %w", err)
		}
	}

	// 8. Update snapshot with final facts: READY only if COMPLETE, PARTIAL if capped
	_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageFinalizing, 95.0, len(allSymbols), len(allSymbols), "Finalizing snapshot and architectural subsystems")

	finalSnapshotStatus := domain.SnapshotStatusReady
	if fetchRes.Outcome == git.IngestionOutcomePartial {
		finalSnapshotStatus = domain.SnapshotStatusPartial
	}

	if err := s.finalizeSnapshot(ctx, snapshot.ID, len(files), totalLines, primaryLang, finalSnapshotStatus); err != nil {
		s.failJobAndSnapshot(jobID, &snapshot.ID, err.Error())
		return fmt.Errorf("failed to finalize snapshot: %w", err)
	}

	// 9. Complete job and broadcast completion event
	durationMs := time.Since(startTime).Milliseconds()
	return s.jobManager.CompleteJob(ctx, jobID, snapshot.ID.String(), snapshot.CommitSHA, durationMs)
}

func (s *Service) getOrCreateRepo(ctx context.Context, owner, name, defaultBranch string) (*domain.Repository, error) {
	// Deterministic pseudo-github ID for repo if not fetched
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(owner + "/" + name)))
	githubID := int64(h.Sum64() & 0x7FFFFFFFFFFFFFFF)

	query := `
		INSERT INTO repositories (github_id, owner, name, default_branch, is_private, created_at, updated_at)
		VALUES ($1, $2, $3, $4, false, NOW(), NOW())
		ON CONFLICT (owner, name) DO UPDATE
		SET updated_at = NOW()
		RETURNING id, github_id, owner, name, default_branch, is_private, created_at, updated_at`

	row := s.db.QueryRowContext(ctx, query, githubID, owner, name, defaultBranch)

	var r domain.Repository
	err := row.Scan(&r.ID, &r.GitHubID, &r.Owner, &r.Name, &r.DefaultBranch, &r.IsPrivate, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) getSnapshot(ctx context.Context, repoID uuid.UUID, commitSHA string) (*domain.RepositorySnapshot, error) {
	query := `
		SELECT id, repository_id, commit_sha, ref_name, status, total_files, total_lines, primary_language, analyzed_at, expires_at, created_at
		FROM repository_snapshots
		WHERE repository_id = $1 AND commit_sha = $2`

	row := s.db.QueryRowContext(ctx, query, repoID, commitSHA)
	return scanSnapshot(row)
}

func (s *Service) getSnapshotAndRepoByID(ctx context.Context, snapshotID uuid.UUID) (*domain.RepositorySnapshot, *domain.Repository, error) {
	query := `
		SELECT s.id, s.repository_id, s.commit_sha, s.ref_name, s.status, s.total_files, s.total_lines, s.primary_language, s.analyzed_at, s.expires_at, s.created_at,
		       r.id, r.github_id, r.owner, r.name, r.default_branch, r.is_private, r.created_at, r.updated_at
		FROM repository_snapshots s
		JOIN repositories r ON s.repository_id = r.id
		WHERE s.id = $1`

	row := s.db.QueryRowContext(ctx, query, snapshotID)

	var snap domain.RepositorySnapshot
	var repo domain.Repository
	var primLang sql.NullString

	err := row.Scan(
		&snap.ID, &snap.RepositoryID, &snap.CommitSHA, &snap.RefName, &snap.Status,
		&snap.TotalFiles, &snap.TotalLines, &primLang, &snap.AnalyzedAt, &snap.ExpiresAt, &snap.CreatedAt,
		&repo.ID, &repo.GitHubID, &repo.Owner, &repo.Name, &repo.DefaultBranch, &repo.IsPrivate, &repo.CreatedAt, &repo.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}
	if primLang.Valid {
		snap.PrimaryLanguage = primLang.String
	}
	return &snap, &repo, nil
}

func (s *Service) createOrResetSnapshot(ctx context.Context, repoID uuid.UUID, commitSHA, refName string) (*domain.RepositorySnapshot, error) {
	query := `
		INSERT INTO repository_snapshots (repository_id, commit_sha, ref_name, status, total_files, total_lines, created_at)
		VALUES ($1, $2, $3, 'QUEUED', 0, 0, NOW())
		ON CONFLICT (repository_id, commit_sha) DO UPDATE
		SET status = CASE 
		        WHEN repository_snapshots.status = 'FAILED' THEN 'QUEUED' 
		        ELSE repository_snapshots.status 
		    END,
		    ref_name = EXCLUDED.ref_name,
		    analyzed_at = CASE 
		        WHEN repository_snapshots.status = 'FAILED' THEN NULL 
		        ELSE repository_snapshots.analyzed_at 
		    END
		RETURNING id, repository_id, commit_sha, ref_name, status, total_files, total_lines, primary_language, analyzed_at, expires_at, created_at`

	row := s.db.QueryRowContext(ctx, query, repoID, commitSHA, refName)
	return scanSnapshot(row)
}

func (s *Service) getActiveJobForSnapshot(ctx context.Context, snapshotID uuid.UUID) (*domain.AnalysisJob, error) {
	query := `
		SELECT id, type, snapshot_id, status, stage, progress_percent, retry_count, created_at, updated_at
		FROM analysis_jobs
		WHERE snapshot_id = $1 AND status IN ('QUEUED', 'PROCESSING')
		ORDER BY created_at DESC
		LIMIT 1`

	row := s.db.QueryRowContext(ctx, query, snapshotID)

	var job domain.AnalysisJob
	var snapID sql.NullString
	err := row.Scan(&job.ID, &job.Type, &snapID, &job.Status, &job.Stage, &job.ProgressPercent, &job.RetryCount, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if snapID.Valid {
		parsed, _ := uuid.Parse(snapID.String)
		job.SnapshotID = &parsed
	}
	return &job, nil
}

func (s *Service) createJob(ctx context.Context, snapshotID uuid.UUID) (*domain.AnalysisJob, error) {
	query := `
		INSERT INTO analysis_jobs (type, snapshot_id, status, stage, progress_percent, created_at, updated_at)
		VALUES ('SNAPSHOT_INGEST', $1, 'QUEUED', 'INITIALIZING', 0.0, NOW(), NOW())
		ON CONFLICT (snapshot_id) WHERE status IN ('QUEUED', 'PROCESSING') AND snapshot_id IS NOT NULL DO NOTHING
		RETURNING id, type, snapshot_id, status, stage, progress_percent, retry_count, created_at, updated_at`

	row := s.db.QueryRowContext(ctx, query, snapshotID)

	var job domain.AnalysisJob
	var snapID sql.NullString
	err := row.Scan(&job.ID, &job.Type, &snapID, &job.Status, &job.Stage, &job.ProgressPercent, &job.RetryCount, &job.CreatedAt, &job.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Concurrent request won the insert race; reuse the existing active job
		existingJob, getErr := s.getActiveJobForSnapshot(ctx, snapshotID)
		if getErr == nil && existingJob != nil {
			return existingJob, nil
		}
		return nil, fmt.Errorf("active job exists but failed to retrieve: %w", getErr)
	}
	if err != nil {
		return nil, err
	}
	if snapID.Valid {
		parsed, _ := uuid.Parse(snapID.String)
		job.SnapshotID = &parsed
	}
	return &job, nil
}

func (s *Service) updateSnapshotStatus(ctx context.Context, snapshotID uuid.UUID, status domain.SnapshotStatus) error {
	query := `UPDATE repository_snapshots SET status = $2 WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, snapshotID, string(status))
	return err
}

func (s *Service) finalizeSnapshot(ctx context.Context, snapshotID uuid.UUID, totalFiles, totalLines int, primaryLang string, status domain.SnapshotStatus) error {
	query := `
		UPDATE repository_snapshots
		SET status = $2,
		    total_files = $3,
		    total_lines = $4,
		    primary_language = $5,
		    analyzed_at = NOW()
		WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, snapshotID, string(status), totalFiles, totalLines, primaryLang)
	return err
}

func (s *Service) persistFiles(ctx context.Context, snapshotID uuid.UUID, files []git.FileEntry) (map[string]uuid.UUID, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO repository_files (snapshot_id, path, extension, language, size_bytes, line_count, sha256_hash, content, is_binary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (snapshot_id, path) DO UPDATE
		SET size_bytes = EXCLUDED.size_bytes,
		    line_count = EXCLUDED.line_count,
		    sha256_hash = EXCLUDED.sha256_hash,
		    content = EXCLUDED.content,
		    is_binary = EXCLUDED.is_binary
		RETURNING id, path`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	pathToFileID := make(map[string]uuid.UUID)
	for _, f := range files {
		var contentVal *string
		if !f.IsBinary && f.Content != "" {
			c := f.Content
			contentVal = &c
		}

		var fileID uuid.UUID
		var retPath string
		err := stmt.QueryRowContext(ctx,
			snapshotID,
			f.Path,
			f.Extension,
			f.Language,
			f.SizeBytes,
			f.LineCount,
			f.SHA256Hash,
			contentVal,
			f.IsBinary,
		).Scan(&fileID, &retPath)
		if err != nil {
			return nil, fmt.Errorf("failed to insert file %s: %w", f.Path, err)
		}
		pathToFileID[retPath] = fileID
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return pathToFileID, nil
}

func (s *Service) persistSymbols(ctx context.Context, snapshotID uuid.UUID, symbols []domain.CodeSymbol) error {
	if len(symbols) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO code_symbols (id, snapshot_id, file_id, name, kind, start_line, end_line, signature, is_exported, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (file_id, name, kind, start_line) DO UPDATE
		SET end_line = EXCLUDED.end_line,
		    signature = EXCLUDED.signature,
		    is_exported = EXCLUDED.is_exported`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range symbols {
		symID := symbols[i].ID
		if symID == uuid.Nil {
			symID = uuid.New()
			symbols[i].ID = symID
		}
		_, err := stmt.ExecContext(ctx,
			symID,
			snapshotID,
			symbols[i].FileID,
			symbols[i].Name,
			string(symbols[i].Kind),
			symbols[i].StartLine,
			symbols[i].EndLine,
			symbols[i].Signature,
			symbols[i].IsExported,
		)
		if err != nil {
			return fmt.Errorf("failed to insert code symbol %s: %w", symbols[i].Name, err)
		}
	}

	return tx.Commit()
}

// GetSnapshotSymbols retrieves code symbols recorded for a commit snapshot with optional path filter.
func (s *Service) GetSnapshotSymbols(ctx context.Context, owner, repoName, commitSHA, pathFilter string) ([]domain.CodeSymbol, error) {
	query := `
		SELECT cs.id, cs.snapshot_id, cs.file_id, rf.path, cs.name, cs.kind, cs.start_line, cs.end_line, cs.signature, cs.is_exported, cs.created_at
		FROM code_symbols cs
		JOIN repository_files rf ON cs.file_id = rf.id
		JOIN repository_snapshots s ON cs.snapshot_id = s.id
		JOIN repositories r ON s.repository_id = r.id
		WHERE r.owner = $1 AND r.name = $2 AND s.commit_sha = $3
		  AND ($4 = '' OR rf.path = $4)
		ORDER BY rf.path ASC, cs.start_line ASC`

	rows, err := s.db.QueryContext(ctx, query, owner, repoName, commitSHA, pathFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to query code symbols: %w", err)
	}
	defer rows.Close()

	var symbols []domain.CodeSymbol
	for rows.Next() {
		var sym domain.CodeSymbol
		var kindStr string
		var sig sql.NullString
		err := rows.Scan(
			&sym.ID, &sym.SnapshotID, &sym.FileID, &sym.FilePath,
			&sym.Name, &kindStr, &sym.StartLine, &sym.EndLine,
			&sig, &sym.IsExported, &sym.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sym.Kind = domain.SymbolKind(kindStr)
		if sig.Valid {
			sym.Signature = sig.String
		}
		symbols = append(symbols, sym)
	}
	return symbols, nil
}

// GetSubsystems computes explainable path-based subsystems from the snapshot files and symbols,
// enriched with deterministic inter-subsystem connection edges.
func (s *Service) GetSubsystems(ctx context.Context, owner, repoName, ref string) ([]domain.SubsystemNode, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}

	files, err := s.GetSnapshotFiles(ctx, owner, repoName, snap.CommitSHA)
	if err != nil {
		return nil, err
	}

	symbols, err := s.GetSnapshotSymbols(ctx, owner, repoName, snap.CommitSHA, "")
	if err != nil {
		return nil, err
	}

	subsystems := analysis.ClassifySubsystems(files, symbols)

	// Enrich with real graph connections if graph service is active
	if s.graphSvc != nil {
		if conns, err := s.graphSvc.GetSubsystemConnections(ctx, snap.ID); err == nil && len(conns) > 0 {
			for i := range subsystems {
				if targets, ok := conns[subsystems[i].ID]; ok && len(targets) > 0 {
					subsystems[i].Connections = targets
				}
			}
		}
	}

	return subsystems, nil
}

// GetCandidateImpact computes candidate blast-radius impact analysis for a file in a snapshot.
func (s *Service) GetCandidateImpact(ctx context.Context, owner, repoName, ref, filePath string) (*domain.CandidateImpactReport, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}

	if s.graphSvc == nil {
		return nil, errors.New("graph service is not configured")
	}

	return s.graphSvc.GetCandidateImpact(ctx, snap.ID, filePath)
}

// GetFeatureTraces returns verifiable end-to-end execution flows across subsystems.
func (s *Service) GetFeatureTraces(ctx context.Context, owner, repoName, ref string) ([]domain.FeatureTrace, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}

	if s.traceBuilder == nil {
		return nil, errors.New("feature trace builder is not configured")
	}

	return s.traceBuilder.BuildFeatureTraces(ctx, snap.ID)
}

// GetFeatureTraceByID returns a specific feature trace for a repository snapshot.
func (s *Service) GetFeatureTraceByID(ctx context.Context, owner, repoName, ref, traceID string) (*domain.FeatureTrace, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}

	if s.traceBuilder == nil {
		return nil, errors.New("feature trace builder is not configured")
	}

	return s.traceBuilder.GetFeatureTraceByID(ctx, snap.ID, traceID)
}

// GetTree returns a hierarchical directory tree with extracted symbol counts.
func (s *Service) GetTree(ctx context.Context, owner, repoName, ref string) ([]domain.RepoTreeItem, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}

	files, err := s.GetSnapshotFiles(ctx, owner, repoName, snap.CommitSHA)
	if err != nil {
		return nil, err
	}

	symbols, err := s.GetSnapshotSymbols(ctx, owner, repoName, snap.CommitSHA, "")
	if err != nil {
		return nil, err
	}

	return analysis.BuildRepoTree(files, symbols), nil
}

// Search executes tiered hybrid code search combining exact symbols, FTS text, and dense vectors.
func (s *Service) Search(ctx context.Context, owner, repoName, ref, query string, topK int) ([]domain.SearchResult, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}

	if s.retrievalSvc == nil {
		return nil, errors.New("retrieval service is not configured")
	}

	return s.retrievalSvc.HybridSearch(ctx, snap.ID, query, topK)
}

// AssembleEvidence hydrates and bundles evidence references into an EvidencePackage with verified citations.
func (s *Service) AssembleEvidence(ctx context.Context, owner, repoName, ref, query string, refs []domain.EvidenceRef) (*domain.EvidencePackage, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}

	if s.evidenceStore == nil {
		return nil, errors.New("evidence store is not configured")
	}

	return s.evidenceStore.AssembleEvidencePackage(ctx, snap.ID, snap.CommitSHA, query, refs)
}

// Chat processes an AI mentor question grounded against repository evidence.
func (s *Service) Chat(ctx context.Context, owner, repoName, ref string, req domain.ChatRequest) (*domain.ChatResponse, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}
	if s.mentorSvc == nil {
		return nil, errors.New("mentor service is not configured")
	}
	return s.mentorSvc.Chat(ctx, snap.ID, snap.CommitSHA, req)
}

// ChatStream initiates a streaming response for interactive mentor chat.
func (s *Service) ChatStream(ctx context.Context, owner, repoName, ref string, req domain.ChatRequest) (<-chan domain.ChatStreamChunk, error) {
	snap, err := s.resolveSnapshotForRef(ctx, owner, repoName, ref)
	if err != nil {
		return nil, err
	}
	if s.mentorSvc == nil {
		return nil, errors.New("mentor service is not configured")
	}
	return s.mentorSvc.ChatStream(ctx, snap.ID, snap.CommitSHA, req)
}

// GetSessionMessages fetches chat message history for a mentorship session.
func (s *Service) GetSessionMessages(ctx context.Context, sessionID uuid.UUID) ([]domain.MentorMessage, error) {
	if s.mentorSvc == nil {
		return nil, errors.New("mentor service is not configured")
	}
	return s.mentorSvc.GetSessionMessages(ctx, sessionID)
}

func (s *Service) resolveSnapshotForRef(ctx context.Context, owner, repoName, ref string) (*domain.RepositorySnapshot, error) {
	if ref == "" {
		ref = "main"
	}

	repo, err := s.GetRepository(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("repository %s/%s not found: %w", owner, repoName, err)
	}

	var commitSHA string
	if s.fetcher != nil {
		sha, err := s.fetcher.ResolveCommitSHA(ctx, owner, repoName, ref)
		if err == nil && sha != "" {
			commitSHA = sha
		}
	}

	if commitSHA != "" {
		snap, err := s.getSnapshot(ctx, repo.ID, commitSHA)
		if err == nil && snap != nil && (snap.Status == domain.SnapshotStatusReady || snap.Status == domain.SnapshotStatusPartial) {
			return snap, nil
		}
	}

	// Fallback to latest ready snapshot for this repository
	query := `
		SELECT id, repository_id, commit_sha, ref_name, status, total_files, total_lines, primary_language, analyzed_at, expires_at, created_at
		FROM repository_snapshots
		WHERE repository_id = $1 AND status IN ('READY', 'PARTIAL')
		ORDER BY created_at DESC
		LIMIT 1`

	row := s.db.QueryRowContext(ctx, query, repo.ID)
	snap, err := scanSnapshot(row)
	if err != nil {
		return nil, fmt.Errorf("no ready snapshot found for repository %s/%s: %w", owner, repoName, err)
	}
	return snap, nil
}

func (s *Service) failJobDirect(ctx context.Context, jobID uuid.UUID, errMsg string) error {
	query := `UPDATE analysis_jobs SET status = 'FAILED', error_message = $2, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, jobID, errMsg)
	return err
}

func (s *Service) failJobAndSnapshot(jobID uuid.UUID, snapshotID *uuid.UUID, errMsg string) {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if snapshotID != nil {
		if err := s.updateSnapshotStatus(cleanupCtx, *snapshotID, domain.SnapshotStatusFailed); err != nil {
			s.logger.Error("Failed to update snapshot status during error cleanup",
				slog.String("snapshot_id", snapshotID.String()),
				slog.String("job_id", jobID.String()),
				slog.String("original_error", errMsg),
				slog.String("cleanup_error", err.Error()),
			)
		}
	}
	if s.jobManager != nil {
		if err := s.jobManager.FailJob(cleanupCtx, jobID, errMsg); err != nil {
			s.logger.Error("Failed to update job status during error cleanup",
				slog.String("job_id", jobID.String()),
				slog.String("original_error", errMsg),
				slog.String("cleanup_error", err.Error()),
			)
		}
	}
}

// GetSnapshotFiles retrieves the list of files recorded in an immutable snapshot.
func (s *Service) GetSnapshotFiles(ctx context.Context, owner, repoName, commitSHA string) ([]domain.RepositoryFile, error) {
	query := `
		SELECT f.id, f.snapshot_id, f.path, f.extension, f.language, f.size_bytes, f.line_count, f.sha256_hash, f.is_binary, f.created_at
		FROM repository_files f
		JOIN repository_snapshots s ON f.snapshot_id = s.id
		JOIN repositories r ON s.repository_id = r.id
		WHERE r.owner = $1 AND r.name = $2 AND s.commit_sha = $3
		ORDER BY f.path ASC`

	rows, err := s.db.QueryContext(ctx, query, owner, repoName, commitSHA)
	if err != nil {
		return nil, fmt.Errorf("failed to query repository files: %w", err)
	}
	defer rows.Close()

	var files []domain.RepositoryFile
	for rows.Next() {
		var rf domain.RepositoryFile
		var ext, lang sql.NullString
		err := rows.Scan(
			&rf.ID, &rf.SnapshotID, &rf.Path, &ext, &lang,
			&rf.SizeBytes, &rf.LineCount, &rf.SHA256Hash, &rf.IsBinary, &rf.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if ext.Valid {
			rf.Extension = ext.String
		}
		if lang.Valid {
			rf.Language = lang.String
		}
		files = append(files, rf)
	}
	return files, nil
}

// GetFileContent retrieves content and provenance for a single file in a snapshot.
func (s *Service) GetFileContent(ctx context.Context, snapshotID uuid.UUID, filePath string) (*domain.RepositoryFile, error) {
	query := `
		SELECT id, snapshot_id, path, extension, language, size_bytes, line_count, sha256_hash, content, is_binary, created_at
		FROM repository_files
		WHERE snapshot_id = $1 AND path = $2`

	row := s.db.QueryRowContext(ctx, query, snapshotID, filePath)

	var rf domain.RepositoryFile
	var ext, lang, content sql.NullString

	err := row.Scan(
		&rf.ID, &rf.SnapshotID, &rf.Path, &ext, &lang,
		&rf.SizeBytes, &rf.LineCount, &rf.SHA256Hash, &content, &rf.IsBinary, &rf.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ext.Valid {
		rf.Extension = ext.String
	}
	if lang.Valid {
		rf.Language = lang.String
	}
	if content.Valid {
		rf.Content = &content.String
	}
	return &rf, nil
}

// GetRepository reads repository info by owner and name.
func (s *Service) GetRepository(ctx context.Context, owner, name string) (*domain.Repository, error) {
	query := `
		SELECT id, github_id, owner, name, default_branch, is_private, created_at, updated_at
		FROM repositories
		WHERE owner = $1 AND name = $2`

	row := s.db.QueryRowContext(ctx, query, owner, name)

	var r domain.Repository
	err := row.Scan(&r.ID, &r.GitHubID, &r.Owner, &r.Name, &r.DefaultBranch, &r.IsPrivate, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func scanSnapshot(row *sql.Row) (*domain.RepositorySnapshot, error) {
	var snap domain.RepositorySnapshot
	var primLang sql.NullString

	err := row.Scan(
		&snap.ID, &snap.RepositoryID, &snap.CommitSHA, &snap.RefName, &snap.Status,
		&snap.TotalFiles, &snap.TotalLines, &primLang, &snap.AnalyzedAt, &snap.ExpiresAt, &snap.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if primLang.Valid {
		snap.PrimaryLanguage = primLang.String
	}
	return &snap, nil
}
