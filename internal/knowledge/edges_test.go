package knowledge

import (
	"context"
	"testing"
)

func TestCreateEdge(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	a := &Node{Title: "A"}
	b := &Node{Title: "B"}
	store.CreateNode(ctx, a)
	store.CreateNode(ctx, b)

	edge := &Edge{FromNodeID: a.ID, ToNodeID: b.ID, Type: "motivated_by", Explanation: "because reasons"}
	if err := store.CreateEdge(ctx, edge); err != nil {
		t.Fatalf("CreateEdge error: %v", err)
	}
	if edge.ID == "" {
		t.Error("edge.ID should be set")
	}
}

func TestCreateEdge_NonexistentNode(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	a := &Node{Title: "A"}
	store.CreateNode(ctx, a)

	err := store.CreateEdge(ctx, &Edge{FromNodeID: a.ID, ToNodeID: "nonexistent", Type: "related"})
	if err == nil {
		t.Error("expected error for nonexistent target node")
	}
}

func TestCreateEdge_Duplicate(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	a := &Node{Title: "A"}
	b := &Node{Title: "B"}
	store.CreateNode(ctx, a)
	store.CreateNode(ctx, b)

	store.CreateEdge(ctx, &Edge{FromNodeID: a.ID, ToNodeID: b.ID, Type: "related"})
	err := store.CreateEdge(ctx, &Edge{FromNodeID: a.ID, ToNodeID: b.ID, Type: "related"})
	if err == nil {
		t.Error("expected error for duplicate edge")
	}
}

func TestGetEdges_BothDirections(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	a := &Node{Title: "A"}
	b := &Node{Title: "B"}
	c := &Node{Title: "C"}
	store.CreateNode(ctx, a)
	store.CreateNode(ctx, b)
	store.CreateNode(ctx, c)

	store.CreateEdge(ctx, &Edge{FromNodeID: a.ID, ToNodeID: b.ID, Type: "related"})
	store.CreateEdge(ctx, &Edge{FromNodeID: c.ID, ToNodeID: a.ID, Type: "depends_on"})

	edges, err := store.GetEdges(ctx, a.ID)
	if err != nil {
		t.Fatalf("GetEdges error: %v", err)
	}
	if len(edges) != 2 {
		t.Errorf("got %d edges, want 2 (both directions)", len(edges))
	}
}

func TestDeleteEdge(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	a := &Node{Title: "A"}
	b := &Node{Title: "B"}
	store.CreateNode(ctx, a)
	store.CreateNode(ctx, b)

	edge := &Edge{FromNodeID: a.ID, ToNodeID: b.ID, Type: "related"}
	store.CreateEdge(ctx, edge)

	if err := store.DeleteEdge(ctx, edge.ID); err != nil {
		t.Fatalf("DeleteEdge error: %v", err)
	}

	// Verify edge is gone
	edges, _ := store.GetEdges(ctx, a.ID)
	if len(edges) != 0 {
		t.Errorf("got %d edges after delete, want 0", len(edges))
	}

	// Verify nodes still exist
	_, err := store.GetNode(ctx, a.ID)
	if err != nil {
		t.Error("node A should still exist after edge delete")
	}
}

func TestDeleteEdge_NotFound(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	err := store.DeleteEdge(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent edge")
	}
}
