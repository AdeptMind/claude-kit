package knowledge

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	_ "modernc.org/sqlite"
)

// Store provides access to the knowledge graph SQLite database.
type Store struct {
	db       *sql.DB
	vecAvail bool // true if sqlite-vec extension loaded
	embedder embedder.Embedder
}

// Open opens (or creates) the knowledge database at dbPath.
// It creates the parent directory if needed, applies the schema,
// and enables WAL mode and foreign keys.
func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Execute schema (all statements are IF NOT EXISTS, safe to re-run)
	for _, stmt := range splitStatements(schemaSQL) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("applying schema (%s): %w", truncate(stmt, 60), err)
		}
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying *sql.DB for advanced queries.
func (s *Store) DB() *sql.DB {
	return s.db
}

// VecAvailable reports whether sqlite-vec extension is loaded.
func (s *Store) VecAvailable() bool {
	return s.vecAvail
}

// splitStatements splits a SQL script on semicolons, respecting
// multi-line triggers (END;) and PRAGMA statements.
func splitStatements(sqlScript string) []string {
	var stmts []string
	var current strings.Builder
	inTrigger := false

	for _, line := range strings.Split(sqlScript, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")

		upper := strings.ToUpper(trimmed)
		if strings.Contains(upper, "CREATE TRIGGER") || strings.Contains(upper, "BEGIN") {
			inTrigger = true
		}
		if inTrigger {
			if strings.HasSuffix(upper, "END;") {
				inTrigger = false
				stmts = append(stmts, current.String())
				current.Reset()
			}
		} else if strings.HasSuffix(trimmed, ";") {
			stmts = append(stmts, current.String())
			current.Reset()
		}
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		stmts = append(stmts, s)
	}
	return stmts
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
