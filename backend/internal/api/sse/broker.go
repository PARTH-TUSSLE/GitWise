package sse

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Event represents a typed Server-Sent Event envelope as specified in V2.1 Section 2.7.
type Event struct {
	Type      string      `json:"type"`
	JobID     string      `json:"jobId"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// ProgressPayload describes job execution progress.
type ProgressPayload struct {
	Stage          string  `json:"stage"`
	CompletedFiles int     `json:"completedFiles"`
	TotalFiles     int     `json:"totalFiles"`
	Percent        float64 `json:"percent"`
	Message        string  `json:"message,omitempty"`
}

// CompletePayload describes job completion details.
type CompletePayload struct {
	Status     string `json:"status"`
	DurationMs int64  `json:"durationMs"`
	SnapshotID string `json:"snapshotId"`
	CommitSHA  string `json:"commitSha"`
}

// ErrorPayload describes a fatal processing failure.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WarningPayload describes a non-fatal analyzer alert.
type WarningPayload struct {
	Code    string `json:"code"`
	File    string `json:"file,omitempty"`
	Message string `json:"message"`
}

// Client represents an active SSE subscriber.
type Client struct {
	JobID   string
	Channel chan Event
}

const (
	ClientBufferSize = 64
	MaxRingHistory   = 16
	PingInterval     = 15 * time.Second
)

// Broker coordinates Server-Sent Events broadcasting, client buffers, and heartbeats.
type Broker struct {
	logger    *slog.Logger
	clients   map[chan Event]string
	history   map[string][]Event
	clientsMu sync.RWMutex
}

// NewBroker initializes an SSE Broker.
func NewBroker(logger *slog.Logger) *Broker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Broker{
		logger:  logger,
		clients: make(map[chan Event]string),
		history: make(map[string][]Event),
	}
}

// Subscribe registers a new subscriber channel for a specific jobID or all jobs ("*").
func (b *Broker) Subscribe(jobID string) (chan Event, func()) {
	ch := make(chan Event, ClientBufferSize)

	b.clientsMu.Lock()
	b.clients[ch] = jobID
	b.clientsMu.Unlock()

	unsubscribe := func() {
		b.clientsMu.Lock()
		if _, ok := b.clients[ch]; ok {
			delete(b.clients, ch)
			close(ch)
		}
		b.clientsMu.Unlock()
	}

	return ch, unsubscribe
}

// GetRecentEvents returns up to MaxRingHistory cached events for micro-reconnect recovery.
func (b *Broker) GetRecentEvents(jobID string) []Event {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	hist, ok := b.history[jobID]
	if !ok || len(hist) == 0 {
		return nil
	}
	res := make([]Event, len(hist))
	copy(res, hist)
	return res
}

// Broadcast distributes an event envelope to all matching subscribers and caches in ring buffer.
func (b *Broker) Broadcast(event Event) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	// Append to bounded in-memory ring buffer (up to 16 events)
	if event.JobID != "" {
		h := b.history[event.JobID]
		if len(h) >= MaxRingHistory {
			h = h[1:]
		}
		b.history[event.JobID] = append(h, event)
	}

	// Dispatch to subscribers
	for ch, targetJob := range b.clients {
		if targetJob == "*" || targetJob == event.JobID {
			select {
			case ch <- event:
			default:
				// Slow client: buffer full, skip or close to prevent goroutine memory leaks
				b.logger.Warn("SSE subscriber channel full, dropping event to prevent block",
					slog.String("job_id", event.JobID),
					slog.String("event_type", event.Type),
				)
			}
		}
	}
}

// BroadcastJobProgress helper to broadcast a job.progress event.
func (b *Broker) BroadcastJobProgress(jobID string, stage string, completed, total int, percent float64, msg string) {
	b.Broadcast(Event{
		Type:      "job.progress",
		JobID:     jobID,
		Timestamp: time.Now().UTC(),
		Payload: ProgressPayload{
			Stage:          stage,
			CompletedFiles: completed,
			TotalFiles:     total,
			Percent:        percent,
			Message:        msg,
		},
	})
}

// BroadcastJobComplete helper to broadcast a complete event.
func (b *Broker) BroadcastJobComplete(jobID string, snapshotID, commitSHA string, durationMs int64) {
	b.Broadcast(Event{
		Type:      "complete",
		JobID:     jobID,
		Timestamp: time.Now().UTC(),
		Payload: CompletePayload{
			Status:     "COMPLETED",
			DurationMs: durationMs,
			SnapshotID: snapshotID,
			CommitSHA:  commitSHA,
		},
	})
}

// BroadcastJobError helper to broadcast an error event.
func (b *Broker) BroadcastJobError(jobID string, code, msg string) {
	b.Broadcast(Event{
		Type:      "error",
		JobID:     jobID,
		Timestamp: time.Now().UTC(),
		Payload: ErrorPayload{
			Code:    code,
			Message: msg,
		},
	})
}

// ServeHTTP handles an incoming SSE client connection for a specific jobID.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request, jobID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported by client connection", http.StatusInternalServerError)
		return
	}

	// SSE response headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Subscribe client
	eventChan, unsubscribe := b.Subscribe(jobID)
	defer unsubscribe()

	// Initial flush to establish connection
	flusher.Flush()

	// Flush any initial buffered events for reconnects
	recent := b.GetRecentEvents(jobID)
	for _, ev := range recent {
		if err := writeSSE(w, ev); err != nil {
			return
		}
		flusher.Flush()
	}

	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected, exit loop cleanly
			return
		case <-ticker.C:
			// Send heartbeat comment to keep proxy connections alive
			if _, err := fmt.Fprintf(w, ":ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case ev, ok := <-eventChan:
			if !ok {
				return
			}
			if err := writeSSE(w, ev); err != nil {
				return
			}
			flusher.Flush()

			// If complete or error event, close stream after delivering
			if ev.Type == "complete" || ev.Type == "error" {
				return
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, ev Event) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(data))
	return err
}
