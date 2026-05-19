package embedder

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
)

// maxSafetensorsHeaderBytes caps the safetensors header JSON to prevent
// adversarial or corrupt files from triggering huge allocations. 100 MB is
// far above any legitimate header size (real models are well under 1 MB).
const maxSafetensorsHeaderBytes = 100 << 20

// Tensor holds a dense float32 matrix loaded from a safetensors file.
type Tensor struct {
	Data []float32
	Rows int
	Cols int
}

// SafetensorsFile holds all tensors parsed from a safetensors file.
type SafetensorsFile struct {
	Tensors map[string]Tensor
}

type safetensorsHeader struct {
	Dtype       string `json:"dtype"`
	Shape       []int  `json:"shape"`
	DataOffsets [2]int `json:"data_offsets"`
}

// LoadSafetensors reads a safetensors file and returns all tensors.
// Format: 8-byte LE header size + JSON header + raw tensor data.
func LoadSafetensors(path string) (*SafetensorsFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open safetensors: %w", err)
	}
	defer f.Close()

	// Read 8-byte header size (little-endian uint64)
	var headerSize uint64
	if err := binary.Read(f, binary.LittleEndian, &headerSize); err != nil {
		return nil, fmt.Errorf("read header size: %w", err)
	}
	if headerSize == 0 || headerSize > maxSafetensorsHeaderBytes {
		return nil, fmt.Errorf("safetensors header size %d out of range (max %d) — file is corrupt or not safetensors format", headerSize, maxSafetensorsHeaderBytes)
	}

	// Read JSON header
	headerBuf := make([]byte, headerSize)
	if _, err := f.Read(headerBuf); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	// Parse header — may contain __metadata__ key which is not a tensor
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(headerBuf, &raw); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	dataOffset := int64(8 + headerSize) // absolute offset where tensor data begins

	result := &SafetensorsFile{Tensors: make(map[string]Tensor)}

	for name, rawEntry := range raw {
		if name == "__metadata__" {
			continue
		}

		var entry safetensorsHeader
		if err := json.Unmarshal(rawEntry, &entry); err != nil {
			return nil, fmt.Errorf("parse tensor %q: %w", name, err)
		}

		if entry.Dtype != "F32" {
			return nil, fmt.Errorf("tensor %q: unsupported dtype %q (only F32)", name, entry.Dtype)
		}

		byteLen := entry.DataOffsets[1] - entry.DataOffsets[0]
		numFloats := byteLen / 4

		// Read raw float32 data
		buf := make([]byte, byteLen)
		if _, err := f.ReadAt(buf, dataOffset+int64(entry.DataOffsets[0])); err != nil {
			return nil, fmt.Errorf("read tensor %q data: %w", name, err)
		}

		data := make([]float32, numFloats)
		for i := range data {
			data[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf[i*4:]))
		}

		// Derive rows and cols from shape
		rows, cols := shapeToRowsCols(entry.Shape)

		result.Tensors[name] = Tensor{Data: data, Rows: rows, Cols: cols}
	}

	return result, nil
}

// shapeToRowsCols converts a tensor shape to (rows, cols).
// [N, M] → (N, M), [N] → (N, 1), [] → (1, 1).
func shapeToRowsCols(shape []int) (int, int) {
	switch len(shape) {
	case 0:
		return 1, 1
	case 1:
		return shape[0], 1
	default:
		return shape[0], shape[1]
	}
}
