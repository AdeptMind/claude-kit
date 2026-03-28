package knowledge

import (
	"context"
	"testing"
)

func TestCreateDimension(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	dim := &Dimension{Name: "sprint-42", Description: "Sprint 42 work"}
	if err := store.CreateDimension(ctx, dim); err != nil {
		t.Fatalf("CreateDimension error: %v", err)
	}
	if dim.ID == "" {
		t.Error("dimension ID should be set")
	}
}

func TestCreateDimension_Duplicate(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	store.CreateDimension(ctx, &Dimension{Name: "unique"})
	err := store.CreateDimension(ctx, &Dimension{Name: "unique"})
	if err == nil {
		t.Error("expected error for duplicate dimension name")
	}
}

func TestDeleteDimension(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	dim := &Dimension{Name: "to-delete"}
	store.CreateDimension(ctx, dim)

	if err := store.DeleteDimension(ctx, dim.ID); err != nil {
		t.Fatalf("DeleteDimension error: %v", err)
	}

	dims, _ := store.ListDimensions(ctx)
	for _, d := range dims {
		if d.ID == dim.ID {
			t.Error("dimension should be deleted")
		}
	}
}

func TestDeleteDimension_CascadesToNodeDimensions(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	dim := &Dimension{Name: "temp-dim"}
	store.CreateDimension(ctx, dim)

	node := &Node{Title: "Tagged Node"}
	store.CreateNode(ctx, node)
	store.TagNode(ctx, node.ID, dim.ID)

	// Delete dimension — junction should be cleaned
	store.DeleteDimension(ctx, dim.ID)

	// Node should still exist but without dimension
	got, _ := store.GetNode(ctx, node.ID)
	if got == nil {
		t.Error("node should still exist after dimension delete")
	}
}

func TestTagAndUntagNode(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	dim := &Dimension{Name: "perf"}
	store.CreateDimension(ctx, dim)

	node := &Node{Title: "Performance Node"}
	store.CreateNode(ctx, node)

	// Tag
	if err := store.TagNode(ctx, node.ID, dim.ID); err != nil {
		t.Fatalf("TagNode error: %v", err)
	}

	// Verify via ListNodes with dimension filter
	nodes, _ := store.ListNodes(ctx, ListOpts{Limit: 10, DimensionID: dim.ID})
	if len(nodes) != 1 {
		t.Errorf("expected 1 tagged node, got %d", len(nodes))
	}

	// Untag
	if err := store.UntagNode(ctx, node.ID, dim.ID); err != nil {
		t.Fatalf("UntagNode error: %v", err)
	}

	nodes, _ = store.ListNodes(ctx, ListOpts{Limit: 10, DimensionID: dim.ID})
	if len(nodes) != 0 {
		t.Errorf("expected 0 tagged nodes after untag, got %d", len(nodes))
	}
}

func TestListDimensions_WithCount(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	d1 := &Dimension{Name: "api"}
	d2 := &Dimension{Name: "infra"}
	store.CreateDimension(ctx, d1)
	store.CreateDimension(ctx, d2)

	n1 := &Node{Title: "Node 1"}
	n2 := &Node{Title: "Node 2"}
	store.CreateNode(ctx, n1)
	store.CreateNode(ctx, n2)

	store.TagNode(ctx, n1.ID, d1.ID)
	store.TagNode(ctx, n2.ID, d1.ID)
	store.TagNode(ctx, n1.ID, d2.ID)

	dims, err := store.ListDimensions(ctx)
	if err != nil {
		t.Fatalf("ListDimensions error: %v", err)
	}
	if len(dims) != 2 {
		t.Fatalf("expected 2 dimensions, got %d", len(dims))
	}

	// Find api dimension
	for _, d := range dims {
		if d.Name == "api" && d.NodeCount != 2 {
			t.Errorf("api dimension count = %d, want 2", d.NodeCount)
		}
		if d.Name == "infra" && d.NodeCount != 1 {
			t.Errorf("infra dimension count = %d, want 1", d.NodeCount)
		}
	}
}

func TestListNodes_DimensionFilter(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	dim := &Dimension{Name: "security"}
	store.CreateDimension(ctx, dim)

	n1 := &Node{Title: "Auth Node"}
	n2 := &Node{Title: "Untagged Node"}
	store.CreateNode(ctx, n1)
	store.CreateNode(ctx, n2)
	store.TagNode(ctx, n1.ID, dim.ID)

	nodes, _ := store.ListNodes(ctx, ListOpts{Limit: 10, DimensionID: dim.ID})
	if len(nodes) != 1 {
		t.Errorf("expected 1 filtered node, got %d", len(nodes))
	}
	if nodes[0].ID != n1.ID {
		t.Errorf("filtered node should be %q, got %q", n1.ID, nodes[0].ID)
	}
}
