package knowledge

const schemaSQL = `
CREATE TABLE IF NOT EXISTS nodes (
	id         TEXT PRIMARY KEY,
	title      TEXT NOT NULL,
	content    TEXT NOT NULL DEFAULT '',
	type       TEXT NOT NULL DEFAULT 'note',
	embedding  BLOB,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
CREATE INDEX IF NOT EXISTS idx_nodes_created_at ON nodes(created_at);

CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
	title,
	content,
	content='nodes',
	content_rowid='rowid'
);

-- Triggers to keep FTS index in sync with nodes table
CREATE TRIGGER IF NOT EXISTS nodes_fts_insert AFTER INSERT ON nodes BEGIN
	INSERT INTO nodes_fts(rowid, title, content) VALUES (new.rowid, new.title, new.content);
END;

CREATE TRIGGER IF NOT EXISTS nodes_fts_delete AFTER DELETE ON nodes BEGIN
	INSERT INTO nodes_fts(nodes_fts, rowid, title, content) VALUES ('delete', old.rowid, old.title, old.content);
END;

CREATE TRIGGER IF NOT EXISTS nodes_fts_update AFTER UPDATE ON nodes BEGIN
	INSERT INTO nodes_fts(nodes_fts, rowid, title, content) VALUES ('delete', old.rowid, old.title, old.content);
	INSERT INTO nodes_fts(rowid, title, content) VALUES (new.rowid, new.title, new.content);
END;

CREATE TABLE IF NOT EXISTS edges (
	id           TEXT PRIMARY KEY,
	from_node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	to_node_id   TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	type         TEXT NOT NULL DEFAULT 'related',
	explanation  TEXT NOT NULL DEFAULT '',
	created_at   TEXT NOT NULL,
	UNIQUE(from_node_id, to_node_id, type)
);

CREATE INDEX IF NOT EXISTS idx_edges_from ON edges(from_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_to ON edges(to_node_id);

CREATE TABLE IF NOT EXISTS dimensions (
	id          TEXT PRIMARY KEY,
	name        TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL DEFAULT '',
	created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS node_dimensions (
	node_id      TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	dimension_id TEXT NOT NULL REFERENCES dimensions(id) ON DELETE CASCADE,
	PRIMARY KEY (node_id, dimension_id)
);

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
`

const schemaVersion = 1
