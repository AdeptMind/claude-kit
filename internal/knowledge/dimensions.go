package knowledge

import (
	"context"
	"fmt"
	"time"
)

// CreateDimension creates a new dimension. Name must be unique.
func (s *Store) CreateDimension(ctx context.Context, dim *Dimension) error {
	dim.ID = newID()
	ts := now()
	dim.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO dimensions (id, name, description, created_at) VALUES (?, ?, ?, ?)`,
		dim.ID, dim.Name, dim.Description, ts,
	)
	return err
}

// DeleteDimension removes a dimension and its node associations.
func (s *Store) DeleteDimension(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM dimensions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("dimension %q not found", id)
	}
	return nil
}

// TagNode associates a node with a dimension.
func (s *Store) TagNode(ctx context.Context, nodeID, dimensionID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO node_dimensions (node_id, dimension_id) VALUES (?, ?)`,
		nodeID, dimensionID,
	)
	return err
}

// UntagNode removes the association between a node and a dimension.
func (s *Store) UntagNode(ctx context.Context, nodeID, dimensionID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM node_dimensions WHERE node_id = ? AND dimension_id = ?`,
		nodeID, dimensionID,
	)
	return err
}
