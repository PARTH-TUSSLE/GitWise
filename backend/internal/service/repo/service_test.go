package repo_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/api/sse"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/git"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/gitwise/backend/internal/service/repo"
	"github.com/google/uuid"
)

func TestService_IngestValidation(t *testing.T) {
	svc := repo.NewService(nil, nil, nil, nil)
	ctx := context.Background()

	// Missing owner
	_, _, err := svc.Ingest(ctx, "", "repo", "main")
	if err == nil {
		t.Fatal("expected error for empty owner, got nil")
	}

	// Missing repo
	_, _, err = svc.Ingest(ctx, "owner", "", "main")
	if err == nil {
		t.Fatal("expected error for empty repo, got nil")
	}
}

func TestService_WithMockFetcher_Validation(t *testing.T) {
	expectedSHA := "0123456789abcdef0123456789abcdef01234567"
	content := "package main\n\nfunc main() {}\n"
	h := sha256.Sum256([]byte(content))
	hashHex := hex.EncodeToString(h[:])

	files := []git.FileEntry{
		{
			Path:       "main.go",
			Extension:  ".go",
			Language:   "Go",
			SizeBytes:  len(content),
			LineCount:  3,
			SHA256Hash: hashHex,
			Content:    content,
			IsBinary:   false,
		},
	}

	fetcher := git.NewMockFetcher(expectedSHA, files)
	svc := repo.NewService(nil, fetcher, nil, nil)

	ctx := context.Background()
	sha, err := fetcher.ResolveCommitSHA(ctx, "testowner", "testrepo", "main")
	if err != nil {
		t.Fatalf("unexpected error resolving commit: %v", err)
	}
	if sha != expectedSHA {
		t.Errorf("expected sha %s, got %s", expectedSHA, sha)
	}

	// With nil DB, Ingest should return a database error safely
	_, _, err = svc.Ingest(ctx, "testowner", "testrepo", "main")
	if err == nil {
		t.Fatal("expected error with nil database, got nil")
	}
}

func TestService_TriggerIngestion_ReusesActiveJob_Queued(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_queued_reuse")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	activeJobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "INSERT INTO repositories") {
			return &fakeRepoRows{
				cols: []string{"id", "github_id", "owner", "name", "default_branch", "is_private", "created_at", "updated_at"},
				data: [][]driver.Value{
					{repoID.String(), int64(12345), "owner", "repo", "main", false, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots") && strings.Contains(query, "WHERE repository_id = $1") {
			// Existing snapshot is in QUEUED state
			return &fakeRepoRows{
				cols: []string{"id", "repository_id", "commit_sha", "ref_name", "status", "total_files", "total_lines", "primary_language", "analyzed_at", "expires_at", "created_at"},
				data: [][]driver.Value{
					{snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusQueued), 0, 0, nil, nil, nil, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE snapshot_id = $1") {
			// Active job exists in QUEUED state
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{activeJobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	fetcher := git.NewMockFetcher(commitSHA, nil)
	jm := jobs.NewJobManager(db, 2, 10, nil, nil)
	svc := repo.NewService(db, fetcher, jm, nil)

	job, snap, err := svc.Ingest(context.Background(), "owner", "repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.ID != activeJobID {
		t.Errorf("expected reused job ID %s, got %s", activeJobID, job.ID)
	}
	if snap.ID != snapshotID {
		t.Errorf("expected snapshot ID %s, got %s", snapshotID, snap.ID)
	}
}

func TestService_TriggerIngestion_ReusesActiveJob_Processing(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_processing_reuse")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	activeJobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "INSERT INTO repositories") {
			return &fakeRepoRows{
				cols: []string{"id", "github_id", "owner", "name", "default_branch", "is_private", "created_at", "updated_at"},
				data: [][]driver.Value{
					{repoID.String(), int64(12345), "owner", "repo", "main", false, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots") && strings.Contains(query, "WHERE repository_id = $1") {
			return &fakeRepoRows{
				cols: []string{"id", "repository_id", "commit_sha", "ref_name", "status", "total_files", "total_lines", "primary_language", "analyzed_at", "expires_at", "created_at"},
				data: [][]driver.Value{
					{snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusProcessing), 0, 0, nil, nil, nil, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE snapshot_id = $1") {
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{activeJobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusProcessing), string(domain.JobStageFetchingTree), 25.0, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	fetcher := git.NewMockFetcher(commitSHA, nil)
	jm := jobs.NewJobManager(db, 2, 10, nil, nil)
	svc := repo.NewService(db, fetcher, jm, nil)

	job, snap, err := svc.Ingest(context.Background(), "owner", "repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.ID != activeJobID {
		t.Errorf("expected reused job ID %s, got %s", activeJobID, job.ID)
	}
	if job.Status != domain.JobStatusProcessing {
		t.Errorf("expected job status PROCESSING, got %s", job.Status)
	}
	if snap.ID != snapshotID {
		t.Errorf("expected snapshot ID %s, got %s", snapshotID, snap.ID)
	}
}

func TestService_ProcessIngestion_Complete_BecomesReady(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_process_complete")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	jobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	var updatedSnapshotStatus string
	var completedJobID string

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE id = $1") {
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "error_message", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{jobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, nil, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots s") {
			return &fakeRepoRows{
				cols: []string{
					"s.id", "s.repository_id", "s.commit_sha", "s.ref_name", "s.status", "s.total_files", "s.total_lines", "primary_language", "s.analyzed_at", "s.expires_at", "s.created_at",
					"r.id", "r.github_id", "r.owner", "r.name", "r.default_branch", "r.is_private", "r.created_at", "r.updated_at",
				},
				data: [][]driver.Value{
					{
						snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusQueued), 0, 0, nil, nil, nil, now,
						repoID.String(), int64(123), "owner", "repo", "main", false, now, now,
					},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "UPDATE repository_snapshots") && strings.Contains(query, "SET status = $2,") {
			// Finalize snapshot query: args[0] is snapshotID, args[1] is status
			for _, arg := range args {
				if s, ok := arg.Value.(string); ok && (s == string(domain.SnapshotStatusReady) || s == string(domain.SnapshotStatusPartial) || s == string(domain.SnapshotStatusFailed)) {
					updatedSnapshotStatus = s
				}
			}
			return nil, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "UPDATE analysis_jobs") && strings.Contains(query, "SET status = 'COMPLETED'") {
			completedJobID = jobID.String()
			return nil, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	files := []git.FileEntry{
		{Path: "main.go", Extension: ".go", Language: "Go", SizeBytes: 50, LineCount: 5, SHA256Hash: "hash1", Content: "package main"},
	}
	fetcher := git.NewMockFetcher(commitSHA, files)
	fetcher.Outcome = git.IngestionOutcomeComplete

	broker := sse.NewBroker(nil)
	jm := jobs.NewJobManager(db, 1, 5, nil, broker)
	svc := repo.NewService(db, fetcher, jm, nil)

	err = svc.ProcessIngestion(context.Background(), jobID)
	if err != nil {
		t.Fatalf("unexpected error during ProcessIngestion: %v", err)
	}

	if updatedSnapshotStatus != string(domain.SnapshotStatusReady) {
		t.Errorf("expected snapshot finalized as READY, got %s", updatedSnapshotStatus)
	}
	if completedJobID != jobID.String() {
		t.Errorf("expected job %s to be marked COMPLETED, got %s", jobID, completedJobID)
	}
}

func TestService_ProcessIngestion_Capped_BecomesPartial(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_process_partial")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	jobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	var updatedSnapshotStatus string

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE id = $1") {
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "error_message", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{jobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, nil, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots s") {
			return &fakeRepoRows{
				cols: []string{
					"s.id", "s.repository_id", "s.commit_sha", "s.ref_name", "s.status", "s.total_files", "s.total_lines", "primary_language", "s.analyzed_at", "s.expires_at", "s.created_at",
					"r.id", "r.github_id", "r.owner", "r.name", "r.default_branch", "r.is_private", "r.created_at", "r.updated_at",
				},
				data: [][]driver.Value{
					{
						snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusQueued), 0, 0, nil, nil, nil, now,
						repoID.String(), int64(123), "owner", "repo", "main", false, now, now,
					},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "UPDATE repository_snapshots") && strings.Contains(query, "SET status = $2,") {
			for _, arg := range args {
				if s, ok := arg.Value.(string); ok && (s == string(domain.SnapshotStatusReady) || s == string(domain.SnapshotStatusPartial) || s == string(domain.SnapshotStatusFailed)) {
					updatedSnapshotStatus = s
				}
			}
			return nil, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	files := []git.FileEntry{
		{Path: "main.go", Extension: ".go", Language: "Go", SizeBytes: 50, LineCount: 5, SHA256Hash: "hash1", Content: "package main"},
	}
	fetcher := git.NewMockFetcher(commitSHA, files)
	fetcher.Outcome = git.IngestionOutcomePartial
	fetcher.CappedReason = "repository exceeded total size cap of 150 MB"

	broker := sse.NewBroker(nil)
	jm := jobs.NewJobManager(db, 1, 5, nil, broker)
	svc := repo.NewService(db, fetcher, jm, nil)

	err = svc.ProcessIngestion(context.Background(), jobID)
	if err != nil {
		t.Fatalf("unexpected error during ProcessIngestion: %v", err)
	}

	if updatedSnapshotStatus != string(domain.SnapshotStatusPartial) {
		t.Errorf("expected snapshot finalized as PARTIAL, got %s", updatedSnapshotStatus)
	}
}

func TestService_ProcessIngestion_TruncatedTree_FailsAndNotReady(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_process_truncated")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	jobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	var snapshotStatuses []string
	var jobFailed bool

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE id = $1") {
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "error_message", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{jobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, nil, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots s") {
			return &fakeRepoRows{
				cols: []string{
					"s.id", "s.repository_id", "s.commit_sha", "s.ref_name", "s.status", "s.total_files", "s.total_lines", "primary_language", "s.analyzed_at", "s.expires_at", "s.created_at",
					"r.id", "r.github_id", "r.owner", "r.name", "r.default_branch", "r.is_private", "r.created_at", "r.updated_at",
				},
				data: [][]driver.Value{
					{
						snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusQueued), 0, 0, nil, nil, nil, now,
						repoID.String(), int64(123), "owner", "repo", "main", false, now, now,
					},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "UPDATE repository_snapshots SET status = $2 WHERE id = $1") {
			for _, arg := range args {
				if s, ok := arg.Value.(string); ok {
					snapshotStatuses = append(snapshotStatuses, s)
				}
			}
			return nil, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "UPDATE analysis_jobs") && strings.Contains(query, "SET status = 'FAILED'") {
			jobFailed = true
			return nil, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	fetcher := git.NewMockFetcher(commitSHA, nil)
	fetcher.Err = git.ErrTreeTruncated

	broker := sse.NewBroker(nil)
	jm := jobs.NewJobManager(db, 1, 5, nil, broker)
	svc := repo.NewService(db, fetcher, jm, nil)

	err = svc.ProcessIngestion(context.Background(), jobID)
	if err == nil {
		t.Fatal("expected error from ProcessIngestion when tree is truncated, got nil")
	}

	if !errors.Is(err, git.ErrTreeTruncated) {
		t.Errorf("expected ErrTreeTruncated in error chain, got %v", err)
	}

	// Snapshot must NEVER be marked READY; it must be marked FAILED
	for _, status := range snapshotStatuses {
		if status == string(domain.SnapshotStatusReady) {
			t.Fatal("snapshot must never be marked READY on truncated tree")
		}
	}

	lastStatus := snapshotStatuses[len(snapshotStatuses)-1]
	if lastStatus != string(domain.SnapshotStatusFailed) {
		t.Errorf("expected snapshot to end in FAILED, got %s", lastStatus)
	}
	if !jobFailed {
		t.Error("expected job to be marked FAILED")
	}
}

func TestService_ProcessIngestion_BlobFetchFailure_FailsAndNotReady(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_process_blob_failure")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	jobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	var snapshotStatuses []string
	var jobFailed bool

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE id = $1") {
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "error_message", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{jobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, nil, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots s") {
			return &fakeRepoRows{
				cols: []string{
					"s.id", "s.repository_id", "s.commit_sha", "s.ref_name", "s.status", "s.total_files", "s.total_lines", "primary_language", "s.analyzed_at", "s.expires_at", "s.created_at",
					"r.id", "r.github_id", "r.owner", "r.name", "r.default_branch", "r.is_private", "r.created_at", "r.updated_at",
				},
				data: [][]driver.Value{
					{
						snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusQueued), 0, 0, nil, nil, nil, now,
						repoID.String(), int64(123), "owner", "repo", "main", false, now, now,
					},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "UPDATE repository_snapshots SET status = $2 WHERE id = $1") {
			for _, arg := range args {
				if s, ok := arg.Value.(string); ok {
					snapshotStatuses = append(snapshotStatuses, s)
				}
			}
			return nil, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "UPDATE analysis_jobs") && strings.Contains(query, "SET status = 'FAILED'") {
			jobFailed = true
			return nil, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	fetcher := git.NewMockFetcher(commitSHA, nil)
	fetcher.Err = fmt.Errorf("failed to fetch blob content: network timeout")

	broker := sse.NewBroker(nil)
	jm := jobs.NewJobManager(db, 1, 5, nil, broker)
	svc := repo.NewService(db, fetcher, jm, nil)

	err = svc.ProcessIngestion(context.Background(), jobID)
	if err == nil {
		t.Fatal("expected error on blob fetch failure, got nil")
	}

	for _, status := range snapshotStatuses {
		if status == string(domain.SnapshotStatusReady) {
			t.Fatal("snapshot must never be marked READY on blob fetch failure")
		}
	}

	lastStatus := snapshotStatuses[len(snapshotStatuses)-1]
	if lastStatus != string(domain.SnapshotStatusFailed) {
		t.Errorf("expected snapshot to end in FAILED, got %s", lastStatus)
	}
	if !jobFailed {
		t.Error("expected job to be marked FAILED")
	}
}

func TestService_TriggerIngestion_ConcurrentInsertConflictReusesActiveJob(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_concurrent_insert_conflict")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	existingActiveJobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "INSERT INTO repositories") {
			return &fakeRepoRows{
				cols: []string{"id", "github_id", "owner", "name", "default_branch", "is_private", "created_at", "updated_at"},
				data: [][]driver.Value{
					{repoID.String(), int64(12345), "owner", "repo", "main", false, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots") && strings.Contains(query, "WHERE repository_id = $1") {
			// Initially no snapshot returned by SELECT (simulating race before insert)
			return &fakeRepoRows{}, driver.RowsAffected(0), nil
		}
		if strings.Contains(query, "INSERT INTO repository_snapshots") {
			return &fakeRepoRows{
				cols: []string{"id", "repository_id", "commit_sha", "ref_name", "status", "total_files", "total_lines", "primary_language", "analyzed_at", "expires_at", "created_at"},
				data: [][]driver.Value{
					{snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusQueued), 0, 0, nil, nil, nil, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "INSERT INTO analysis_jobs") {
			// Simulate ON CONFLICT DO NOTHING by returning 0 rows (empty rows)
			return &fakeRepoRows{}, driver.RowsAffected(0), nil
		}
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE snapshot_id = $1") {
			// Returns the winner of the race
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{existingActiveJobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	fetcher := git.NewMockFetcher(commitSHA, nil)
	jm := jobs.NewJobManager(db, 2, 10, nil, nil)
	svc := repo.NewService(db, fetcher, jm, nil)

	job, snap, err := svc.Ingest(context.Background(), "owner", "repo", "main")
	if err != nil {
		t.Fatalf("unexpected error during concurrent insert conflict: %v", err)
	}

	if job.ID != existingActiveJobID {
		t.Errorf("expected reused active job ID %s, got %s", existingActiveJobID, job.ID)
	}
	if snap.ID != snapshotID {
		t.Errorf("expected snapshot ID %s, got %s", snapshotID, snap.ID)
	}
}

func TestService_TriggerIngestion_FailedSnapshotAllowsNewJob(t *testing.T) {
	db, err := sql.Open("fake_repo_driver", "test_failed_allows_new")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	snapshotID := uuid.New()
	newJobID := uuid.New()
	commitSHA := "0123456789abcdef0123456789abcdef01234567"
	now := time.Now()

	testRepoDriver.mu.Lock()
	testRepoDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "INSERT INTO repositories") {
			return &fakeRepoRows{
				cols: []string{"id", "github_id", "owner", "name", "default_branch", "is_private", "created_at", "updated_at"},
				data: [][]driver.Value{
					{repoID.String(), int64(12345), "owner", "repo", "main", false, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM repository_snapshots") && strings.Contains(query, "WHERE repository_id = $1") {
			// Snapshot was previously FAILED
			return &fakeRepoRows{
				cols: []string{"id", "repository_id", "commit_sha", "ref_name", "status", "total_files", "total_lines", "primary_language", "analyzed_at", "expires_at", "created_at"},
				data: [][]driver.Value{
					{snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusFailed), 0, 0, nil, nil, nil, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "FROM analysis_jobs") && strings.Contains(query, "WHERE snapshot_id = $1") {
			// No active job (previous jobs were FAILED or COMPLETED, not QUEUED or PROCESSING)
			return &fakeRepoRows{}, driver.RowsAffected(0), nil
		}
		if strings.Contains(query, "INSERT INTO repository_snapshots") {
			// Reset to QUEUED
			return &fakeRepoRows{
				cols: []string{"id", "repository_id", "commit_sha", "ref_name", "status", "total_files", "total_lines", "primary_language", "analyzed_at", "expires_at", "created_at"},
				data: [][]driver.Value{
					{snapshotID.String(), repoID.String(), commitSHA, "main", string(domain.SnapshotStatusQueued), 0, 0, nil, nil, nil, now},
				},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "INSERT INTO analysis_jobs") {
			return &fakeRepoRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{newJobID.String(), string(domain.JobTypeSnapshotIngest), snapshotID.String(), string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		return &fakeRepoRows{}, driver.RowsAffected(1), nil
	}
	testRepoDriver.mu.Unlock()

	fetcher := git.NewMockFetcher(commitSHA, nil)
	jm := jobs.NewJobManager(db, 2, 10, nil, nil)
	svc := repo.NewService(db, fetcher, jm, nil)

	job, snap, err := svc.Ingest(context.Background(), "owner", "repo", "main")
	if err != nil {
		t.Fatalf("unexpected error re-ingesting failed snapshot: %v", err)
	}

	if job.ID != newJobID {
		t.Errorf("expected new job ID %s, got %s", newJobID, job.ID)
	}
	if job.Status != domain.JobStatusQueued {
		t.Errorf("expected job status QUEUED, got %s", job.Status)
	}
	if snap.ID != snapshotID {
		t.Errorf("expected snapshot ID %s, got %s", snapshotID, snap.ID)
	}
}
