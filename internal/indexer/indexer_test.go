package indexer

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestImportToKG_MissingDBReturnsError(t *testing.T) {
	cfg := Config{IndexDBPath: filepath.Join(t.TempDir(), "does-not-exist.sqlite")}

	_, err := ImportToKG(context.Background(), cfg, nil, nil)
	if err == nil {
		t.Fatal("expected error for missing index db, got nil")
	}
	if !strings.Contains(err.Error(), "index db not found") {
		t.Errorf("expected 'index db not found' message, got: %v", err)
	}
}

func TestImportToKG_SchemaMismatchReturnsError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "bad-schema.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	// Create a chunks table missing the expected columns
	if _, err := db.Exec(`CREATE TABLE chunks (wrong_column TEXT)`); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	db.Close()

	cfg := Config{IndexDBPath: dbPath}
	_, err = ImportToKG(context.Background(), cfg, nil, nil)
	if err == nil {
		t.Fatal("expected error for schema mismatch, got nil")
	}
}

func TestEnsurePipelineWritesTemplate(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{PipelineDir: dir}

	path, err := EnsurePipeline(cfg)
	if err != nil {
		t.Fatalf("EnsurePipeline: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("expected pipeline in %s, got %s", dir, path)
	}

	// Second call must be idempotent (file already exists, no error)
	if _, err := EnsurePipeline(cfg); err != nil {
		t.Errorf("second EnsurePipeline call failed: %v", err)
	}
}
