package failures

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store handles persistence of failure data using SQLite
type Store struct {
	db *sql.DB
}

// Failure represents a single failure record
type Failure struct {
	ID        int
	TaskID    string
	Category  string
	Details   string
	Source    string
	CreatedAt time.Time
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

// UpsertCategoryStats inserts or updates category statistics in the database.
// If the category already exists, it updates the occurrence count and timestamps.
// If the category doesn't exist, it creates a new record.
func (s *Store) UpsertCategoryStats(category string, count int, firstSeen, lastSeen time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO category_stats (category, occurrence_count, first_seen, last_seen)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(category) DO UPDATE SET
			occurrence_count = excluded.occurrence_count,
			first_seen = excluded.first_seen,
			last_seen = excluded.last_seen
	`, category, count, firstSeen, lastSeen)

	if err != nil {
		return fmt.Errorf("upserting category stats for %q: %w", category, err)
	}

	return nil
}

// Insert inserts a new failure record into the failures table.
// The created_at timestamp is automatically set to the current time if not provided.
func (s *Store) Insert(failure Failure) error {
	createdAt := failure.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err := s.db.Exec(`
		INSERT INTO failures (task_id, category, details, source, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, failure.TaskID, failure.Category, failure.Details, failure.Source, createdAt)

	if err != nil {
		return fmt.Errorf("inserting failure for task %q: %w", failure.TaskID, err)
	}

	return nil
}

// GetByCategory retrieves all failure records for the specified category.
// Returns an empty slice (not nil) if no failures are found.
func (s *Store) GetByCategory(category string) ([]Failure, error) {
	rows, err := s.db.Query(`
		SELECT id, task_id, category, details, source, created_at
		FROM failures
		WHERE category = ?
		ORDER BY created_at DESC
	`, category)
	if err != nil {
		return nil, fmt.Errorf("querying failures for category %q: %w", category, err)
	}
	defer rows.Close()

	var failures []Failure
	for rows.Next() {
		var f Failure
		if err := rows.Scan(&f.ID, &f.TaskID, &f.Category, &f.Details, &f.Source, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning failure row: %w", err)
		}
		failures = append(failures, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating failure rows: %w", err)
	}

	// Return empty slice instead of nil if no results
	if failures == nil {
		failures = []Failure{}
	}

	return failures, nil
}

// GetOccurrenceCount retrieves the occurrence count for the specified category.
// Returns 0 if the category doesn't exist (not an error).
func (s *Store) GetOccurrenceCount(category string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT occurrence_count
		FROM category_stats
		WHERE category = ?
	`, category).Scan(&count)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("getting occurrence count for category %q: %w", category, err)
	}

	return count, nil
}

// IncrementCount increments the occurrence count for the specified category.
// If the category doesn't exist, it creates a new record with count=1.
// Updates last_seen to the current time.
func (s *Store) IncrementCount(category string) error {
	now := time.Now()

	// Try to get existing stats
	var existingCount int
	var firstSeen time.Time
	err := s.db.QueryRow(`
		SELECT occurrence_count, first_seen
		FROM category_stats
		WHERE category = ?
	`, category).Scan(&existingCount, &firstSeen)

	if err == sql.ErrNoRows {
		// Category doesn't exist, insert with count=1
		_, err = s.db.Exec(`
			INSERT INTO category_stats (category, occurrence_count, first_seen, last_seen)
			VALUES (?, 1, ?, ?)
		`, category, now, now)
		if err != nil {
			return fmt.Errorf("inserting new category stats for %q: %w", category, err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("checking existing category stats for %q: %w", category, err)
	}

	// Category exists, increment count and update last_seen
	_, err = s.db.Exec(`
		UPDATE category_stats
		SET occurrence_count = ?, last_seen = ?
		WHERE category = ?
	`, existingCount+1, now, category)

	if err != nil {
		return fmt.Errorf("incrementing count for category %q: %w", category, err)
	}

	return nil
}
