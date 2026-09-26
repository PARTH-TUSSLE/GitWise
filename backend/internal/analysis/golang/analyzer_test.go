package golang_test

import (
	"context"
	"os"
	"testing"

	"github.com/gitwise/backend/internal/analysis/golang"
	"github.com/gitwise/backend/internal/domain"
)

func TestGoAnalyzer_CapabilityReport(t *testing.T) {
	analyzer := golang.NewGoAnalyzer()
	cap := analyzer.Capability()

	if cap.Language != "Go" {
		t.Errorf("expected language Go, got %s", cap.Language)
	}
	if !cap.IsDeterministic {
		t.Error("expected Go analyzer to be deterministic")
	}
	if len(cap.Features) == 0 {
		t.Error("expected non-empty capability features")
	}
}

func TestGoAnalyzer_GoldenFixture(t *testing.T) {
	fixtureBytes, err := os.ReadFile("../../../testdata/fixtures/sample-go/main.go")
	if err != nil {
		t.Fatalf("failed to read golden Go fixture: %v", err)
	}

	analyzer := golang.NewGoAnalyzer()
	res, err := analyzer.AnalyzeFile(context.Background(), "sample-go/main.go", string(fixtureBytes))
	if err != nil {
		t.Fatalf("unexpected error analyzing Go fixture: %v", err)
	}

	if len(res.ParseErrors) > 0 {
		t.Errorf("expected 0 parse errors on valid fixture, got: %v", res.ParseErrors)
	}

	// Verify imports
	expectedImports := map[string]bool{"context": true, "fmt": true, "sync": true}
	for _, imp := range res.Imports {
		delete(expectedImports, imp)
	}
	if len(expectedImports) > 0 {
		t.Errorf("missing expected imports: %v", expectedImports)
	}

	// Index symbols by name
	symMap := make(map[string]domain.SymbolKind)
	exportMap := make(map[string]bool)
	lineMap := make(map[string][2]int)

	for _, s := range res.Symbols {
		symMap[s.Name] = s.Kind
		exportMap[s.Name] = s.IsExported
		lineMap[s.Name] = [2]int{s.StartLine, s.EndLine}
	}

	// Verify Structs
	if symMap["Config"] != domain.SymbolKindStruct {
		t.Errorf("expected Config to be STRUCT, got %v", symMap["Config"])
	}
	if !exportMap["Config"] {
		t.Error("expected Config to be exported")
	}

	if symMap["Service"] != domain.SymbolKindStruct {
		t.Errorf("expected Service to be STRUCT, got %v", symMap["Service"])
	}
	if !exportMap["Service"] {
		t.Error("expected Service to be exported")
	}

	// Verify Interface
	if symMap["Greeter"] != domain.SymbolKindInterface {
		t.Errorf("expected Greeter to be INTERFACE, got %v", symMap["Greeter"])
	}
	if !exportMap["Greeter"] {
		t.Error("expected Greeter to be exported")
	}

	// Verify Type Alias
	if symMap["UserID"] != domain.SymbolKindType {
		t.Errorf("expected UserID to be TYPE, got %v", symMap["UserID"])
	}

	// Verify Functions & Methods
	if symMap["NewService"] != domain.SymbolKindFunction {
		t.Errorf("expected NewService to be FUNCTION, got %v", symMap["NewService"])
	}
	if !exportMap["NewService"] {
		t.Error("expected NewService to be exported")
	}

	if symMap["Greet"] != domain.SymbolKindMethod {
		t.Errorf("expected Greet to be METHOD, got %v", symMap["Greet"])
	}
	if !exportMap["Greet"] {
		t.Error("expected Greet to be exported")
	}

	// Verify unexported helper
	if symMap["unexportedHelper"] != domain.SymbolKindFunction {
		t.Errorf("expected unexportedHelper to be FUNCTION, got %v", symMap["unexportedHelper"])
	}
	if exportMap["unexportedHelper"] {
		t.Error("expected unexportedHelper to be unexported")
	}

	// Verify line ranges
	for name, lines := range lineMap {
		if lines[0] <= 0 || lines[1] <= 0 {
			t.Errorf("symbol %s has invalid 0 or negative line range: %v", name, lines)
		}
		if lines[0] > lines[1] {
			t.Errorf("symbol %s has startLine %d > endLine %d", name, lines[0], lines[1])
		}
	}
}

func TestGoAnalyzer_EmptyAndWhitespaceFile(t *testing.T) {
	analyzer := golang.NewGoAnalyzer()
	res, err := analyzer.AnalyzeFile(context.Background(), "empty.go", "   \n\t\n")
	if err != nil {
		t.Fatalf("unexpected error on empty file: %v", err)
	}
	if len(res.Symbols) != 0 {
		t.Errorf("expected 0 symbols for empty file, got %d", len(res.Symbols))
	}
}

func TestGoAnalyzer_MalformedSyntaxGracefulDegradation(t *testing.T) {
	malformedCode := `package broken

func GoodFunction() string {
	return "ok"
}

func BrokenFunction( { // syntax error
`
	analyzer := golang.NewGoAnalyzer()
	res, err := analyzer.AnalyzeFile(context.Background(), "broken.go", malformedCode)
	if err != nil {
		t.Fatalf("analyzer should not return fatal error on malformed file: %v", err)
	}

	if len(res.ParseErrors) == 0 {
		t.Error("expected parse error captured for malformed file")
	}

	// Should still have extracted the preceding valid function
	var foundGood bool
	for _, s := range res.Symbols {
		if s.Name == "GoodFunction" {
			foundGood = true
		}
	}
	if !foundGood {
		t.Error("expected partial recovery of GoodFunction from malformed file")
	}
}

func TestGoAnalyzer_ContextCancellation(t *testing.T) {
	analyzer := golang.NewGoAnalyzer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := analyzer.AnalyzeFile(ctx, "test.go", "package main\nfunc main() {}\n")
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}
