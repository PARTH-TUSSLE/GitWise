package retrieval

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Embedder provides embedding vector generation for semantic retrieval.
type Embedder interface {
	ModelName() string
	Dimensions() int
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// MockEmbedder produces fast, deterministic, unit-normalized 768-dimensional float32
// embeddings for unit tests, benchmarking, and offline local development.
type MockEmbedder struct {
	modelName  string
	dimensions int
}

// NewMockEmbedder creates a mock embedder with standard 768 dimensions.
func NewMockEmbedder() *MockEmbedder {
	return &MockEmbedder{
		modelName:  "mock-768",
		dimensions: 768,
	}
}

func (m *MockEmbedder) ModelName() string {
	return m.modelName
}

func (m *MockEmbedder) Dimensions() int {
	return m.dimensions
}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	vecs, err := m.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return vecs[0], nil
}

func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	results := make([][]float32, len(texts))
	for i, text := range texts {
		results[i] = m.embedSingle(text)
	}
	return results, nil
}

func (m *MockEmbedder) embedSingle(text string) []float32 {
	vec := make([]float32, m.dimensions)
	clean := strings.ToLower(strings.TrimSpace(text))
	if clean == "" {
		vec[0] = 1.0
		return vec
	}

	// 1. Hash-based pseudo-random generator seeded with text content
	h := sha256.Sum256([]byte(clean))
	seed := binary.BigEndian.Uint64(h[:8])

	// 2. Token-level feature projection
	words := strings.Fields(clean)
	for i, word := range words {
		wordHash := sha256.Sum256([]byte(word))
		slot := int(binary.BigEndian.Uint32(wordHash[:4]) % uint32(m.dimensions))
		weight := float32(1.0 / math.Sqrt(float64(i+1)))
		vec[slot] += weight
	}

	// 3. Fill remaining coordinates with deterministic pseudo-spectral noise
	cur := seed
	for j := 0; j < m.dimensions; j++ {
		cur = cur*6364136223846793005 + 1442695040888963407
		noise := float32(float64(cur%10000)/10000.0 - 0.5)
		vec[j] += noise * 0.1
	}

	// 4. L2 Normalize vector so cosine distance ||u - v|| corresponds to unit sphere
	var sumSq float64
	for _, val := range vec {
		sumSq += float64(val * val)
	}
	if sumSq > 0 {
		norm := float32(math.Sqrt(sumSq))
		for j := range vec {
			vec[j] /= norm
		}
	} else {
		vec[0] = 1.0
	}

	return vec
}

// FormatVector converts a float32 slice into a PostgreSQL pgvector literal string "[v1,v2,v3...]".
func FormatVector(vec []float32) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, v := range vec {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatFloat(float64(v), 'f', 6, 32))
	}
	sb.WriteByte(']')
	return sb.String()
}

// ParseVector converts a PostgreSQL pgvector literal string "[v1,v2,v3...]" into a float32 slice.
func ParseVector(s string) ([]float32, error) {
	trimmed := strings.Trim(s, "[] \t\n\r")
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ",")
	vec := make([]float32, len(parts))
	for i, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return nil, fmt.Errorf("invalid vector component at index %d: %w", i, err)
		}
		vec[i] = float32(val)
	}
	return vec, nil
}
