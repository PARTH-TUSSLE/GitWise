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

// Client defines the contract for LLM generation.
type Client interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string) (<-chan string, error)
}

// MockClient produces deterministic grounded answers for tests, offline dev, and CI.
type MockClient struct{}

// NewMockClient creates a new mock AI client.
func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) Generate(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Parse available evidence anchors in prompt e.g. [ev_01]
	hasEv01 := strings.Contains(prompt, "ev_01")
	hasEv02 := strings.Contains(prompt, "ev_02")

	var sb strings.Builder
	sb.WriteString("Based on the repository source structure and verified code citations:\n\n")

	if hasEv01 {
		sb.WriteString("The primary component logic is declared in [ev_01], defining the core execution behavior.\n")
	}
	if hasEv02 {
		sb.WriteString("Supporting interactions and data dependencies are coordinated through [ev_02].\n")
	}
	if !hasEv01 && !hasEv02 {
		sb.WriteString("The requested area implements standard modular entrypoints and interface contracts.\n")
	}

	sb.WriteString("\nYou can explore the exact file lines in the citations panel below.")
	return sb.String(), nil
}

func (m *MockClient) GenerateStream(ctx context.Context, prompt string) (<-chan string, error) {
	text, err := m.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	out := make(chan string, 16)
	words := strings.Split(text, " ")

	go func() {
		defer close(out)
		for i, w := range words {
			select {
			case <-ctx.Done():
				return
			default:
				if i > 0 {
					out <- " " + w
				} else {
					out <- w
				}
				time.Sleep(10 * time.Millisecond) // Smooth token pacing
			}
		}
	}()

	return out, nil
}

// GeminiClient connects to the Google Gemini API.
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiClient creates a Gemini API client.
func NewGeminiClient(apiKey, model string) *GeminiClient {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiCandidate struct {
	Content struct {
		Parts []geminiPart `json:"parts"`
	} `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

func (g *GeminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	if g.apiKey == "" {
		return "", errors.New("gemini api key is required")
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model, g.apiKey,
	)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("gemini api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(errBytes))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from gemini model")
	}

	var sb strings.Builder
	for _, p := range geminiResp.Candidates[0].Content.Parts {
		sb.WriteString(p.Text)
	}

	return sb.String(), nil
}

func (g *GeminiClient) GenerateStream(ctx context.Context, prompt string) (<-chan string, error) {
	if g.apiKey == "" {
		return nil, errors.New("gemini api key is required")
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s",
		g.model, g.apiKey,
	)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini stream request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini stream request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("gemini stream api error (status %d): %s", resp.StatusCode, string(errBytes))
	}

	out := make(chan string, 16)
	go func() {
		defer resp.Body.Close()
		defer close(out)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				var chunkResp geminiResponse
				if err := json.Unmarshal([]byte(dataStr), &chunkResp); err == nil {
					for _, cand := range chunkResp.Candidates {
						for _, p := range cand.Content.Parts {
							if p.Text != "" {
								out <- p.Text
							}
						}
					}
				}
			}
		}
	}()

	return out, nil
}
