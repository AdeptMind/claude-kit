package knowledge

import (
	"context"
	"fmt"
	"time"
)

// CreateEdge creates a relationship between two nodes.
// Both nodes must exist. The combination (from, to, type) must be unique.
func (s *Store) CreateEdge(ctx context.Context, edge *Edge) error {
	// Validate both nodes exist
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM nodes WHERE id IN (?, ?)`, edge.FromNodeID, edge.ToNodeID,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count < 2 {
		return fmt.Errorf("one or both nodes not found (from=%q, to=%q)", edge.FromNodeID, edge.ToNodeID)
	}

	edge.ID = newID()
	ts := now()
	edge.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)

	if edge.Type == "" {
		edge.Type = "related"
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO edges (id, from_node_id, to_node_id, type, explanation, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		edge.ID, edge.FromNodeID, edge.ToNodeID, edge.Type, edge.Explanation, ts,
	)
	return err
}

// GetEdges returns all edges connected to a node (both directions).
func (s *Store) GetEdges(ctx context.Context, nodeID string) ([]Edge, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, from_node_id, to_node_id, type, explanation, created_at
		 FROM edges WHERE from_node_id = ? OR to_node_id = ?
		 ORDER BY created_at DESC`, nodeID, nodeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		var createdAt string
		if err := rows.Scan(&e.ID, &e.FromNodeID, &e.ToNodeID, &e.Type, &e.Explanation, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// DeleteEdge removes an edge by ID.
func (s *Store) DeleteEdge(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM edges WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("edge %q not found", id)
	}
	return nil
}
