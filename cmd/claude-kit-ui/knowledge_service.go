package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/knowledge"
)

// KnowledgeService provides Wails RPC bindings for the knowledge graph.
type KnowledgeService struct {
	store    *knowledge.Store
	emb      embedder.Embedder
	ctx      context.Context
	cacheMu  sync.Mutex
	cache    map[string]cacheEntry
}

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

func (k *KnowledgeService) startup(ctx context.Context) {
	k.ctx = ctx
	k.cache = make(map[string]cacheEntry)

	dbPath := filepath.Join(config.DataDir(), "knowledge.sqlite")
	store, err := knowledge.Open(dbPath)
	if err != nil {
		log.Printf("knowledge-service: failed to open store: %v", err)
		return
	}
	k.store = store

	k.emb = embedder.AutoDetect(ctx)
	if k.emb != nil {
		store.SetEmbedder(k.emb)
	}

	log.Printf("knowledge-service: ready (db=%s, embedder=%v)", dbPath, k.emb != nil)
}

func (k *KnowledgeService) shutdown() {
	if k.store != nil {
		k.store.Close()
	}
}

// ListNodes returns paginated node summaries.
func (k *KnowledgeService) ListNodes(limit, offset int) ([]knowledge.NodeSummary, error) {
	if k.store == nil {
		return nil, fmt.Errorf("knowledge store not initialized")
	}
	return k.store.ListNodes(k.ctx, knowledge.ListOpts{Limit: limit, Offset: offset})
}

// SearchNodes performs hybrid search (keyword + semantic if available).
func (k *KnowledgeService) SearchNodes(query string, semantic bool, limit int) ([]knowledge.SearchResult, error) {
	if k.store == nil {
		return nil, fmt.Errorf("knowledge store not initialized")
	}
	if limit <= 0 {
		limit = 20
	}

	if semantic && k.emb != nil {
		vec, err := k.emb.Embed(k.ctx, query)
		if err != nil {
			return k.store.SearchKeyword(k.ctx, query, limit)
		}
		return k.store.SearchHybrid(k.ctx, query, vec, limit)
	}
	return k.store.SearchKeyword(k.ctx, query, limit)
}

// GetNode returns a full node by ID.
func (k *KnowledgeService) GetNode(id string) (*knowledge.Node, error) {
	if k.store == nil {
		return nil, fmt.Errorf("knowledge store not initialized")
	}
	return k.store.GetNode(k.ctx, id)
}

// GetEdges returns all edges for a node.
func (k *KnowledgeService) GetEdges(nodeID string) ([]knowledge.Edge, error) {
	if k.store == nil {
		return nil, fmt.Errorf("knowledge store not initialized")
	}
	return k.store.GetEdges(k.ctx, nodeID)
}

// GetStats returns graph statistics (cached for 10s).
func (k *KnowledgeService) GetStats() (*knowledge.Stats, error) {
	if k.store == nil {
		return nil, fmt.Errorf("knowledge store not initialized")
	}
	if cached := k.getCache("stats"); cached != nil {
		return cached.(*knowledge.Stats), nil
	}
	stats, err := k.store.GetStats(k.ctx)
	if err != nil {
		return nil, err
	}
	k.setCache("stats", stats, 10*time.Second)
	return stats, nil
}

// ListDimensions returns all dimensions with node counts (cached for 10s).
func (k *KnowledgeService) ListDimensions() ([]knowledge.DimensionWithCount, error) {
	if k.store == nil {
		return nil, fmt.Errorf("knowledge store not initialized")
	}
	if cached := k.getCache("dimensions"); cached != nil {
		return cached.([]knowledge.DimensionWithCount), nil
	}
	dims, err := k.store.ListDimensions(k.ctx)
	if err != nil {
		return nil, err
	}
	k.setCache("dimensions", dims, 10*time.Second)
	return dims, nil
}

// CreateNode creates a new node.
func (k *KnowledgeService) CreateNode(title, content, nodeType string) (string, error) {
	if k.store == nil {
		return "", fmt.Errorf("knowledge store not initialized")
	}
	node := &knowledge.Node{Title: title, Content: content, Type: nodeType}
	if err := k.store.CreateNode(k.ctx, node); err != nil {
		return "", err
	}
	// Auto-embed
	if k.emb != nil && (content != "" || title != "") {
		if vec, err := k.emb.Embed(k.ctx, title+" "+content); err == nil {
			k.store.UpdateNodeEmbedding(k.ctx, node.ID, vec)
		}
	}
	return node.ID, nil
}

// CreateEdge creates a relationship between two nodes.
func (k *KnowledgeService) CreateEdge(from, to, edgeType, explanation string) (string, error) {
	if k.store == nil {
		return "", fmt.Errorf("knowledge store not initialized")
	}
	edge := &knowledge.Edge{FromNodeID: from, ToNodeID: to, Type: edgeType, Explanation: explanation}
	if err := k.store.CreateEdge(k.ctx, edge); err != nil {
		return "", err
	}
	return edge.ID, nil
}

// DeleteNode removes a node.
func (k *KnowledgeService) DeleteNode(id string) error {
	if k.store == nil {
		return fmt.Errorf("knowledge store not initialized")
	}
	return k.store.DeleteNode(k.ctx, id)
}

// HasEmbedder reports whether an embedding backend is available.
func (k *KnowledgeService) HasEmbedder() bool {
	return k.emb != nil
}

// --- cache helpers ---

func (k *KnowledgeService) getCache(key string) any {
	k.cacheMu.Lock()
	defer k.cacheMu.Unlock()
	entry, ok := k.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil
	}
	return entry.data
}

func (k *KnowledgeService) setCache(key string, data any, ttl time.Duration) {
	k.cacheMu.Lock()
	defer k.cacheMu.Unlock()
	k.cache[key] = cacheEntry{data: data, expiresAt: time.Now().Add(ttl)}
}
