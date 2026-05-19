package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
)

// modelDownloadClient bounds the HuggingFace fetch so a stalled connection
// cannot hang the CLI indefinitely. 64 MB at 100 KB/s = ~11 minutes worst case.
var modelDownloadClient = &http.Client{Timeout: 15 * time.Minute}

const (
	hfBaseURL = "https://huggingface.co/minishlab/potion-code-16M/resolve/main"
)

var modelFiles = []struct {
	name string
	desc string
}{
	{"model.safetensors", "embedding matrix (64 MB)"},
	{"tokenizer.json", "WordPiece vocabulary"},
	{"config.json", "model configuration"},
}

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage embedding models",
}

var modelDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download the potion-code-16M model for semantic search",
	Long: `Download the Model2Vec potion-code-16M model from HuggingFace.

Files are saved to ~/.claude-kit/models/potion-code-16M/.
This model enables local semantic code search — CPU only, no external service required.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runModelDownload()
	},
}

func init() {
	modelCmd.AddCommand(modelDownloadCmd)
	rootCmd.AddCommand(modelCmd)
}

func runModelDownload() error {
	modelDir := embedder.DefaultModelDir()
	if modelDir == "" {
		return fmt.Errorf("cannot determine home directory")
	}

	fmt.Println(banner())
	fmt.Println(subtitleStyle.Render("  Download potion-code-16M model"))
	fmt.Println()

	// Check if already downloaded
	allExist := true
	for _, f := range modelFiles {
		if _, err := os.Stat(filepath.Join(modelDir, f.name)); err != nil {
			allExist = false
			break
		}
	}

	if allExist {
		fmt.Println(fmt.Sprintf("  %s Model already downloaded at %s", checkMark, dimStyle.Render(modelDir)))
		return nil
	}

	// Create directory
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("create model directory: %w", err)
	}

	// Download each file
	for _, f := range modelFiles {
		dest := filepath.Join(modelDir, f.name)
		if _, err := os.Stat(dest); err == nil {
			fmt.Println(fmt.Sprintf("  %s %s (already exists)", checkMark, dimStyle.Render(f.name)))
			continue
		}

		fmt.Print(fmt.Sprintf("  %s Downloading %s — %s... ", arrow, accentStyle.Render(f.name), dimStyle.Render(f.desc)))

		if err := downloadFile(hfBaseURL+"/"+f.name, dest); err != nil {
			fmt.Println(errorStyle.Render("FAILED"))
			return fmt.Errorf("download %s: %w", f.name, err)
		}

		fmt.Println(successStyle.Render("OK"))
	}

	fmt.Println()
	fmt.Println(successStyle.Render(fmt.Sprintf("  %s Model ready at %s", arrow, modelDir)))
	fmt.Println(dimStyle.Render("  Semantic search is now enabled for 'ck knowledge search --semantic'."))

	return nil
}

func downloadFile(url, dest string) error {
	resp, err := modelDownloadClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()

	return os.Rename(tmp, dest)
}
