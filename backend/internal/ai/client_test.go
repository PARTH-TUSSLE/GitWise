package ai_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/ai"
)

func TestMockClient_GenerateAndStream(t *testing.T) {
	client := ai.NewMockClient()
	ctx := context.Background()

	prompt := "Explain the architecture based on evidence ev_01 and ev_02."
	resp, err := client.Generate(ctx, prompt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(resp, "[ev_01]") {
		t.Errorf("expected response to cite [ev_01], got: %s", resp)
	}
	if !strings.Contains(resp, "[ev_02]") {
		t.Errorf("expected response to cite [ev_02], got: %s", resp)
	}

	// Test streaming
	streamCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	stream, err := client.GenerateStream(streamCtx, prompt)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var collected strings.Builder
	for chunk := range stream {
		collected.WriteString(chunk)
	}

	if collected.String() != resp {
		t.Errorf("streamed text does not match generated text:\nStream: %s\nFull: %s", collected.String(), resp)
	}
}
