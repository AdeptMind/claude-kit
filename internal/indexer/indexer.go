package indexer

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/knowledge"
)

//go:embed pipeline.py.tmpl
var pipelineTemplate string

// Config holds indexing configuration.
type Config struct {
	Source     string // local path or s3:// URI
	Full       bool   // force full reindex
	PipelineDir string // directory for generated pipeline file
	IndexDBPath string // path to CocoIndex's output SQLite
}

// DefaultConfig returns config with sensible defaults.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Source:      ".",
		PipelineDir: filepath.Join(home, ".claude-kit"),
		IndexDBPath: filepath.Join(home, ".claude-kit", "index.sqlite"),
	}
}

// EnsurePipeline writes the Python pipeline template if absent.
func EnsurePipeline(cfg Config) (string, error) {
	path := filepath.Join(cfg.PipelineDir, "cocoindex_pipeline.py")
	if _, err := os.Stat(path); err == nil {
		return path, nil // already exists
	}

	if err := os.MkdirAll(cfg.PipelineDir, 0755); err != nil {
		return "", fmt.Errorf("create pipeline dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(pipelineTemplate), 0644); err != nil {
		return "", fmt.Errorf("write pipeline: %w", err)
	}
	return path, nil
}

// RunCocoIndex invokes cocoindex update on the pipeline file.
func RunCocoIndex(cfg Config, pipelinePath string) error {
	args := []string{"update", pipelinePath}
	if cfg.Full {
		args = append(args, "--full-reprocess")
	}

	cmd := exec.Command("cocoindex", args...)
	cmd.Env = append(os.Environ(),
		"CK_INDEX_SOURCE="+cfg.Source,
		"CK_INDEX_DB="+cfg.IndexDBPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cocoindex update: %w", err)
	}
	return nil
}

// chunk represents a row from CocoIndex's output SQLite.
type chunk struct {
	filename   string
	sourceType string
	chunkText  string
	chunkStart int
	chunkEnd   int
}

// ImportToKG reads chunks from CocoIndex's SQLite and imports them as KG nodes.
// Each chunk is embedded with Model2Vec and stored in the knowledge graph.
func ImportToKG(ctx context.Context, cfg Config, store *knowledge.Store, emb embedder.Embedder) (int, error) {
	indexDB, err := sql.Open("sqlite", cfg.IndexDBPath+"?mode=ro")
	if err != nil {
		return 0, fmt.Errorf("open index db: %w", err)
	}
	defer indexDB.Close()

	rows, err := indexDB.QueryContext(ctx,
		`SELECT filename, source_type, chunk_text, chunk_start, chunk_end FROM chunks`)
	if err != nil {
		return 0, fmt.Errorf("query chunks: %w", err)
	}
	defer rows.Close()

	var chunks []chunk
	for rows.Next() {
		var c chunk
		if err := rows.Scan(&c.filename, &c.sourceType, &c.chunkText, &c.chunkStart, &c.chunkEnd); err != nil {
			continue
		}
		chunks = append(chunks, c)
	}

	imported := 0
	for _, c := range chunks {
		title := fmt.Sprintf("%s [%d:%d]", c.filename, c.chunkStart, c.chunkEnd)

		node := &knowledge.Node{
			Title:   title,
			Content: c.chunkText,
			Type:    "chunk",
		}
		if err := store.CreateNode(ctx, node); err != nil {
			continue
		}

		// Embed and store vector
		if emb != nil {
			if vec, err := emb.Embed(ctx, c.chunkText); err == nil {
				store.UpdateNodeEmbedding(ctx, node.ID, vec)
			}
		}

		imported++
	}

	return imported, nil
}
