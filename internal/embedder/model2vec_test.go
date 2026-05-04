package embedder

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// buildTestModel creates a minimal Model2Vec model directory for testing.
// Returns the directory path containing model.safetensors, tokenizer.json, config.json.
func buildTestModel(t *testing.T, vocabSize, dims int) string {
	t.Helper()
	dir := t.TempDir()

	// Build vocabulary: [UNK], hello, world, ##lo, he
	vocab := map[string]int{
		"[UNK]": 0,
		"hello": 1,
		"world": 2,
		"##lo":  3,
		"he":    4,
	}
	if vocabSize > len(vocab) {
		for i := len(vocab); i < vocabSize; i++ {
			vocab[fmt.Sprintf("tok%d", i)] = i
		}
	}

	// Write tokenizer.json
	tokJSON := map[string]interface{}{
		"model": map[string]interface{}{
			"type":                       "WordPiece",
			"unk_token":                  "[UNK]",
			"continuing_subword_prefix":  "##",
			"vocab":                      vocab,
		},
	}
	tokData, _ := json.Marshal(tokJSON)
	os.WriteFile(filepath.Join(dir, "tokenizer.json"), tokData, 0644)

	// Build embedding matrix: vocabSize x dims
	embData := make([]float32, vocabSize*dims)
	for i := range embData {
		embData[i] = float32(i) * 0.01
	}

	// Write model.safetensors
	header := map[string]interface{}{
		"embeddings": map[string]interface{}{
			"dtype":        "F32",
			"shape":        []int{vocabSize, dims},
			"data_offsets": []int{0, vocabSize * dims * 4},
		},
	}
	headerJSON, _ := json.Marshal(header)

	raw := make([]byte, vocabSize*dims*4)
	for i, v := range embData {
		binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(v))
	}

	f, _ := os.Create(filepath.Join(dir, "model.safetensors"))
	sizeBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBuf, uint64(len(headerJSON)))
	f.Write(sizeBuf)
	f.Write(headerJSON)
	f.Write(raw)
	f.Close()

	// Write config.json
	cfg := map[string]interface{}{
		"normalize":      true,
		"embedding_dtype": "float32",
	}
	cfgData, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(dir, "config.json"), cfgData, 0644)

	return dir
}

func TestModel2VecEmbedder_Load(t *testing.T) {
	dir := buildTestModel(t, 5, 8)

	emb, err := NewModel2Vec(dir)
	if err != nil {
		t.Fatalf("NewModel2Vec: %v", err)
	}

	if emb.Dimensions() != 8 {
		t.Fatalf("dims: got %d, want 8", emb.Dimensions())
	}

	if emb.Name() != "model2vec/potion-code-16M" {
		t.Fatalf("name: got %q", emb.Name())
	}
}

func TestModel2VecEmbedder_Embed(t *testing.T) {
	dir := buildTestModel(t, 5, 4)

	emb, err := NewModel2Vec(dir)
	if err != nil {
		t.Fatalf("NewModel2Vec: %v", err)
	}

	vec, err := emb.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}

	if len(vec) != 4 {
		t.Fatalf("vector length: got %d, want 4", len(vec))
	}

	// Verify the output is L2-normalized (magnitude ≈ 1.0)
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 0.001 {
		t.Fatalf("L2 norm: got %f, want ~1.0", norm)
	}
}

func TestModel2VecEmbedder_EmbedUnknownToken(t *testing.T) {
	dir := buildTestModel(t, 5, 4)

	emb, err := NewModel2Vec(dir)
	if err != nil {
		t.Fatalf("NewModel2Vec: %v", err)
	}

	vec, err := emb.Embed(context.Background(), "xyzxyz")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}

	// Should still return a valid vector (from [UNK] embedding)
	if len(vec) != 4 {
		t.Fatalf("vector length: got %d, want 4", len(vec))
	}
}

func TestModel2VecEmbedder_EmbedEmpty(t *testing.T) {
	dir := buildTestModel(t, 5, 4)

	emb, err := NewModel2Vec(dir)
	if err != nil {
		t.Fatalf("NewModel2Vec: %v", err)
	}

	_, err = emb.Embed(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestModel2VecEmbedder_DeterministicOutput(t *testing.T) {
	dir := buildTestModel(t, 5, 4)

	emb, err := NewModel2Vec(dir)
	if err != nil {
		t.Fatalf("NewModel2Vec: %v", err)
	}

	ctx := context.Background()
	v1, _ := emb.Embed(ctx, "hello")
	v2, _ := emb.Embed(ctx, "hello")

	for i := range v1 {
		if v1[i] != v2[i] {
			t.Fatalf("non-deterministic: v1[%d]=%f != v2[%d]=%f", i, v1[i], i, v2[i])
		}
	}
}
