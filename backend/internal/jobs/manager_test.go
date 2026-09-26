package jobs_test

import (
	"context"
	"testing"

	"github.com/gitwise/backend/internal/api/sse"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/google/uuid"
)

func TestJobManager_LifecycleAndEnqueue(t *testing.T) {
	broker := sse.NewBroker(nil)
	jm := jobs.NewJobManager(nil, 2, 10, nil, broker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Enqueue before start
	testID := uuid.New()
	err := jm.Enqueue(ctx, testID)
	if err != nil {
		t.Fatalf("unexpected error enqueuing job: %v", err)
	}

	// Register test handler
	handled := make(chan uuid.UUID, 1)
	jm.RegisterHandler(domain.JobTypeSnapshotIngest, func(ctx context.Context, jobID uuid.UUID) error {
		handled <- jobID
		return nil
	})

	// Stop without start should be a no-op
	jm.Stop()
}

func TestJobManager_ReconcileStaleJobs_NilDB(t *testing.T) {
	jm := jobs.NewJobManager(nil, 2, 10, nil, nil)
	ctx := context.Background()

	err := jm.ReconcileStaleJobs(ctx)
	if err == nil {
		t.Fatal("expected error when DB is nil, got nil")
	}
}

func TestJobManager_QueueFullError(t *testing.T) {
	jm := jobs.NewJobManager(nil, 1, 1, nil, nil)
	ctx := context.Background()

	err1 := jm.Enqueue(ctx, uuid.New())
	if err1 != nil {
		t.Fatalf("first enqueue failed: %v", err1)
	}

	// Second enqueue should fail because queueSize is 1
	err2 := jm.Enqueue(ctx, uuid.New())
	if err2 == nil {
		t.Fatal("expected queue full error, got nil")
	}
}

func TestJobManager_ContextCancellation(t *testing.T) {
	jm := jobs.NewJobManager(nil, 1, 1, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := jm.Enqueue(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestJobManager_DuplicateEnqueue(t *testing.T) {
	jm := jobs.NewJobManager(nil, 1, 5, nil, nil)
	ctx := context.Background()
	jobID := uuid.New()

	if err := jm.Enqueue(ctx, jobID); err != nil {
		t.Fatalf("first enqueue failed: %v", err)
	}

	// Second enqueue with exact same job ID should be suppressed cleanly without error
	if err := jm.Enqueue(ctx, jobID); err != nil {
		t.Fatalf("duplicate enqueue should succeed as no-op: %v", err)
	}
}

func TestJobManager_RecoverQueuedJobs_NilDB(t *testing.T) {
	jm := jobs.NewJobManager(nil, 1, 5, nil, nil)
	ctx := context.Background()

	err := jm.RecoverQueuedJobs(ctx)
	if err == nil {
		t.Fatal("expected error with nil DB, got nil")
	}
}
