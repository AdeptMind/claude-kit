package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/knowledge"
)

var knowledgeCmd = &cobra.Command{
	Use:   "knowledge",
	Short: "Manage the knowledge graph — search, add, show, link, stats, reindex",
	Aliases: []string{"kg"},
}

// Shared state for subcommands
func openKnowledgeStore() (*knowledge.Store, error) {
	dbPath := filepath.Join(config.DataDir(), "knowledge.sqlite")
	return knowledge.Open(dbPath)
}

// --- search ---

var searchSemantic bool
var searchLimit int
var searchDimension string

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search the knowledge graph",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openKnowledgeStore()
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		query := strings.Join(args, " ")

		var results []knowledge.SearchResult

		if searchSemantic {
			emb := embedder.AutoDetect(ctx)
			if emb == nil {
				return fmt.Errorf("no embedding backend available — install Ollama and run: ollama pull all-minilm")
			}
			vec, err := emb.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("embedding query: %w", err)
			}
			results, err = store.SearchHybrid(ctx, query, vec, searchLimit)
		} else {
			results, err = store.SearchKeyword(ctx, query, searchLimit)
		}
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			if !searchSemantic {
				fmt.Println("Tip: try --semantic for AI-powered search")
			}
			return nil
		}

		printSearchResults(results)
		return nil
	},
}

func printSearchResults(results []knowledge.SearchResult) {
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Printf("%s  %s  %s  %s\n",
		header.Render(fmt.Sprintf("%-26s", "ID")),
		header.Render(fmt.Sprintf("%-40s", "Title")),
		header.Render(fmt.Sprintf("%-8s", "Score")),
		header.Render(fmt.Sprintf("%-8s", "Type")),
	)
	fmt.Println(strings.Repeat("─", 90))

	for _, r := range results {
		title := r.Node.Title
		if len(title) > 38 {
			title = title[:38] + "…"
		}
		fmt.Printf("%-26s  %-40s  %-8.4f  %s\n",
			dim.Render(r.Node.ID),
			title,
			r.Score,
			r.MatchType,
		)
	}
}

// --- add ---

var kgAddType string

var kgAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a knowledge node",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openKnowledgeStore()
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		title := strings.Join(args, " ")

		// Read content from stdin if piped
		var content string
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			buf, _ := os.ReadFile("/dev/stdin")
			content = strings.TrimSpace(string(buf))
		}

		if kgAddType == "" {
			kgAddType = "note"
		}

		node := &knowledge.Node{Title: title, Content: content, Type: kgAddType}
		if err := store.CreateNode(ctx, node); err != nil {
			return err
		}

		// Auto-embed
		emb := embedder.AutoDetect(ctx)
		if emb != nil && (node.Content != "" || node.Title != "") {
			if vec, err := emb.Embed(ctx, node.Title+" "+node.Content); err == nil {
				store.UpdateNodeEmbedding(ctx, node.ID, vec)
			}
		}

		fmt.Printf("Created node %s\n", node.ID)
		return nil
	},
}

// --- show ---

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a knowledge node with its connections",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openKnowledgeStore()
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		node, err := store.GetNode(ctx, args[0])
		if err != nil {
			return err
		}

		bold := lipgloss.NewStyle().Bold(true)
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		fmt.Println(bold.Render(node.Title))
		fmt.Printf("%s  %s  %s\n",
			dim.Render("Type: "+node.Type),
			dim.Render("ID: "+node.ID),
			dim.Render("Created: "+node.CreatedAt.Format("2006-01-02")),
		)
		if node.Content != "" {
			fmt.Println()
			fmt.Println(node.Content)
		}

		edges, _ := store.GetEdges(ctx, node.ID)
		if len(edges) > 0 {
			fmt.Println()
			fmt.Println(bold.Render("Connections"))
			for _, e := range edges {
				arrow := "→"
				otherID := e.ToNodeID
				if e.ToNodeID == node.ID {
					arrow = "←"
					otherID = e.FromNodeID
				}
				other, _ := store.GetNode(ctx, otherID)
				otherTitle := otherID
				if other != nil {
					otherTitle = other.Title
				}
				fmt.Printf("  %s %s %s", arrow, dim.Render("["+e.Type+"]"), otherTitle)
				if e.Explanation != "" {
					fmt.Printf(" — %s", e.Explanation)
				}
				fmt.Println()
			}
		}

		return nil
	},
}

// --- link ---

var linkType string
var linkExplain string

var linkCmd = &cobra.Command{
	Use:   "link <from-id> <to-id>",
	Short: "Create a relationship between two nodes",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openKnowledgeStore()
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		edge := &knowledge.Edge{
			FromNodeID:  args[0],
			ToNodeID:    args[1],
			Type:        linkType,
			Explanation: linkExplain,
		}
		if err := store.CreateEdge(ctx, edge); err != nil {
			return err
		}
		fmt.Printf("Created edge %s\n", edge.ID)
		return nil
	},
}

// --- stats ---

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show knowledge graph statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openKnowledgeStore()
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		stats, err := store.GetStats(ctx)
		if err != nil {
			return err
		}

		bold := lipgloss.NewStyle().Bold(true)
		fmt.Println(bold.Render("Knowledge Graph Stats"))
		fmt.Printf("  Nodes:      %d\n", stats.Nodes)
		fmt.Printf("  Edges:      %d\n", stats.Edges)
		fmt.Printf("  Dimensions: %d\n", stats.Dimensions)

		coverage := 0
		if stats.Nodes > 0 {
			coverage = stats.NodesWithEmbedding * 100 / stats.Nodes
		}
		fmt.Printf("  Embeddings: %d/%d (%d%%)\n", stats.NodesWithEmbedding, stats.Nodes, coverage)
		return nil
	},
}

// --- reindex ---

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Re-generate embeddings for nodes missing them",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openKnowledgeStore()
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		emb := embedder.AutoDetect(ctx)
		if emb == nil {
			return fmt.Errorf("no embedding backend available — install Ollama and run: ollama pull all-minilm")
		}

		// Get nodes without embeddings
		rows, err := store.DB().QueryContext(ctx,
			`SELECT id, title, content FROM nodes WHERE embedding IS NULL`)
		if err != nil {
			return err
		}
		defer rows.Close()

		type nodeInfo struct{ id, title, content string }
		var nodes []nodeInfo
		for rows.Next() {
			var n nodeInfo
			if err := rows.Scan(&n.id, &n.title, &n.content); err != nil {
				continue
			}
			nodes = append(nodes, n)
		}

		if len(nodes) == 0 {
			fmt.Println("All nodes already have embeddings.")
			return nil
		}

		fmt.Printf("Reindexing %d nodes...\n", len(nodes))
		for i, n := range nodes {
			vec, err := emb.Embed(ctx, n.title+" "+n.content)
			if err != nil {
				fmt.Printf("  [%d/%d] %s — error: %v\n", i+1, len(nodes), n.id, err)
				continue
			}
			store.UpdateNodeEmbedding(ctx, n.id, vec)
			fmt.Printf("  [%d/%d] %s ✓\n", i+1, len(nodes), n.id)
		}
		fmt.Println("Done.")
		return nil
	},
}

// --- dimensions ---

var dimensionsCmd = &cobra.Command{
	Use:   "dimensions",
	Short: "List all dimensions",
	Aliases: []string{"dims"},
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openKnowledgeStore()
		if err != nil {
			return err
		}
		defer store.Close()

		ctx := context.Background()
		dims, err := store.ListDimensions(ctx)
		if err != nil {
			return err
		}

		if len(dims) == 0 {
			fmt.Println("No dimensions defined.")
			return nil
		}

		bold := lipgloss.NewStyle().Bold(true)
		fmt.Println(bold.Render("Dimensions"))
		for _, d := range dims {
			fmt.Printf("  %-20s  %d nodes  %s\n", d.Name, d.NodeCount, d.ID)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchSemantic, "semantic", false, "Use AI-powered semantic search")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum number of results")
	searchCmd.Flags().StringVar(&searchDimension, "dimension", "", "Filter by dimension name")

	kgAddCmd.Flags().StringVar(&kgAddType, "type", "note", "Node type (note, decision, pattern, concept, bug, reference)")

	linkCmd.Flags().StringVar(&linkType, "type", "related", "Edge type (related, motivated_by, depends_on, supersedes, implements, contradicts)")
	linkCmd.Flags().StringVar(&linkExplain, "explain", "", "Explanation for the relationship")

	knowledgeCmd.AddCommand(searchCmd, kgAddCmd, showCmd, linkCmd, statsCmd, reindexCmd, dimensionsCmd)
	rootCmd.AddCommand(knowledgeCmd)
}
