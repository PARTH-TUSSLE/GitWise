package mentor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gitwise/backend/internal/ai"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/evidence"
	"github.com/gitwise/backend/internal/retrieval"
	"github.com/google/uuid"
)

// Service provides grounded AI mentorship and repository chat with strict citation verification.
type Service struct {
	db              *sql.DB
	retrievalSvc    *retrieval.Service
	evidenceStore   *evidence.Store
	aiClient        ai.Client
	defaultAIConfig ai.Config
	validator       *ai.Validator
	logger          *slog.Logger
}

// NewService creates a new mentor service instance.
func NewService(
	db *sql.DB,
	retrievalSvc *retrieval.Service,
	evidenceStore *evidence.Store,
	aiClient ai.Client,
	logger *slog.Logger,
	defaultAIConfig ...ai.Config,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if aiClient == nil {
		aiClient = ai.NewMockClient()
	}
	var defCfg ai.Config
	if len(defaultAIConfig) > 0 {
		defCfg = defaultAIConfig[0]
	}
	return &Service{
		db:              db,
		retrievalSvc:    retrievalSvc,
		evidenceStore:   evidenceStore,
		aiClient:        aiClient,
		defaultAIConfig: defCfg,
		validator:       ai.NewValidator(),
		logger:          logger,
	}
}

// SetAIClient updates the active AI client at runtime.
func (s *Service) SetAIClient(client ai.Client) {
	if client != nil {
		s.aiClient = client
	}
}

// AIClient returns the active AI client.
func (s *Service) AIClient() ai.Client {
	return s.aiClient
}

// resolveClient handles dynamic per-request model or provider overrides.
func (s *Service) resolveClient(req domain.ChatRequest) ai.Client {
	if req.Provider == "" && req.Model == "" && req.BaseURL == "" {
		return s.aiClient
	}

	// If using OpenAICompatibleClient and only model changed, reuse client instance
	if req.Provider == "" && req.BaseURL == "" && req.Model != "" {
		if oai, ok := s.aiClient.(*ai.OpenAICompatibleClient); ok {
			return oai.WithModel(req.Model)
		}
	}

	provider := ai.ProviderType(strings.ToLower(strings.TrimSpace(req.Provider)))
	if provider == "" {
		provider = s.defaultAIConfig.Provider
		if provider == "" {
			provider = ai.ProviderGroq
		}
	}

	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = s.defaultAIConfig.BaseURL
	}

	apiKey := s.defaultAIConfig.APIKey

	model := req.Model
	if model == "" {
		model = s.defaultAIConfig.Model
	}

	return ai.NewClient(ai.Config{
		Provider: provider,
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
	})
}

// BuildGroundedPrompt formats the user query alongside retrieved evidence snippets.
func BuildGroundedPrompt(query string, pkg *domain.EvidencePackage) string {
	var sb strings.Builder
	sb.WriteString("You are the GitWise Open Source Mentor, an AI guide helping developers understand and contribute to this repository.\n")
	sb.WriteString("CRITICAL RULE: Base your explanation solely on the verified code evidence below.\n")
	sb.WriteString("Whenever you cite a fact, reference the exact evidence tag (e.g. [ev_01], [ev_02]).\n")
	sb.WriteString("Do not invent non-existent file names, line coordinates, or evidence tags.\n\n")

	if pkg != nil && len(pkg.Items) > 0 {
		sb.WriteString("=== VERIFIED CODE EVIDENCE ===\n")
		for _, item := range pkg.Items {
			sb.WriteString(fmt.Sprintf("[%s] File: %s (lines %d-%d)\n", item.ID, item.FilePath, item.StartLine, item.EndLine))
			sb.WriteString("```\n")
			sb.WriteString(item.Snippet)
			sb.WriteString("\n```\n\n")
		}
	} else {
		sb.WriteString("No specific code snippets were retrieved for this query. Explain based on general repository entrypoints.\n\n")
	}

	sb.WriteString("=== CONTRIBUTOR QUESTION ===\n")
	sb.WriteString(query)
	sb.WriteString("\n\n=== GROUNDED MENTOR RESPONSE ===\n")
	return sb.String()
}

// Chat processes a synchronous question, retrieving evidence, calling the LLM, and validating citations.
func (s *Service) Chat(
	ctx context.Context,
	snapshotID uuid.UUID,
	commitSHA string,
	req domain.ChatRequest,
) (*domain.ChatResponse, error) {
	cleanMsg := strings.TrimSpace(req.Message)
	if cleanMsg == "" {
		return nil, errors.New("chat message cannot be empty")
	}

	// 1. Resolve or create session
	session, err := s.GetOrCreateSession(ctx, snapshotID, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create mentor session: %w", err)
	}

	// 2. Persist user message
	_ = s.persistMessage(ctx, session.ID, "user", cleanMsg, nil)

	// 3. Retrieve code evidence via tiered hybrid search
	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}
	evidencePkg, err := s.retrieveEvidence(ctx, snapshotID, commitSHA, cleanMsg, topK)
	if err != nil {
		s.logger.WarnContext(ctx, "retrieval failed for mentor chat, proceeding with fallback", "error", err)
	}

	// 4. Construct prompt and query AI client
	client := s.resolveClient(req)
	prompt := BuildGroundedPrompt(cleanMsg, evidencePkg)
	rawAnswer, err := client.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// 5. Strict citation validation: drop any hallucinated evidence IDs
	validationRes := s.validator.ValidateAndHydrate(rawAnswer, evidencePkg)
	if len(validationRes.RejectedEvidenceIDs) > 0 {
		s.logger.WarnContext(ctx, "rejected hallucinated evidence IDs",
			"rejected", validationRes.RejectedEvidenceIDs,
		)
	}

	// 6. Persist mentor response
	_ = s.persistMessage(ctx, session.ID, "mentor", validationRes.SanitizedText, validationRes.ValidatedCitations)

	return &domain.ChatResponse{
		SessionID: session.ID,
		Message:   validationRes.SanitizedText,
		Citations: validationRes.ValidatedCitations,
	}, nil
}

// ChatStream processes a question with live token streaming and citation validation upon stream completion.
func (s *Service) ChatStream(
	ctx context.Context,
	snapshotID uuid.UUID,
	commitSHA string,
	req domain.ChatRequest,
) (<-chan domain.ChatStreamChunk, error) {
	cleanMsg := strings.TrimSpace(req.Message)
	if cleanMsg == "" {
		return nil, errors.New("chat message cannot be empty")
	}

	session, err := s.GetOrCreateSession(ctx, snapshotID, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create mentor session: %w", err)
	}

	_ = s.persistMessage(ctx, session.ID, "user", cleanMsg, nil)

	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}
	evidencePkg, _ := s.retrieveEvidence(ctx, snapshotID, commitSHA, cleanMsg, topK)

	client := s.resolveClient(req)
	prompt := BuildGroundedPrompt(cleanMsg, evidencePkg)
	tokenStream, err := client.GenerateStream(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI stream generation failed: %w", err)
	}

	out := make(chan domain.ChatStreamChunk, 32)
	go func() {
		defer close(out)
		var fullText strings.Builder

		for chunk := range tokenStream {
			fullText.WriteString(chunk)
			out <- domain.ChatStreamChunk{
				SessionID: session.ID,
				Delta:     chunk,
				Done:      false,
			}
		}

		// Stream ended: Validate citations across complete generated text
		validationRes := s.validator.ValidateAndHydrate(fullText.String(), evidencePkg)
		_ = s.persistMessage(ctx, session.ID, "mentor", validationRes.SanitizedText, validationRes.ValidatedCitations)

		out <- domain.ChatStreamChunk{
			SessionID: session.ID,
			Done:      true,
			Citations: validationRes.ValidatedCitations,
		}
	}()

	return out, nil
}

func (s *Service) retrieveEvidence(
	ctx context.Context,
	snapshotID uuid.UUID,
	commitSHA string,
	query string,
	topK int,
) (*domain.EvidencePackage, error) {
	if s.retrievalSvc == nil || s.evidenceStore == nil {
		return nil, nil
	}

	results, err := s.retrievalSvc.HybridSearch(ctx, snapshotID, query, topK)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	var refs []domain.EvidenceRef
	for i, r := range results {
		refs = append(refs, domain.EvidenceRef{
			ID:          fmt.Sprintf("ev_%02d", i+1),
			SnapshotID:  snapshotID,
			FileID:      r.Chunk.FileID,
			FilePath:    r.Chunk.FilePath,
			StartLine:   r.Chunk.StartLine,
			EndLine:     r.Chunk.EndLine,
			ContentHash: evidence.ComputeHash(r.Chunk.Content),
			Provenance:  r.RankTier,
			Snippet:     r.Chunk.Content,
		})
	}

	return s.evidenceStore.AssembleEvidencePackage(ctx, snapshotID, commitSHA, query, refs)
}

// GetOrCreateSession retrieves an existing session or creates a new one for the snapshot.
func (s *Service) GetOrCreateSession(
	ctx context.Context,
	snapshotID uuid.UUID,
	sessionID *uuid.UUID,
) (*domain.MentorSession, error) {
	if s.db == nil {
		id := uuid.New()
		if sessionID != nil && *sessionID != uuid.Nil {
			id = *sessionID
		}
		return &domain.MentorSession{
			ID:         id,
			SnapshotID: snapshotID,
			Title:      "Mentorship Session",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}, nil
	}

	if sessionID != nil && *sessionID != uuid.Nil {
		var sess domain.MentorSession
		err := s.db.QueryRowContext(ctx, `
			SELECT id, snapshot_id, title, created_at, updated_at
			FROM mentor_sessions
			WHERE id = $1`, *sessionID).Scan(
			&sess.ID, &sess.SnapshotID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt,
		)
		if err == nil {
			return &sess, nil
		}
	}

	// Create new session
	newID := uuid.New()
	var sess domain.MentorSession
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO mentor_sessions (id, snapshot_id, title, created_at, updated_at)
		VALUES ($1, $2, 'Mentorship Session', NOW(), NOW())
		RETURNING id, snapshot_id, title, created_at, updated_at`,
		newID, snapshotID).Scan(
		&sess.ID, &sess.SnapshotID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert mentor session: %w", err)
	}

	return &sess, nil
}

func (s *Service) persistMessage(
	ctx context.Context,
	sessionID uuid.UUID,
	role, content string,
	citations []domain.ChatCitation,
) error {
	if s.db == nil {
		return nil
	}

	citationsJSON, err := json.Marshal(citations)
	if err != nil {
		citationsJSON = []byte("[]")
	}

	query := `
		INSERT INTO chat_messages (session_id, role, content, citations, created_at)
		VALUES ($1, $2, $3, $4::jsonb, NOW())`

	_, err = s.db.ExecContext(ctx, query, sessionID, role, content, string(citationsJSON))
	return err
}

// GetSessionMessages retrieves conversation turns for a session.
func (s *Service) GetSessionMessages(ctx context.Context, sessionID uuid.UUID) ([]domain.MentorMessage, error) {
	if s.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, session_id, role, content, citations, created_at
		FROM chat_messages
		WHERE session_id = $1
		ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.MentorMessage
	for rows.Next() {
		var msg domain.MentorMessage
		var citStr string
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &citStr, &msg.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(citStr), &msg.Citations)
		messages = append(messages, msg)
	}

	return messages, nil
}
