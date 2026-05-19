package embedder

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// writeSafetensors creates a minimal safetensors file for testing.
// Format: 8-byte LE header size + JSON header + raw tensor data.
func writeSafetensors(t *testing.T, dir string, tensors map[string][]float32) string {
	t.Helper()

	header := make(map[string]interface{})
	offset := 0
	tensorData := make([]byte, 0)

	for name, data := range tensors {
		raw := make([]byte, len(data)*4)
		for i, v := range data {
			binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(v))
		}
		header[name] = map[string]interface{}{
			"dtype":        "F32",
			"shape":        []int{len(data) / 4, 4}, // rows x cols for 4-dim embeddings
			"data_offsets": []int{offset, offset + len(raw)},
		}
		tensorData = append(tensorData, raw...)
		offset += len(raw)
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "model.safetensors")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Write 8-byte LE header size
	sizeBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBuf, uint64(len(headerJSON)))
	f.Write(sizeBuf)
	f.Write(headerJSON)
	f.Write(tensorData)

	return path
}

func TestLoadSafetensors_HeaderSizeOverflowRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evil.safetensors")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// Advertise a header that's larger than maxSafetensorsHeaderBytes (100 MB).
	// Writing 8 bytes of size + a few bytes of garbage is enough — we must reject
	// before attempting the allocation.
	sizeBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBuf, 1<<40) // 1 TB advertised header
	f.Write(sizeBuf)
	f.Write([]byte("garbage"))
	f.Close()

	_, err = LoadSafetensors(path)
	if err == nil {
		t.Fatal("expected error for oversized header, got nil — allocation would have OOMed the process")
	}
}

func TestLoadSafetensors_ZeroHeaderRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.safetensors")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// 8 bytes of zeros = headerSize 0
	f.Write(make([]byte, 8))
	f.Close()

	if _, err := LoadSafetensors(path); err == nil {
		t.Fatal("expected error for zero header size, got nil")
	}
}

func TestLoadSafetensors_SingleTensor(t *testing.T) {
	dir := t.TempDir()

	// 3 rows x 4 cols embedding matrix
	data := []float32{
		1.0, 2.0, 3.0, 4.0,
		5.0, 6.0, 7.0, 8.0,
		9.0, 10.0, 11.0, 12.0,
	}

	path := writeSafetensors(t, dir, map[string][]float32{
		"embeddings": data,
	})

	sf, err := LoadSafetensors(path)
	if err != nil {
		t.Fatalf("LoadSafetensors: %v", err)
	}

	emb, ok := sf.Tensors["embeddings"]
	if !ok {
		t.Fatal("tensor 'embeddings' not found")
	}

	if emb.Rows != 3 || emb.Cols != 4 {
		t.Fatalf("shape: got %dx%d, want 3x4", emb.Rows, emb.Cols)
	}

	for i, want := range data {
		if emb.Data[i] != want {
			t.Fatalf("data[%d]: got %f, want %f", i, emb.Data[i], want)
		}
	}
}

func TestLoadSafetensors_MultipleTensors(t *testing.T) {
	dir := t.TempDir()

	embeddings := []float32{1.0, 2.0, 3.0, 4.0}
	weights := []float32{0.5, 0.3, 0.8, 0.1}

	// Build manually to control order and shapes
	header := map[string]interface{}{
		"embeddings": map[string]interface{}{
			"dtype":        "F32",
			"shape":        []int{1, 4},
			"data_offsets": []int{0, 16},
		},
		"weights": map[string]interface{}{
			"dtype":        "F32",
			"shape":        []int{4},
			"data_offsets": []int{16, 32},
		},
	}
	headerJSON, _ := json.Marshal(header)

	raw := make([]byte, 32)
	for i, v := range embeddings {
		binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(v))
	}
	for i, v := range weights {
		binary.LittleEndian.PutUint32(raw[16+i*4:], math.Float32bits(v))
	}

	path := filepath.Join(dir, "model.safetensors")
	f, _ := os.Create(path)
	sizeBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBuf, uint64(len(headerJSON)))
	f.Write(sizeBuf)
	f.Write(headerJSON)
	f.Write(raw)
	f.Close()

	sf, err := LoadSafetensors(path)
	if err != nil {
		t.Fatalf("LoadSafetensors: %v", err)
	}

	if len(sf.Tensors) != 2 {
		t.Fatalf("got %d tensors, want 2", len(sf.Tensors))
	}

	w, ok := sf.Tensors["weights"]
	if !ok {
		t.Fatal("tensor 'weights' not found")
	}
	// 1D tensor: shape [4] → Rows=4, Cols=1
	if w.Rows != 4 || w.Cols != 1 {
		t.Fatalf("weights shape: got %dx%d, want 4x1", w.Rows, w.Cols)
	}
}

func TestLoadSafetensors_FileNotFound(t *testing.T) {
	_, err := LoadSafetensors("/nonexistent/model.safetensors")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
