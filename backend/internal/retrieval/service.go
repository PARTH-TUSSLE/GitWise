package retrieval

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/gitwise/backend/internal/domain"
	"github.com/google/uuid"
)

// Service coordinates code chunk storage, vector indexing, and tiered hybrid retrieval.
type Service struct {
	db       *sql.DB
	embedder Embedder
	logger   *slog.Logger
}

// NewService creates a new retrieval service.
func NewService(db *sql.DB, embedder Embedder, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if embedder == nil {
		embedder = NewMockEmbedder()
	}
	return &Service{
		db:       db,
		embedder: embedder,
		logger:   logger,
	}
}

// PersistChunks writes semantic code chunks and their embeddings to PostgreSQL.
func (s *Service) PersistChunks(ctx context.Context, snapshotID uuid.UUID, chunks []domain.CodeChunk) error {
	if len(chunks) == 0 || s.db == nil {
		return nil
	}

	// 1. Compute missing embeddings in batch
	var textsToEmbed []string
	var unindexedIndices []int

	for i, c := range chunks {
		if len(c.Embedding) == 0 {
			textsToEmbed = append(textsToEmbed, c.Content)
			unindexedIndices = append(unindexedIndices, i)
		}
	}

	if len(textsToEmbed) > 0 {
		embeddings, err := s.embedder.EmbedBatch(ctx, textsToEmbed)
		if err != nil {
			return fmt.Errorf("failed to generate chunk embeddings: %w", err)
		}
		for j, emb := range embeddings {
			idx := unindexedIndices[j]
			chunks[idx].Embedding = emb
			chunks[idx].EmbeddingModel = s.embedder.ModelName()
		}
	}

	// 2. Insert chunks inside transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for chunk persistence: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO code_chunks (
			id, snapshot_id, file_id, symbol_id, start_line, end_line, scope, content, embedding, embedding_model, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::vector, $9, NOW())`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare chunk insert statement: %w", err)
	}
	defer stmt.Close()

	for _, c := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		vecStr := FormatVector(c.Embedding)
		modelName := c.EmbeddingModel
		if modelName == "" {
			modelName = s.embedder.ModelName()
		}

		_, err := stmt.ExecContext(ctx,
			c.ID,
			snapshotID,
			c.FileID,
			c.SymbolID,
			c.StartLine,
			c.EndLine,
			c.Scope,
			c.Content,
			vecStr,
			modelName,
		)
		if err != nil {
			return fmt.Errorf("failed to insert code chunk: %w", err)
		}
	}

	return tx.Commit()
}

// HybridSearch executes tiered hybrid retrieval combining:
// Tier 1: Exact Symbol Matching (highest priority anchor)
// Tier 2: Full-Text Lexical Search (ts_rank)
// Tier 3: Dense Vector Similarity (cosine distance)
// Fused via weighted reciprocal rank fusion without arbitrary magic numbers dominating exact symbols.
func (s *Service) HybridSearch(ctx context.Context, snapshotID uuid.UUID, query string, topK int) ([]domain.SearchResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return nil, nil
	}
	if topK <= 0 {
		topK = 5
	}
	if topK > 20 {
		topK = 20
	}

	seenChunks := make(map[uuid.UUID]bool)
	var finalResults []domain.SearchResult

	// -------------------------------------------------------------
	// TIER 1: Exact / In-Query Symbol Matching
	// -------------------------------------------------------------
	symbolQuery := `
		SELECT c.id, c.snapshot_id, c.file_id, rf.path, c.symbol_id, cs.name, c.start_line, c.end_line, c.scope, c.content, c.embedding_model, c.created_at
		FROM code_symbols cs
		JOIN code_chunks c ON c.symbol_id = cs.id
		JOIN repository_files rf ON c.file_id = rf.id
		WHERE c.snapshot_id = $1
		  AND (LOWER(cs.name) = LOWER($2) OR LOWER($2) LIKE '%' || LOWER(cs.name) || '%')
		LIMIT $3;`

	symRows, err := s.db.QueryContext(ctx, symbolQuery, snapshotID, cleanQuery, topK)
	if err == nil {
		defer symRows.Close()
		for symRows.Next() {
			var chunk domain.CodeChunk
			var scope sql.NullString
			if err := symRows.Scan(
				&chunk.ID, &chunk.SnapshotID, &chunk.FileID, &chunk.FilePath,
				&chunk.SymbolID, &chunk.SymbolName, &chunk.StartLine, &chunk.EndLine,
				&scope, &chunk.Content, &chunk.EmbeddingModel, &chunk.CreatedAt,
			); err != nil {
				continue
			}
			if scope.Valid {
				chunk.Scope = scope.String
			}
			seenChunks[chunk.ID] = true
			finalResults = append(finalResults, domain.SearchResult{
				Chunk:       chunk,
				Score:       1.0,
				RankTier:    "SYMBOL_EXACT",
				SymbolMatch: true,
			})
		}
	}

	// -------------------------------------------------------------
	// TIER 2: Full-Text Lexical Search (PostgreSQL ts_rank)
	// -------------------------------------------------------------
	type ftsHit struct {
		chunk domain.CodeChunk
		rank  float64
	}
	var ftsHits []ftsHit

	ftsQuery := `
		SELECT c.id, c.snapshot_id, c.file_id, rf.path, c.symbol_id, COALESCE(cs.name, ''), c.start_line, c.end_line, c.scope, c.content, c.embedding_model, c.created_at,
		       ts_rank(c.content_tsv, plainto_tsquery('english', $2)) AS lexical_score
		FROM code_chunks c
		JOIN repository_files rf ON c.file_id = rf.id
		LEFT JOIN code_symbols cs ON c.symbol_id = cs.id
		WHERE c.snapshot_id = $1
		  AND c.content_tsv @@ plainto_tsquery('english', $2)
		ORDER BY lexical_score DESC
		LIMIT $3;`

	ftsRows, err := s.db.QueryContext(ctx, ftsQuery, snapshotID, cleanQuery, topK*2)
	if err == nil {
		defer ftsRows.Close()
		for ftsRows.Next() {
			var chunk domain.CodeChunk
			var scope sql.NullString
			var symName string
			var score float64
			if err := ftsRows.Scan(
				&chunk.ID, &chunk.SnapshotID, &chunk.FileID, &chunk.FilePath,
				&chunk.SymbolID, &symName, &chunk.StartLine, &chunk.EndLine,
				&scope, &chunk.Content, &chunk.EmbeddingModel, &chunk.CreatedAt,
				&score,
			); err != nil {
				continue
			}
			if scope.Valid {
				chunk.Scope = scope.String
			}
			chunk.SymbolName = symName
			ftsHits = append(ftsHits, ftsHit{chunk: chunk, rank: score})
		}
	}

	// -------------------------------------------------------------
	// TIER 3: Dense Vector Similarity Search (pgvector HNSW cosine)
	// -------------------------------------------------------------
	type vecHit struct {
		chunk    domain.CodeChunk
		distance float64
	}
	var vecHits []vecHit

	queryVecs, err := s.embedder.EmbedBatch(ctx, []string{cleanQuery})
	if err == nil && len(queryVecs) > 0 {
		qVecStr := FormatVector(queryVecs[0])
		vecQuery := `
			SELECT c.id, c.snapshot_id, c.file_id, rf.path, c.symbol_id, COALESCE(cs.name, ''), c.start_line, c.end_line, c.scope, c.content, c.embedding_model, c.created_at,
			       (c.embedding <=> $2::vector) AS cosine_distance
			FROM code_chunks c
			JOIN repository_files rf ON c.file_id = rf.id
			LEFT JOIN code_symbols cs ON c.symbol_id = cs.id
			WHERE c.snapshot_id = $1
			ORDER BY cosine_distance ASC
			LIMIT $3;`

		vecRows, err := s.db.QueryContext(ctx, vecQuery, snapshotID, qVecStr, topK*2)
		if err == nil {
			defer vecRows.Close()
			for vecRows.Next() {
				var chunk domain.CodeChunk
				var scope sql.NullString
				var symName string
				var dist float64
				if err := vecRows.Scan(
					&chunk.ID, &chunk.SnapshotID, &chunk.FileID, &chunk.FilePath,
					&chunk.SymbolID, &symName, &chunk.StartLine, &chunk.EndLine,
					&scope, &chunk.Content, &chunk.EmbeddingModel, &chunk.CreatedAt,
					&dist,
				); err != nil {
					continue
				}
				if scope.Valid {
					chunk.Scope = scope.String
				}
				chunk.SymbolName = symName
				vecHits = append(vecHits, vecHit{chunk: chunk, distance: dist})
			}
		}
	}

	// -------------------------------------------------------------
	// Fusion: Reciprocal Rank Fusion (RRF) for Tiers 2 & 3
	// -------------------------------------------------------------
	type fusedItem struct {
		chunk    domain.CodeChunk
		rrfScore float64
		ftsRank  float64
		vecDist  float64
	}
	fusedMap := make(map[uuid.UUID]*fusedItem)

	for rank, hit := range ftsHits {
		if seenChunks[hit.chunk.ID] {
			continue
		}
		item, exists := fusedMap[hit.chunk.ID]
		if !exists {
			item = &fusedItem{chunk: hit.chunk, vecDist: 2.0}
			fusedMap[hit.chunk.ID] = item
		}
		item.rrfScore += 1.0 / float64(60+rank+1)
		item.ftsRank = hit.rank
	}

	for rank, hit := range vecHits {
		if seenChunks[hit.chunk.ID] {
			continue
		}
		item, exists := fusedMap[hit.chunk.ID]
		if !exists {
			item = &fusedItem{chunk: hit.chunk}
			fusedMap[hit.chunk.ID] = item
		}
		item.rrfScore += 1.0 / float64(60+rank+1)
		item.vecDist = hit.distance
	}

	var nonSymbolList []*fusedItem
	for _, item := range fusedMap {
		nonSymbolList = append(nonSymbolList, item)
	}

	sort.Slice(nonSymbolList, func(i, j int) bool {
		return nonSymbolList[i].rrfScore > nonSymbolList[j].rrfScore
	})

	for _, item := range nonSymbolList {
		if len(finalResults) >= topK {
			break
		}
		tier := "VECTOR_SEMANTIC"
		if item.ftsRank > 0 && item.vecDist < 0.5 {
			tier = "HYBRID_FUSION"
		} else if item.ftsRank > 0 {
			tier = "FTS_TEXT"
		}

		finalResults = append(finalResults, domain.SearchResult{
			Chunk:          item.chunk,
			Score:          item.rrfScore,
			RankTier:       tier,
			SymbolMatch:    false,
			LexicalRank:    item.ftsRank,
			VectorDistance: item.vecDist,
		})
	}

	return finalResults, nil
}
