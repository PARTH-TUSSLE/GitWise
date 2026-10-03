package ai_test

import (
	"strings"
	"testing"

	"github.com/gitwise/backend/internal/ai"
	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

func TestValidator_ValidGroundedCitations(t *testing.T) {
	validator := ai.NewValidator()

	snapID := uuid.New()
	fileID := uuid.New()
	pkg := &domain.EvidencePackage{
		SnapshotID: snapID,
		CommitSHA:  "abc1234",
		Query:      "authentication",
		Items: []domain.EvidenceRef{
			{
				ID:          "ev_01",
				SnapshotID:  snapID,
				FileID:      fileID,
				FilePath:    "internal/auth/token.go",
				StartLine:   15,
				EndLine:     30,
				Snippet:     "func ValidateToken() bool { return true }",
				ContentHash: "hash1",
			},
			{
				ID:          "ev_02",
				SnapshotID:  snapID,
				FileID:      fileID,
				FilePath:    "internal/auth/handler.go",
				StartLine:   40,
				EndLine:     55,
				Snippet:     "func LoginHandler(w http.ResponseWriter, r *http.Request)",
				ContentHash: "hash2",
			},
		},
		TotalItems: 2,
	}

	rawText := "Tokens are validated in [ev_01], and login requests enter through [ev_02]."
	res := validator.ValidateAndHydrate(rawText, pkg)

	if len(res.ValidatedCitations) != 2 {
		t.Fatalf("expected 2 validated citations, got %d", len(res.ValidatedCitations))
	}
	if len(res.RejectedEvidenceIDs) != 0 {
		t.Errorf("expected 0 rejected IDs, got %d", len(res.RejectedEvidenceIDs))
	}

	// Verify hydration of exact line coordinates
	c1 := res.ValidatedCitations[0]
	if c1.EvidenceID != "ev_01" || c1.File != "internal/auth/token.go" || c1.StartLine != 15 || c1.EndLine != 30 {
		t.Errorf("citation 1 mismatch: %+v", c1)
	}
	if c1.Snippet == nil || *c1.Snippet != "func ValidateToken() bool { return true }" {
		t.Errorf("citation 1 snippet mismatch: %v", c1.Snippet)
	}

	c2 := res.ValidatedCitations[1]
	if c2.EvidenceID != "ev_02" || c2.File != "internal/auth/handler.go" || c2.StartLine != 40 || c2.EndLine != 55 {
		t.Errorf("citation 2 mismatch: %+v", c2)
	}
}

func TestValidator_RejectsHallucinatedEvidenceIDs(t *testing.T) {
	validator := ai.NewValidator()

	snapID := uuid.New()
	fileID := uuid.New()
	pkg := &domain.EvidencePackage{
		SnapshotID: snapID,
		CommitSHA:  "abc1234",
		Items: []domain.EvidenceRef{
			{
				ID:        "ev_01",
				FileID:    fileID,
				FilePath:  "cmd/main.go",
				StartLine: 1,
				EndLine:   10,
				Snippet:   "package main",
			},
		},
		TotalItems: 1,
	}

	// Model hallucinates [ev_99] which does not exist in the package
	rawText := "The main func is in [ev_01], while the database pool is configured in [ev_99]."
	res := validator.ValidateAndHydrate(rawText, pkg)

	if len(res.ValidatedCitations) != 1 {
		t.Fatalf("expected 1 validated citation, got %d", len(res.ValidatedCitations))
	}
	if res.ValidatedCitations[0].EvidenceID != "ev_01" {
		t.Errorf("expected ev_01, got %s", res.ValidatedCitations[0].EvidenceID)
	}

	// Verify ev_99 was rejected and stripped
	if len(res.RejectedEvidenceIDs) != 1 || res.RejectedEvidenceIDs[0] != "ev_99" {
		t.Errorf("expected rejected ev_99, got %+v", res.RejectedEvidenceIDs)
	}
	if strings.Contains(res.SanitizedText, "[ev_99]") {
		t.Errorf("expected hallucinated citation [ev_99] to be stripped from text: %s", res.SanitizedText)
	}
	if !strings.Contains(res.SanitizedText, "[ev_01]") {
		t.Errorf("expected valid citation [ev_01] to remain in text: %s", res.SanitizedText)
	}
}

func TestValidator_EmptyEvidencePackage(t *testing.T) {
	validator := ai.NewValidator()

	rawText := "Check [ev_01] for details."
	res := validator.ValidateAndHydrate(rawText, nil)

	if len(res.ValidatedCitations) != 0 {
		t.Errorf("expected 0 citations for nil package, got %d", len(res.ValidatedCitations))
	}
	if strings.Contains(res.SanitizedText, "[ev_01]") {
		t.Errorf("expected marker to be stripped when package is nil: %s", res.SanitizedText)
	}
}
