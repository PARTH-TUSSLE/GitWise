package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/ai"
)

func TestOpenAICompatibleClient_Generate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization Bearer test-key, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req["model"] != "llama-3.3-70b-versatile" {
			t.Errorf("expected model llama-3.3-70b-versatile, got %v", req["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"id": "chatcmpl-test",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Explanation with evidence [ev_01]."
					}
				}
			]
		}`)
	}))
	defer ts.Close()

	client := ai.NewOpenAICompatibleClient("test-key", ts.URL, "llama-3.3-70b-versatile")
	resp, err := client.Generate(context.Background(), "Explain this code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(resp, "[ev_01]") {
		t.Errorf("expected response to contain [ev_01], got %q", resp)
	}
}

func TestOpenAICompatibleClient_GenerateStream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
		}

		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}\n\n")
		flusher.Flush()
		time.Sleep(10 * time.Millisecond)

		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"world \"}}]}\n\n")
		flusher.Flush()
		time.Sleep(10 * time.Millisecond)

		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"[ev_02]\"}}]}\n\n")
		flusher.Flush()
		time.Sleep(10 * time.Millisecond)

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer ts.Close()

	client := ai.NewGroqClient("test-key", "llama-3.1-8b-instant")
	// Re-route to test server
	client = ai.NewOpenAICompatibleClient("test-key", ts.URL, "llama-3.1-8b-instant")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.GenerateStream(ctx, "Stream test")
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var sb strings.Builder
	for token := range stream {
		sb.WriteString(token)
	}

	expected := "Hello world [ev_02]"
	if sb.String() != expected {
		t.Errorf("expected full stream %q, got %q", expected, sb.String())
	}
}

func TestOpenAICompatibleClient_ErrorHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, `{"error":{"message":"Rate limit reached"}}`)
	}))
	defer ts.Close()

	client := ai.NewOpenAICompatibleClient("test-key", ts.URL, "test-model")
	_, err := client.Generate(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "429") && !strings.Contains(err.Error(), "Rate limit") {
		t.Errorf("expected error to mention status or rate limit, got %v", err)
	}
}

func TestOpenAICompatibleClient_WithModel(t *testing.T) {
	client := ai.NewGroqClient("test-key", "llama-3.3-70b-versatile")
	if client.Model() != "llama-3.3-70b-versatile" {
		t.Errorf("expected initial model llama-3.3-70b-versatile, got %s", client.Model())
	}

	custom := client.WithModel("llama-3.1-8b-instant")
	if custom.Model() != "llama-3.1-8b-instant" {
		t.Errorf("expected updated model llama-3.1-8b-instant, got %s", custom.Model())
	}
	if custom.BaseURL() != "https://api.groq.com/openai/v1" {
		t.Errorf("expected groq base url, got %s", custom.BaseURL())
	}
}
