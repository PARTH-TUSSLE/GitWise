package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatibleClient connects to any OpenAI-compliant Chat Completions API.
// This supports free & high-speed providers such as Groq, Ollama (local free),
// OpenRouter (free tier models), Together AI, Mistral, DeepSeek, or standard OpenAI.
type OpenAICompatibleClient struct {
	apiKey      string
	baseURL     string
	model       string
	temperature float64
	httpClient  *http.Client
}

// NewOpenAICompatibleClient creates a generic OpenAI-compatible LLM client.
func NewOpenAICompatibleClient(apiKey, baseURL, model string) *OpenAICompatibleClient {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &OpenAICompatibleClient{
		apiKey:      apiKey,
		baseURL:     baseURL,
		model:       model,
		temperature: 0.2,
		httpClient:  &http.Client{Timeout: 90 * time.Second},
	}
}

// NewGroqClient creates an OpenAICompatibleClient pre-configured for Groq's high-speed API.
// Free tier supports llama-3.3-70b-versatile, llama-3.1-8b-instant, etc.
func NewGroqClient(apiKey, model string) *OpenAICompatibleClient {
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return NewOpenAICompatibleClient(apiKey, "https://api.groq.com/openai/v1", model)
}

// NewOllamaClient creates an OpenAICompatibleClient pre-configured for local Ollama instances
// (100% free, offline, no API key needed). Default: http://localhost:11434/v1
func NewOllamaClient(baseURL, model string) *OpenAICompatibleClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	if model == "" {
		model = "llama3.2"
	}
	return NewOpenAICompatibleClient("", baseURL, model)
}

// NewOpenRouterClient creates an OpenAICompatibleClient pre-configured for OpenRouter.
// Supports free tier models (e.g. meta-llama/llama-3.3-70b-instruct:free).
func NewOpenRouterClient(apiKey, model string) *OpenAICompatibleClient {
	if model == "" {
		model = "meta-llama/llama-3.3-70b-instruct:free"
	}
	return NewOpenAICompatibleClient(apiKey, "https://openrouter.ai/api/v1", model)
}

// WithModel returns a copy of the client targeting a different model.
func (c *OpenAICompatibleClient) WithModel(model string) *OpenAICompatibleClient {
	if model == "" {
		model = c.model
	}
	return &OpenAICompatibleClient{
		apiKey:      c.apiKey,
		baseURL:     c.baseURL,
		model:       model,
		temperature: c.temperature,
		httpClient:  c.httpClient,
	}
}

// Model returns the configured model identifier.
func (c *OpenAICompatibleClient) Model() string {
	return c.model
}

// BaseURL returns the configured base URL.
func (c *OpenAICompatibleClient) BaseURL() string {
	return c.baseURL
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
	Stream      bool                `json:"stream"`
}

type openAIChatChoice struct {
	Index   int               `json:"index"`
	Message openAIChatMessage `json:"message"`
}

type openAIChatResponse struct {
	ID      string             `json:"id"`
	Choices []openAIChatChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type openAIStreamChoice struct {
	Index int `json:"index"`
	Delta struct {
		Role    string `json:"role,omitempty"`
		Content string `json:"content,omitempty"`
	} `json:"delta"`
}

type openAIStreamResponse struct {
	ID      string               `json:"id"`
	Choices []openAIStreamChoice `json:"choices"`
}

// Generate sends a unary completion request to the OpenAI-compatible endpoint.
func (c *OpenAICompatibleClient) Generate(ctx context.Context, prompt string) (string, error) {
	endpoint := fmt.Sprintf("%s/chat/completions", c.baseURL)

	reqBody := openAIChatRequest{
		Model: c.model,
		Messages: []openAIChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: c.temperature,
		Stream:      false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal completion request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ai completion request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ai provider error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(errBytes)))
	}

	var chatResp openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode ai response: %w", err)
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("ai provider error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", errors.New("empty choices received from ai model")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GenerateStream sends a streaming request and yields delta tokens via a channel.
func (c *OpenAICompatibleClient) GenerateStream(ctx context.Context, prompt string) (<-chan string, error) {
	endpoint := fmt.Sprintf("%s/chat/completions", c.baseURL)

	reqBody := openAIChatRequest{
		Model: c.model,
		Messages: []openAIChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: c.temperature,
		Stream:      true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal streaming request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ai stream request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ai provider stream error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(errBytes)))
	}

	out := make(chan string, 32)
	go func() {
		defer resp.Body.Close()
		defer close(out)

		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				// Empty line or SSE ping/comment
				continue
			}

			if line == "data: [DONE]" || line == "[DONE]" {
				return
			}

			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				var chunkResp openAIStreamResponse
				if err := json.Unmarshal([]byte(dataStr), &chunkResp); err == nil {
					for _, choice := range chunkResp.Choices {
						if choice.Delta.Content != "" {
							select {
							case <-ctx.Done():
								return
							case out <- choice.Delta.Content:
							}
						}
					}
				}
			}
		}
	}()

	return out, nil
}
