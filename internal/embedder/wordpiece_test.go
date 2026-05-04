package embedder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeTokenizerJSON creates a minimal HuggingFace tokenizer.json for testing.
func writeTokenizerJSON(t *testing.T, dir string, vocab map[string]int) string {
	t.Helper()

	tokJSON := map[string]interface{}{
		"model": map[string]interface{}{
			"type":             "WordPiece",
			"unk_token":        "[UNK]",
			"continuing_subword_prefix": "##",
			"vocab":            vocab,
		},
	}

	data, err := json.Marshal(tokJSON)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "tokenizer.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestWordPieceTokenize_BasicWords(t *testing.T) {
	dir := t.TempDir()
	vocab := map[string]int{
		"[UNK]": 0,
		"hello": 1,
		"world": 2,
		"he":    3,
		"##llo": 4,
	}
	path := writeTokenizerJSON(t, dir, vocab)

	wp, err := LoadWordPiece(path)
	if err != nil {
		t.Fatalf("LoadWordPiece: %v", err)
	}

	ids := wp.Tokenize("hello world")

	// "hello" → exact match → [1]
	// "world" → exact match → [2]
	want := []int{1, 2}
	if len(ids) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(ids), len(want), ids)
	}
	for i, w := range want {
		if ids[i] != w {
			t.Fatalf("token[%d]: got %d, want %d", i, ids[i], w)
		}
	}
}

func TestWordPieceTokenize_SubwordSplit(t *testing.T) {
	dir := t.TempDir()
	vocab := map[string]int{
		"[UNK]": 0,
		"un":    1,
		"##kno": 2,
		"##wn":  3,
	}
	path := writeTokenizerJSON(t, dir, vocab)

	wp, err := LoadWordPiece(path)
	if err != nil {
		t.Fatalf("LoadWordPiece: %v", err)
	}

	ids := wp.Tokenize("unknown")
	// "unknown" → "un" + "##kno" + "##wn" → [1, 2, 3]
	want := []int{1, 2, 3}
	if len(ids) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(ids), len(want), ids)
	}
	for i, w := range want {
		if ids[i] != w {
			t.Fatalf("token[%d]: got %d, want %d", i, ids[i], w)
		}
	}
}

func TestWordPieceTokenize_UnknownToken(t *testing.T) {
	dir := t.TempDir()
	vocab := map[string]int{
		"[UNK]": 0,
		"hello": 1,
	}
	path := writeTokenizerJSON(t, dir, vocab)

	wp, err := LoadWordPiece(path)
	if err != nil {
		t.Fatalf("LoadWordPiece: %v", err)
	}

	ids := wp.Tokenize("xyz")
	// "xyz" → not in vocab, no subword match → [UNK] → [0]
	want := []int{0}
	if len(ids) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(ids), len(want), ids)
	}
	if ids[0] != 0 {
		t.Fatalf("expected [UNK] (0), got %d", ids[0])
	}
}

func TestWordPieceTokenize_Lowercase(t *testing.T) {
	dir := t.TempDir()
	vocab := map[string]int{
		"[UNK]": 0,
		"hello": 1,
	}
	path := writeTokenizerJSON(t, dir, vocab)

	wp, err := LoadWordPiece(path)
	if err != nil {
		t.Fatalf("LoadWordPiece: %v", err)
	}

	ids := wp.Tokenize("HELLO")
	want := []int{1}
	if len(ids) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(ids), len(want), ids)
	}
	if ids[0] != 1 {
		t.Fatalf("expected 1, got %d", ids[0])
	}
}

func TestWordPieceTokenize_EmptyInput(t *testing.T) {
	dir := t.TempDir()
	vocab := map[string]int{
		"[UNK]": 0,
	}
	path := writeTokenizerJSON(t, dir, vocab)

	wp, err := LoadWordPiece(path)
	if err != nil {
		t.Fatalf("LoadWordPiece: %v", err)
	}

	ids := wp.Tokenize("")
	if len(ids) != 0 {
		t.Fatalf("expected empty, got %v", ids)
	}
}

func TestLoadWordPiece_FileNotFound(t *testing.T) {
	_, err := LoadWordPiece("/nonexistent/tokenizer.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
