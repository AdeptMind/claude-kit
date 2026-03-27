package knowledge

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
)

// SearchKeyword performs a full-text search using FTS5.
// Returns results ranked by relevance with snippet extraction.
func (s *Store) SearchKeyword(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT n.id, n.title, n.type, n.created_at,
		       snippet(nodes_fts, 1, '<b>', '</b>', '...', 64) AS snip,
		       rank
		FROM nodes_fts
		JOIN nodes n ON n.rowid = nodes_fts.rowid
		WHERE nodes_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var createdAt string
		var rank float64
		if err := rows.Scan(&r.Node.ID, &r.Node.Title, &r.Node.Type, &createdAt, &r.Snippet, &rank); err != nil {
			return nil, err
		}
		r.Node.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		r.Score = -rank // FTS5 rank is negative (lower = better), invert for display
		r.MatchType = "keyword"
		r.Node.Dimensions = s.getNodeDimensionNames(ctx, r.Node.ID)
		results = append(results, r)
	}
	return results, rows.Err()
}

// SetEmbedder configures the embedding backend for the store.
// When set, CreateNode auto-computes embeddings.
func (s *Store) SetEmbedder(e embedder.Embedder) {
	s.embedder = e
}

// GetEmbedder returns the configured embedder (may be nil).
func (s *Store) GetEmbedder() embedder.Embedder {
	return s.embedder
}

// SearchSemantic performs vector similarity search using cosine distance.
// Requires sqlite-vec extension and pre-computed embeddings.
// Falls back to error if sqlite-vec is not available.
func (s *Store) SearchSemantic(ctx context.Context, vector []float32, limit int) ([]SearchResult, error) {
	if !s.vecAvail {
		return nil, fmt.Errorf("semantic search unavailable: sqlite-vec extension not loaded")
	}
	if limit <= 0 {
		limit = 10
	}

	// Brute-force cosine similarity in Go since sqlite-vec may not be available.
	// When sqlite-vec IS available, we could use vec_distance_cosine.
	// For now, use the Go-based approach which works universally.
	return s.searchSemanticBruteForce(ctx, vector, limit)
}

func (s *Store) searchSemanticBruteForce(ctx context.Context, queryVec []float32, limit int) ([]SearchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, type, content, embedding, created_at FROM nodes WHERE embedding IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		result SearchResult
		score  float64
	}
	var candidates []scored

	for rows.Next() {
		var id, title, typ, content, createdAt string
		var embBlob []byte
		if err := rows.Scan(&id, &title, &typ, &content, &embBlob, &createdAt); err != nil {
			return nil, err
		}
		nodeVec := decodeFloat32s(embBlob)
		if len(nodeVec) != len(queryVec) {
			continue
		}
		sim := cosineSimilarity(queryVec, nodeVec)
		t, _ := time.Parse(time.RFC3339Nano, createdAt)

		snippet := content
		if len(snippet) > 128 {
			snippet = snippet[:128] + "..."
		}

		candidates = append(candidates, scored{
			result: SearchResult{
				Node:      NodeSummary{ID: id, Title: title, Type: typ, CreatedAt: t},
				Snippet:   snippet,
				Score:     sim,
				MatchType: "semantic",
			},
			score: sim,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	var results []SearchResult
	for i, c := range candidates {
		if i >= limit {
			break
		}
		c.result.Node.Dimensions = s.getNodeDimensionNames(ctx, c.result.Node.ID)
		results = append(results, c.result)
	}
	return results, nil
}

// SearchHybrid merges keyword and semantic results using Reciprocal Rank Fusion.
// If no embedder is available, falls back to keyword-only search.
func (s *Store) SearchHybrid(ctx context.Context, query string, vector []float32, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	kwResults, err := s.SearchKeyword(ctx, query, limit*2)
	if err != nil {
		return nil, err
	}

	// If no vector provided or no embeddings, return keyword results
	if len(vector) == 0 {
		if len(kwResults) > limit {
			kwResults = kwResults[:limit]
		}
		return kwResults, nil
	}

	semResults, _ := s.searchSemanticBruteForce(ctx, vector, limit*2)

	// RRF merge: score = sum(1 / (k + rank)) where k=60
	const k = 60.0
	scores := make(map[string]float64)
	nodeMap := make(map[string]SearchResult)

	for rank, r := range kwResults {
		scores[r.Node.ID] += 1.0 / (k + float64(rank+1))
		nodeMap[r.Node.ID] = r
	}
	for rank, r := range semResults {
		scores[r.Node.ID] += 1.0 / (k + float64(rank+1))
		if _, exists := nodeMap[r.Node.ID]; !exists {
			nodeMap[r.Node.ID] = r
		}
	}

	type scored struct {
		id    string
		score float64
	}
	var merged []scored
	for id, score := range scores {
		merged = append(merged, scored{id, score})
	}
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].score > merged[j].score
	})

	var results []SearchResult
	for i, m := range merged {
		if i >= limit {
			break
		}
		r := nodeMap[m.id]
		r.Score = m.score
		r.MatchType = "hybrid"
		results = append(results, r)
	}
	return results, nil
}

func decodeFloat32s(b []byte) []float32 {
	if len(b) == 0 {
		return nil
	}
	n := len(b) / 4
	v := make([]float32, n)
	for i := range v {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return v
}

func cosineSimilarity(a, b []float32) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// GetStats returns aggregate statistics about the knowledge graph.
func (s *Store) GetStats(ctx context.Context) (*Stats, error) {
	var st Stats
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&st.Nodes)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM edges`).Scan(&st.Edges)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dimensions`).Scan(&st.Dimensions)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes WHERE embedding IS NOT NULL`).Scan(&st.NodesWithEmbedding)
	return &st, nil
}
