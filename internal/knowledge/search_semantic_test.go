package knowledge

import (
	"context"
	"testing"
)

func TestSearchSemantic_BruteForce(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	// Create nodes with known embeddings
	// Vector [1,0,0] is similar to [0.9,0.1,0] but not to [0,0,1]
	n1 := &Node{Title: "Auth decision", Content: "JWT tokens", Embedding: []float32{1.0, 0.0, 0.0}}
	n2 := &Node{Title: "Similar auth", Content: "OAuth flow", Embedding: []float32{0.9, 0.1, 0.0}}
	n3 := &Node{Title: "Unrelated", Content: "Database stuff", Embedding: []float32{0.0, 0.0, 1.0}}
	store.CreateNode(ctx, n1)
	store.CreateNode(ctx, n2)
	store.CreateNode(ctx, n3)

	// Force brute-force mode (vecAvail=false is default)
	results, err := store.searchSemanticBruteForce(ctx, []float32{1.0, 0.0, 0.0}, 10)
	if err != nil {
		t.Fatalf("searchSemanticBruteForce error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}

	// First result should be most similar (n1, exact match)
	if results[0].Node.ID != n1.ID {
		t.Errorf("first result should be %q, got %q", n1.ID, results[0].Node.ID)
	}
	// Second should be n2 (high similarity)
	if results[1].Node.ID != n2.ID {
		t.Errorf("second result should be %q, got %q", n2.ID, results[1].Node.ID)
	}
	// Third should be n3 (orthogonal, score ~0)
	if results[2].Node.ID != n3.ID {
		t.Errorf("third result should be %q, got %q", n3.ID, results[2].Node.ID)
	}
	// Scores should be descending
	if results[0].Score < results[1].Score {
		t.Error("scores should be descending")
	}
	if results[0].MatchType != "semantic" {
		t.Errorf("matchType = %q, want %q", results[0].MatchType, "semantic")
	}
}

func TestSearchSemantic_NoVec(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	// vecAvail is false by default
	_, err := store.SearchSemantic(ctx, []float32{1.0}, 10)
	if err == nil {
		t.Error("expected error when sqlite-vec not available")
	}
}

func TestSearchHybrid_KeywordOnly(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	store.CreateNode(ctx, &Node{Title: "Rate limiting decision", Content: "Token bucket algorithm"})
	store.CreateNode(ctx, &Node{Title: "Redis cache", Content: "Cache invalidation"})

	// No vector → falls back to keyword
	results, err := store.SearchHybrid(ctx, "rate limiting", nil, 10)
	if err != nil {
		t.Fatalf("SearchHybrid error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1 (keyword-only fallback)", len(results))
	}
}

func TestSearchHybrid_WithEmbeddings(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	// Node with both text and embedding
	store.CreateNode(ctx, &Node{
		Title: "Rate limiting", Content: "throttle API requests",
		Embedding: []float32{1.0, 0.0, 0.0},
	})
	store.CreateNode(ctx, &Node{
		Title: "API quotas", Content: "quota management",
		Embedding: []float32{0.9, 0.1, 0.0},
	})
	store.CreateNode(ctx, &Node{
		Title: "Unrelated", Content: "database migration",
		Embedding: []float32{0.0, 0.0, 1.0},
	})

	// Hybrid search: keyword "rate" + semantic vector similar to first two
	results, err := store.SearchHybrid(ctx, "rate", []float32{1.0, 0.0, 0.0}, 10)
	if err != nil {
		t.Fatalf("SearchHybrid error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results from hybrid search")
	}
	// First result should be "Rate limiting" (matches both keyword and semantic)
	if results[0].MatchType != "hybrid" {
		t.Errorf("matchType = %q, want %q", results[0].MatchType, "hybrid")
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float32
		want float64
	}{
		{"identical", []float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{"orthogonal", []float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{"opposite", []float32{1, 0, 0}, []float32{-1, 0, 0}, -1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cosineSimilarity(tt.a, tt.b)
			if abs(got-tt.want) > 0.001 {
				t.Errorf("cosineSimilarity = %f, want %f", got, tt.want)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
