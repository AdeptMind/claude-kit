package embedder

import (
	"context"
	"testing"
)

func TestAutoDetect_NoModel(t *testing.T) {
	// Without model files installed, should return nil (not panic)
	e := AutoDetect(context.Background())
	// Can't guarantee model isn't installed, so just check it doesn't panic
	_ = e
}

func TestDetectModel2Vec_MissingDir(t *testing.T) {
	_, err := DetectModel2Vec("/nonexistent/model/dir")
	if err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestDetectModel2Vec_ValidModel(t *testing.T) {
	dir := buildTestModel(t, 5, 8)

	e, err := DetectModel2Vec(dir)
	if err != nil {
		t.Fatalf("DetectModel2Vec: %v", err)
	}
	if e.Dimensions() != 8 {
		t.Fatalf("dims: got %d, want 8", e.Dimensions())
	}
}
