package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gitwise/backend/internal/storage/postgres"
)

type HealthHandler struct {
	db        *postgres.DB
	startTime time.Time
	version   string
}

func NewHealthHandler(db *postgres.DB, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbStatus := "disconnected"
	if h.db != nil {
		if err := h.db.CheckHealth(r.Context()); err == nil {
			dbStatus = "connected"
		}
	}

	resp := HealthResponse{
		Status:    "ok",
		Version:   h.version,
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		Database:  dbStatus,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
