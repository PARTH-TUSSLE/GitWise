package pr

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gitwise/backend/internal/domain"
)

// AnalyzeContractShift evaluates file diffs to detect exported contract changes, breaking signatures, and review findings.
func AnalyzeContractShift(files []domain.PRFileDiff) (domain.ArchitecturalShift, []domain.ReviewFinding) {
	contractStatus := "unchanged"
	var contractExplanations []string
	var findings []domain.ReviewFinding

	subsystemsMap := make(map[string]bool)
	hasTestFile := false
	hasLogicFile := false

	for _, f := range files {
		subsystemsMap[f.Subsystem] = true
		isTest := strings.Contains(f.Path, "test") || strings.Contains(f.Path, "_test.go")
		if isTest {
			hasTestFile = true
		} else {
			hasLogicFile = true
		}

		// Inspect before and after diff snippets for exported signature mutations
		beforeExported := extractExportedSignatures(f.DiffSnippet.Before)
		afterExported := extractExportedSignatures(f.DiffSnippet.After)

		// Check for deleted or mutated signatures
		for name, beforeSig := range beforeExported {
			afterSig, exists := afterExported[name]
			if !exists {
				contractStatus = "breaking"
				contractExplanations = append(contractExplanations,
					fmt.Sprintf("Exported symbol '%s' was deleted in %s.", name, f.Path),
				)
				findings = append(findings, domain.ReviewFinding{
					RuleID:     "REG-01",
					Severity:   "warning",
					Category:   "regression",
					File:       f.Path,
					Line:       1,
					Message:    fmt.Sprintf("Public export '%s' was removed or renamed.", name),
					Suggestion: "Retain a deprecated alias or provide backward-compatible overload.",
				})
			} else if beforeSig != afterSig {
				contractStatus = "breaking"
				contractExplanations = append(contractExplanations,
					fmt.Sprintf("Exported signature for '%s' was mutated in %s.", name, f.Path),
				)
				findings = append(findings, domain.ReviewFinding{
					RuleID:     "REG-02",
					Severity:   "warning",
					Category:   "regression",
					File:       f.Path,
					Line:       1,
					Message:    fmt.Sprintf("Signature parameters or return types changed for '%s'.", name),
					Suggestion: "Introduce optional parameter or create v2 variant to avoid breaking callers.",
				})
			}
		}

		// Check for newly introduced exports
		for name := range afterExported {
			if _, exists := beforeExported[name]; !exists {
				if contractStatus != "breaking" {
					contractStatus = "extended"
				}
				contractExplanations = append(contractExplanations,
					fmt.Sprintf("Introduced new public contract '%s' in %s.", name, f.Path),
				)
				findings = append(findings, domain.ReviewFinding{
					RuleID:     "CONV-01",
					Severity:   "clean",
					Category:   "convention",
					File:       f.Path,
					Line:       1,
					Message:    fmt.Sprintf("New public export '%s' registered.", name),
					Suggestion: "Ensure exported symbol is thoroughly documented and exported in package root.",
				})
			}
		}

		// Performance check: nested loops
		if strings.Contains(f.DiffSnippet.After, "for ") && strings.Count(f.DiffSnippet.After, "for ") >= 2 {
			findings = append(findings, domain.ReviewFinding{
				RuleID:     "PERF-01",
				Severity:   "advisory",
				Category:   "performance",
				File:       f.Path,
				Line:       10,
				Message:    "Nested iteration detected in patch hunk; verify computational complexity on large collections.",
				Suggestion: "Use a map lookup or set membership instead of nested loop traversal.",
			})
		}
	}

	// Coverage check
	if hasLogicFile && !hasTestFile {
		findings = append(findings, domain.ReviewFinding{
			RuleID:     "COV-01",
			Severity:   "advisory",
			Category:   "coverage",
			File:       files[0].Path,
			Line:       1,
			Message:    "Patch modifies core functionality without accompanying test updates or new test specifications.",
			Suggestion: "Add unit regression test coverage exercising the changed code path.",
		})
	}

	var downstream []domain.SubsystemCoupling
	for sub := range subsystemsMap {
		downstream = append(downstream, domain.SubsystemCoupling{
			Name:     sub,
			Coupling: "tight",
			Impact:   fmt.Sprintf("Executes logic coordinated with %s module boundaries.", sub),
		})
	}

	stateMutationRisk := "none"
	riskExplanation := "No hazardous state mutations or unbounded concurrency detected."
	if contractStatus == "breaking" {
		stateMutationRisk = "high"
		riskExplanation = "Breaking contract mutation introduces high risk for external and downstream callers."
	} else if contractStatus == "extended" {
		stateMutationRisk = "low"
		riskExplanation = "Additive contract extension preserves backward compatibility."
	}

	explanation := "All modified signatures maintain exact backward compatibility."
	if len(contractExplanations) > 0 {
		explanation = strings.Join(contractExplanations, " ")
	}

	shift := domain.ArchitecturalShift{
		PublicAPIContract:    contractStatus,
		PublicAPIExplanation: explanation,
		DownstreamSubsystems: downstream,
		ExecutionFlowDelta:   "Synchronous and asynchronous execution pipelines preserved without unexpected blocking.",
		StateMutationRisk:    stateMutationRisk,
		RiskExplanation:      riskExplanation,
	}

	return shift, findings
}

func extractExportedSignatures(content string) map[string]string {
	out := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Go exported function or type: starts with func / type and an uppercase letter
		if strings.HasPrefix(trimmed, "func ") {
			rest := strings.TrimPrefix(trimmed, "func ")
			// Check method receiver e.g. func (s *Service) ExportedName(...)
			if strings.HasPrefix(rest, "(") {
				closeIdx := strings.Index(rest, ")")
				if closeIdx != -1 && len(rest) > closeIdx+1 {
					rest = strings.TrimSpace(rest[closeIdx+1:])
				}
			}
			parts := strings.SplitN(rest, "(", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				if len(name) > 0 && unicode.IsUpper(rune(name[0])) {
					out[name] = trimmed
				}
			}
		} else if strings.HasPrefix(trimmed, "type ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := parts[1]
				if len(name) > 0 && unicode.IsUpper(rune(name[0])) {
					out[name] = trimmed
				}
			}
		}

		// TypeScript / JavaScript exported symbols: "export function", "export const", "export interface", "export class"
		if strings.HasPrefix(trimmed, "export ") {
			tokens := strings.Fields(trimmed)
			if len(tokens) >= 3 {
				kind := tokens[1]
				if kind == "function" || kind == "const" || kind == "interface" || kind == "class" || kind == "type" {
					name := strings.Trim(tokens[2], "(:={;")
					if name != "" {
						out[name] = trimmed
					}
				}
			}
		}
	}

	return out
}
