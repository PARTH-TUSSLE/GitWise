package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gitwise/backend/internal/api"
	"github.com/gitwise/backend/internal/config"
	"github.com/gitwise/backend/internal/storage/migrations"
	"github.com/gitwise/backend/internal/storage/postgres"
	"github.com/gitwise/backend/pkg/logger"
)

const AppVersion = "2.1.0-production"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger from pkg/logger
	appLog := logger.New(logger.Config{
		Environment: cfg.AppEnv,
		Level:       cfg.LogLevel,
		AddSource:   cfg.AppEnv == "production",
	})
	slog.SetDefault(appLog)

	appLog.Info("Starting GitWise Backend Server",
		slog.String("version", AppVersion),
		slog.String("environment", cfg.AppEnv),
		slog.String("host", cfg.HTTPHost),
		slog.String("port", cfg.HTTPPort),
		slog.String("frontend_origin", cfg.FrontendOrigin),
	)

	// Context for bootstrap & shutdown
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// Connect to PostgreSQL (fail clearly if unreachable)
	db, err := postgres.New(rootCtx, cfg.DatabaseURL, appLog)
	if err != nil {
		appLog.Error("FATAL: Failed to connect to PostgreSQL database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLog.Error("Error closing database connection", slog.String("error", err.Error()))
		}
	}()

	// Auto-apply pending database schema migrations in version order
	if err := migrations.Run(rootCtx, db.DB, appLog); err != nil {
		appLog.Error("FATAL: Failed to apply database schema migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize Router
	router := api.NewRouter(cfg, db, appLog, AppVersion)

	// Start JobManager workers & reconcile stale in-flight jobs on startup (Phase 3)
	if router.JobManager != nil {
		if err := router.JobManager.Start(rootCtx); err != nil {
			appLog.Error("FATAL: Failed to start background job manager", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer router.JobManager.Stop()
	}

	// Configure HTTP Server with robust timeouts
	serverAddr := cfg.ServerAddress()
	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Channel to listen for shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start server in background goroutine
	serverErrors := make(chan error, 1)
	go func() {
		appLog.Info("HTTP listener active", slog.String("addr", serverAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Wait for shutdown signal or fatal error
	select {
	case err := <-serverErrors:
		appLog.Error("Fatal server error occurred", slog.String("error", err.Error()))
		os.Exit(1)
	case sig := <-shutdownChan:
		appLog.Info("Received termination signal, initiating graceful shutdown", slog.String("signal", sig.String()))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			appLog.Error("Graceful shutdown failed, forcing close", slog.String("error", err.Error()))
			_ = srv.Close()
		} else {
			appLog.Info("Server stopped cleanly")
		}
	}
}
