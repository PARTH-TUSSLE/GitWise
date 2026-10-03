package ai

import (
	"strings"
)

// ProviderType identifies supported LLM backends.
type ProviderType string

const (
	ProviderGroq             ProviderType = "groq"
	ProviderOllama           ProviderType = "ollama"
	ProviderOpenRouter       ProviderType = "openrouter"
	ProviderOpenAI           ProviderType = "openai"
	ProviderOpenAICompatible ProviderType = "openai_compatible"
	ProviderGemini           ProviderType = "gemini"
	ProviderMock             ProviderType = "mock"
)

// Config defines connection details for constructing an AI Client.
type Config struct {
	Provider ProviderType
	APIKey   string
	BaseURL  string
	Model    string
}

// NewClient returns a Client implementation based on the specified provider and configuration.
// If credentials are not supplied for paid or remote providers, it gracefully falls back
// to MockClient to ensure local development, tests, and CI succeed seamlessly.
func NewClient(cfg Config) Client {
	provider := ProviderType(strings.ToLower(strings.TrimSpace(string(cfg.Provider))))
	apiKey := strings.TrimSpace(cfg.APIKey)
	baseURL := strings.TrimSpace(cfg.BaseURL)
	model := strings.TrimSpace(cfg.Model)

	// Auto-detect provider if unspecified
	if provider == "" {
		if strings.HasPrefix(apiKey, "gsk_") {
			provider = ProviderGroq
		} else if strings.HasPrefix(apiKey, "AIzaSy") {
			provider = ProviderGemini
		} else if strings.HasPrefix(apiKey, "sk-or-") {
			provider = ProviderOpenRouter
		} else if strings.HasPrefix(apiKey, "sk-") {
			provider = ProviderOpenAI
		} else if baseURL != "" {
			provider = ProviderOpenAICompatible
		} else if apiKey != "" {
			provider = ProviderGroq
		} else {
			provider = ProviderMock
		}
	}

	switch provider {
	case ProviderGroq:
		if apiKey == "" {
			return NewMockClient()
		}
		if model == "" {
			model = "llama-3.3-70b-versatile"
		}
		return NewGroqClient(apiKey, model)

	case ProviderOllama:
		// Ollama is 100% free and local, requires no API key
		if baseURL == "" {
			baseURL = "http://localhost:11434/v1"
		}
		if model == "" {
			model = "llama3.2"
		}
		return NewOllamaClient(baseURL, model)

	case ProviderOpenRouter:
		if apiKey == "" {
			return NewMockClient()
		}
		if model == "" {
			model = "meta-llama/llama-3.3-70b-instruct:free"
		}
		return NewOpenRouterClient(apiKey, model)

	case ProviderOpenAI:
		if apiKey == "" {
			return NewMockClient()
		}
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		if model == "" {
			model = "gpt-4o-mini"
		}
		return NewOpenAICompatibleClient(apiKey, baseURL, model)

	case ProviderOpenAICompatible:
		if baseURL == "" {
			baseURL = "https://api.groq.com/openai/v1"
		}
		if model == "" {
			model = "llama-3.3-70b-versatile"
		}
		return NewOpenAICompatibleClient(apiKey, baseURL, model)

	case ProviderGemini:
		if apiKey == "" {
			return NewMockClient()
		}
		if model == "" {
			model = "gemini-1.5-flash"
		}
		return NewGeminiClient(apiKey, model)

	case ProviderMock:
		fallthrough
	default:
		return NewMockClient()
	}
}
