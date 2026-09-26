package config_test

import (
	"os"
	"testing"

	"github.com/gitwise/backend/internal/config"
)

func TestConfig_Defaults(t *testing.T) {
	// Clear any overrides for test
	os.Unsetenv("APP_ENV")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("DATABASE_URL")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected successful config load, got: %v", err)
	}

	if cfg.HTTPPort != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.HTTPPort)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected default env development, got %s", cfg.AppEnv)
	}
	if cfg.ServerAddress() != "0.0.0.0:8080" {
		t.Errorf("expected 0.0.0.0:8080, got %s", cfg.ServerAddress())
	}
	if cfg.GitHubAPIBaseURL != "https://api.github.com" {
		t.Errorf("expected default github api url https://api.github.com, got %s", cfg.GitHubAPIBaseURL)
	}
}

func TestConfig_CustomEnv(t *testing.T) {
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("APP_ENV", "production")
	defer func() {
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("APP_ENV")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected successful load, got: %v", err)
	}

	if cfg.HTTPPort != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.HTTPPort)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("expected env production, got %s", cfg.AppEnv)
	}
}
