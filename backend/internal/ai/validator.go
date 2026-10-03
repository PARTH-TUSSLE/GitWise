package ai

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gitwise/backend/internal/domain"
)

var (
	// Regex matching citations like [ev_01], [ev_02], [ev_foo_1]
	citationRegex = regexp.MustCompile(`\[(ev_[a-zA-Z0-9_-]+)\]`)
)

// Validator validates and hydrates citation references emitted by LLM reasoning.
// Invariant: Hallucinated evidence IDs not present in the supplied EvidencePackage
// are strictly rejected and stripped from the grounded response citations.
type Validator struct{}

// NewValidator creates a new citation validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ExtractionResult holds the validated text and hydrated citations.
type ExtractionResult struct {
	SanitizedText       string
	ValidatedCitations  []domain.ChatCitation
	RejectedEvidenceIDs []string
}

// ValidateAndHydrate parses evidence markers from text, verifies them against the package,
// and produces verified citations with exact line coordinates and code snippets.
func (v *Validator) ValidateAndHydrate(rawText string, pkg *domain.EvidencePackage) ExtractionResult {
	if pkg == nil || len(pkg.Items) == 0 {
		// No evidence in package: all [ev_XX] markers are hallucinated
		sanitized := citationRegex.ReplaceAllString(rawText, "")
		return ExtractionResult{
			SanitizedText: strings.TrimSpace(sanitized),
		}
	}

	// Index valid evidence items by their ID
	validMap := make(map[string]domain.EvidenceRef)
	for _, item := range pkg.Items {
		validMap[strings.ToLower(item.ID)] = item
	}

	matches := citationRegex.FindAllStringSubmatch(rawText, -1)
	seen := make(map[string]bool)
	var validated []domain.ChatCitation
	var rejected []string

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		rawID := match[1]
		normID := strings.ToLower(rawID)

		if seen[normID] {
			continue
		}
		seen[normID] = true

		ref, exists := validMap[normID]
		if !exists {
			// Hallucinated ID
			rejected = append(rejected, rawID)
			continue
		}

		lineVal := ref.StartLine
		snippetVal := ref.Snippet
		desc := fmt.Sprintf("%s (lines %d-%d)", ref.FilePath, ref.StartLine, ref.EndLine)

		validated = append(validated, domain.ChatCitation{
			EvidenceID:  ref.ID,
			File:        ref.FilePath,
			Line:        &lineVal,
			StartLine:   ref.StartLine,
			EndLine:     ref.EndLine,
			Snippet:     &snippetVal,
			Description: desc,
		})
	}

	// Clean out hallucinated citations from text while retaining valid citations
	sanitized := citationRegex.ReplaceAllStringFunc(rawText, func(match string) string {
		sub := citationRegex.FindStringSubmatch(match)
		if len(sub) >= 2 {
			if _, ok := validMap[strings.ToLower(sub[1])]; !ok {
				return "" // Strip hallucinated reference
			}
		}
		return match
	})

	return ExtractionResult{
		SanitizedText:       strings.TrimSpace(sanitized),
		ValidatedCitations:  validated,
		RejectedEvidenceIDs: rejected,
	}
}
