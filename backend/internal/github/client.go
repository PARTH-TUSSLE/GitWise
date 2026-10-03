package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitInfo captures current GitHub API quota status.
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
	Resource  string    `json:"resource"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Client provides resilient access to GitHub REST and GraphQL APIs.
type Client struct {
	httpClient *http.Client
	baseURL    string
	graphqlURL string
	token      string
	logger     *slog.Logger

	mu        sync.RWMutex
	rateLimit RateLimitInfo
}

// NewClient initializes a new GitHub client.
func NewClient(baseURL, token string, logger *slog.Logger) *Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	graphqlURL := "https://api.github.com/graphql"
	if baseURL != "https://api.github.com" {
		// Allows tests to route graphql to the same mock server
		graphqlURL = baseURL + "/graphql"
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL:    baseURL,
		graphqlURL: graphqlURL,
		token:      token,
		logger:     logger,
	}
}

// Do executes an authenticated or anonymous HTTP request to GitHub and tracks rate limits.
// It applies exponential backoff retry on transient 5xx server errors.
func (c *Client) Do(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	url := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		url = fmt.Sprintf("%s/%s", c.baseURL, strings.TrimLeft(endpoint, "/"))
	}

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	const maxRetries = 2
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to construct request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "GitWise-Backend/2.1")
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		c.updateRateLimit(resp.Header)

		// Retry on transient gateway/server errors (502, 503, 504)
		if resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			_ = resp.Body.Close()
			lastResp = resp
			continue
		}

		return resp, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("github api request failed after retries: %w", lastErr)
	}
	return lastResp, nil
}

// updateRateLimit extracts rate-limit tracking headers from the response.
func (c *Client) updateRateLimit(h http.Header) {
	limitStr := h.Get("X-RateLimit-Limit")
	remStr := h.Get("X-RateLimit-Remaining")
	resetStr := h.Get("X-RateLimit-Reset")
	resource := h.Get("X-RateLimit-Resource")

	if limitStr == "" && remStr == "" && resetStr == "" {
		return
	}

	limit, _ := strconv.Atoi(limitStr)
	remaining, _ := strconv.Atoi(remStr)

	var resetTime time.Time
	if resetEpoch, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
		resetTime = time.Unix(resetEpoch, 0)
	}

	c.mu.Lock()
	c.rateLimit = RateLimitInfo{
		Limit:     limit,
		Remaining: remaining,
		Reset:     resetTime,
		Resource:  resource,
		UpdatedAt: time.Now(),
	}
	c.mu.Unlock()

	if remaining <= 5 && limit > 0 {
		c.logger.Warn("GitHub API rate limit nearly exhausted",
			slog.Int("remaining", remaining),
			slog.Int("limit", limit),
			slog.Time("reset", resetTime),
		)
	}
}

// GetRateLimit returns the most recently observed rate limit status.
func (c *Client) GetRateLimit() RateLimitInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rateLimit
}

// HasToken indicates whether the client has an authentication token configured.
func (c *Client) HasToken() bool {
	return c.token != ""
}

// checkResponseStatus validates response code and decodes GitHub errors.
func (c *Client) checkResponseStatus(resp *http.Response, resource, id string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	var ghErr struct {
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url"`
	}
	_ = json.Unmarshal(bodyBytes, &ghErr)
	msg := ghErr.Message
	if msg == "" {
		msg = string(bodyBytes)
	}

	// 404 Not Found
	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{
			Resource:   resource,
			Identifier: id,
			Message:    msg,
		}
	}

	// 403 or 429 Rate Limit Exhaustion
	rl := c.GetRateLimit()
	if resp.StatusCode == http.StatusTooManyRequests ||
		(resp.StatusCode == http.StatusForbidden && (rl.Remaining == 0 || strings.Contains(strings.ToLower(msg), "rate limit"))) {
		return &RateLimitError{
			Limit:     rl.Limit,
			Remaining: rl.Remaining,
			Reset:     rl.Reset,
			Resource:  rl.Resource,
			Message:   msg,
		}
	}

	return fmt.Errorf("github api error (%d %s): %s", resp.StatusCode, resp.Status, msg)
}
