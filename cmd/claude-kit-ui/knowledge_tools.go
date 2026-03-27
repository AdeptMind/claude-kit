package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/knowledge"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterKnowledgeTools registers all kg_* tools on the MCP server.
func RegisterKnowledgeTools(server *mcp.Server, store *knowledge.Store, emb embedder.Embedder) {
	type CreateNodeArgs struct {
		Title      string   `json:"title"`
		Content    string   `json:"content,omitempty"`
		Type       string   `json:"type,omitempty"`
		Dimensions []string `json:"dimensions,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_createNode",
		Description: "Create a new knowledge node. Returns the node ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateNodeArgs) (*mcp.CallToolResult, any, error) {
		node := &knowledge.Node{Title: args.Title, Content: args.Content, Type: args.Type}
		if err := store.CreateNode(ctx, node); err != nil {
			return nil, nil, err
		}
		// Auto-embed if embedder available
		if emb != nil && node.Content != "" {
			if vec, err := emb.Embed(ctx, node.Title+" "+node.Content); err == nil {
				node.Embedding = vec
				store.UpdateNodeEmbedding(ctx, node.ID, vec)
			}
		}
		return textResult(fmt.Sprintf("Created node %s", node.ID)), nil, nil
	})

	type QueryNodesArgs struct {
		Query     string `json:"query,omitempty"`
		Dimension string `json:"dimension,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_queryNodes",
		Description: "Search knowledge nodes by keyword. Returns matching nodes with snippets.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args QueryNodesArgs) (*mcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}
		if args.Query != "" {
			results, err := store.SearchKeyword(ctx, args.Query, limit)
			if err != nil {
				return nil, nil, err
			}
			return jsonResult(results), nil, nil
		}
		nodes, err := store.ListNodes(ctx, knowledge.ListOpts{Limit: limit, DimensionID: args.Dimension})
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(nodes), nil, nil
	})

	type GetNodeArgs struct {
		ID string `json:"id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_getNode",
		Description: "Get a knowledge node by ID, including its content and connections.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetNodeArgs) (*mcp.CallToolResult, any, error) {
		node, err := store.GetNode(ctx, args.ID)
		if err != nil {
			return nil, nil, err
		}
		edges, _ := store.GetEdges(ctx, args.ID)
		result := map[string]any{"node": node, "edges": edges}
		return jsonResult(result), nil, nil
	})

	type UpdateNodeArgs struct {
		ID      string `json:"id"`
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
		Type    string `json:"type,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_updateNode",
		Description: "Update an existing knowledge node's title, content, or type.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UpdateNodeArgs) (*mcp.CallToolResult, any, error) {
		node, err := store.GetNode(ctx, args.ID)
		if err != nil {
			return nil, nil, err
		}
		if args.Title != "" {
			node.Title = args.Title
		}
		if args.Content != "" {
			node.Content = args.Content
		}
		if args.Type != "" {
			node.Type = args.Type
		}
		if err := store.UpdateNode(ctx, node); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Updated node %s", node.ID)), nil, nil
	})

	type DeleteNodeArgs struct {
		ID string `json:"id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_deleteNode",
		Description: "Delete a knowledge node and its connections.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args DeleteNodeArgs) (*mcp.CallToolResult, any, error) {
		if err := store.DeleteNode(ctx, args.ID); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Deleted node %s", args.ID)), nil, nil
	})

	type CreateEdgeArgs struct {
		From        string `json:"from"`
		To          string `json:"to"`
		Type        string `json:"type,omitempty"`
		Explanation string `json:"explanation,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_createEdge",
		Description: "Create a relationship between two knowledge nodes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateEdgeArgs) (*mcp.CallToolResult, any, error) {
		edge := &knowledge.Edge{
			FromNodeID:  args.From,
			ToNodeID:    args.To,
			Type:        args.Type,
			Explanation: args.Explanation,
		}
		if err := store.CreateEdge(ctx, edge); err != nil {
			return nil, nil, err
		}
		return textResult(fmt.Sprintf("Created edge %s", edge.ID)), nil, nil
	})

	type QueryEdgesArgs struct {
		NodeID string `json:"nodeId"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_queryEdges",
		Description: "Get all connections for a knowledge node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args QueryEdgesArgs) (*mcp.CallToolResult, any, error) {
		edges, err := store.GetEdges(ctx, args.NodeID)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(edges), nil, nil
	})

	type SearchSemanticArgs struct {
		Query string `json:"query"`
		Limit int    `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_searchSemantic",
		Description: "Search knowledge nodes by meaning using AI embeddings. Finds related concepts even with different wording.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SearchSemanticArgs) (*mcp.CallToolResult, any, error) {
		if emb == nil {
			return textResult("Semantic search unavailable — no embedding backend configured"), nil, nil
		}
		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}
		vec, err := emb.Embed(ctx, args.Query)
		if err != nil {
			return nil, nil, fmt.Errorf("embedding query: %w", err)
		}
		results, err := store.SearchHybrid(ctx, args.Query, vec, limit)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(results), nil, nil
	})

	type EmptyArgs struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_getContext",
		Description: "Get knowledge graph overview — stats, recent nodes, and available dimensions.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyArgs) (*mcp.CallToolResult, any, error) {
		stats, _ := store.GetStats(ctx)
		recent, _ := store.ListNodes(ctx, knowledge.ListOpts{Limit: 5})
		result := map[string]any{
			"stats":       stats,
			"recentNodes": recent,
		}
		return jsonResult(result), nil, nil
	})
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func jsonResult(v any) *mcp.CallToolResult {
	b, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}
}
