package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/knowledge"
)

var mcpRegister bool

var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Run the knowledge graph MCP server (stdio transport)",
	Long: `Starts an MCP server over stdin/stdout for use with Claude Desktop, Claude Code, or any MCP client.

Add to claude_desktop_config.json:
  {
    "mcpServers": {
      "claude-kit": {
        "command": "ck",
        "args": ["mcp-server"]
      }
    }
  }

Or for Claude Code (~/.claude.json):
  {
    "mcpServers": {
      "claude-kit": {
        "command": "ck",
        "args": ["mcp-server"]
      }
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpRegister {
			return registerMCPServer()
		}
		return runMCPServer()
	},
}

func runMCPServer() error {
	// Suppress log output — stdio transport uses stdout for JSON-RPC
	log.SetOutput(os.Stderr)

	dbPath := filepath.Join(config.DataDir(), "knowledge.sqlite")
	store, err := knowledge.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening knowledge store: %w", err)
	}
	defer store.Close()

	ctx := context.Background()
	emb := embedder.AutoDetect(ctx)
	if emb != nil {
		store.SetEmbedder(emb)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "claude-kit",
		Version: version,
	}, nil)

	// Register knowledge tools (reuse the same function from ck-ui)
	registerKnowledgeToolsStdio(server, store, emb)

	log.Printf("mcp-server: starting stdio transport (db=%s)", dbPath)
	return server.Run(ctx, &mcp.StdioTransport{})
}

// registerKnowledgeToolsStdio registers all kg_* tools on a stdio MCP server.
// This duplicates the tool registration from knowledge_tools.go in cmd/claude-kit-ui/
// because that file is in a different main package. The handlers delegate to the
// same knowledge.Store methods.
func registerKnowledgeToolsStdio(server *mcp.Server, store *knowledge.Store, emb embedder.Embedder) {
	type CreateNodeArgs struct {
		Title   string `json:"title"`
		Content string `json:"content,omitempty"`
		Type    string `json:"type,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_createNode",
		Description: "Create a new knowledge node. Returns the node ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateNodeArgs) (*mcp.CallToolResult, any, error) {
		node := &knowledge.Node{Title: args.Title, Content: args.Content, Type: args.Type}
		if err := store.CreateNode(ctx, node); err != nil {
			return nil, nil, err
		}
		if emb != nil && (node.Content != "" || node.Title != "") {
			if vec, err := emb.Embed(ctx, node.Title+" "+node.Content); err == nil {
				store.UpdateNodeEmbedding(ctx, node.ID, vec)
			}
		}
		return textRes(fmt.Sprintf("Created node %s", node.ID)), nil, nil
	})

	type QueryNodesArgs struct {
		Query string `json:"query,omitempty"`
		Limit int    `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_queryNodes",
		Description: "Search knowledge nodes by keyword.",
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
			return jsonRes(results), nil, nil
		}
		nodes, err := store.ListNodes(ctx, knowledge.ListOpts{Limit: limit})
		if err != nil {
			return nil, nil, err
		}
		return jsonRes(nodes), nil, nil
	})

	type GetNodeArgs struct {
		ID string `json:"id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_getNode",
		Description: "Get a knowledge node by ID with connections.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetNodeArgs) (*mcp.CallToolResult, any, error) {
		node, err := store.GetNode(ctx, args.ID)
		if err != nil {
			return nil, nil, err
		}
		edges, _ := store.GetEdges(ctx, args.ID)
		return jsonRes(map[string]any{"node": node, "edges": edges}), nil, nil
	})

	type UpdateNodeArgs struct {
		ID      string `json:"id"`
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
		Type    string `json:"type,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_updateNode",
		Description: "Update a knowledge node.",
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
		return textRes(fmt.Sprintf("Updated node %s", node.ID)), nil, nil
	})

	type DeleteNodeArgs struct {
		ID string `json:"id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_deleteNode",
		Description: "Delete a knowledge node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args DeleteNodeArgs) (*mcp.CallToolResult, any, error) {
		if err := store.DeleteNode(ctx, args.ID); err != nil {
			return nil, nil, err
		}
		return textRes(fmt.Sprintf("Deleted node %s", args.ID)), nil, nil
	})

	type CreateEdgeArgs struct {
		From        string `json:"from"`
		To          string `json:"to"`
		Type        string `json:"type,omitempty"`
		Explanation string `json:"explanation,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_createEdge",
		Description: "Create a relationship between two nodes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateEdgeArgs) (*mcp.CallToolResult, any, error) {
		edge := &knowledge.Edge{FromNodeID: args.From, ToNodeID: args.To, Type: args.Type, Explanation: args.Explanation}
		if err := store.CreateEdge(ctx, edge); err != nil {
			return nil, nil, err
		}
		return textRes(fmt.Sprintf("Created edge %s", edge.ID)), nil, nil
	})

	type QueryEdgesArgs struct {
		NodeID string `json:"nodeId"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_queryEdges",
		Description: "Get all connections for a node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args QueryEdgesArgs) (*mcp.CallToolResult, any, error) {
		edges, err := store.GetEdges(ctx, args.NodeID)
		if err != nil {
			return nil, nil, err
		}
		return jsonRes(edges), nil, nil
	})

	type SearchSemanticArgs struct {
		Query string `json:"query"`
		Limit int    `json:"limit,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_searchSemantic",
		Description: "Search knowledge by meaning using AI embeddings.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SearchSemanticArgs) (*mcp.CallToolResult, any, error) {
		if emb == nil {
			return textRes("Semantic search unavailable — no embedding backend"), nil, nil
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
		return jsonRes(results), nil, nil
	})

	type EmptyArgs struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_getContext",
		Description: "Get knowledge graph overview — stats and recent nodes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyArgs) (*mcp.CallToolResult, any, error) {
		stats, _ := store.GetStats(ctx)
		recent, _ := store.ListNodes(ctx, knowledge.ListOpts{Limit: 5})
		return jsonRes(map[string]any{"stats": stats, "recentNodes": recent}), nil, nil
	})

	// Dimension tools
	type CreateDimensionArgs struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_createDimension",
		Description: "Create a dimension (tag/category) for organizing nodes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateDimensionArgs) (*mcp.CallToolResult, any, error) {
		dim := &knowledge.Dimension{Name: args.Name, Description: args.Description}
		if err := store.CreateDimension(ctx, dim); err != nil {
			return nil, nil, err
		}
		return textRes(fmt.Sprintf("Created dimension %s (%s)", dim.ID, dim.Name)), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kg_queryDimensions",
		Description: "List all dimensions with node counts.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyArgs) (*mcp.CallToolResult, any, error) {
		dims, err := store.ListDimensions(ctx)
		if err != nil {
			return nil, nil, err
		}
		return jsonRes(dims), nil, nil
	})
}

func textRes(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

func jsonRes(v any) *mcp.CallToolResult {
	b, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}
}

// registerMCPServer writes the stdio config to claude_desktop_config.json
func registerMCPServer() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")

	cfg := map[string]any{}
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	servers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
	}

	// Find ck binary path
	exe, _ := os.Executable()
	servers["claude-kit"] = map[string]any{
		"command": exe,
		"args":    []string{"mcp-server"},
	}
	cfg["mcpServers"] = servers

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	out, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(configPath, append(out, '\n'), 0o644); err != nil {
		return err
	}

	fmt.Printf("Registered claude-kit MCP server in %s\n", configPath)
	fmt.Printf("Command: %s mcp-server\n", exe)
	fmt.Println("Restart Claude Desktop to activate.")
	return nil
}

func init() {
	mcpServerCmd.Flags().BoolVar(&mcpRegister, "register", false, "Register in Claude Desktop config instead of running")
	rootCmd.AddCommand(mcpServerCmd)
}
