package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gitwise/backend/internal/api/sse"
	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

// TaskFunc represents an executable background job task.
type TaskFunc func(ctx context.Context, jobID uuid.UUID) error

// JobManager coordinates in-process background worker execution, status persistence, and event notifications.
type JobManager struct {
	db          *sql.DB
	logger      *slog.Logger
	jobQueue    chan uuid.UUID
	maxWorkers  int
	wg          sync.WaitGroup
	sseBroker   *sse.Broker
	handlers    map[domain.JobType]TaskFunc
	handlersMu  sync.RWMutex
	cancelFunc  context.CancelFunc
	runningLock sync.Mutex
	isStarted   bool
	enqueuedMu  sync.Mutex
	enqueued    map[uuid.UUID]bool
}

const (
	DefaultWorkerCount = 4
	DefaultQueueSize   = 128
)

// NewJobManager creates a new background job coordinator.
func NewJobManager(db *sql.DB, workers int, queueSize int, logger *slog.Logger, sseBroker *sse.Broker) *JobManager {
	if workers <= 0 {
		workers = DefaultWorkerCount
	}
	if queueSize <= 0 {
		queueSize = DefaultQueueSize
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &JobManager{
		db:         db,
		logger:     logger,
		jobQueue:   make(chan uuid.UUID, queueSize),
		maxWorkers: workers,
		sseBroker:  sseBroker,
		handlers:   make(map[domain.JobType]TaskFunc),
		enqueued:   make(map[uuid.UUID]bool),
	}
}

// RegisterHandler registers a task executor for a given JobType.
func (jm *JobManager) RegisterHandler(jobType domain.JobType, handler TaskFunc) {
	jm.handlersMu.Lock()
	defer jm.handlersMu.Unlock()
	jm.handlers[jobType] = handler
}

// ReconcileStaleJobs sweeps any job that was in-flight (PROCESSING) when the server restarted.
func (jm *JobManager) ReconcileStaleJobs(ctx context.Context) error {
	if jm.db == nil {
		return errors.New("database is not configured")
	}

	query := `
		UPDATE analysis_jobs
		SET status = 'FAILED',
		    error_message = 'Server process restarted while job was in-flight',
		    updated_at = NOW()
		WHERE status = 'PROCESSING'`

	res, err := jm.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to reconcile stale in-flight jobs: %w", err)
	}

	if res != nil {
		rows, _ := res.RowsAffected()
		if rows > 0 {
			jm.logger.Warn("Reconciled stale in-flight jobs on startup",
				slog.Int64("abandoned_jobs_reconciled", rows),
			)
		}
	}

	// Also fail any repository_snapshots that were left in PROCESSING state
	snapQuery := `
		UPDATE repository_snapshots
		SET status = 'FAILED'
		WHERE status = 'PROCESSING'`
	_, _ = jm.db.ExecContext(ctx, snapQuery)

	return nil
}

// Start boots the worker pool after sweeping stale jobs from prior crashes.
func (jm *JobManager) Start(ctx context.Context) error {
	jm.runningLock.Lock()
	defer jm.runningLock.Unlock()

	if jm.isStarted {
		return nil
	}

	// 1. Reconcile stale jobs from prior crash
	if err := jm.ReconcileStaleJobs(ctx); err != nil {
		jm.logger.Error("Failed to reconcile stale jobs during startup", slog.String("error", err.Error()))
	}

	// 2. Launch worker goroutines
	workerCtx, cancel := context.WithCancel(ctx)
	jm.cancelFunc = cancel
	jm.isStarted = true

	for i := 0; i < jm.maxWorkers; i++ {
		jm.wg.Add(1)
		go jm.workerLoop(workerCtx, i+1)
	}

	// 3. Recover persisted QUEUED jobs after workers are active
	if jm.db != nil {
		if err := jm.RecoverQueuedJobs(ctx); err != nil {
			jm.logger.Error("Failed to recover persisted QUEUED jobs during startup", slog.String("error", err.Error()))
		}
	}

	jm.logger.Info("JobManager worker pool started",
		slog.Int("workers", jm.maxWorkers),
		slog.Int("queue_capacity", cap(jm.jobQueue)),
	)
	return nil
}

// RecoverQueuedJobs finds persisted jobs in QUEUED status and enqueues them into the worker queue.
func (jm *JobManager) RecoverQueuedJobs(ctx context.Context) error {
	if jm.db == nil {
		return errors.New("database is not configured")
	}

	query := `
		SELECT id
		FROM analysis_jobs
		WHERE status = 'QUEUED'
		ORDER BY created_at ASC`

	rows, err := jm.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query queued jobs for recovery: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var jobID uuid.UUID
		if err := rows.Scan(&jobID); err != nil {
			return err
		}

		if err := jm.Enqueue(ctx, jobID); err != nil {
			jm.logger.Error("Failed to re-enqueue recovered job",
				slog.String("job_id", jobID.String()),
				slog.String("error", err.Error()),
			)
		} else {
			count++
		}
	}

	if count > 0 {
		jm.logger.Info("Recovered persisted QUEUED jobs on startup", slog.Int("count", count))
	}
	return nil
}

// Stop gracefully stops workers and waits for in-flight tasks to terminate.
func (jm *JobManager) Stop() {
	jm.runningLock.Lock()
	defer jm.runningLock.Unlock()

	if !jm.isStarted {
		return
	}

	if jm.cancelFunc != nil {
		jm.cancelFunc()
	}
	close(jm.jobQueue)
	jm.wg.Wait()
	jm.isStarted = false
	jm.logger.Info("JobManager worker pool cleanly stopped")
}

// Enqueue submits a job ID into the in-process worker queue, preventing duplicate queueing.
func (jm *JobManager) Enqueue(ctx context.Context, jobID uuid.UUID) error {
	jm.enqueuedMu.Lock()
	if jm.enqueued[jobID] {
		jm.enqueuedMu.Unlock()
		jm.logger.Debug("Job already enqueued, skipping duplicate enqueue", slog.String("job_id", jobID.String()))
		return nil
	}

	if err := ctx.Err(); err != nil {
		jm.enqueuedMu.Unlock()
		return err
	}

	select {
	case <-ctx.Done():
		jm.enqueuedMu.Unlock()
		return ctx.Err()
	case jm.jobQueue <- jobID:
		jm.enqueued[jobID] = true
		jm.enqueuedMu.Unlock()
		jm.logger.Debug("Job enqueued for background execution", slog.String("job_id", jobID.String()))
		return nil
	default:
		jm.enqueuedMu.Unlock()
		return errors.New("job queue is full, unable to accept job")
	}
}

func (jm *JobManager) workerLoop(ctx context.Context, workerID int) {
	defer jm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case jobID, ok := <-jm.jobQueue:
			if !ok {
				return
			}
			jm.enqueuedMu.Lock()
			delete(jm.enqueued, jobID)
			jm.enqueuedMu.Unlock()

			jm.executeJob(ctx, jobID, workerID)
		}
	}
}

func (jm *JobManager) executeJob(ctx context.Context, jobID uuid.UUID, workerID int) {
	// 1. Fetch job definition from database
	job, err := jm.GetJob(ctx, jobID)
	if err != nil {
		jm.logger.Error("Worker failed to load job from database",
			slog.Int("worker_id", workerID),
			slog.String("job_id", jobID.String()),
			slog.String("error", err.Error()),
		)
		return
	}

	// 2. Locate registered handler for this job type
	jm.handlersMu.RLock()
	handler, exists := jm.handlers[job.Type]
	jm.handlersMu.RUnlock()

	if !exists {
		errMsg := fmt.Sprintf("no handler registered for job type %s", job.Type)
		jm.logger.Error("Job execution failed", slog.String("job_id", jobID.String()), slog.String("error", errMsg))
		_ = jm.FailJob(ctx, jobID, errMsg)
		return
	}

	// 3. Mark job as PROCESSING
	startTime := time.Now()
	if err := jm.UpdateProgress(ctx, jobID, domain.JobStageInitializing, 5.0, 0, 0, "Job processing started"); err != nil {
		jm.logger.Error("Failed to mark job as processing", slog.String("job_id", jobID.String()), slog.String("error", err.Error()))
	}

	// 4. Safely invoke task handler
	var taskErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				taskErr = fmt.Errorf("worker panic during job execution: %v", r)
			}
		}()
		taskErr = handler(ctx, jobID)
	}()

	durationMs := time.Since(startTime).Milliseconds()

	// 5. Commit final status
	if taskErr != nil {
		jm.logger.Error("Job processing failed",
			slog.String("job_id", jobID.String()),
			slog.Int64("duration_ms", durationMs),
			slog.String("error", taskErr.Error()),
		)
		_ = jm.FailJob(ctx, jobID, taskErr.Error())
	} else {
		jm.logger.Info("Job processing completed successfully",
			slog.String("job_id", jobID.String()),
			slog.Int64("duration_ms", durationMs),
		)
	}
}

// UpdateProgress updates job stage, percent, and broadcasts progress SSE event.
func (jm *JobManager) UpdateProgress(ctx context.Context, jobID uuid.UUID, stage domain.JobStage, percent float64, completed, total int, msg string) error {
	query := `
		UPDATE analysis_jobs
		SET status = 'PROCESSING',
		    stage = $2,
		    progress_percent = $3,
		    updated_at = NOW()
		WHERE id = $1`

	_, err := jm.db.ExecContext(ctx, query, jobID, string(stage), percent)
	if err != nil {
		return fmt.Errorf("failed to update job progress in db: %w", err)
	}

	if jm.sseBroker != nil {
		jm.sseBroker.BroadcastJobProgress(jobID.String(), string(stage), completed, total, percent, msg)
	}
	return nil
}

// CompleteJob transitions a job to COMPLETED state and emits SSE completion event.
func (jm *JobManager) CompleteJob(ctx context.Context, jobID uuid.UUID, snapshotID, commitSHA string, durationMs int64) error {
	query := `
		UPDATE analysis_jobs
		SET status = 'COMPLETED',
		    stage = 'DONE',
		    progress_percent = 100.0,
		    updated_at = NOW()
		WHERE id = $1`

	_, err := jm.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to complete job in db: %w", err)
	}

	if jm.sseBroker != nil {
		jm.sseBroker.BroadcastJobComplete(jobID.String(), snapshotID, commitSHA, durationMs)
	}
	return nil
}

// FailJob transitions a job to FAILED state and emits SSE error event.
func (jm *JobManager) FailJob(ctx context.Context, jobID uuid.UUID, errMsg string) error {
	query := `
		UPDATE analysis_jobs
		SET status = 'FAILED',
		    error_message = $2,
		    updated_at = NOW()
		WHERE id = $1`

	_, err := jm.db.ExecContext(ctx, query, jobID, errMsg)
	if err != nil {
		return fmt.Errorf("failed to fail job in db: %w", err)
	}

	if jm.sseBroker != nil {
		jm.sseBroker.BroadcastJobError(jobID.String(), "JOB_EXECUTION_FAILED", errMsg)
	}
	return nil
}

// GetJob reads a single AnalysisJob record by ID.
func (jm *JobManager) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error) {
	query := `
		SELECT id, type, snapshot_id, status, stage, progress_percent, error_message, retry_count, created_at, updated_at
		FROM analysis_jobs
		WHERE id = $1`

	row := jm.db.QueryRowContext(ctx, query, jobID)

	var job domain.AnalysisJob
	var snapID sql.NullString
	var errStr sql.NullString

	err := row.Scan(
		&job.ID,
		&job.Type,
		&snapID,
		&job.Status,
		&job.Stage,
		&job.ProgressPercent,
		&errStr,
		&job.RetryCount,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if snapID.Valid {
		parsed, err := uuid.Parse(snapID.String)
		if err == nil {
			job.SnapshotID = &parsed
		}
	}
	if errStr.Valid {
		job.ErrorMessage = &errStr.String
	}

	return &job, nil
}
