package typescript_test

import (
	"context"
	"os"
	"testing"

	"github.com/gitwise/backend/internal/analysis/typescript"
	"github.com/gitwise/backend/internal/domain"
)

func TestTypeScriptAnalyzer_CapabilityReport(t *testing.T) {
	analyzer := typescript.NewTypeScriptAnalyzer()
	cap := analyzer.Capability()

	if cap.Language != "TypeScript/JavaScript" {
		t.Errorf("expected language TypeScript/JavaScript, got %s", cap.Language)
	}
	if cap.IsDeterministic {
		t.Error("expected TypeScript analyzer to be explicitly heuristic (IsDeterministic=false)")
	}
	if len(cap.Limitations) == 0 {
		t.Error("expected non-empty documented limitations")
	}
}

func TestTypeScriptAnalyzer_GoldenFixture(t *testing.T) {
	fixtureBytes, err := os.ReadFile("../../../testdata/fixtures/sample-ts/service.ts")
	if err != nil {
		t.Fatalf("failed to read golden TypeScript fixture: %v", err)
	}

	analyzer := typescript.NewTypeScriptAnalyzer()
	res, err := analyzer.AnalyzeFile(context.Background(), "sample-ts/service.ts", string(fixtureBytes))
	if err != nil {
		t.Fatalf("unexpected error analyzing TypeScript fixture: %v", err)
	}

	// Verify imports
	expectedImports := map[string]bool{"express": true, "path": true, "fs": true}
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

	// Verify Interface
	if symMap["UserProfile"] != domain.SymbolKindInterface {
		t.Errorf("expected UserProfile to be INTERFACE, got %v", symMap["UserProfile"])
	}
	if !exportMap["UserProfile"] {
		t.Error("expected UserProfile to be exported")
	}

	// Verify Type Alias
	if symMap["AccountStatus"] != domain.SymbolKindType {
		t.Errorf("expected AccountStatus to be TYPE, got %v", symMap["AccountStatus"])
	}
	if !exportMap["AccountStatus"] {
		t.Error("expected AccountStatus to be exported")
	}

	// Verify Class
	if symMap["UserService"] != domain.SymbolKindClass {
		t.Errorf("expected UserService to be CLASS, got %v", symMap["UserService"])
	}
	if !exportMap["UserService"] {
		t.Error("expected UserService to be exported")
	}

	// Verify Class Methods
	if symMap["UserService.getUser"] != domain.SymbolKindMethod {
		t.Errorf("expected UserService.getUser to be METHOD, got %v", symMap["UserService.getUser"])
	}
	if symMap["UserService.clear"] != domain.SymbolKindMethod {
		t.Errorf("expected UserService.clear to be METHOD, got %v", symMap["UserService.clear"])
	}

	// Verify Functions
	if symMap["fetchRemoteUser"] != domain.SymbolKindFunction {
		t.Errorf("expected fetchRemoteUser to be FUNCTION, got %v", symMap["fetchRemoteUser"])
	}
	if !exportMap["fetchRemoteUser"] {
		t.Error("expected fetchRemoteUser to be exported")
	}

	// Verify Arrow Function
	if symMap["formatUserName"] != domain.SymbolKindFunction {
		t.Errorf("expected formatUserName to be FUNCTION, got %v", symMap["formatUserName"])
	}
	if !exportMap["formatUserName"] {
		t.Error("expected formatUserName to be exported")
	}

	// Verify local unexported function
	if symMap["internalLocalHelper"] != domain.SymbolKindFunction {
		t.Errorf("expected internalLocalHelper to be FUNCTION, got %v", symMap["internalLocalHelper"])
	}
	if exportMap["internalLocalHelper"] {
		t.Error("expected internalLocalHelper to be unexported")
	}

	// Verify default export function
	if symMap["defaultHandler"] != domain.SymbolKindFunction {
		t.Errorf("expected defaultHandler to be FUNCTION, got %v", symMap["defaultHandler"])
	}
	if !exportMap["defaultHandler"] {
		t.Error("expected defaultHandler to be exported")
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

func TestTypeScriptAnalyzer_EmptyAndWhitespaceFile(t *testing.T) {
	analyzer := typescript.NewTypeScriptAnalyzer()
	res, err := analyzer.AnalyzeFile(context.Background(), "empty.ts", "   \n\t\n")
	if err != nil {
		t.Fatalf("unexpected error on empty file: %v", err)
	}
	if len(res.Symbols) != 0 {
		t.Errorf("expected 0 symbols for empty file, got %d", len(res.Symbols))
	}
}

func TestTypeScriptAnalyzer_TSXComponent(t *testing.T) {
	tsxCode := `import React from 'react';

export interface ButtonProps {
  label: string;
  onClick: () => void;
}

export const Button: React.FC<ButtonProps> = ({ label, onClick }) => {
  return <button onClick={onClick}>{label}</button>;
};
`
	analyzer := typescript.NewTypeScriptAnalyzer()
	res, err := analyzer.AnalyzeFile(context.Background(), "Button.tsx", tsxCode)
	if err != nil {
		t.Fatalf("unexpected error parsing TSX: %v", err)
	}

	var hasInterface, hasButton bool
	for _, s := range res.Symbols {
		if s.Name == "ButtonProps" && s.Kind == domain.SymbolKindInterface {
			hasInterface = true
		}
		if s.Name == "Button" && s.Kind == domain.SymbolKindFunction {
			hasButton = true
		}
	}

	if !hasInterface {
		t.Error("expected ButtonProps interface extracted from TSX")
	}
	if !hasButton {
		t.Error("expected Button function component extracted from TSX")
	}
}

func TestTypeScriptAnalyzer_MalformedSyntaxGracefulDegradation(t *testing.T) {
	malformed := `export function ValidOne() {
  return 1;
}

// Unclosed or chaotic block
export class Chaotic {
  brokenMethod( {
`
	analyzer := typescript.NewTypeScriptAnalyzer()
	res, err := analyzer.AnalyzeFile(context.Background(), "malformed.ts", malformed)
	if err != nil {
		t.Fatalf("analyzer should not crash on malformed typescript: %v", err)
	}

	var foundValid bool
	for _, s := range res.Symbols {
		if s.Name == "ValidOne" {
			foundValid = true
		}
	}
	if !foundValid {
		t.Error("expected ValidOne extracted despite trailing syntax chaos")
	}
}
