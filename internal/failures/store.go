package failures

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Store handles persistence of failure data using SQLite
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store with the specified database path.
// It opens the SQLite database, creates tables if they don't exist,
// and returns the store instance.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	store := &Store{db: db}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// initSchema creates the necessary tables and indexes if they don't exist.
// This method is idempotent and safe to run multiple times.
func (s *Store) initSchema() error {
	// Create failures table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS failures (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			category TEXT NOT NULL,
			details TEXT NOT NULL,
			source TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("creating failures table: %w", err)
	}

	// Create category_stats table
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS category_stats (
			category TEXT PRIMARY KEY,
			occurrence_count INTEGER NOT NULL,
			last_seen DATETIME NOT NULL,
			first_seen DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("creating category_stats table: %w", err)
	}

	// Create index on category column for faster lookups
	_, err = s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_failures_category
		ON failures(category)
	`)
	if err != nil {
		return fmt.Errorf("creating category index: %w", err)
	}

	return nil
}
