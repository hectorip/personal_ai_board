package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config represents database configuration
type Config struct {
	Path            string        `json:"path"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	EnableWAL       bool          `json:"enable_wal"`
	EnableForeignKeys bool        `json:"enable_foreign_keys"`
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	return &Config{
		Path:            "personal_ai_board.db",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 5,
		EnableWAL:       true,
		EnableForeignKeys: true,
	}
}

// Database wraps the sql.DB with additional functionality
type Database struct {
	*sql.DB
	config *Config
}

// Connect establishes a connection to the SQLite database
func Connect(config *Config) (*Database, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Ensure the directory exists
	dir := filepath.Dir(config.Path)
	if dir != "." && dir != "" {
		// Note: In a real implementation, you'd create the directory
		// For now, we assume the directory exists
	}

	// Build connection string with SQLite options
	dsn := config.Path
	params := []string{}

	if config.EnableWAL {
		params = append(params, "_journal_mode=WAL")
	}

	if config.EnableForeignKeys {
		params = append(params, "_foreign_keys=on")
	}

	// Add other SQLite performance optimizations
	params = append(params, "_synchronous=NORMAL")
	params = append(params, "_cache_size=10000")
	params = append(params, "_temp_store=memory")

	if len(params) > 0 {
		dsn += "?"
		for i, param := range params {
			if i > 0 {
				dsn += "&"
			}
			dsn += param
		}
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		DB:     db,
		config: config,
	}

	// Run initial setup
	if err := database.setup(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return database, nil
}

// setup performs initial database setup
func (db *Database) setup() error {
	// Enable SQLite optimizations
	pragmas := []string{
		"PRAGMA busy_timeout = 5000",  // 5 second timeout for busy database
		"PRAGMA temp_store = memory",  // Store temp tables in memory
		"PRAGMA mmap_size = 268435456", // 256MB memory-mapped I/O
	}

	if db.config.EnableForeignKeys {
		pragmas = append(pragmas, "PRAGMA foreign_keys = ON")
	}

	if db.config.EnableWAL {
		pragmas = append(pragmas, "PRAGMA journal_mode = WAL")
		pragmas = append(pragmas, "PRAGMA synchronous = NORMAL")
		pragmas = append(pragmas, "PRAGMA wal_autocheckpoint = 1000")
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	return nil
}

// Migrate runs database migrations
func (db *Database) Migrate() error {
	migrator := NewMigrator(db.DB)
	return migrator.RunMigrations()
}

// Reset drops all tables and recreates the schema
func (db *Database) Reset() error {
	migrator := NewMigrator(db.DB)
	return migrator.ResetDatabase()
}

// ValidateSchema validates the database schema
func (db *Database) ValidateSchema() error {
	migrator := NewMigrator(db.DB)
	return migrator.ValidateSchema()
}

// GetConfig returns the database configuration
func (db *Database) GetConfig() *Config {
	return db.config
}

// HealthCheck performs a health check on the database
func (db *Database) HealthCheck() error {
	// Test basic connectivity
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test a simple query
	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("test query returned unexpected result: %d", result)
	}

	return nil
}

// GetStats returns database statistics
func (db *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Connection pool stats
	dbStats := db.Stats()
	stats["open_connections"] = dbStats.OpenConnections
	stats["in_use"] = dbStats.InUse
	stats["idle"] = dbStats.Idle
	stats["wait_count"] = dbStats.WaitCount
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = dbStats.MaxIdleClosed
	stats["max_idle_time_closed"] = dbStats.MaxIdleTimeClosed
	stats["max_lifetime_closed"] = dbStats.MaxLifetimeClosed

	// Database size
	var dbSize int64
	err := db.QueryRow("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").Scan(&dbSize)
	if err == nil {
		stats["database_size_bytes"] = dbSize
		stats["database_size_mb"] = float64(dbSize) / (1024 * 1024)
	}

	// Table counts
	tables := []string{"personas", "boards", "projects", "analysis_sessions", "llm_interaction_logs"}
	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err == nil {
			stats[fmt.Sprintf("%s_count", table)] = count
		}
	}

	// WAL mode info if enabled
	if db.config.EnableWAL {
		var walMode string
		err := db.QueryRow("PRAGMA journal_mode").Scan(&walMode)
		if err == nil {
			stats["journal_mode"] = walMode
		}

		var walSize int64
		err = db.QueryRow("PRAGMA wal_checkpoint(PASSIVE)").Scan(&walSize)
		if err == nil {
			stats["wal_size_pages"] = walSize
		}
	}

	return stats, nil
}

// Backup creates a backup of the database
func (db *Database) Backup(destPath string) error {
	// For SQLite, we can use the backup API
	destDB, err := sql.Open("sqlite3", destPath)
	if err != nil {
		return fmt.Errorf("failed to open destination database: %w", err)
	}
	defer destDB.Close()

	// Simple backup by copying all data
	// In a production environment, you'd use sqlite3_backup_init/step/finish
	// For now, we'll use a simple approach
	
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// This is a simplified backup - in production you'd use the SQLite backup API
	query := "VACUUM INTO ?"
	_, err = db.Exec(query, destPath)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	return nil
}

// Vacuum optimizes the database by rebuilding it
func (db *Database) Vacuum() error {
	_, err := db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("vacuum failed: %w", err)
	}
	return nil
}

// Analyze updates database statistics for the query planner
func (db *Database) Analyze() error {
	_, err := db.Exec("ANALYZE")
	if err != nil {
		return fmt.Errorf("analyze failed: %w", err)
	}
	return nil
}

// CleanupOldLogs removes old LLM interaction logs based on retention policy
func (db *Database) CleanupOldLogs(retentionDays int) error {
	query := `
		DELETE FROM llm_interaction_logs 
		WHERE created_at < datetime('now', '-' || ? || ' days')
	`
	result, err := db.Exec(query, retentionDays)
	if err != nil {
		return fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// Log the cleanup operation
		fmt.Printf("Cleaned up %d old LLM interaction logs\n", rowsAffected)
	}

	return nil
}

// GetSystemConfig retrieves a system configuration value
func (db *Database) GetSystemConfig(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM system_config WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSystemConfig sets a system configuration value
func (db *Database) SetSystemConfig(key, value, description string) error {
	query := `
		INSERT OR REPLACE INTO system_config (key, value, description, updated_at)
		VALUES (?, ?, ?, datetime('now'))
	`
	_, err := db.Exec(query, key, value, description)
	return err
}

// GetAllSystemConfig retrieves all system configuration values
func (db *Database) GetAllSystemConfig() (map[string]string, error) {
	rows, err := db.Query("SELECT key, value FROM system_config")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		config[key] = value
	}

	return config, rows.Err()
}

// WithTransaction executes a function within a database transaction
func (db *Database) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// Close closes the database connection
func (db *Database) Close() error {
	if db.config.EnableWAL {
		// Checkpoint the WAL file before closing
		_, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
		if err != nil {
			fmt.Printf("Warning: failed to checkpoint WAL file: %v\n", err)
		}
	}

	return db.DB.Close()
}

// DatabaseInfo contains information about the database
type DatabaseInfo struct {
	Path            string            `json:"path"`
	Size            int64             `json:"size"`
	Tables          map[string]int    `json:"tables"`
	Config          *Config           `json:"config"`
	JournalMode     string            `json:"journal_mode"`
	SchemaVersion   int               `json:"schema_version"`
	Stats           map[string]interface{} `json:"stats"`
}

// GetDatabaseInfo returns comprehensive information about the database
func (db *Database) GetDatabaseInfo() (*DatabaseInfo, error) {
	info := &DatabaseInfo{
		Path:   db.config.Path,
		Config: db.config,
		Tables: make(map[string]int),
	}

	// Get database size
	err := db.QueryRow("SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()").Scan(&info.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	// Get journal mode
	err = db.QueryRow("PRAGMA journal_mode").Scan(&info.JournalMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get journal mode: %w", err)
	}

	// Get schema version
	migrator := NewMigrator(db.DB)
	info.SchemaVersion, err = migrator.getCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get schema version: %w", err)
	}

	// Get table counts
	tables := []string{
		"personas", "boards", "board_personas", "projects", "ideas",
		"analysis_sessions", "analysis_responses", "llm_interaction_logs",
		"documents", "analysis_insights",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err == nil {
			info.Tables[table] = count
		}
	}

	// Get database stats
	info.Stats, err = db.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	return info, nil
}