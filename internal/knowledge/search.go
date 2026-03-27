package knowledge

import (
	"context"
	"time"
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

// GetStats returns aggregate statistics about the knowledge graph.
func (s *Store) GetStats(ctx context.Context) (*Stats, error) {
	var st Stats
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&st.Nodes)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM edges`).Scan(&st.Edges)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dimensions`).Scan(&st.Dimensions)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes WHERE embedding IS NOT NULL`).Scan(&st.NodesWithEmbedding)
	return &st, nil
}
