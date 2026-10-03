package api

import (
	"log/slog"
	"net/http"

	"github.com/gitwise/backend/internal/api/handlers"
	"github.com/gitwise/backend/internal/api/middleware"
	"github.com/gitwise/backend/internal/api/sse"
	"github.com/gitwise/backend/internal/config"
	"github.com/gitwise/backend/internal/git"
	"github.com/gitwise/backend/internal/github"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/gitwise/backend/internal/service/gitstat"
	"github.com/gitwise/backend/internal/service/repo"
	"github.com/gitwise/backend/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	chi.Router
	logger     *slog.Logger
	cfg        *config.Config
	db         *postgres.DB
	JobManager *jobs.JobManager
	SSEBroker  *sse.Broker
}

func NewRouter(cfg *config.Config, db *postgres.DB, logger *slog.Logger, version string) *Router {
	r := chi.NewRouter()

	// Base Chi middlewares
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.StructuredLogger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORSConfig([]string{cfg.FrontendOrigin}))

	// Health check route
	healthH := handlers.NewHealthHandler(db, version)
	r.Get("/healthz", healthH.ServeHTTP)

	// API v1 services & handlers
	ghClient := github.NewClient(cfg.GitHubAPIBaseURL, cfg.GitHubToken, logger)
	gitstatSvc := gitstat.NewService(ghClient, db, logger)

	// Phase 3 Ingestion, Jobs & SSE Infrastructure
	sseBroker := sse.NewBroker(logger)
	gitFetcher := git.NewGitHubFetcher(cfg.GitHubAPIBaseURL, cfg.GitHubToken, logger)
	var jobManager *jobs.JobManager
	var repoSvc *repo.Service

	if db != nil && db.DB != nil {
		jobManager = jobs.NewJobManager(db.DB, 4, 128, logger, sseBroker)
		repoSvc = repo.NewService(db.DB, gitFetcher, jobManager, logger)
	}

	gitstatH := handlers.NewGitStatHandler(gitstatSvc, logger)
	repoH := handlers.NewRepoHandler(repoSvc, jobManager, sseBroker, logger)

	r.Route("/api/v1", func(v1 chi.Router) {
		// Telemetry & GITSTAT routes
		v1.Get("/gitstat/{username}", gitstatH.GetProfile)

		// Repository intelligence & Ingestion routes (Phase 3 & Phase 4)
		v1.Post("/repositories/ingest", repoH.IngestRepository)
		v1.Get("/repositories/{owner}/{repo}", repoH.GetRepository)
		v1.Get("/repositories/{owner}/{repo}/snapshots/{commitSha}/files", repoH.GetSnapshotFiles)
		v1.Get("/repositories/{owner}/{repo}/snapshots/{commitSha}/symbols", repoH.GetSnapshotSymbols)
		v1.Get("/repositories/{owner}/{repo}/subsystems", repoH.GetSubsystems)
		v1.Get("/repositories/{owner}/{repo}/tree", repoH.GetTree)

		// Code Intelligence Graph & Candidate Impact (Phase 5)
		v1.Get("/repositories/{owner}/{repo}/traces", repoH.GetFeatureTraces)
		v1.Get("/repositories/{owner}/{repo}/traces/{traceId}", repoH.GetFeatureTraceByID)
		v1.Get("/repositories/{owner}/{repo}/impact", repoH.GetCandidateImpact)

		// Tiered Hybrid Retrieval & Search (Phase 6)
		v1.Get("/repositories/{owner}/{repo}/search", repoH.GetSearch)

		// Background Jobs & SSE streaming (Phase 3)
		v1.Get("/jobs/{id}", repoH.GetJob)
		v1.Get("/jobs/{id}/stream", repoH.GetJobStream)
	})

	return &Router{
		Router:     r,
		logger:     logger,
		cfg:        cfg,
		db:         db,
		JobManager: jobManager,
		SSEBroker:  sseBroker,
	}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.Router.ServeHTTP(w, r)
}
