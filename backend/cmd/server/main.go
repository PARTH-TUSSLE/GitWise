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
)

const AppVersion = "2.1.0-phase2"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	if cfg.AppEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("Starting GitWise Backend Server",
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
	db, err := postgres.New(rootCtx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("FATAL: Failed to connect to PostgreSQL database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error closing database connection", slog.String("error", err.Error()))
		}
	}()

	// Auto-apply Phase 1 & Phase 2 database schema migrations
	if err := db.ApplyMigrations(rootCtx, migrations.InitSchemaUp); err != nil {
		logger.Error("FATAL: Failed to apply Phase 1 database migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := db.ApplyMigrations(rootCtx, migrations.GithubGitstatUp); err != nil {
		logger.Error("FATAL: Failed to apply Phase 2 database migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize Router
	router := api.NewRouter(cfg, db, logger, AppVersion)

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
		logger.Info("HTTP listener active", slog.String("addr", serverAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Wait for shutdown signal or fatal error
	select {
	case err := <-serverErrors:
		logger.Error("Fatal server error occurred", slog.String("error", err.Error()))
		os.Exit(1)
	case sig := <-shutdownChan:
		logger.Info("Received termination signal, initiating graceful shutdown", slog.String("signal", sig.String()))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("Graceful shutdown failed, forcing close", slog.String("error", err.Error()))
			_ = srv.Close()
		} else {
			logger.Info("Server stopped cleanly")
		}
	}
}
