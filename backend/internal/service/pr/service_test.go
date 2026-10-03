package pr_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/ai"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/service/pr"
	"github.com/google/uuid"
)

func TestService_GetPullRequests(t *testing.T) {
	db, err := sql.Open("fake_pr_driver", "test_get_prs")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	prID := uuid.New()

	testPRDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "SELECT id FROM repositories") {
			return newRows([]string{"id"}, [][]driver.Value{{repoID.String()}}), nil, nil
		}
		if strings.Contains(query, "FROM pull_requests p") {
			return newRows(
				[]string{
					"id", "number", "title", "author", "status", "created_at",
					"stats", "summary", "architectural_shift", "files", "review_findings", "test_verification",
				},
				[][]driver.Value{
					{
						prID.String(), 42, "Refactor cache eviction", "bob", "ready_for_review", time.Now(),
						[]byte(`{"additions":10,"deletions":5,"filesChanged":1,"commitsCount":1}`),
						"Eviction summary",
						[]byte(`{"publicApiContract":"unchanged","publicApiExplanation":"safe","downstreamSubsystems":[],"executionFlowDelta":"","stateMutationRisk":"none","riskExplanation":""}`),
						[]byte(`[{"path":"cache.go","additions":10,"deletions":5,"status":"modified","subsystem":"Cache","behavioralSummary":"Refactor","diffSnippet":{"target":"cache.go","after":"code"}}]`),
						[]byte(`[]`),
						[]byte(`{"command":"go test ./...","targetSuite":"","expectedResult":"PASS","coverageDelta":"0%"}`),
					},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	svc := pr.NewService(db, nil, ai.NewMockClient(), nil)
	prs, err := svc.GetPullRequests(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prs) != 1 {
		t.Fatalf("expected 1 pr, got %d", len(prs))
	}
	if prs[0].Number != 42 {
		t.Errorf("expected PR number 42, got %d", prs[0].Number)
	}
	if prs[0].ArchitecturalShift.PublicAPIContract != "unchanged" {
		t.Errorf("expected unchanged contract, got %s", prs[0].ArchitecturalShift.PublicAPIContract)
	}
}

func TestService_ReviewPullRequest(t *testing.T) {
	db, err := sql.Open("fake_pr_driver", "test_review_pr")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	repoID := uuid.New()
	prID := uuid.New()

	testPRDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "SELECT id FROM repositories") {
			return newRows([]string{"id"}, [][]driver.Value{{repoID.String()}}), nil, nil
		}
		if strings.Contains(query, "SELECT id, title, author, status, created_at") {
			return newRows(
				[]string{"id", "title", "author", "status", "created_at"},
				[][]driver.Value{
					{prID.String(), "Add streaming backpressure", "alice", "ready_for_review", time.Now()},
				},
			), nil, nil
		}
		if strings.Contains(query, "INSERT INTO pr_reviews") {
			return nil, driver.RowsAffected(1), nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	rawPatch := `diff --git a/internal/stream/pipe.go b/internal/stream/pipe.go
index 1111111..2222222 100644
--- a/internal/stream/pipe.go
+++ b/internal/stream/pipe.go
@@ -1,5 +1,6 @@
 package stream
-func Drain() bool {
+func Drain(maxBytes int) bool {
+    // Added maxBytes parameter
     return true
 }
`

	svc := pr.NewService(db, nil, ai.NewMockClient(), nil)
	model, err := svc.ReviewPullRequest(context.Background(), "owner", "repo", 99, domain.ReviewPullRequestRequest{
		Patch: rawPatch,
	})
	if err != nil {
		t.Fatalf("unexpected review error: %v", err)
	}

	if model == nil {
		t.Fatal("expected non-nil pull request model")
	}
	if model.Number != 99 {
		t.Errorf("expected number 99, got %d", model.Number)
	}
	if model.ArchitecturalShift.PublicAPIContract != "breaking" {
		t.Errorf("expected breaking contract for mutated Drain signature, got %s", model.ArchitecturalShift.PublicAPIContract)
	}
	if len(model.Files) != 1 || model.Files[0].Path != "internal/stream/pipe.go" {
		t.Errorf("expected file internal/stream/pipe.go, got %+v", model.Files)
	}
	if len(model.ReviewFindings) == 0 {
		t.Error("expected review findings for breaking change and missing test")
	}
}
