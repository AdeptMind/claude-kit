package embedder

import (
	"context"
	"log"
	"os"
	"path/filepath"
)

// Embedder generates vector embeddings from text.
type Embedder interface {
	// Embed returns a float32 vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)
	// Dimensions returns the number of dimensions in the output vector.
	Dimensions() int
	// Name returns a human-readable name for this backend.
	Name() string
}

// DefaultModelDir returns the default path for the potion-code-16M model.
func DefaultModelDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude-kit", "models", "potion-code-16M")
}

// AutoDetect tries to load the local Model2Vec model.
// Returns nil if the model is not installed.
func AutoDetect(ctx context.Context) Embedder {
	modelDir := DefaultModelDir()
	if modelDir == "" {
		log.Printf("embedder: cannot determine home directory")
		return nil
	}

	e, err := DetectModel2Vec(modelDir)
	if err != nil {
		log.Printf("embedder: Model2Vec not available — run: ck model download (%v)", err)
		return nil
	}

	log.Printf("embedder: using Model2Vec backend (dims=%d)", e.Dimensions())
	return e
}

// DetectModel2Vec tries to load a Model2Vec model from the given directory.
func DetectModel2Vec(modelDir string) (*Model2VecEmbedder, error) {
	// Check that required files exist
	for _, f := range []string{"model.safetensors", "tokenizer.json"} {
		if _, err := os.Stat(filepath.Join(modelDir, f)); err != nil {
			return nil, err
		}
	}
	return NewModel2Vec(modelDir)
}
