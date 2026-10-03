package retrieval_test

import (
	"context"
	"math"
	"testing"

	"github.com/gitwise/backend/internal/retrieval"
)

func TestMockEmbedder_DeterminismAndDimension(t *testing.T) {
	embedder := retrieval.NewMockEmbedder()
	if embedder.ModelName() != "mock-768" {
		t.Errorf("expected model mock-768, got %s", embedder.ModelName())
	}
	if embedder.Dimensions() != 768 {
		t.Errorf("expected dimensions 768, got %d", embedder.Dimensions())
	}

	ctx := context.Background()
	vec1, err := embedder.Embed(ctx, "func HandleLogin(w http.ResponseWriter, r *http.Request)")
	if err != nil {
		t.Fatalf("unexpected embed error: %v", err)
	}
	if len(vec1) != 768 {
		t.Fatalf("expected 768 length, got %d", len(vec1))
	}

	// Determinism check: same input yields identical vector
	vec2, err := embedder.Embed(ctx, "func HandleLogin(w http.ResponseWriter, r *http.Request)")
	if err != nil {
		t.Fatalf("unexpected embed error: %v", err)
	}
	for i := range vec1 {
		if vec1[i] != vec2[i] {
			t.Fatalf("determinism failure at index %d: %f != %f", i, vec1[i], vec2[i])
		}
	}

	// Unit normalization check: norm ≈ 1.0
	var sumSq float64
	for _, v := range vec1 {
		sumSq += float64(v * v)
	}
	norm := math.Sqrt(sumSq)
	if math.Abs(norm-1.0) > 1e-4 {
		t.Errorf("expected unit norm ≈ 1.0, got %f", norm)
	}

	// Different input produces different vector
	vec3, err := embedder.Embed(ctx, "func Logout()")
	if err != nil {
		t.Fatalf("unexpected embed error: %v", err)
	}
	equal := true
	for i := range vec1 {
		if vec1[i] != vec3[i] {
			equal = false
			break
		}
	}
	if equal {
		t.Error("expected different vectors for different inputs")
	}
}

func TestVector_FormatAndParse(t *testing.T) {
	orig := []float32{0.123, -0.456, 0.789, 0.0}
	str := retrieval.FormatVector(orig)
	expectedStr := "[0.123000,-0.456000,0.789000,0.000000]"
	if str != expectedStr {
		t.Errorf("expected %s, got %s", expectedStr, str)
	}

	parsed, err := retrieval.ParseVector(str)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(parsed) != len(orig) {
		t.Fatalf("expected len %d, got %d", len(orig), len(parsed))
	}
	for i := range orig {
		if math.Abs(float64(orig[i]-parsed[i])) > 1e-5 {
			t.Errorf("mismatch at %d: %f != %f", i, orig[i], parsed[i])
		}
	}

	// Invalid format
	_, err = retrieval.ParseVector("invalid")
	if err == nil {
		t.Error("expected error for invalid vector string")
	}
}
