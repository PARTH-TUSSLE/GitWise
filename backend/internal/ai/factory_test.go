package ai_test

import (
	"testing"

	"github.com/gitwise/backend/internal/ai"
)

func TestFactory_NewClient(t *testing.T) {
	tests := []struct {
		name         string
		cfg          ai.Config
		wantType     string
		checkDetails func(t *testing.T, client ai.Client)
	}{
		{
			name: "Groq provider with key",
			cfg: ai.Config{
				Provider: ai.ProviderGroq,
				APIKey:   "gsk_test123",
				Model:    "llama-3.3-70b-versatile",
			},
			wantType: "*ai.OpenAICompatibleClient",
			checkDetails: func(t *testing.T, client ai.Client) {
				oai, ok := client.(*ai.OpenAICompatibleClient)
				if !ok {
					t.Fatalf("expected *ai.OpenAICompatibleClient, got %T", client)
				}
				if oai.Model() != "llama-3.3-70b-versatile" {
					t.Errorf("expected model llama-3.3-70b-versatile, got %s", oai.Model())
				}
				if oai.BaseURL() != "https://api.groq.com/openai/v1" {
					t.Errorf("expected groq base url, got %s", oai.BaseURL())
				}
			},
		},
		{
			name: "Ollama provider without key (local, free)",
			cfg: ai.Config{
				Provider: ai.ProviderOllama,
				BaseURL:  "http://localhost:11434/v1",
				Model:    "llama3.2",
			},
			wantType: "*ai.OpenAICompatibleClient",
			checkDetails: func(t *testing.T, client ai.Client) {
				oai, ok := client.(*ai.OpenAICompatibleClient)
				if !ok {
					t.Fatalf("expected *ai.OpenAICompatibleClient, got %T", client)
				}
				if oai.Model() != "llama3.2" {
					t.Errorf("expected model llama3.2, got %s", oai.Model())
				}
				if oai.BaseURL() != "http://localhost:11434/v1" {
					t.Errorf("expected ollama url, got %s", oai.BaseURL())
				}
			},
		},
		{
			name: "OpenRouter provider",
			cfg: ai.Config{
				Provider: ai.ProviderOpenRouter,
				APIKey:   "sk-or-test",
			},
			wantType: "*ai.OpenAICompatibleClient",
			checkDetails: func(t *testing.T, client ai.Client) {
				oai, ok := client.(*ai.OpenAICompatibleClient)
				if !ok {
					t.Fatalf("expected *ai.OpenAICompatibleClient, got %T", client)
				}
				if oai.Model() != "meta-llama/llama-3.3-70b-instruct:free" {
					t.Errorf("expected free model, got %s", oai.Model())
				}
			},
		},
		{
			name: "Gemini provider with key",
			cfg: ai.Config{
				Provider: ai.ProviderGemini,
				APIKey:   "AIzaSyTestKey",
				Model:    "gemini-1.5-flash",
			},
			wantType: "*ai.GeminiClient",
		},
		{
			name: "Groq provider without key falls back to MockClient safely",
			cfg: ai.Config{
				Provider: ai.ProviderGroq,
				APIKey:   "",
			},
			wantType: "*ai.MockClient",
		},
		{
			name: "Auto-detect Groq from key prefix gsk_",
			cfg: ai.Config{
				APIKey: "gsk_auto_detect",
			},
			wantType: "*ai.OpenAICompatibleClient",
			checkDetails: func(t *testing.T, client ai.Client) {
				oai, ok := client.(*ai.OpenAICompatibleClient)
				if !ok {
					t.Fatalf("expected *ai.OpenAICompatibleClient, got %T", client)
				}
				if oai.BaseURL() != "https://api.groq.com/openai/v1" {
					t.Errorf("expected groq url, got %s", oai.BaseURL())
				}
			},
		},
		{
			name:     "Empty config defaults to MockClient",
			cfg:      ai.Config{},
			wantType: "*ai.MockClient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := ai.NewClient(tt.cfg)
			if client == nil {
				t.Fatal("expected non-nil client")
			}
			if tt.checkDetails != nil {
				tt.checkDetails(t, client)
			}
		})
	}
}
