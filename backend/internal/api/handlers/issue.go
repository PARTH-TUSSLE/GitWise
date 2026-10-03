package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/service/issue"
	"github.com/go-chi/chi/v5"
)

// IssueHandler handles issue intelligence and blueprint endpoints.
type IssueHandler struct {
	issueSvc *issue.Service
	logger   *slog.Logger
}

// NewIssueHandler creates a new issue handler.
func NewIssueHandler(issueSvc *issue.Service, logger *slog.Logger) *IssueHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &IssueHandler{
		issueSvc: issueSvc,
		logger:   logger,
	}
}

// GetIssues handles GET /api/v1/issues/{owner}/{repo}
func (h *IssueHandler) GetIssues(w http.ResponseWriter, r *http.Request) {
	if h.issueSvc == nil {
		http.Error(w, `{"error":"Issue service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	if owner == "" || repo == "" {
		http.Error(w, `{"error":"owner and repo are required"}`, http.StatusBadRequest)
		return
	}

	issues, err := h.issueSvc.GetIssues(r.Context(), owner, repo)
	if err != nil {
		h.logger.WarnContext(r.Context(), "failed to get issues", "owner", owner, "repo", repo, "error", err)
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(issues)
}

// GetIssue handles GET /api/v1/issues/{owner}/{repo}/{number}
func (h *IssueHandler) GetIssue(w http.ResponseWriter, r *http.Request) {
	if h.issueSvc == nil {
		http.Error(w, `{"error":"Issue service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	numStr := chi.URLParam(r, "number")

	number, err := strconv.Atoi(numStr)
	if err != nil || number <= 0 {
		http.Error(w, `{"error":"invalid issue number"}`, http.StatusBadRequest)
		return
	}

	issueModel, err := h.issueSvc.GetIssue(r.Context(), owner, repo, number)
	if err != nil {
		h.logger.WarnContext(r.Context(), "failed to get issue", "owner", owner, "repo", repo, "number", number, "error", err)
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(issueModel)
}

// GenerateBlueprint handles POST /api/v1/issues/{owner}/{repo}/{number}/blueprint
func (h *IssueHandler) GenerateBlueprint(w http.ResponseWriter, r *http.Request) {
	if h.issueSvc == nil {
		http.Error(w, `{"error":"Issue service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	numStr := chi.URLParam(r, "number")

	number, err := strconv.Atoi(numStr)
	if err != nil || number <= 0 {
		http.Error(w, `{"error":"invalid issue number"}`, http.StatusBadRequest)
		return
	}

	var req domain.CreateBlueprintRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	issueModel, err := h.issueSvc.GenerateBlueprint(r.Context(), owner, repo, number, req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to generate blueprint", "owner", owner, "repo", repo, "number", number, "error", err)
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(issueModel)
}
