package db

import (
	"database/sql"
	"fmt"
)

const schemaVersion = 1

const createTablesSQL = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS conversations (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS provider_configs (
	id TEXT PRIMARY KEY,
	provider TEXT NOT NULL UNIQUE,
	config TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	prompt_template TEXT NOT NULL,
	input_schema_json TEXT NOT NULL DEFAULT '[]',
	output_style TEXT NOT NULL DEFAULT 'markdown',
	version INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS task_versions (
	id TEXT PRIMARY KEY,
	task_id TEXT NOT NULL,
	version INTEGER NOT NULL,
	snapshot TEXT NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS runs (
	id TEXT PRIMARY KEY,
	task_id TEXT NOT NULL,
	task_version INTEGER NOT NULL,
	input_values_json TEXT NOT NULL DEFAULT '{}',
	prompt_final TEXT NOT NULL DEFAULT '',
	output TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'running',
	error_text TEXT NOT NULL DEFAULT '',
	model TEXT NOT NULL DEFAULT '',
	input_tokens INTEGER NOT NULL DEFAULT 0,
	output_tokens INTEGER NOT NULL DEFAULT 0,
	cost_usd REAL NOT NULL DEFAULT 0,
	duration_ms INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS memory (
	task_id TEXT NOT NULL,
	key TEXT NOT NULL,
	value TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL,
	PRIMARY KEY (task_id, key),
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
`

// Migrate ensures the database schema is up to date.
func Migrate(db *sql.DB) error {
	// Check if schema_version table exists
	var tableCount int
	err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_version'`).Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("checking schema_version table: %w", err)
	}

	if tableCount == 0 {
		// Fresh database — create all tables
		if _, err := db.Exec(createTablesSQL); err != nil {
			return fmt.Errorf("creating tables: %w", err)
		}
		if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (?)`, schemaVersion); err != nil {
			return fmt.Errorf("inserting schema version: %w", err)
		}
		return nil
	}

	// Database exists — check version and apply migrations
	var currentVersion int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&currentVersion); err != nil {
		return fmt.Errorf("reading schema version: %w", err)
	}

	if currentVersion < schemaVersion {
		// Future migrations go here:
		// if currentVersion < 2 { migrateV1toV2(db) }
		if _, err := db.Exec(`UPDATE schema_version SET version = ?`, schemaVersion); err != nil {
			return fmt.Errorf("updating schema version: %w", err)
		}
	}

	return nil
}
