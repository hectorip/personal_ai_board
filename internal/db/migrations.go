package db

import (
	"database/sql"
	"fmt"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// GetMigrations returns all available migrations in order
func (m *Migrator) GetMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_personas_table",
			Up: `
				CREATE TABLE IF NOT EXISTS personas (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					description TEXT,
					traits_config TEXT NOT NULL,
					memory_data TEXT,
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_personas_updated_at ON personas(updated_at);
				CREATE INDEX IF NOT EXISTS idx_personas_name ON personas(name);
			`,
			Down: `DROP TABLE IF EXISTS personas;`,
		},
		{
			Version: 2,
			Name:    "create_boards_table",
			Up: `
				CREATE TABLE IF NOT EXISTS boards (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					description TEXT,
					is_template BOOLEAN DEFAULT FALSE,
					metadata TEXT,
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_boards_updated_at ON boards(updated_at);
				CREATE INDEX IF NOT EXISTS idx_boards_is_template ON boards(is_template);
				CREATE INDEX IF NOT EXISTS idx_boards_name ON boards(name);
			`,
			Down: `DROP TABLE IF EXISTS boards;`,
		},
		{
			Version: 3,
			Name:    "create_board_personas_table",
			Up: `
				CREATE TABLE IF NOT EXISTS board_personas (
					board_id TEXT NOT NULL,
					persona_id TEXT NOT NULL,
					role TEXT,
					position INTEGER DEFAULT 0,
					added_at DATETIME NOT NULL,
					PRIMARY KEY (board_id, persona_id),
					FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
					FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_board_personas_board_id ON board_personas(board_id);
				CREATE INDEX IF NOT EXISTS idx_board_personas_persona_id ON board_personas(persona_id);
			`,
			Down: `DROP TABLE IF EXISTS board_personas;`,
		},
		{
			Version: 4,
			Name:    "create_projects_table",
			Up: `
				CREATE TABLE IF NOT EXISTS projects (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					description TEXT,
					metadata TEXT,
					status TEXT DEFAULT 'active',
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_projects_updated_at ON projects(updated_at);
				CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
				CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
			`,
			Down: `DROP TABLE IF EXISTS projects;`,
		},
		{
			Version: 5,
			Name:    "create_ideas_table",
			Up: `
				CREATE TABLE IF NOT EXISTS ideas (
					id TEXT PRIMARY KEY,
					project_id TEXT NOT NULL,
					title TEXT NOT NULL,
					content TEXT,
					metadata TEXT,
					status TEXT DEFAULT 'draft',
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL,
					FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_ideas_project_id ON ideas(project_id);
				CREATE INDEX IF NOT EXISTS idx_ideas_updated_at ON ideas(updated_at);
				CREATE INDEX IF NOT EXISTS idx_ideas_status ON ideas(status);
				CREATE INDEX IF NOT EXISTS idx_ideas_title ON ideas(title);
			`,
			Down: `DROP TABLE IF EXISTS ideas;`,
		},
		{
			Version: 6,
			Name:    "create_analysis_sessions_table",
			Up: `
				CREATE TABLE IF NOT EXISTS analysis_sessions (
					id TEXT PRIMARY KEY,
					project_id TEXT NOT NULL,
					board_id TEXT NOT NULL,
					mode TEXT NOT NULL,
					status TEXT DEFAULT 'pending',
					context_data TEXT,
					results_data TEXT,
					created_at DATETIME NOT NULL,
					started_at DATETIME,
					completed_at DATETIME,
					FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
					FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_analysis_sessions_project_id ON analysis_sessions(project_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_sessions_board_id ON analysis_sessions(board_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_sessions_status ON analysis_sessions(status);
				CREATE INDEX IF NOT EXISTS idx_analysis_sessions_created_at ON analysis_sessions(created_at);
				CREATE INDEX IF NOT EXISTS idx_analysis_sessions_mode ON analysis_sessions(mode);
			`,
			Down: `DROP TABLE IF EXISTS analysis_sessions;`,
		},
		{
			Version: 7,
			Name:    "create_analysis_responses_table",
			Up: `
				CREATE TABLE IF NOT EXISTS analysis_responses (
					id TEXT PRIMARY KEY,
					session_id TEXT NOT NULL,
					persona_id TEXT NOT NULL,
					response_content TEXT NOT NULL,
					reasoning TEXT,
					confidence REAL DEFAULT 0.5,
					emotional_tone TEXT,
					response_order INTEGER DEFAULT 0,
					created_at DATETIME NOT NULL,
					FOREIGN KEY (session_id) REFERENCES analysis_sessions(id) ON DELETE CASCADE,
					FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_analysis_responses_session_id ON analysis_responses(session_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_responses_persona_id ON analysis_responses(persona_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_responses_created_at ON analysis_responses(created_at);
				CREATE INDEX IF NOT EXISTS idx_analysis_responses_order ON analysis_responses(response_order);
			`,
			Down: `DROP TABLE IF EXISTS analysis_responses;`,
		},
		{
			Version: 8,
			Name:    "create_llm_interaction_logs_table",
			Up: `
				CREATE TABLE IF NOT EXISTS llm_interaction_logs (
					id TEXT PRIMARY KEY,
					persona_id TEXT,
					session_id TEXT,
					prompt TEXT NOT NULL,
					system_message TEXT,
					response TEXT NOT NULL,
					model_name TEXT NOT NULL,
					temperature REAL,
					max_tokens INTEGER,
					tokens_used INTEGER,
					duration_ms INTEGER,
					context_data TEXT,
					created_at DATETIME NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_llm_logs_persona_id ON llm_interaction_logs(persona_id);
				CREATE INDEX IF NOT EXISTS idx_llm_logs_session_id ON llm_interaction_logs(session_id);
				CREATE INDEX IF NOT EXISTS idx_llm_logs_created_at ON llm_interaction_logs(created_at);
				CREATE INDEX IF NOT EXISTS idx_llm_logs_model_name ON llm_interaction_logs(model_name);
			`,
			Down: `DROP TABLE IF EXISTS llm_interaction_logs;`,
		},
		{
			Version: 9,
			Name:    "create_documents_table",
			Up: `
				CREATE TABLE IF NOT EXISTS documents (
					id TEXT PRIMARY KEY,
					project_id TEXT NOT NULL,
					filename TEXT NOT NULL,
					file_path TEXT NOT NULL,
					file_type TEXT NOT NULL,
					file_size INTEGER NOT NULL,
					content_hash TEXT,
					processed_content TEXT,
					metadata TEXT,
					status TEXT DEFAULT 'pending',
					created_at DATETIME NOT NULL,
					processed_at DATETIME,
					FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
				CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
				CREATE INDEX IF NOT EXISTS idx_documents_file_type ON documents(file_type);
				CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);
				CREATE INDEX IF NOT EXISTS idx_documents_content_hash ON documents(content_hash);
			`,
			Down: `DROP TABLE IF EXISTS documents;`,
		},
		{
			Version: 10,
			Name:    "create_migrations_table",
			Up: `
				CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					name TEXT NOT NULL,
					applied_at DATETIME NOT NULL
				);
			`,
			Down: `DROP TABLE IF EXISTS schema_migrations;`,
		},
		{
			Version: 11,
			Name:    "create_system_config_table",
			Up: `
				CREATE TABLE IF NOT EXISTS system_config (
					key TEXT PRIMARY KEY,
					value TEXT NOT NULL,
					description TEXT,
					updated_at DATETIME NOT NULL
				);

				-- Insert default configuration values
				INSERT OR IGNORE INTO system_config (key, value, description, updated_at) VALUES
				('schema_version', '11', 'Current database schema version', datetime('now')),
				('llm_default_provider', 'openai', 'Default LLM provider', datetime('now')),
				('llm_default_model', 'gpt-4', 'Default LLM model', datetime('now')),
				('analysis_max_concurrent', '5', 'Maximum concurrent analysis sessions', datetime('now')),
				('memory_retention_days', '90', 'Days to retain interaction logs', datetime('now'));
			`,
			Down: `DROP TABLE IF EXISTS system_config;`,
		},
		{
			Version: 12,
			Name:    "add_persona_statistics",
			Up: `
				-- Add columns for persona usage statistics
				ALTER TABLE personas ADD COLUMN total_interactions INTEGER DEFAULT 0;
				ALTER TABLE personas ADD COLUMN last_interaction_at DATETIME;
				ALTER TABLE personas ADD COLUMN average_confidence REAL DEFAULT 0.5;

				CREATE INDEX IF NOT EXISTS idx_personas_last_interaction ON personas(last_interaction_at);
				CREATE INDEX IF NOT EXISTS idx_personas_total_interactions ON personas(total_interactions);
			`,
			Down: `
				-- SQLite doesn't support DROP COLUMN, so we recreate the table
				CREATE TABLE personas_backup AS SELECT
					id, name, description, traits_config, memory_data, created_at, updated_at
				FROM personas;

				DROP TABLE personas;

				CREATE TABLE personas (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					description TEXT,
					traits_config TEXT NOT NULL,
					memory_data TEXT,
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL
				);

				INSERT INTO personas SELECT * FROM personas_backup;
				DROP TABLE personas_backup;

				CREATE INDEX idx_personas_updated_at ON personas(updated_at);
				CREATE INDEX idx_personas_name ON personas(name);
			`,
		},
		{
			Version: 13,
			Name:    "add_analysis_insights_table",
			Up: `
				CREATE TABLE IF NOT EXISTS analysis_insights (
					id TEXT PRIMARY KEY,
					session_id TEXT NOT NULL,
					insight_text TEXT NOT NULL,
					insight_type TEXT NOT NULL,
					confidence REAL DEFAULT 0.5,
					persona_id TEXT,
					created_at DATETIME NOT NULL,
					FOREIGN KEY (session_id) REFERENCES analysis_sessions(id) ON DELETE CASCADE,
					FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE SET NULL
				);

				CREATE INDEX IF NOT EXISTS idx_insights_session_id ON analysis_insights(session_id);
				CREATE INDEX IF NOT EXISTS idx_insights_type ON analysis_insights(insight_type);
				CREATE INDEX IF NOT EXISTS idx_insights_persona_id ON analysis_insights(persona_id);
				CREATE INDEX IF NOT EXISTS idx_insights_created_at ON analysis_insights(created_at);
			`,
			Down: `DROP TABLE IF EXISTS analysis_insights;`,
		},
		{
			Version: 14,
			Name:    "create_project_ideas_table",
			Up: `
				CREATE TABLE IF NOT EXISTS project_ideas (
					id TEXT PRIMARY KEY,
					project_id TEXT NOT NULL,
					title TEXT NOT NULL,
					description TEXT,
					content TEXT,
					tags TEXT,
					priority INTEGER DEFAULT 0,
					status TEXT DEFAULT 'draft',
					metadata TEXT,
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL,
					FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_project_ideas_project_id ON project_ideas(project_id);
				CREATE INDEX IF NOT EXISTS idx_project_ideas_updated_at ON project_ideas(updated_at);
				CREATE INDEX IF NOT EXISTS idx_project_ideas_status ON project_ideas(status);
				CREATE INDEX IF NOT EXISTS idx_project_ideas_priority ON project_ideas(priority);
			`,
			Down: `DROP TABLE IF EXISTS project_ideas;`,
		},
		{
			Version: 15,
			Name:    "create_project_documents_table",
			Up: `
				CREATE TABLE IF NOT EXISTS project_documents (
					id TEXT PRIMARY KEY,
					project_id TEXT NOT NULL,
					name TEXT NOT NULL,
					file_path TEXT NOT NULL,
					content_type TEXT NOT NULL,
					size INTEGER NOT NULL,
					processed_at DATETIME,
					knowledge_id TEXT,
					metadata TEXT,
					created_at DATETIME NOT NULL,
					updated_at DATETIME NOT NULL,
					FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_project_documents_project_id ON project_documents(project_id);
				CREATE INDEX IF NOT EXISTS idx_project_documents_created_at ON project_documents(created_at);
				CREATE INDEX IF NOT EXISTS idx_project_documents_content_type ON project_documents(content_type);
			`,
			Down: `DROP TABLE IF EXISTS project_documents;`,
		},
		{
			Version: 16,
			Name:    "create_analysis_requests_table",
			Up: `
				CREATE TABLE IF NOT EXISTS analysis_requests (
					id TEXT PRIMARY KEY,
					project_id TEXT NOT NULL,
					board_id TEXT NOT NULL,
					mode TEXT NOT NULL,
					config TEXT,
					created_at DATETIME NOT NULL,
					FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
					FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_analysis_requests_project_id ON analysis_requests(project_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_requests_board_id ON analysis_requests(board_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_requests_mode ON analysis_requests(mode);
				CREATE INDEX IF NOT EXISTS idx_analysis_requests_created_at ON analysis_requests(created_at);
			`,
			Down: `DROP TABLE IF EXISTS analysis_requests;`,
		},
		{
			Version: 17,
			Name:    "create_analysis_results_table",
			Up: `
				CREATE TABLE IF NOT EXISTS analysis_results (
					id TEXT PRIMARY KEY,
					request_id TEXT NOT NULL,
					project_id TEXT NOT NULL,
					board_id TEXT NOT NULL,
					mode TEXT NOT NULL,
					status TEXT DEFAULT 'pending',
					summary TEXT,
					insights TEXT,
					responses TEXT,
					metrics TEXT,
					metadata TEXT,
					started_at DATETIME NOT NULL,
					completed_at DATETIME,
					duration_ms INTEGER DEFAULT 0,
					created_at DATETIME NOT NULL,
					FOREIGN KEY (request_id) REFERENCES analysis_requests(id) ON DELETE CASCADE,
					FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
					FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_analysis_results_request_id ON analysis_results(request_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_results_project_id ON analysis_results(project_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_results_board_id ON analysis_results(board_id);
				CREATE INDEX IF NOT EXISTS idx_analysis_results_status ON analysis_results(status);
				CREATE INDEX IF NOT EXISTS idx_analysis_results_mode ON analysis_results(mode);
				CREATE INDEX IF NOT EXISTS idx_analysis_results_created_at ON analysis_results(created_at);
			`,
			Down: `DROP TABLE IF EXISTS analysis_results;`,
		},
	}
}

// RunMigrations executes all pending migrations
func (m *Migrator) RunMigrations() error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current schema version
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Get all migrations
	migrations := m.GetMigrations()

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := m.runMigration(migration); err != nil {
				return fmt.Errorf("failed to run migration %d (%s): %w", migration.Version, migration.Name, err)
			}
		}
	}

	return nil
}

// RollbackMigration rolls back to the specified version
func (m *Migrator) RollbackMigration(targetVersion int) error {
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if targetVersion >= currentVersion {
		return fmt.Errorf("target version %d is not less than current version %d", targetVersion, currentVersion)
	}

	migrations := m.GetMigrations()

	// Run rollbacks in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if migration.Version > targetVersion && migration.Version <= currentVersion {
			if err := m.rollbackMigration(migration); err != nil {
				return fmt.Errorf("failed to rollback migration %d (%s): %w", migration.Version, migration.Name, err)
			}
		}
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL
		)
	`
	_, err := m.db.Exec(query)
	return err
}

// getCurrentVersion gets the current schema version
func (m *Migrator) getCurrentVersion() (int, error) {
	var version int
	err := m.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// runMigration executes a single migration
func (m *Migrator) runMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute the migration SQL
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
		migration.Version, migration.Name, "datetime('now')",
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute the rollback SQL
	if _, err := tx.Exec(migration.Down); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove the migration record
	if _, err := tx.Exec(
		"DELETE FROM schema_migrations WHERE version = ?",
		migration.Version,
	); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return tx.Commit()
}

// GetMigrationStatus returns the status of all migrations
func (m *Migrator) GetMigrationStatus() ([]MigrationStatus, error) {
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return nil, err
	}

	migrations := m.GetMigrations()
	status := make([]MigrationStatus, len(migrations))

	for i, migration := range migrations {
		status[i] = MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: migration.Version <= currentVersion,
		}
	}

	return status, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version int    `json:"version"`
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
}

// ResetDatabase drops all tables and recreates the schema
func (m *Migrator) ResetDatabase() error {
	// Get all table names
	rows, err := m.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	// Drop all tables
	for _, table := range tables {
		if _, err := m.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Run all migrations
	return m.RunMigrations()
}

// ValidateSchema validates that the database schema matches the expected structure
func (m *Migrator) ValidateSchema() error {
	expectedTables := []string{
		"personas", "boards", "board_personas", "projects", "ideas",
		"analysis_sessions", "analysis_responses", "llm_interaction_logs",
		"documents", "schema_migrations", "system_config", "analysis_insights",
	}

	for _, tableName := range expectedTables {
		var count int
		err := m.db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			tableName,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", tableName, err)
		}
		if count == 0 {
			return fmt.Errorf("missing required table: %s", tableName)
		}
	}

	return nil
}
