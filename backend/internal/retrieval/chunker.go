package retrieval

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

// Chunker slices repository source files into semantically meaningful code chunks.
// Boundaries: Uses AST declaration boundaries (functions, methods, classes, structs)
// to prevent arbitrary splitting in the middle of statements.
type Chunker struct {
	maxLinesPerChunk int
	gapChunkSize     int
}

// NewChunker creates a new AST-guided semantic code chunker.
func NewChunker() *Chunker {
	return &Chunker{
		maxLinesPerChunk: 100,
		gapChunkSize:     50,
	}
}

// ChunkFile slices a file's content into semantic CodeChunks using extracted symbols.
func (c *Chunker) ChunkFile(file domain.RepositoryFile, symbols []domain.CodeSymbol) []domain.CodeChunk {
	if file.Content == nil || strings.TrimSpace(*file.Content) == "" {
		return nil
	}

	lines := strings.Split(*file.Content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	totalLines := len(lines)
	if totalLines == 0 {
		return nil
	}

	// Filter and sort symbols belonging to this file by StartLine ascending
	var fileSyms []domain.CodeSymbol
	for _, s := range symbols {
		if s.FileID == file.ID || s.FilePath == file.Path {
			fileSyms = append(fileSyms, s)
		}
	}
	sort.Slice(fileSyms, func(i, j int) bool {
		return fileSyms[i].StartLine < fileSyms[j].StartLine
	})

	var chunks []domain.CodeChunk

	// Case 1: File has extracted symbols -> AST declaration boundary chunking
	if len(fileSyms) > 0 {
		lastProcessedLine := 1

		for _, sym := range fileSyms {
			startLine := sym.StartLine
			endLine := sym.EndLine

			if startLine < 1 {
				startLine = 1
			}
			if endLine > totalLines {
				endLine = totalLines
			}
			if startLine > totalLines || startLine > endLine {
				continue
			}

			// Capture preamble/gap before this symbol (e.g. imports, comments, type definitions)
			if startLine > lastProcessedLine {
				gapChunks := c.chunkLineRange(file, lines, lastProcessedLine, startLine-1, "preamble", nil)
				chunks = append(chunks, gapChunks...)
			}

			// Symbol declaration chunk
			symContent := extractLines(lines, startLine, endLine)
			if strings.TrimSpace(symContent) != "" {
				symID := sym.ID
				scope := fmt.Sprintf("%s %s", sym.Kind, sym.Name)
				chunks = append(chunks, domain.CodeChunk{
					ID:         uuid.New(),
					SnapshotID: file.SnapshotID,
					FileID:     file.ID,
					FilePath:   file.Path,
					SymbolID:   &symID,
					SymbolName: sym.Name,
					StartLine:  startLine,
					EndLine:    endLine,
					Scope:      scope,
					Content:    symContent,
				})
			}

			if endLine >= lastProcessedLine {
				lastProcessedLine = endLine + 1
			}
		}

		// Capture any trailing lines after the last symbol
		if lastProcessedLine <= totalLines {
			trailingChunks := c.chunkLineRange(file, lines, lastProcessedLine, totalLines, "trailing", nil)
			chunks = append(chunks, trailingChunks...)
		}

		return chunks
	}

	// Case 2: File has no structural symbols (config, markdown, small scripts) -> Windowed chunking
	scope := filepath.Base(file.Path)
	return c.chunkLineRange(file, lines, 1, totalLines, scope, nil)
}

func (c *Chunker) chunkLineRange(
	file domain.RepositoryFile,
	lines []string,
	startLine, endLine int,
	scope string,
	symID *uuid.UUID,
) []domain.CodeChunk {
	if startLine > endLine || startLine > len(lines) {
		return nil
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}

	var chunks []domain.CodeChunk
	cur := startLine

	for cur <= endLine {
		nextEnd := cur + c.gapChunkSize - 1
		if nextEnd > endLine {
			nextEnd = endLine
		}

		content := extractLines(lines, cur, nextEnd)
		if strings.TrimSpace(content) != "" {
			chunks = append(chunks, domain.CodeChunk{
				ID:         uuid.New(),
				SnapshotID: file.SnapshotID,
				FileID:     file.ID,
				FilePath:   file.Path,
				SymbolID:   symID,
				StartLine:  cur,
				EndLine:    nextEnd,
				Scope:      scope,
				Content:    content,
			})
		}

		cur = nextEnd + 1
	}

	return chunks
}

func extractLines(lines []string, startLine, endLine int) string {
	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > endLine {
		return ""
	}

	// 1-indexed to 0-indexed slice
	slice := lines[startLine-1 : endLine]
	return strings.Join(slice, "\n")
}
