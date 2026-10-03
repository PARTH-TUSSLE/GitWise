package domain

import (
	"time"

	"github.com/google/uuid"
)

// MentorSession models an interactive mentorship conversation session bound to a snapshot.
type MentorSession struct {
	ID         uuid.UUID `json:"id"`
	SnapshotID uuid.UUID `json:"snapshotId"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// MentorMessage models a persistent chat message within a mentor session.
type MentorMessage struct {
	ID        uuid.UUID      `json:"id"`
	SessionID uuid.UUID      `json:"sessionId"`
	Role      string         `json:"role"` // "user", "assistant", "mentor", "gitwise"
	Content   string         `json:"content"`
	Citations []ChatCitation `json:"citations,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

// ChatRequest represents the payload for chat interactions.
type ChatRequest struct {
	SessionID *uuid.UUID `json:"sessionId,omitempty"`
	Ref       string     `json:"ref,omitempty"`
	Message   string     `json:"message"`
	Stream    bool       `json:"stream,omitempty"`
	TopK      int        `json:"topK,omitempty"`

	// Optional runtime model/provider overrides
	Provider string `json:"provider,omitempty"` // "groq", "ollama", "openrouter", "openai", "gemini"
	Model    string `json:"model,omitempty"`    // e.g. "llama-3.3-70b-versatile", "llama-3.1-8b-instant", "llama3.2"
	BaseURL  string `json:"baseUrl,omitempty"`  // custom proxy or local endpoint
}

// ChatResponse is the response returned for synchronous chat interactions.
type ChatResponse struct {
	SessionID uuid.UUID      `json:"sessionId"`
	Message   string         `json:"message"`
	Citations []ChatCitation `json:"citations"`
}

// ChatStreamChunk models the SSE payload during token/chunk streaming.
type ChatStreamChunk struct {
	SessionID uuid.UUID      `json:"sessionId"`
	Delta     string         `json:"delta,omitempty"`
	Done      bool           `json:"done"`
	Citations []ChatCitation `json:"citations,omitempty"`
	Error     string         `json:"error,omitempty"`
}
