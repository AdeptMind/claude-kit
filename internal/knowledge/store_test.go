package knowledge

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpen_CreatesNewDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	// Verify tables exist
	tables := []string{"nodes", "edges", "dimensions", "node_dimensions"}
	for _, table := range tables {
		var name string
		err := store.DB().QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}

	// Verify FTS5 virtual table
	var ftsName string
	err = store.DB().QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='nodes_fts'",
	).Scan(&ftsName)
	if err != nil {
		t.Errorf("FTS5 table not found: %v", err)
	}

	// Verify WAL mode
	var journalMode string
	err = store.DB().QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("PRAGMA journal_mode error: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	// Verify foreign keys enabled
	var fk int
	err = store.DB().QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil {
		t.Fatalf("PRAGMA foreign_keys error: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
}

func TestOpen_ExistingDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")

	// First open: create
	store1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first Open() error: %v", err)
	}

	// Insert a row
	_, err = store1.DB().Exec(
		"INSERT INTO nodes (id, title, content, type, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"test-id", "Test", "content", "note", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	store1.Close()

	// Second open: should preserve data
	store2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open() error: %v", err)
	}
	defer store2.Close()

	var title string
	err = store2.DB().QueryRow("SELECT title FROM nodes WHERE id = ?", "test-id").Scan(&title)
	if err != nil {
		t.Fatalf("query after reopen: %v", err)
	}
	if title != "Test" {
		t.Errorf("title = %q, want %q", title, "Test")
	}
}

func TestOpen_SchemaColumns(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	// Verify nodes columns
	expectedNodes := map[string]bool{
		"id": true, "title": true, "content": true, "type": true,
		"embedding": true, "created_at": true, "updated_at": true,
	}
	checkColumns(t, store.DB(), "nodes", expectedNodes)

	// Verify edges columns
	expectedEdges := map[string]bool{
		"id": true, "from_node_id": true, "to_node_id": true,
		"type": true, "explanation": true, "created_at": true,
	}
	checkColumns(t, store.DB(), "edges", expectedEdges)

	// Verify dimensions columns
	expectedDims := map[string]bool{
		"id": true, "name": true, "description": true, "created_at": true,
	}
	checkColumns(t, store.DB(), "dimensions", expectedDims)

	// Verify node_dimensions columns
	expectedND := map[string]bool{
		"node_id": true, "dimension_id": true,
	}
	checkColumns(t, store.DB(), "node_dimensions", expectedND)
}

func checkColumns(t *testing.T, db *sql.DB, table string, expected map[string]bool) {
	t.Helper()
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s) error: %v", table, err)
	}
	defer rows.Close()

	found := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan error: %v", err)
		}
		found[name] = true
	}

	for col := range expected {
		if !found[col] {
			t.Errorf("table %s: missing column %q", table, col)
		}
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}
