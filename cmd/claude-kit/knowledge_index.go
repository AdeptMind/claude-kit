package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/indexer"
)

var indexFull bool

var indexCmd = &cobra.Command{
	Use:   "index [source]",
	Short: "Index sources into the knowledge graph via CocoIndex",
	Long: `Incrementally index local directories or S3 buckets into the knowledge graph.

Uses CocoIndex for incremental source processing (only changed files are re-indexed),
then embeds chunks with Model2Vec and stores them in the knowledge graph.

Examples:
  ck knowledge index .                    # index current directory
  ck knowledge index ./src                # index specific directory
  ck knowledge index s3://my-bucket/      # index S3 bucket
  ck knowledge index . --full             # force full re-index`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runIndex(args)
	},
}

func init() {
	indexCmd.Flags().BoolVar(&indexFull, "full", false, "Force full re-index (ignore incremental state)")
	knowledgeCmd.AddCommand(indexCmd)
}

func runIndex(args []string) error {
	// Check cocoindex is installed
	if _, err := exec.LookPath("cocoindex"); err != nil {
		fmt.Println(errorStyle.Render("  CocoIndex not found."))
		fmt.Println(dimStyle.Render("  Run 'ck dep install' and select cocoindex, or: pip3 install -U 'cocoindex[sqlite]'"))
		return nil
	}

	cfg := indexer.DefaultConfig()
	if len(args) > 0 {
		cfg.Source = args[0]
	}
	cfg.Full = indexFull

	ctx := context.Background()

	// 1. Ensure pipeline template exists
	fmt.Println(sectionHeader("Preparing pipeline"))
	pipelinePath, err := indexer.EnsurePipeline(cfg)
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("  %s Pipeline: %s", checkMark, dimStyle.Render(pipelinePath)))

	// 2. Run CocoIndex
	fmt.Println(sectionHeader("Indexing sources"))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Source: %s", cfg.Source)))
	if cfg.Full {
		fmt.Println(warnStyle.Render("  Mode: full re-index"))
	} else {
		fmt.Println(dimStyle.Render("  Mode: incremental"))
	}
	fmt.Println()

	if err := indexer.RunCocoIndex(cfg, pipelinePath); err != nil {
		return err
	}

	// 3. Import chunks into KG
	fmt.Println()
	fmt.Println(sectionHeader("Importing to knowledge graph"))

	store, err := openKnowledgeStore()
	if err != nil {
		return err
	}
	defer store.Close()

	emb := embedder.AutoDetect(ctx)
	if emb == nil {
		fmt.Println(warnStyle.Render("  Model2Vec not available — importing without embeddings."))
		fmt.Println(dimStyle.Render("  Run 'ck model download' then 'ck knowledge reindex' to add embeddings later."))
	} else {
		fmt.Println(fmt.Sprintf("  %s Embedder: %s", checkMark, dimStyle.Render(emb.Name())))
	}

	imported, err := indexer.ImportToKG(ctx, cfg, store, emb)
	if err != nil {
		return err
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("  %s %d chunks imported into knowledge graph", arrow, imported)))
	return nil
}
