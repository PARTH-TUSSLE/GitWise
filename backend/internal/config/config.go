package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config represents strongly-typed configuration for the GitWise backend.
type Config struct {
	AppEnv           string `env:"APP_ENV" envDefault:"development"`
	HTTPHost         string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	HTTPPort         string `env:"HTTP_PORT" envDefault:"8080"`
	DatabaseURL      string `env:"DATABASE_URL" envDefault:"postgres://gitwise:gitwise@localhost:5432/gitwise?sslmode=disable"`
	FrontendOrigin   string `env:"FRONTEND_ORIGIN" envDefault:"http://localhost:3000"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
	GitHubToken      string `env:"GITHUB_TOKEN"`
	GitHubAPIBaseURL string `env:"GITHUB_API_BASE_URL" envDefault:"https://api.github.com"`

	// AI Provider & Model Configuration (Groq, Ollama, OpenRouter, OpenAI, Gemini, etc.)
	AIProvider   string `env:"AI_PROVIDER" envDefault:"groq"`
	AIAPIKey     string `env:"AI_API_KEY"`
	AIBaseURL    string `env:"AI_BASE_URL"`
	AIModel      string `env:"AI_MODEL"`
	GroqAPIKey   string `env:"GROQ_API_KEY"`
	GeminiAPIKey string `env:"GEMINI_API_KEY"`
}

// Load parses environment variables into Config with fail-fast validation.
func Load() (*Config, error) {
	// Attempt loading local .env if present; ignore if missing
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment configuration: %w", err)
	}

	// Auto-detect and normalize AI provider credentials
	if cfg.AIAPIKey == "" {
		if cfg.GroqAPIKey != "" {
			cfg.AIAPIKey = cfg.GroqAPIKey
			if cfg.AIProvider == "" {
				cfg.AIProvider = "groq"
			}
		} else if cfg.GeminiAPIKey != "" {
			cfg.AIAPIKey = cfg.GeminiAPIKey
			if cfg.AIProvider == "" || cfg.AIProvider == "groq" {
				cfg.AIProvider = "gemini"
			}
		}
	}

	if cfg.AIProvider == "groq" && cfg.AIModel == "" {
		cfg.AIModel = "llama-3.3-70b-versatile"
	}

	if cfg.HTTPPort == "" {
		return nil, fmt.Errorf("HTTP_PORT cannot be empty")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL cannot be empty")
	}

	return cfg, nil
}

// ServerAddress returns the listen address formatted as host:port.
func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%s", c.HTTPHost, c.HTTPPort)
}
