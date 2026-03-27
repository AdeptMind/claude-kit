package embedder

import (
	"context"
	"log"
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

// AutoDetect tries embedding backends in order: Ollama → ONNX → nil.
// Returns the first available backend, or nil if none are available.
func AutoDetect(ctx context.Context) Embedder {
	if e := DetectOllama(ctx); e != nil {
		log.Printf("embedder: using Ollama backend (model=%s, dims=%d)", e.model, e.dims)
		return e
	}
	if e := DetectOnnx(ctx); e != nil {
		log.Printf("embedder: using ONNX Runtime backend")
		return e
	}
	log.Printf("embedder: no embedding backend available — semantic search disabled")
	return nil
}
