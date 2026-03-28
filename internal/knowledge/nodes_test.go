package knowledge

import (
	"context"
	"testing"
	"time"
)

func TestCreateAndGetNode(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	node := &Node{Title: "Test Node", Content: "Hello world", Type: "decision"}
	if err := store.CreateNode(ctx, node); err != nil {
		t.Fatalf("CreateNode error: %v", err)
	}
	if node.ID == "" {
		t.Fatal("node.ID should be set after create")
	}

	got, err := store.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatalf("GetNode error: %v", err)
	}
	if got.Title != "Test Node" {
		t.Errorf("title = %q, want %q", got.Title, "Test Node")
	}
	if got.Content != "Hello world" {
		t.Errorf("content = %q, want %q", got.Content, "Hello world")
	}
	if got.Type != "decision" {
		t.Errorf("type = %q, want %q", got.Type, "decision")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestCreateNode_DefaultType(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	node := &Node{Title: "No Type"}
	if err := store.CreateNode(ctx, node); err != nil {
		t.Fatalf("CreateNode error: %v", err)
	}

	got, err := store.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatalf("GetNode error: %v", err)
	}
	if got.Type != "note" {
		t.Errorf("default type = %q, want %q", got.Type, "note")
	}
}

func TestUpdateNode(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	node := &Node{Title: "Original", Content: "v1", Type: "note"}
	store.CreateNode(ctx, node)

	time.Sleep(10 * time.Millisecond) // RFC3339Nano has nanosecond resolution

	node.Title = "Updated"
	node.Content = "v2"
	node.Type = "decision"
	if err := store.UpdateNode(ctx, node); err != nil {
		t.Fatalf("UpdateNode error: %v", err)
	}

	got, _ := store.GetNode(ctx, node.ID)
	if got.Title != "Updated" {
		t.Errorf("title = %q, want %q", got.Title, "Updated")
	}
	if got.Content != "v2" {
		t.Errorf("content = %q, want %q", got.Content, "v2")
	}
	if got.Type != "decision" {
		t.Errorf("type = %q, want %q", got.Type, "decision")
	}
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Error("updated_at should be after created_at")
	}
}

func TestUpdateNode_NotFound(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	err := store.UpdateNode(ctx, &Node{ID: "nonexistent", Title: "x"})
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestDeleteNode(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	node := &Node{Title: "To Delete"}
	store.CreateNode(ctx, node)

	if err := store.DeleteNode(ctx, node.ID); err != nil {
		t.Fatalf("DeleteNode error: %v", err)
	}

	_, err := store.GetNode(ctx, node.ID)
	if err == nil {
		t.Error("expected error getting deleted node")
	}
}

func TestDeleteNode_CascadesEdges(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	a := &Node{Title: "A"}
	b := &Node{Title: "B"}
	store.CreateNode(ctx, a)
	store.CreateNode(ctx, b)

	edge := &Edge{FromNodeID: a.ID, ToNodeID: b.ID, Type: "related"}
	store.CreateEdge(ctx, edge)

	// Delete node A — edge should be cascaded
	store.DeleteNode(ctx, a.ID)

	edges, err := store.GetEdges(ctx, b.ID)
	if err != nil {
		t.Fatalf("GetEdges error: %v", err)
	}
	if len(edges) != 0 {
		t.Errorf("expected 0 edges after cascade delete, got %d", len(edges))
	}
}

func TestDeleteNode_NotFound(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	err := store.DeleteNode(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestListNodes(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	// Create 5 nodes
	for i := 0; i < 5; i++ {
		store.CreateNode(ctx, &Node{Title: "Node"})
		time.Sleep(5 * time.Millisecond)
	}

	// List all
	nodes, err := store.ListNodes(ctx, ListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("ListNodes error: %v", err)
	}
	if len(nodes) != 5 {
		t.Errorf("got %d nodes, want 5", len(nodes))
	}

	// Most recent first
	for i := 1; i < len(nodes); i++ {
		if nodes[i].CreatedAt.After(nodes[i-1].CreatedAt) {
			t.Error("nodes should be sorted by most recent first")
		}
	}

	// Pagination
	page1, _ := store.ListNodes(ctx, ListOpts{Limit: 2, Offset: 0})
	page2, _ := store.ListNodes(ctx, ListOpts{Limit: 2, Offset: 2})
	if len(page1) != 2 || len(page2) != 2 {
		t.Errorf("pagination: page1=%d, page2=%d, want 2,2", len(page1), len(page2))
	}
	if page1[0].ID == page2[0].ID {
		t.Error("pages should not overlap")
	}
}

func TestGetNode_NotFound(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	_, err := store.GetNode(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestEmbeddingRoundTrip(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	emb := []float32{0.1, -0.2, 0.3, 0.0, 1.0}
	node := &Node{Title: "With Embedding", Embedding: emb}
	store.CreateNode(ctx, node)

	got, _ := store.GetNode(ctx, node.ID)
	if len(got.Embedding) != len(emb) {
		t.Fatalf("embedding length = %d, want %d", len(got.Embedding), len(emb))
	}
	for i, v := range emb {
		if got.Embedding[i] != v {
			t.Errorf("embedding[%d] = %f, want %f", i, got.Embedding[i], v)
		}
	}
}
