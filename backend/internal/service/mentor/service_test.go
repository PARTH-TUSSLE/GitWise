package mentor_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/ai"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/evidence"
	"github.com/gitwise/backend/internal/retrieval"
	"github.com/gitwise/backend/internal/service/mentor"
	"github.com/google/uuid"
)

func TestService_Chat_GroundedWithCitations(t *testing.T) {
	db, err := sql.Open("fake_mentor_driver", "test_chat")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	snapID := uuid.New()
	sessID := uuid.New()
	fileID := uuid.New()

	testMentorDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "INSERT INTO mentor_sessions") {
			return newRows(
				[]string{"id", "snapshot_id", "title", "created_at", "updated_at"},
				[][]driver.Value{
					{sessID.String(), snapID.String(), "Mentorship Session", time.Now(), time.Now()},
				},
			), nil, nil
		}
		if strings.Contains(query, "INSERT INTO chat_messages") {
			return nil, driver.RowsAffected(1), nil
		}
		// Code chunks retrieval
		if strings.Contains(query, "code_symbols cs") || strings.Contains(query, "LOWER(cs.name)") {
			return newRows(
				[]string{
					"id", "snapshot_id", "file_id", "path", "symbol_id", "name",
					"start_line", "end_line", "scope", "content", "embedding_model", "created_at",
				},
				[][]driver.Value{
					{
						uuid.New().String(), snapID.String(), fileID.String(), "auth/token.go",
						uuid.New().String(), "ValidateToken", int64(10), int64(25), "FUNCTION ValidateToken",
						"func ValidateToken() bool { return true }", "mock-768", time.Now(),
					},
				},
			), nil, nil
		}
		// Evidence hydration
		if strings.Contains(query, "SELECT path, content FROM repository_files") {
			return newRows(
				[]string{"path", "content"},
				[][]driver.Value{
					{"auth/token.go", "package auth\n\nfunc ValidateToken() bool { return true }\n"},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), nil, nil
	}

	embedder := retrieval.NewMockEmbedder()
	retrievalSvc := retrieval.NewService(db, embedder, nil)
	evidenceStore := evidence.NewStore(db, nil)
	aiClient := ai.NewMockClient()

	svc := mentor.NewService(db, retrievalSvc, evidenceStore, aiClient, nil)
	ctx := context.Background()

	req := domain.ChatRequest{
		Message: "How does token validation work?",
		TopK:    3,
	}

	resp, err := svc.Chat(ctx, snapID, "commit123", req)
	if err != nil {
		t.Fatalf("unexpected chat error: %v", err)
	}

	if resp.SessionID != sessID {
		t.Errorf("expected session ID %s, got %s", sessID, resp.SessionID)
	}
	if len(resp.Message) == 0 {
		t.Error("expected non-empty response message")
	}
	if len(resp.Citations) == 0 {
		t.Fatal("expected at least 1 validated citation, got 0")
	}

	cit := resp.Citations[0]
	if cit.EvidenceID != "ev_01" {
		t.Errorf("expected citation evidenceId ev_01, got %s", cit.EvidenceID)
	}
	if cit.File != "auth/token.go" {
		t.Errorf("expected file auth/token.go, got %s", cit.File)
	}
	if cit.StartLine != 10 || cit.EndLine != 25 {
		t.Errorf("expected lines 10-25, got %d-%d", cit.StartLine, cit.EndLine)
	}
}

func TestService_ChatStream(t *testing.T) {
	db, err := sql.Open("fake_mentor_driver", "test_chat_stream")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	snapID := uuid.New()
	sessID := uuid.New()

	testMentorDriver.handler = func(query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
		if strings.Contains(query, "INSERT INTO mentor_sessions") {
			return newRows(
				[]string{"id", "snapshot_id", "title", "created_at", "updated_at"},
				[][]driver.Value{
					{sessID.String(), snapID.String(), "Mentorship Session", time.Now(), time.Now()},
				},
			), nil, nil
		}
		return newRows([]string{}, nil), driver.RowsAffected(1), nil
	}

	svc := mentor.NewService(db, nil, nil, ai.NewMockClient(), nil)
	ctx := context.Background()

	stream, err := svc.ChatStream(ctx, snapID, "commit123", domain.ChatRequest{
		Message: "Explain the architecture",
	})
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}

	var chunks []domain.ChatStreamChunk
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected multiple stream chunks, got %d", len(chunks))
	}

	lastChunk := chunks[len(chunks)-1]
	if !lastChunk.Done {
		t.Error("expected final chunk to have Done: true")
	}
}
