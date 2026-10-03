package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/service/pr"
	"github.com/go-chi/chi/v5"
)

// PRHandler handles PR intelligence and semantic review endpoints.
type PRHandler struct {
	prSvc  *pr.Service
	logger *slog.Logger
}

// NewPRHandler creates a new PR handler instance.
func NewPRHandler(prSvc *pr.Service, logger *slog.Logger) *PRHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PRHandler{
		prSvc:  prSvc,
		logger: logger,
	}
}

// GetPullRequests handles GET /api/v1/pr/{owner}/{repo}
func (h *PRHandler) GetPullRequests(w http.ResponseWriter, r *http.Request) {
	if h.prSvc == nil {
		http.Error(w, `{"error":"PR service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	if owner == "" || repo == "" {
		http.Error(w, `{"error":"owner and repo are required"}`, http.StatusBadRequest)
		return
	}

	prs, err := h.prSvc.GetPullRequests(r.Context(), owner, repo)
	if err != nil {
		h.logger.WarnContext(r.Context(), "failed to get pull requests", "owner", owner, "repo", repo, "error", err)
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prs)
}

// GetPullRequest handles GET /api/v1/pr/{owner}/{repo}/{number}
func (h *PRHandler) GetPullRequest(w http.ResponseWriter, r *http.Request) {
	if h.prSvc == nil {
		http.Error(w, `{"error":"PR service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	numStr := chi.URLParam(r, "number")

	number, err := strconv.Atoi(numStr)
	if err != nil || number <= 0 {
		http.Error(w, `{"error":"invalid pull request number"}`, http.StatusBadRequest)
		return
	}

	prModel, err := h.prSvc.GetPullRequest(r.Context(), owner, repo, number)
	if err != nil {
		h.logger.WarnContext(r.Context(), "failed to get pull request", "owner", owner, "repo", repo, "number", number, "error", err)
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prModel)
}

// ReviewPullRequest handles POST /api/v1/pr/{owner}/{repo}/{number}/review
func (h *PRHandler) ReviewPullRequest(w http.ResponseWriter, r *http.Request) {
	if h.prSvc == nil {
		http.Error(w, `{"error":"PR service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	numStr := chi.URLParam(r, "number")

	number, err := strconv.Atoi(numStr)
	if err != nil || number <= 0 {
		http.Error(w, `{"error":"invalid pull request number"}`, http.StatusBadRequest)
		return
	}

	var req domain.ReviewPullRequestRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	prModel, err := h.prSvc.ReviewPullRequest(r.Context(), owner, repo, number, req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to review pull request", "owner", owner, "repo", repo, "number", number, "error", err)
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prModel)
}
