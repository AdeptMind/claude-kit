package knowledge

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func newID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// CreateNode inserts a new node. ID and timestamps are auto-generated.
func (s *Store) CreateNode(ctx context.Context, node *Node) error {
	node.ID = newID()
	ts := now()
	node.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)
	node.UpdatedAt = node.CreatedAt

	if node.Type == "" {
		node.Type = "note"
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO nodes (id, title, content, type, embedding, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		node.ID, node.Title, node.Content, node.Type, EncodeEmbedding(node.Embedding), ts, ts,
	)
	return err
}

// GetNode retrieves a node by ID.
func (s *Store) GetNode(ctx context.Context, id string) (*Node, error) {
	var n Node
	var createdAt, updatedAt string
	var embBlob []byte

	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, content, type, embedding, created_at, updated_at
		 FROM nodes WHERE id = ?`, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.Type, &embBlob, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("node %q not found", id)
	}
	if err != nil {
		return nil, err
	}

	n.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	n.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	n.Embedding = DecodeEmbedding(embBlob)
	return &n, nil
}

// UpdateNode modifies an existing node's title, content, and type.
func (s *Store) UpdateNode(ctx context.Context, node *Node) error {
	ts := now()
	node.UpdatedAt, _ = time.Parse(time.RFC3339Nano, ts)

	res, err := s.db.ExecContext(ctx,
		`UPDATE nodes SET title = ?, content = ?, type = ?, updated_at = ? WHERE id = ?`,
		node.Title, node.Content, node.Type, ts, node.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("node %q not found", node.ID)
	}
	return nil
}

// DeleteNode removes a node and cascades to edges and dimension tags.
func (s *Store) DeleteNode(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("node %q not found", id)
	}
	return nil
}

// UpdateNodeEmbedding updates only the embedding column for a node.
func (s *Store) UpdateNodeEmbedding(ctx context.Context, id string, vec []float32) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE nodes SET embedding = ? WHERE id = ?`,
		EncodeEmbedding(vec), id,
	)
	return err
}

// ListNodes returns a paginated list of node summaries, sorted by most recent.
func (s *Store) ListNodes(ctx context.Context, opts ListOpts) ([]NodeSummary, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}

	query := `SELECT n.id, n.title, n.type, n.created_at FROM nodes n`
	var args []any

	if opts.DimensionID != "" {
		query += ` JOIN node_dimensions nd ON nd.node_id = n.id WHERE nd.dimension_id = ?`
		args = append(args, opts.DimensionID)
	}

	query += ` ORDER BY n.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []NodeSummary
	for rows.Next() {
		var ns NodeSummary
		var createdAt string
		if err := rows.Scan(&ns.ID, &ns.Title, &ns.Type, &createdAt); err != nil {
			return nil, err
		}
		ns.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		ns.Dimensions = s.getNodeDimensionNames(ctx, ns.ID)
		results = append(results, ns)
	}
	return results, rows.Err()
}

func (s *Store) getNodeDimensionNames(ctx context.Context, nodeID string) []string {
	rows, err := s.db.QueryContext(ctx,
		`SELECT d.name FROM dimensions d
		 JOIN node_dimensions nd ON nd.dimension_id = d.id
		 WHERE nd.node_id = ?`, nodeID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			names = append(names, name)
		}
	}
	return names
}

// EncodeEmbedding serializes a float32 slice to little-endian bytes.
func EncodeEmbedding(v []float32) []byte {
	if len(v) == 0 {
		return nil
	}
	buf := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

// DecodeEmbedding deserializes little-endian bytes to a float32 slice.
func DecodeEmbedding(b []byte) []float32 {
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
