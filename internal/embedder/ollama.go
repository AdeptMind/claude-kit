package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultOllamaModel = "all-minilm"
	ollamaTimeout      = 5 * time.Second
)

// OllamaEmbedder generates embeddings via a local Ollama instance.
type OllamaEmbedder struct {
	url    string
	model  string
	dims   int
	client *http.Client
}

type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// DetectOllama checks if Ollama is running and the model is available.
// Returns nil if Ollama is not reachable or the model is missing.
func DetectOllama(ctx context.Context) *OllamaEmbedder {
	url := defaultOllamaURL
	model := defaultOllamaModel
	client := &http.Client{Timeout: ollamaTimeout}

	// Ping Ollama
	req, _ := http.NewRequestWithContext(ctx, "GET", url+"/api/tags", nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	resp.Body.Close()

	// Try a test embedding to verify model is available and get dimensions
	e := &OllamaEmbedder{url: url, model: model, client: client}
	vec, err := e.Embed(ctx, "test")
	if err != nil {
		log.Printf("embedder: Ollama running but model %q not available — run: ollama pull %s", model, model)
		return nil
	}
	e.dims = len(vec)
	return e
}

func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	body, _ := json.Marshal(ollamaEmbedRequest{Model: e.model, Prompt: text})

	req, err := http.NewRequestWithContext(ctx, "POST", e.url+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(b))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding ollama response: %w", err)
	}

	// Convert float64 → float32
	vec := make([]float32, len(result.Embedding))
	for i, v := range result.Embedding {
		vec[i] = float32(v)
	}
	return vec, nil
}

func (e *OllamaEmbedder) Dimensions() int { return e.dims }
func (e *OllamaEmbedder) Name() string    { return "ollama/" + e.model }
