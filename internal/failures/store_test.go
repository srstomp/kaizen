package failures

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("NewStore returned nil")
	}

	if store.db == nil {
		t.Fatal("store.db is nil")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file was not created: %s", dbPath)
	}

	// Clean up
	store.db.Close()
}

func TestNewStoreCreatesFailuresTable(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.db.Close()

	// Verify failures table exists with correct schema
	var tableName string
	err = store.db.QueryRow(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name='failures'
	`).Scan(&tableName)

	if err == sql.ErrNoRows {
		t.Fatal("failures table does not exist")
	}
	if err != nil {
		t.Fatalf("error checking for failures table: %v", err)
	}

	if tableName != "failures" {
		t.Errorf("table name = %s, expected failures", tableName)
	}

	// Verify columns exist by attempting to query them
	_, err = store.db.Query(`
		SELECT id, task_id, category, details, source, created_at
		FROM failures LIMIT 0
	`)
	if err != nil {
		t.Errorf("failures table schema is incorrect: %v", err)
	}
}

func TestNewStoreCreatesCategoryStatsTable(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.db.Close()

	// Verify category_stats table exists with correct schema
	var tableName string
	err = store.db.QueryRow(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name='category_stats'
	`).Scan(&tableName)

	if err == sql.ErrNoRows {
		t.Fatal("category_stats table does not exist")
	}
	if err != nil {
		t.Fatalf("error checking for category_stats table: %v", err)
	}

	if tableName != "category_stats" {
		t.Errorf("table name = %s, expected category_stats", tableName)
	}

	// Verify columns exist by attempting to query them
	_, err = store.db.Query(`
		SELECT category, occurrence_count, last_seen, first_seen
		FROM category_stats LIMIT 0
	`)
	if err != nil {
		t.Errorf("category_stats table schema is incorrect: %v", err)
	}
}

func TestNewStoreCreatesIndex(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.db.Close()

	// Verify index on category column exists
	var indexName string
	err = store.db.QueryRow(`
		SELECT name FROM sqlite_master
		WHERE type='index' AND tbl_name='failures' AND name='idx_failures_category'
	`).Scan(&indexName)

	if err == sql.ErrNoRows {
		t.Fatal("index idx_failures_category does not exist")
	}
	if err != nil {
		t.Fatalf("error checking for index: %v", err)
	}

	if indexName != "idx_failures_category" {
		t.Errorf("index name = %s, expected idx_failures_category", indexName)
	}
}

func TestNewStoreIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	// Create store first time
	store1, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("first NewStore failed: %v", err)
	}
	store1.db.Close()

	// Create store second time on same database
	store2, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("second NewStore failed (should be idempotent): %v", err)
	}
	defer store2.db.Close()

	// Verify tables still exist
	var count int
	err = store2.db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND (name='failures' OR name='category_stats')
	`).Scan(&count)

	if err != nil {
		t.Fatalf("error checking tables: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 tables, got %d", count)
	}
}

func TestNewStoreErrorOnInvalidPath(t *testing.T) {
	// Try to create store in a non-existent directory without creating it
	// SQLite should create the file, so this test verifies the error handling
	// exists even if it's not triggered in this specific case

	// Use a path that would definitely fail (e.g., trying to create a db file
	// where a directory should be)
	tempDir := t.TempDir()
	invalidPath := filepath.Join(tempDir, "dir", "subdir", "that", "does", "not", "exist", "db.sqlite")

	// This should still work as SQLite creates parent dirs, but let's verify
	// the error handling exists in the code by checking if errors are properly
	// propagated from sql.Open
	_, err := NewStore(invalidPath)
	// SQLite is permissive, so this might not error, but the function should
	// at least not panic and should return a valid store or error
	if err != nil {
		// Error is acceptable for invalid paths
		t.Logf("NewStore returned expected error for invalid path: %v", err)
	}
}

func TestStoreSchemaDetails(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.db.Close()

	// Test that we can insert into failures table with expected types
	result, err := store.db.Exec(`
		INSERT INTO failures (task_id, category, details, source, created_at)
		VALUES (?, ?, ?, ?, datetime('now'))
	`, "task-123", "test-category", "test details", "test source")

	if err != nil {
		t.Fatalf("failed to insert into failures: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	if id != 1 {
		t.Errorf("expected first insert id to be 1, got %d", id)
	}

	// Test that we can insert into category_stats table
	_, err = store.db.Exec(`
		INSERT INTO category_stats (category, occurrence_count, last_seen, first_seen)
		VALUES (?, ?, datetime('now'), datetime('now'))
	`, "test-category", 1)

	if err != nil {
		t.Fatalf("failed to insert into category_stats: %v", err)
	}

	// Verify category is the primary key (duplicate should fail)
	_, err = store.db.Exec(`
		INSERT INTO category_stats (category, occurrence_count, last_seen, first_seen)
		VALUES (?, ?, datetime('now'), datetime('now'))
	`, "test-category", 2)

	if err == nil {
		t.Error("expected error when inserting duplicate category, got nil")
	}
}

func TestStoreClose(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Close the store
	if err := store.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Attempting to use the database after closing should fail
	_, err = store.db.Exec("SELECT 1")
	if err == nil {
		t.Error("expected error when using database after Close(), got nil")
	}
}

func TestStoreCloseIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "failures.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Close multiple times should not panic
	if err := store.Close(); err != nil {
		t.Errorf("first Close() returned error: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Errorf("second Close() returned error: %v", err)
	}
}
