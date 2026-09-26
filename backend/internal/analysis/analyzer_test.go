package analysis_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gitwise/backend/internal/analysis"
	"github.com/gitwise/backend/internal/analysis/golang"
	"github.com/gitwise/backend/internal/analysis/typescript"
)

func TestRegistry_CancellationPropagated(t *testing.T) {
	reg := analysis.NewRegistry()
	reg.Register(golang.NewGoAnalyzer())
	reg.Register(typescript.NewTypeScriptAnalyzer())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Testing Go cancellation
	res, err := reg.Analyze(ctx, "main.go", "package main\nfunc main() {}")
	if err == nil {
		t.Fatal("expected cancellation error for Go file, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	if res != nil {
		t.Errorf("expected nil result on cancellation, got %v", res)
	}

	// Testing TS cancellation
	resTS, errTS := reg.Analyze(ctx, "app.ts", "class Foo { run() {} }")
	if errTS == nil {
		t.Fatal("expected cancellation error for TS file, got nil")
	}
	if !errors.Is(errTS, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", errTS)
	}
	if resTS != nil {
		t.Errorf("expected nil result on cancellation, got %v", resTS)
	}
}

func TestRegistry_SyntaxErrorDegradesGracefully(t *testing.T) {
	reg := analysis.NewRegistry()
	reg.Register(golang.NewGoAnalyzer())
	reg.Register(typescript.NewTypeScriptAnalyzer())

	ctx := context.Background()

	// Malformed Go syntax should degrade gracefully to warnings, NOT fail with an error
	malformedGo := "package broken\nfunc broken( {"
	res, err := reg.Analyze(ctx, "broken.go", malformedGo)
	if err != nil {
		t.Fatalf("expected nil error on syntax degradation, got: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil FileAnalysisResult on degradation")
	}
	if len(res.ParseErrors) == 0 {
		t.Error("expected parse warning captured in ParseErrors")
	}
}
