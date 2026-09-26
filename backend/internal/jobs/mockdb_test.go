package jobs_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/google/uuid"
)

// fakeJobsDriver implements a minimal driver for JobManager tests.
type fakeJobsDriver struct {
	mu      sync.Mutex
	handler func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error)
}

var (
	testJobsDriver = &fakeJobsDriver{}
	initOnce       sync.Once
)

func init() {
	initOnce.Do(func() {
		sql.Register("fake_jobs_driver", testJobsDriver)
	})
}

func (d *fakeJobsDriver) Open(name string) (driver.Conn, error) {
	return &fakeJobsConn{driver: d}, nil
}

type fakeJobsConn struct {
	driver *fakeJobsDriver
}

func (c *fakeJobsConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeJobsStmt{conn: c, query: query}, nil
}

func (c *fakeJobsConn) Close() error {
	return nil
}

func (c *fakeJobsConn) Begin() (driver.Tx, error) {
	return &fakeJobsTx{}, nil
}

func (c *fakeJobsConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		rows, _, err := c.driver.handler(query, args)
		return rows, err
	}
	return &fakeJobsRows{}, nil
}

func (c *fakeJobsConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	if c.driver.handler != nil {
		_, res, err := c.driver.handler(query, args)
		if res == nil && err == nil {
			res = driver.RowsAffected(1)
		}
		return res, err
	}
	return driver.RowsAffected(1), nil
}

type fakeJobsStmt struct {
	conn  *fakeJobsConn
	query string
}

func (s *fakeJobsStmt) Close() error  { return nil }
func (s *fakeJobsStmt) NumInput() int { return -1 }
func (s *fakeJobsStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (s *fakeJobsStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &fakeJobsRows{}, nil
}

type fakeJobsTx struct{}

func (t *fakeJobsTx) Commit() error   { return nil }
func (t *fakeJobsTx) Rollback() error { return nil }

type fakeJobsRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func (r *fakeJobsRows) Columns() []string { return r.cols }
func (r *fakeJobsRows) Close() error      { return nil }
func (r *fakeJobsRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.idx]
	for i, val := range row {
		dest[i] = val
	}
	r.idx++
	return nil
}

func TestJobManager_RecoverQueuedJobs_Success(t *testing.T) {
	db, err := sql.Open("fake_jobs_driver", "test1")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	queuedJobID1 := uuid.New()
	queuedJobID2 := uuid.New()

	testJobsDriver.mu.Lock()
	testJobsDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		return &fakeJobsRows{
			cols: []string{"id"},
			data: [][]driver.Value{
				{queuedJobID1.String()},
				{queuedJobID2.String()},
			},
		}, nil, nil
	}
	testJobsDriver.mu.Unlock()

	jm := jobs.NewJobManager(db, 2, 10, nil, nil)
	ctx := context.Background()

	if err := jm.RecoverQueuedJobs(ctx); err != nil {
		t.Fatalf("unexpected error recovering queued jobs: %v", err)
	}

	// Verify both jobs are now enqueued in the manager and duplicate enqueue is a no-op
	if err := jm.Enqueue(ctx, queuedJobID1); err != nil {
		t.Fatalf("duplicate enqueue should be suppressed cleanly: %v", err)
	}
	if err := jm.Enqueue(ctx, queuedJobID2); err != nil {
		t.Fatalf("duplicate enqueue should be suppressed cleanly: %v", err)
	}
}

func TestJobManager_ReconcileStaleJobs_Success(t *testing.T) {
	db, err := sql.Open("fake_jobs_driver", "test2")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	var queries []string
	testJobsDriver.mu.Lock()
	testJobsDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		queries = append(queries, query)
		return nil, driver.RowsAffected(3), nil
	}
	testJobsDriver.mu.Unlock()

	jm := jobs.NewJobManager(db, 2, 10, nil, nil)
	ctx := context.Background()

	if err := jm.ReconcileStaleJobs(ctx); err != nil {
		t.Fatalf("unexpected error reconciling stale jobs: %v", err)
	}

	if len(queries) < 2 {
		t.Fatalf("expected at least 2 reconciliation queries (jobs + snapshots), got %d", len(queries))
	}
}

func TestJobManager_StartupRecoveryAndWorkerProcessing(t *testing.T) {
	db, err := sql.Open("fake_jobs_driver", "test3")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	targetJobID := uuid.New()
	processed := make(chan uuid.UUID, 1)

	testJobsDriver.mu.Lock()
	testJobsDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "WHERE status = 'QUEUED'") {
			return &fakeJobsRows{
				cols: []string{"id"},
				data: [][]driver.Value{{targetJobID.String()}},
			}, driver.RowsAffected(1), nil
		}
		if strings.Contains(query, "SELECT") && strings.Contains(query, "analysis_jobs") {
			now := time.Now()
			return &fakeJobsRows{
				cols: []string{"id", "type", "snapshot_id", "status", "stage", "progress_percent", "error_message", "retry_count", "created_at", "updated_at"},
				data: [][]driver.Value{
					{targetJobID.String(), string(domain.JobTypeSnapshotIngest), nil, string(domain.JobStatusQueued), string(domain.JobStageInitializing), 0.0, nil, 0, now, now},
				},
			}, driver.RowsAffected(1), nil
		}
		return &fakeJobsRows{}, driver.RowsAffected(1), nil
	}
	testJobsDriver.mu.Unlock()

	jm := jobs.NewJobManager(db, 1, 5, nil, nil)
	jm.RegisterHandler(domain.JobTypeSnapshotIngest, func(ctx context.Context, jobID uuid.UUID) error {
		processed <- jobID
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := jm.Start(ctx); err != nil {
		t.Fatalf("failed to start JobManager: %v", err)
	}
	defer jm.Stop()

	select {
	case id := <-processed:
		if id != targetJobID {
			t.Fatalf("expected processed job ID %s, got %s", targetJobID, id)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for recovered queued job to be processed by worker")
	}
}
