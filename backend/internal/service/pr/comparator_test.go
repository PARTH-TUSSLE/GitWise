package pr_test

import (
	"testing"

	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/service/pr"
)

func TestAnalyzeContractShift_BreakingSignature(t *testing.T) {
	files := []domain.PRFileDiff{
		{
			Path:      "pkg/auth/token.go",
			Subsystem: "Auth",
			Status:    "modified",
			DiffSnippet: domain.DiffSnippet{
				Target: "pkg/auth/token.go",
				Before: "func Validate(token string) bool",
				After:  "func Validate(token string, salt string) bool",
			},
		},
		{
			Path:      "pkg/auth/token_test.go",
			Subsystem: "Auth",
			Status:    "modified",
			DiffSnippet: domain.DiffSnippet{
				Target: "pkg/auth/token_test.go",
				Before: "func TestValidate(t *testing.T)",
				After:  "func TestValidate(t *testing.T)",
			},
		},
	}

	shift, findings := pr.AnalyzeContractShift(files)

	if shift.PublicAPIContract != "breaking" {
		t.Errorf("expected contract to be 'breaking', got %s", shift.PublicAPIContract)
	}
	if shift.StateMutationRisk != "high" {
		t.Errorf("expected high state mutation risk for breaking changes, got %s", shift.StateMutationRisk)
	}

	hasRegressionFinding := false
	for _, f := range findings {
		if f.Category == "regression" && f.Severity == "warning" {
			hasRegressionFinding = true
			break
		}
	}
	if !hasRegressionFinding {
		t.Error("expected regression warning finding for breaking signature mutation")
	}
}

func TestAnalyzeContractShift_ExtendedContract(t *testing.T) {
	files := []domain.PRFileDiff{
		{
			Path:      "pkg/utils/hash.go",
			Subsystem: "Utils",
			Status:    "modified",
			DiffSnippet: domain.DiffSnippet{
				Target: "pkg/utils/hash.go",
				Before: "func SHA256(data []byte) string",
				After:  "func SHA256(data []byte) string\nfunc SHA512(data []byte) string",
			},
		},
		{
			Path:      "pkg/utils/hash_test.go",
			Subsystem: "Utils",
			Status:    "modified",
			DiffSnippet: domain.DiffSnippet{
				Target: "pkg/utils/hash_test.go",
				After:  "func TestSHA512(t *testing.T)",
			},
		},
	}

	shift, _ := pr.AnalyzeContractShift(files)

	if shift.PublicAPIContract != "extended" {
		t.Errorf("expected contract to be 'extended', got %s", shift.PublicAPIContract)
	}
}

func TestAnalyzeContractShift_UnchangedContractWithMissingTest(t *testing.T) {
	files := []domain.PRFileDiff{
		{
			Path:      "internal/core/worker.go",
			Subsystem: "Core",
			Status:    "modified",
			DiffSnippet: domain.DiffSnippet{
				Target: "internal/core/worker.go",
				Before: "func runInternalLoop()",
				After:  "func runInternalLoop() {\n    // internal refactor\n}",
			},
		},
	}

	shift, findings := pr.AnalyzeContractShift(files)

	if shift.PublicAPIContract != "unchanged" {
		t.Errorf("expected contract to be 'unchanged', got %s", shift.PublicAPIContract)
	}

	hasCovFinding := false
	for _, f := range findings {
		if f.RuleID == "COV-01" {
			hasCovFinding = true
			break
		}
	}
	if !hasCovFinding {
		t.Error("expected coverage advisory finding when PR lacks test updates")
	}
}
