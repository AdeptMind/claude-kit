// This file provides a no-op ONNX detector when the onnx build tag is not set.
// Build with -tags onnx to enable the real ONNX Runtime backend.

//go:build !onnx

package embedder

import (
	"context"
	"log"
)

// DetectOnnx returns nil when built without the onnx tag.
func DetectOnnx(ctx context.Context) Embedder {
	log.Printf("embedder: ONNX backend not compiled (build with -tags onnx)")
	return nil
}
