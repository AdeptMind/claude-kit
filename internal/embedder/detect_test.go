package embedder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAutoDetect_NoBackends(t *testing.T) {
	// With no Ollama running and no ONNX (stub), should return nil
	e := AutoDetect(context.Background())
	// Can't guarantee Ollama isn't running, so just check it doesn't panic
	_ = e
}

func TestDetectOllama_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
		case "/api/embeddings":
			resp := ollamaEmbedResponse{Embedding: make([]float64, 384)}
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Manually create embedder pointing at mock
	e := &OllamaEmbedder{
		url:    server.URL,
		model:  "all-minilm",
		client: server.Client(),
	}

	vec, err := e.Embed(context.Background(), "test")
	if err != nil {
		t.Fatalf("Embed error: %v", err)
	}
	if len(vec) != 384 {
		t.Errorf("dimensions = %d, want 384", len(vec))
	}

	e.dims = len(vec)
	if e.Dimensions() != 384 {
		t.Errorf("Dimensions() = %d, want 384", e.Dimensions())
	}
}

func TestDetectOnnx_Stub(t *testing.T) {
	// Without onnx build tag, should return nil
	e := DetectOnnx(context.Background())
	if e != nil {
		t.Error("DetectOnnx should return nil without onnx build tag")
	}
}
