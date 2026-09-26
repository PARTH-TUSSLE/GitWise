package sse_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/api/sse"
)

func TestBroker_SubscribeAndBroadcast(t *testing.T) {
	broker := sse.NewBroker(nil)
	jobID := "test-job-123"

	ch, unsub := broker.Subscribe(jobID)
	defer unsub()

	// Broadcast progress event
	broker.BroadcastJobProgress(jobID, "FETCHING_TREE", 5, 20, 25.0, "Fetching files")

	select {
	case ev := <-ch:
		if ev.Type != "job.progress" {
			t.Errorf("expected event type 'job.progress', got %s", ev.Type)
		}
		if ev.JobID != jobID {
			t.Errorf("expected job ID %s, got %s", jobID, ev.JobID)
		}
		payload, ok := ev.Payload.(sse.ProgressPayload)
		if !ok {
			t.Fatalf("expected ProgressPayload, got %T", ev.Payload)
		}
		if payload.Percent != 25.0 {
			t.Errorf("expected 25.0 percent, got %f", payload.Percent)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for broadcast event")
	}
}

func TestBroker_RecentHistoryAndFiltering(t *testing.T) {
	broker := sse.NewBroker(nil)
	targetJob := "job-target"
	otherJob := "job-other"

	// Broadcast to other job
	broker.BroadcastJobProgress(otherJob, "INITIALIZING", 0, 10, 0.0, "Init")

	ch, unsub := broker.Subscribe(targetJob)
	defer unsub()

	// Broadcast to target job
	broker.BroadcastJobComplete(targetJob, "snap-1", "sha-abc", 150)

	select {
	case ev := <-ch:
		if ev.JobID != targetJob {
			t.Errorf("expected targetJob, got %s", ev.JobID)
		}
		if ev.Type != "complete" {
			t.Errorf("expected type 'complete', got %s", ev.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for targeted event")
	}

	// Verify history retrieval
	recent := broker.GetRecentEvents(targetJob)
	if len(recent) != 1 {
		t.Fatalf("expected 1 recent event for targetJob, got %d", len(recent))
	}
	if recent[0].Type != "complete" {
		t.Errorf("expected 'complete' in history, got %s", recent[0].Type)
	}
}

func TestBroker_ServeHTTP(t *testing.T) {
	broker := sse.NewBroker(nil)
	jobID := "stream-job-1"

	// Pre-seed an event into history
	broker.BroadcastJobProgress(jobID, "STARTING", 0, 10, 0.0, "Ready")

	req := httptest.NewRequest("GET", "/stream", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	go func() {
		time.Sleep(50 * time.Millisecond)
		broker.BroadcastJobComplete(jobID, "snap-x", "sha-x", 100)
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	broker.ServeHTTP(rec, req, jobID)

	body := rec.Body.String()
	if !strings.Contains(body, "event: job.progress") {
		t.Errorf("expected body to contain 'event: job.progress', got %s", body)
	}
	if !strings.Contains(body, "event: complete") {
		t.Errorf("expected body to contain 'event: complete', got %s", body)
	}
}
