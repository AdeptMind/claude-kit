package embedder

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Model2VecEmbedder generates embeddings using a Model2Vec static model.
// No transformer inference — just lookup + mean pooling + L2 normalize.
type Model2VecEmbedder struct {
	tokenizer  *WordPiece
	embeddings Tensor    // [vocab_size, dims]
	weights    []float32 // SIF weights per token (optional)
	normalize  bool
	dims       int
}

type model2vecConfig struct {
	Normalize bool `json:"normalize"`
}

// NewModel2Vec loads a Model2Vec model from a directory containing:
//   - model.safetensors (embedding matrix)
//   - tokenizer.json (WordPiece vocabulary)
//   - config.json (normalize flag)
func NewModel2Vec(modelDir string) (*Model2VecEmbedder, error) {
	// Load tokenizer
	tok, err := LoadWordPiece(filepath.Join(modelDir, "tokenizer.json"))
	if err != nil {
		return nil, fmt.Errorf("load tokenizer: %w", err)
	}

	// Load safetensors
	sf, err := LoadSafetensors(filepath.Join(modelDir, "model.safetensors"))
	if err != nil {
		return nil, fmt.Errorf("load model: %w", err)
	}

	emb, ok := sf.Tensors["embeddings"]
	if !ok {
		// Some models use "embedding.weight" instead
		emb, ok = sf.Tensors["embedding.weight"]
		if !ok {
			return nil, fmt.Errorf("no embeddings tensor found in model")
		}
	}

	// Load optional SIF weights
	var weights []float32
	if w, ok := sf.Tensors["weights"]; ok {
		weights = w.Data
	}

	// Load config
	normalize := true
	cfgData, err := os.ReadFile(filepath.Join(modelDir, "config.json"))
	if err == nil {
		var cfg model2vecConfig
		if json.Unmarshal(cfgData, &cfg) == nil {
			normalize = cfg.Normalize
		}
	}

	return &Model2VecEmbedder{
		tokenizer:  tok,
		embeddings: emb,
		weights:    weights,
		normalize:  normalize,
		dims:       emb.Cols,
	}, nil
}

// Embed returns a float32 vector for the given text.
// Flow: tokenize → lookup → optional SIF weighting → mean pooling → L2 normalize.
func (m *Model2VecEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	ids := m.tokenizer.Tokenize(text)
	if len(ids) == 0 {
		return nil, fmt.Errorf("empty input produces no tokens")
	}

	// Mean pooling over token embeddings
	vec := make([]float32, m.dims)
	count := 0

	for _, id := range ids {
		if id < 0 || id >= m.embeddings.Rows {
			continue
		}

		offset := id * m.dims
		// Guard against shape mismatch between Rows/Cols header and actual data slice.
		// A corrupt model file can pass the Rows check but still over-read Data.
		if offset+m.dims > len(m.embeddings.Data) {
			continue
		}

		weight := float32(1.0)
		if m.weights != nil && id < len(m.weights) {
			weight = m.weights[id]
		}

		for j := 0; j < m.dims; j++ {
			vec[j] += m.embeddings.Data[offset+j] * weight
		}
		count++
	}

	if count == 0 {
		return nil, fmt.Errorf("no valid token embeddings found")
	}

	// Mean
	for j := range vec {
		vec[j] /= float32(count)
	}

	// L2 normalize
	if m.normalize {
		var norm float64
		for _, v := range vec {
			norm += float64(v) * float64(v)
		}
		norm = math.Sqrt(norm)
		if norm > 0 {
			for j := range vec {
				vec[j] = float32(float64(vec[j]) / norm)
			}
		}
	}

	return vec, nil
}

// Dimensions returns the embedding dimensionality.
func (m *Model2VecEmbedder) Dimensions() int { return m.dims }

// Name returns a human-readable name for this backend.
func (m *Model2VecEmbedder) Name() string { return "model2vec/potion-code-16M" }
