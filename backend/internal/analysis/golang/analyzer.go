package golang

import (
	"bytes"
	"context"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"

	"github.com/gitwise/backend/internal/analysis"
	"github.com/gitwise/backend/internal/domain"
)

// GoAnalyzer provides deterministic Level 2 structural analysis for Go source code.
type GoAnalyzer struct{}

func NewGoAnalyzer() *GoAnalyzer {
	return &GoAnalyzer{}
}

func (a *GoAnalyzer) Language() string {
	return "Go"
}

func (a *GoAnalyzer) SupportsExtension(ext string) bool {
	return ext == ".go"
}

func (a *GoAnalyzer) Capability() analysis.CapabilityReport {
	return analysis.CapabilityReport{
		Language:        "Go",
		Level:           analysis.Level2Symbols,
		IsDeterministic: true,
		Features: []string{
			"Deterministic AST symbol extraction via go/parser and go/ast",
			"Top-level function and method declaration extraction with receiver types",
			"Struct, interface, and type alias declarations",
			"Accurate 1-indexed source line ranges",
			"Standard Go export visibility classification",
			"Explicit import path extraction",
		},
		Limitations: []string{
			"Intra-file syntactic declarations only (no cross-package type solver)",
			"No runtime reflection or dynamic invocation resolution",
		},
	}
}

func (a *GoAnalyzer) AnalyzeFile(ctx context.Context, filePath, content string) (*analysis.FileAnalysisResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(content) == "" {
		return &analysis.FileAnalysisResult{
			Symbols: nil,
			Imports: nil,
			Level:   analysis.Level2Symbols,
		}, nil
	}

	fset := token.NewFileSet()
	// Parse with AllErrors to attempt partial AST recovery on malformed source
	node, parseErr := parser.ParseFile(fset, filePath, content, parser.AllErrors|parser.ParseComments)

	result := &analysis.FileAnalysisResult{
		Symbols: make([]analysis.RawSymbol, 0),
		Imports: make([]string, 0),
		Level:   analysis.Level2Symbols,
	}

	if parseErr != nil {
		result.ParseErrors = append(result.ParseErrors, parseErr.Error())
	}

	if node == nil {
		return result, nil
	}

	// 1. Extract imports
	for _, imp := range node.Imports {
		if imp.Path != nil {
			importPath := strings.Trim(imp.Path.Value, `"`)
			result.Imports = append(result.Imports, importPath)
		}
	}

	// 2. Extract top-level declarations
	for _, decl := range node.Decls {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		switch d := decl.(type) {
		case *ast.FuncDecl:
			sym := a.extractFuncDecl(fset, d)
			result.Symbols = append(result.Symbols, sym)

		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				typeSyms := a.extractTypeDecl(fset, d)
				result.Symbols = append(result.Symbols, typeSyms...)
			}
		}
	}

	return result, nil
}

func (a *GoAnalyzer) extractFuncDecl(fset *token.FileSet, fn *ast.FuncDecl) analysis.RawSymbol {
	startLine := fset.Position(fn.Pos()).Line
	endLine := fset.Position(fn.End()).Line

	name := fn.Name.Name
	isExported := ast.IsExported(name)

	var kind domain.SymbolKind = domain.SymbolKindFunction
	var receiverStr string

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		kind = domain.SymbolKindMethod
		receiverField := fn.Recv.List[0]
		var buf bytes.Buffer
		_ = format.Node(&buf, fset, receiverField.Type)
		receiverStr = "(" + buf.String() + ") "
	}

	// Format signature (func Header without body)
	sigCopy := *fn
	sigCopy.Body = nil
	var sigBuf bytes.Buffer
	if err := format.Node(&sigBuf, fset, &sigCopy); err != nil {
		sigBuf.WriteString(receiverStr + name)
	}

	signature := strings.TrimSpace(sigBuf.String())

	return analysis.RawSymbol{
		Name:       name,
		Kind:       kind,
		StartLine:  startLine,
		EndLine:    endLine,
		Signature:  signature,
		IsExported: isExported,
	}
}

func (a *GoAnalyzer) extractTypeDecl(fset *token.FileSet, gen *ast.GenDecl) []analysis.RawSymbol {
	var symbols []analysis.RawSymbol

	for _, spec := range gen.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		name := typeSpec.Name.Name
		isExported := ast.IsExported(name)
		startLine := fset.Position(typeSpec.Pos()).Line
		endLine := fset.Position(typeSpec.End()).Line

		var kind domain.SymbolKind = domain.SymbolKindType

		switch typeSpec.Type.(type) {
		case *ast.StructType:
			kind = domain.SymbolKindStruct
		case *ast.InterfaceType:
			kind = domain.SymbolKindInterface
		}

		// Clean signature: "type Name struct" or "type Name interface" or "type Name ..."
		var typeBuf bytes.Buffer
		_ = format.Node(&typeBuf, fset, typeSpec)
		sigStr := "type " + strings.TrimSpace(typeBuf.String())
		// If multiline struct/interface, truncate signature to header
		if idx := strings.Index(sigStr, "{"); idx != -1 {
			sigStr = strings.TrimSpace(sigStr[:idx+1]) + " ... }"
		}

		symbols = append(symbols, analysis.RawSymbol{
			Name:       name,
			Kind:       kind,
			StartLine:  startLine,
			EndLine:    endLine,
			Signature:  sigStr,
			IsExported: isExported,
		})
	}

	return symbols
}
