package issue_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/ai"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/service/issue"
	"github.com/google/uuid"
)

func TestService_GetIssues_FromDatabase(t *testing.T) {
	db, err := sql.Open("fake_issue_driver", "test_get_issues")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	issueID := uuid.New()

	testIssueDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "SELECT id, name FROM repositories") {
			return newRows(
				[]string{"id", "name"},
				[][]driver.Value{
					{repoID.String(), "owner/repo"},
				},
			), nil, nil
		}
		if strings.Contains(query, "FROM issues i") {
			return newRows(
				[]string{
					"id", "number", "title", "body", "author", "state", "subsystem", "reported_date",
					"blast_radius", "summary", "prerequisites", "affected_files", "stages", "test_strategy",
				},
				[][]driver.Value{
					{
						issueID.String(), 101, "Memory leak in token cache", "Tokens not freed", "alice", "open", "Auth", time.Now(),
						[]byte(`{"score":2.5,"fileCount":2,"subsystemsAffected":["Auth"],"riskAssessment":"Low"}`),
						"Summary",
						[]byte(`["Prereq 1"]`),
						[]byte(`[{"path":"auth/token.go","linesChanged":20,"role":"core_logic","description":"fix"}]`),
						[]byte(`{"stage01_triage":{"rootCauseAnalysis":"leak","reproductionSteps":["run"],"scopeBoundary":"auth"},"stage02_impacted_paths":{"callChain":[],"criticalSymbols":[],"stateMutations":""},"stage03_blueprint":{"steps":[]}}`),
						[]byte(`{"suites":[{"name":"unit","type":"unit","testFile":"auth/token_test.go","command":"go test","expectedOutput":"PASS"}],"edgeCases":[],"verificationChecklist":[]}`),
					},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	svc := issue.NewService(db, nil, nil, nil, ai.NewMockClient(), nil)
	issues, err := svc.GetIssues(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	item := issues[0]
	if item.Number != 101 {
		t.Errorf("expected issue number 101, got %d", item.Number)
	}
	if item.Subsystem != "Auth" {
		t.Errorf("expected subsystem Auth, got %s", item.Subsystem)
	}
	if len(item.AffectedFiles) != 1 || item.AffectedFiles[0].Path != "auth/token.go" {
		t.Errorf("expected affected file auth/token.go, got %+v", item.AffectedFiles)
	}
}

func TestService_GenerateBlueprint_GroundedSynthesis(t *testing.T) {
	db, err := sql.Open("fake_issue_driver", "test_generate_blueprint")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	issueID := uuid.New()
	snapID := uuid.New()

	testIssueDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "SELECT id, name FROM repositories") {
			return newRows(
				[]string{"id", "name"},
				[][]driver.Value{
					{repoID.String(), "owner/repo"},
				},
			), nil, nil
		}
		if strings.Contains(query, "SELECT id, title, body, author, state, subsystem, reported_date") {
			return newRows(
				[]string{"id", "title", "body", "author", "state", "subsystem", "reported_date"},
				[][]driver.Value{
					{issueID.String(), "Deadlock in worker pool", "Lock held across channel send", "bob", "open", "Worker", time.Now()},
				},
			), nil, nil
		}
		if strings.Contains(query, "SELECT id, commit_sha FROM repository_snapshots") {
			return newRows(
				[]string{"id", "commit_sha"},
				[][]driver.Value{
					{snapID.String(), "sha_commit_123"},
				},
			), nil, nil
		}
		if strings.Contains(query, "INSERT INTO issue_blueprints") {
			return nil, driver.RowsAffected(1), nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	svc := issue.NewService(db, nil, nil, nil, ai.NewMockClient(), nil)
	model, err := svc.GenerateBlueprint(context.Background(), "owner", "repo", 202, domain.CreateBlueprintRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("expected non-nil issue model")
	}
	if model.Number != 202 {
		t.Errorf("expected number 202, got %d", model.Number)
	}
	if model.BlastRadius.Score <= 0 {
		t.Errorf("expected positive blast radius score, got %f", model.BlastRadius.Score)
	}
	if len(model.Stages.Stage03Blueprint.Steps) == 0 {
		t.Error("expected non-empty implementation blueprint steps")
	}
	if len(model.TestStrategy.Suites) == 0 {
		t.Error("expected non-empty test strategy suites")
	}
}
