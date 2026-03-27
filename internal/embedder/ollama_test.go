package embedder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaEmbedder_Embed(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/embeddings" {
			var req ollamaEmbedRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := ollamaEmbedResponse{
				Embedding: []float64{0.1, -0.2, 0.3, 0.0, 1.0},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	e := &OllamaEmbedder{
		url:    server.URL,
		model:  "test-model",
		dims:   5,
		client: server.Client(),
	}

	vec, err := e.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("Embed error: %v", err)
	}
	if len(vec) != 5 {
		t.Fatalf("vec length = %d, want 5", len(vec))
	}
	if vec[0] != 0.1 {
		t.Errorf("vec[0] = %f, want 0.1", vec[0])
	}
	if vec[1] != -0.2 {
		t.Errorf("vec[1] = %f, want -0.2", vec[1])
	}
}

func TestOllamaEmbedder_Dimensions(t *testing.T) {
	e := &OllamaEmbedder{dims: 384}
	if e.Dimensions() != 384 {
		t.Errorf("Dimensions() = %d, want 384", e.Dimensions())
	}
}

func TestOllamaEmbedder_Name(t *testing.T) {
	e := &OllamaEmbedder{model: "all-minilm"}
	if e.Name() != "ollama/all-minilm" {
		t.Errorf("Name() = %q, want %q", e.Name(), "ollama/all-minilm")
	}
}

func TestDetectOllama_NotRunning(t *testing.T) {
	// No server running — should return nil
	e := DetectOllama(context.Background())
	// This test is flaky if Ollama is actually running on the machine.
	// We can't guarantee it, so just verify the function doesn't panic.
	_ = e
}

func TestOllamaEmbedder_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("model not found"))
	}))
	defer server.Close()

	e := &OllamaEmbedder{url: server.URL, model: "bad", client: server.Client()}
	_, err := e.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestOllamaEmbedder_Timeout(t *testing.T) {
	// Server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {} // block forever
	}))
	defer server.Close()

	e := &OllamaEmbedder{
		url:    server.URL,
		model:  "test",
		client: &http.Client{Timeout: 1}, // 1 nanosecond timeout
	}
	_, err := e.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected timeout error")
	}
}
