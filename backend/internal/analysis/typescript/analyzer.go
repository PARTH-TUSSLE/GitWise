package typescript

import (
	"bufio"
	"context"
	"regexp"
	"strings"

	"github.com/gitwise/backend/internal/analysis"
	"github.com/gitwise/backend/internal/domain"
)

// TypeScriptAnalyzer provides lightweight, pure-Go structural analysis for TypeScript and JavaScript.
// Boundaries: This is a structural regex/block-based extractor, NOT a compiler-grade type evaluator.
// Does NOT perform cross-file symbol resolution, runtime routing, or dynamic import evaluation.
type TypeScriptAnalyzer struct{}

func NewTypeScriptAnalyzer() *TypeScriptAnalyzer {
	return &TypeScriptAnalyzer{}
}

func (a *TypeScriptAnalyzer) Language() string {
	return "TypeScript"
}

func (a *TypeScriptAnalyzer) SupportsExtension(ext string) bool {
	switch ext {
	case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func (a *TypeScriptAnalyzer) Capability() analysis.CapabilityReport {
	return analysis.CapabilityReport{
		Language:        "TypeScript/JavaScript",
		Level:           analysis.Level2Symbols,
		IsDeterministic: false, // Explicitly documented as heuristic structural parsing
		Features: []string{
			"Pure-Go CGO-free structural parsing of TS, TSX, JS, JSX",
			"ESM (import ... from) and CommonJS (require(...)) import extraction",
			"Exported and local function declaration extraction",
			"Arrow function variable assignment extraction",
			"Class declaration and method extraction",
			"TypeScript interface and type alias extraction",
			"1-indexed line range tracking via balanced brace scanning",
		},
		Limitations: []string{
			"No compiler-grade type evaluation or cross-file type resolution",
			"No runtime routing, dynamic import resolution, or callback tracking",
			"Heuristic fallback on heavily obfuscated or non-standard syntax transforms",
		},
	}
}

// Regex patterns for top-level structural declarations
var (
	// ESM import: import { a, b } from './module'; or import foo from 'bar';
	esmImportRegex = regexp.MustCompile(`(?m)^\s*import\s+(?:(?:[\w*\s{},$]+)\s+from\s+)?['"]([^'"]+)['"]`)
	// CommonJS require: const foo = require('bar');
	cjsRequireRegex = regexp.MustCompile(`(?:require\s*\(\s*['"]([^'"]+)['"]\s*\))`)

	// function declaration: export (default)? (async)? function name(...)
	funcDeclRegex = regexp.MustCompile(`(?m)^(?:export\s+(?:default\s+)?)?(?:async\s+)?function\s*(\*?\s*[a-zA-Z0-9_$]+)\s*\(`)

	// arrow function / const function: export const name (: Type)? = (async)? (...) =>
	arrowFuncRegex = regexp.MustCompile(`(?m)^(?:export\s+)?(?:const|let|var)\s+([a-zA-Z0-9_$]+)\s*(?::\s*[^=]+)?\s*=\s*(?:async\s*)?(?:\((?:[^)]|\n)*\)|[a-zA-Z0-9_$]+)\s*(?::\s*[^=]+)?\s*=>`)

	// class declaration: export (default)? (abstract)? class Name
	classDeclRegex = regexp.MustCompile(`(?m)^(?:export\s+(?:default\s+)?)?(?:abstract\s+)?class\s+([a-zA-Z0-9_$]+)`)

	// interface declaration: export interface Name
	interfaceDeclRegex = regexp.MustCompile(`(?m)^(?:export\s+)?interface\s+([a-zA-Z0-9_$]+)`)

	// type alias: export type Name = ...
	typeDeclRegex = regexp.MustCompile(`(?m)^(?:export\s+)?type\s+([a-zA-Z0-9_$]+)\s*(?:<[^>]+>)?\s*=`)

	// class method inside class body: (public|private|async|...)* name(...) {
	methodDeclRegex = regexp.MustCompile(`^\s*(?:(?:public|private|protected|static|override|readonly|async|get|set)\s+)*([a-zA-Z0-9_$]+)\s*(?:<[^>]+>)?\s*\((?:[^)]|\n)*\)\s*(?::\s*[^{]+)?\s*\{?`)
)

func (a *TypeScriptAnalyzer) AnalyzeFile(ctx context.Context, filePath, content string) (*analysis.FileAnalysisResult, error) {
	if strings.TrimSpace(content) == "" {
		return &analysis.FileAnalysisResult{
			Symbols: nil,
			Imports: nil,
			Level:   analysis.Level2Symbols,
		}, nil
	}

	result := &analysis.FileAnalysisResult{
		Symbols: make([]analysis.RawSymbol, 0),
		Imports: make([]string, 0),
		Level:   analysis.Level2Symbols,
	}

	// 1. Extract imports
	seenImports := make(map[string]bool)
	for _, match := range esmImportRegex.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 && !seenImports[match[1]] {
			seenImports[match[1]] = true
			result.Imports = append(result.Imports, match[1])
		}
	}
	for _, match := range cjsRequireRegex.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 && !seenImports[match[1]] {
			seenImports[match[1]] = true
			result.Imports = append(result.Imports, match[1])
		}
	}

	// 2. Line-by-line scanning with brace depth tracking
	lines := splitLines(content)
	totalLines := len(lines)

	var currentClass *analysis.RawSymbol
	classBraceDepth := 0
	currentBraceDepth := 0

	for i := 0; i < totalLines; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lineNum := i + 1

		// Skip comments
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Update brace count for line
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")

		// If currently inside a class body, look for methods
		if currentClass != nil {
			if methodMatch := methodDeclRegex.FindStringSubmatch(line); len(methodMatch) > 1 {
				mName := methodMatch[1]
				// Avoid keywords like if, for, while, switch, constructor (or include constructor)
				if !isControlFlowKeyword(mName) {
					endMethodLine := findBlockEnd(lines, i)
					isExported := currentClass.IsExported // inherits class visibility
					sig := cleanSignature(trimmed)
					result.Symbols = append(result.Symbols, analysis.RawSymbol{
						Name:       currentClass.Name + "." + mName,
						Kind:       domain.SymbolKindMethod,
						StartLine:  lineNum,
						EndLine:    endMethodLine,
						Signature:  sig,
						IsExported: isExported,
					})
				}
			}

			currentBraceDepth += (openBraces - closeBraces)
			if currentBraceDepth <= classBraceDepth {
				// Class ended
				currentClass.EndLine = lineNum
				currentClass = nil
			}
			continue
		}

		// Check Interface
		if match := interfaceDeclRegex.FindStringSubmatch(line); len(match) > 1 {
			name := match[1]
			isExported := strings.HasPrefix(trimmed, "export")
			endLine := findBlockEnd(lines, i)
			result.Symbols = append(result.Symbols, analysis.RawSymbol{
				Name:       name,
				Kind:       domain.SymbolKindInterface,
				StartLine:  lineNum,
				EndLine:    endLine,
				Signature:  cleanSignature(trimmed),
				IsExported: isExported,
			})
			continue
		}

		// Check Type Alias
		if match := typeDeclRegex.FindStringSubmatch(line); len(match) > 1 {
			name := match[1]
			isExported := strings.HasPrefix(trimmed, "export")
			endLine := findStatementEnd(lines, i)
			result.Symbols = append(result.Symbols, analysis.RawSymbol{
				Name:       name,
				Kind:       domain.SymbolKindType,
				StartLine:  lineNum,
				EndLine:    endLine,
				Signature:  cleanSignature(trimmed),
				IsExported: isExported,
			})
			continue
		}

		// Check Class
		if match := classDeclRegex.FindStringSubmatch(line); len(match) > 1 {
			name := match[1]
			isExported := strings.HasPrefix(trimmed, "export")
			endLine := findBlockEnd(lines, i)
			classSym := analysis.RawSymbol{
				Name:       name,
				Kind:       domain.SymbolKindClass,
				StartLine:  lineNum,
				EndLine:    endLine,
				Signature:  cleanSignature(trimmed),
				IsExported: isExported,
			}
			result.Symbols = append(result.Symbols, classSym)

			if strings.Contains(line, "{") {
				currentClass = &result.Symbols[len(result.Symbols)-1]
				classBraceDepth = currentBraceDepth
				currentBraceDepth += (openBraces - closeBraces)
			}
			continue
		}

		// Check Function Declaration
		if match := funcDeclRegex.FindStringSubmatch(line); len(match) > 1 {
			name := strings.TrimSpace(match[1])
			name = strings.TrimPrefix(name, "*") // remove generator asterisk
			isExported := strings.HasPrefix(trimmed, "export")
			endLine := findBlockEnd(lines, i)
			result.Symbols = append(result.Symbols, analysis.RawSymbol{
				Name:       name,
				Kind:       domain.SymbolKindFunction,
				StartLine:  lineNum,
				EndLine:    endLine,
				Signature:  cleanSignature(trimmed),
				IsExported: isExported,
			})
			continue
		}

		// Check Arrow Function / Function Expression
		if match := arrowFuncRegex.FindStringSubmatch(line); len(match) > 1 {
			name := match[1]
			isExported := strings.HasPrefix(trimmed, "export")
			endLine := findBlockEnd(lines, i)
			result.Symbols = append(result.Symbols, analysis.RawSymbol{
				Name:       name,
				Kind:       domain.SymbolKindFunction,
				StartLine:  lineNum,
				EndLine:    endLine,
				Signature:  cleanSignature(trimmed),
				IsExported: isExported,
			})
			continue
		}

		currentBraceDepth += (openBraces - closeBraces)
	}

	return result, nil
}

func splitLines(content string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func findBlockEnd(lines []string, startIdx int) int {
	braceDepth := 0
	foundOpen := false

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		for _, ch := range line {
			if ch == '{' {
				braceDepth++
				foundOpen = true
			} else if ch == '}' {
				braceDepth--
				if foundOpen && braceDepth <= 0 {
					return i + 1
				}
			}
		}
		if foundOpen && braceDepth <= 0 {
			return i + 1
		}
	}
	return startIdx + 1
}

func findStatementEnd(lines []string, startIdx int) int {
	for i := startIdx; i < len(lines); i++ {
		if strings.Contains(lines[i], ";") {
			return i + 1
		}
		// If another declaration starts or blank line follows
		if i > startIdx && strings.TrimSpace(lines[i]) == "" {
			return i
		}
	}
	return startIdx + 1
}

func cleanSignature(line string) string {
	sig := strings.TrimSpace(line)
	if idx := strings.Index(sig, "{"); idx != -1 {
		sig = strings.TrimSpace(sig[:idx])
	}
	if strings.HasSuffix(sig, "=>") {
		sig = strings.TrimSpace(strings.TrimSuffix(sig, "=>"))
	}
	return sig
}

func isControlFlowKeyword(word string) bool {
	switch word {
	case "if", "for", "while", "do", "switch", "case", "catch", "return", "throw":
		return true
	default:
		return false
	}
}
