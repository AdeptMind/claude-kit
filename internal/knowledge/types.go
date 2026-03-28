package knowledge

import "time"

// Node represents a knowledge unit in the graph.
type Node struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Embedding []float32 `json:"embedding,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	ID         string    `json:"id"`
	FromNodeID string    `json:"fromNodeId"`
	ToNodeID   string    `json:"toNodeId"`
	Type       string    `json:"type"`
	Explanation string   `json:"explanation"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Dimension represents a tag or category for organizing nodes.
type Dimension struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// DimensionWithCount includes the number of nodes tagged with this dimension.
type DimensionWithCount struct {
	Dimension
	NodeCount int `json:"nodeCount"`
}

// NodeSummary is a lightweight node representation for list views.
type NodeSummary struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Type       string   `json:"type"`
	Dimensions []string `json:"dimensions"`
	CreatedAt  time.Time `json:"createdAt"`
}

// SearchResult wraps a node summary with search metadata.
type SearchResult struct {
	Node      NodeSummary `json:"node"`
	Snippet   string      `json:"snippet"`
	Score     float64     `json:"score"`
	MatchType string      `json:"matchType"` // "keyword", "semantic", "hybrid"
}

// Stats holds aggregate knowledge graph statistics.
type Stats struct {
	Nodes              int `json:"nodes"`
	Edges              int `json:"edges"`
	Dimensions         int `json:"dimensions"`
	NodesWithEmbedding int `json:"nodesWithEmbedding"`
}

// ListOpts configures node listing queries.
type ListOpts struct {
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	DimensionID string `json:"dimensionId,omitempty"`
}
