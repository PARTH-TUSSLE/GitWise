package api

import (
	"log/slog"
	"net/http"

	"github.com/gitwise/backend/internal/api/handlers"
	"github.com/gitwise/backend/internal/api/middleware"
	"github.com/gitwise/backend/internal/config"
	"github.com/gitwise/backend/internal/github"
	"github.com/gitwise/backend/internal/service/gitstat"
	"github.com/gitwise/backend/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	chi.Router
	logger *slog.Logger
	cfg    *config.Config
	db     *postgres.DB
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

	gitstatH := handlers.NewGitStatHandler(gitstatSvc, logger)
	repoH := handlers.NewRepoHandler()

	r.Route("/api/v1", func(v1 chi.Router) {
		// Telemetry & GITSTAT routes
		v1.Get("/gitstat/{username}", gitstatH.GetProfile)

		// Repository intelligence routes
		v1.Get("/repositories/{owner}/{repo}", repoH.GetRepository)
	})

	return &Router{
		Router: r,
		logger: logger,
		cfg:    cfg,
		db:     db,
	}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.Router.ServeHTTP(w, r)
}
