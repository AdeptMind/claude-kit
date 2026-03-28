package knowledge

import (
	"context"
	"testing"
)

func TestSearchKeyword_SingleTerm(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	store.CreateNode(ctx, &Node{Title: "OAuth2 authentication flow", Content: "JWT tokens and refresh"})
	store.CreateNode(ctx, &Node{Title: "Database migrations", Content: "Schema versioning"})
	store.CreateNode(ctx, &Node{Title: "Auth middleware", Content: "Authentication layer for API"})

	results, err := store.SearchKeyword(ctx, "authentication", 10)
	if err != nil {
		t.Fatalf("SearchKeyword error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.MatchType != "keyword" {
			t.Errorf("matchType = %q, want %q", r.MatchType, "keyword")
		}
		if r.Score <= 0 {
			t.Error("score should be positive")
		}
	}
}

func TestSearchKeyword_MultiTerm(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	store.CreateNode(ctx, &Node{Title: "OAuth2 flow for API gateway", Content: "Authentication via OAuth"})
	store.CreateNode(ctx, &Node{Title: "Redis caching", Content: "Cache invalidation"})

	results, err := store.SearchKeyword(ctx, "oauth api", 10)
	if err != nil {
		t.Fatalf("SearchKeyword error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least 1 result for multi-term search")
	}
}

func TestSearchKeyword_Snippet(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	store.CreateNode(ctx, &Node{Title: "Test", Content: "This is about authentication middleware patterns"})

	results, _ := store.SearchKeyword(ctx, "authentication", 10)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Snippet == "" {
		t.Error("snippet should not be empty")
	}
}

func TestSearchKeyword_EmptyQuery(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	store.CreateNode(ctx, &Node{Title: "Anything"})

	results, err := store.SearchKeyword(ctx, "", 10)
	if err != nil {
		t.Fatalf("SearchKeyword error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("empty query should return 0 results, got %d", len(results))
	}
}

func TestGetStats(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	a := &Node{Title: "A"}
	b := &Node{Title: "B", Embedding: []float32{0.1, 0.2}}
	store.CreateNode(ctx, a)
	store.CreateNode(ctx, b)
	store.CreateEdge(ctx, &Edge{FromNodeID: a.ID, ToNodeID: b.ID, Type: "related"})

	stats, err := store.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if stats.Nodes != 2 {
		t.Errorf("nodes = %d, want 2", stats.Nodes)
	}
	if stats.Edges != 1 {
		t.Errorf("edges = %d, want 1", stats.Edges)
	}
	if stats.NodesWithEmbedding != 1 {
		t.Errorf("nodesWithEmbedding = %d, want 1", stats.NodesWithEmbedding)
	}
}
