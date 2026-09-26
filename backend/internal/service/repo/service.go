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

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/git"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/google/uuid"
)

// Service coordinates repository ingestion, commit snapshots, and file persistence.
type Service struct {
	db         *sql.DB
	fetcher    git.Fetcher
	jobManager *jobs.JobManager
	logger     *slog.Logger
}

// NewService creates a new repository service instance.
func NewService(db *sql.DB, fetcher git.Fetcher, jobManager *jobs.JobManager, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	svc := &Service{
		db:         db,
		fetcher:    fetcher,
		jobManager: jobManager,
		logger:     logger,
	}

	// Register the task handler for SNAPSHOT_INGEST
	if jobManager != nil {
		jobManager.RegisterHandler(domain.JobTypeSnapshotIngest, svc.ProcessIngestion)
	}

	return svc
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
		if existingSnap.Status == domain.SnapshotStatusReady {
			// Idempotent: Snapshot already ingested and ready
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
	}

	// 4. Create or reset snapshot record in QUEUED state
	snapshot, err := s.createOrResetSnapshot(ctx, repo.ID, commitSHA, ref)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create snapshot record: %w", err)
	}

	// 5. Create background analysis job
	job, err := s.createJob(ctx, snapshot.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create analysis job: %w", err)
	}

	// 6. Enqueue job for background processing
	if s.jobManager != nil {
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
		return fmt.Errorf("failed to load job: %w", err)
	}
	if job.SnapshotID == nil {
		return errors.New("job has no associated snapshot ID")
	}

	// 2. Load snapshot and repo
	snapshot, repo, err := s.getSnapshotAndRepoByID(ctx, *job.SnapshotID)
	if err != nil {
		return fmt.Errorf("failed to load snapshot or repository: %w", err)
	}

	// 3. Mark snapshot and job as PROCESSING
	_ = s.updateSnapshotStatus(ctx, snapshot.ID, domain.SnapshotStatusProcessing)
	_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageFetchingTree, 15.0, 0, 0, "Fetching git tree from repository")

	// 4. Fetch tree
	files, err := s.fetcher.FetchTree(ctx, repo.Owner, repo.Name, snapshot.CommitSHA)
	if err != nil {
		_ = s.updateSnapshotStatus(ctx, snapshot.ID, domain.SnapshotStatusFailed)
		return fmt.Errorf("failed to fetch git tree: %w", err)
	}

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
		if count > maxLines {
			maxLines = count
			primaryLang = lang
		}
	}

	// 6. Persist files in database transaction
	_ = s.jobManager.UpdateProgress(ctx, jobID, domain.JobStageFinalizing, 80.0, len(files), len(files), "Persisting repository files into snapshot")

	if err := s.persistFiles(ctx, snapshot.ID, files); err != nil {
		_ = s.updateSnapshotStatus(ctx, snapshot.ID, domain.SnapshotStatusFailed)
		return fmt.Errorf("failed to persist files: %w", err)
	}

	// 7. Update snapshot with final facts
	if err := s.finalizeSnapshot(ctx, snapshot.ID, len(files), totalLines, primaryLang); err != nil {
		return fmt.Errorf("failed to finalize snapshot: %w", err)
	}

	// 8. Complete job and broadcast completion event
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
		SET status = 'QUEUED',
		    ref_name = EXCLUDED.ref_name,
		    analyzed_at = NULL
		RETURNING id, repository_id, commit_sha, ref_name, status, total_files, total_lines, primary_language, analyzed_at, expires_at, created_at`

	row := s.db.QueryRowContext(ctx, query, repoID, commitSHA, refName)
	return scanSnapshot(row)
}

func (s *Service) createJob(ctx context.Context, snapshotID uuid.UUID) (*domain.AnalysisJob, error) {
	query := `
		INSERT INTO analysis_jobs (type, snapshot_id, status, stage, progress_percent, created_at, updated_at)
		VALUES ('SNAPSHOT_INGEST', $1, 'QUEUED', 'INITIALIZING', 0.0, NOW(), NOW())
		RETURNING id, type, snapshot_id, status, stage, progress_percent, retry_count, created_at, updated_at`

	row := s.db.QueryRowContext(ctx, query, snapshotID)

	var job domain.AnalysisJob
	var snapID sql.NullString
	err := row.Scan(&job.ID, &job.Type, &snapID, &job.Status, &job.Stage, &job.ProgressPercent, &job.RetryCount, &job.CreatedAt, &job.UpdatedAt)
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

func (s *Service) finalizeSnapshot(ctx context.Context, snapshotID uuid.UUID, totalFiles, totalLines int, primaryLang string) error {
	query := `
		UPDATE repository_snapshots
		SET status = 'READY',
		    total_files = $2,
		    total_lines = $3,
		    primary_language = $4,
		    analyzed_at = NOW()
		WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, snapshotID, totalFiles, totalLines, primaryLang)
	return err
}

func (s *Service) persistFiles(ctx context.Context, snapshotID uuid.UUID, files []git.FileEntry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
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
		    is_binary = EXCLUDED.is_binary`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range files {
		var contentVal *string
		if !f.IsBinary && f.Content != "" {
			c := f.Content
			contentVal = &c
		}

		_, err := stmt.ExecContext(ctx,
			snapshotID,
			f.Path,
			f.Extension,
			f.Language,
			f.SizeBytes,
			f.LineCount,
			f.SHA256Hash,
			contentVal,
			f.IsBinary,
		)
		if err != nil {
			return fmt.Errorf("failed to insert file %s: %w", f.Path, err)
		}
	}

	return tx.Commit()
}

func (s *Service) failJobDirect(ctx context.Context, jobID uuid.UUID, errMsg string) error {
	query := `UPDATE analysis_jobs SET status = 'FAILED', error_message = $2, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, jobID, errMsg)
	return err
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
