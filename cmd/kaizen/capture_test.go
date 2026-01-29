package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/srstomp/kaizen/internal/failures"
)

func TestCaptureCommand(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		category   string
		details    string
		source     string
		wantTaskID string
		wantCat    string
		wantErr    bool
	}{
		{
			name:       "successful capture",
			taskID:     "TASK-123",
			category:   "missing-tests",
			details:    "Task is missing unit tests for the new feature",
			source:     "spec-review",
			wantTaskID: "TASK-123",
			wantCat:    "missing-tests",
			wantErr:    false,
		},
		{
			name:       "capture with different category",
			taskID:     "TASK-456",
			category:   "scope-creep",
			details:    "Task added extra features beyond requirements",
			source:     "quality-review",
			wantTaskID: "TASK-456",
			wantCat:    "scope-creep",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary database for testing
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test-failures.db")

			// Initialize database
			store, err := failures.NewStore(dbPath)
			if err != nil {
				t.Fatalf("Failed to create test store: %v", err)
			}
			store.Close()

			// Run the capture command
			output, err := runCaptureCommandWithConfig(tt.taskID, tt.category, tt.details, tt.source, dbPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("runCaptureCommand error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			// Parse JSON output
			var result CaptureOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Verify output
			if !result.Success {
				t.Errorf("Success = false, want true")
			}
			if result.TaskID != tt.wantTaskID {
				t.Errorf("TaskID = %q, want %q", result.TaskID, tt.wantTaskID)
			}
			if result.Category != tt.wantCat {
				t.Errorf("Category = %q, want %q", result.Category, tt.wantCat)
			}
			if result.Message != "Failure captured successfully" {
				t.Errorf("Message = %q, want %q", result.Message, "Failure captured successfully")
			}
			if result.Error != "" {
				t.Errorf("Error = %q, want empty", result.Error)
			}

			// Verify failure was inserted in database
			store, err = failures.NewStore(dbPath)
			if err != nil {
				t.Fatalf("Failed to open store: %v", err)
			}
			defer store.Close()

			failures, err := store.GetByCategory(tt.category)
			if err != nil {
				t.Fatalf("Failed to get failures: %v", err)
			}

			if len(failures) != 1 {
				t.Fatalf("Expected 1 failure, got %d", len(failures))
			}

			f := failures[0]
			if f.TaskID != tt.taskID {
				t.Errorf("Failure.TaskID = %q, want %q", f.TaskID, tt.taskID)
			}
			if f.Category != tt.category {
				t.Errorf("Failure.Category = %q, want %q", f.Category, tt.category)
			}
			if f.Details != tt.details {
				t.Errorf("Failure.Details = %q, want %q", f.Details, tt.details)
			}
			if f.Source != tt.source {
				t.Errorf("Failure.Source = %q, want %q", f.Source, tt.source)
			}

			// Verify category count was incremented
			count, err := store.GetOccurrenceCount(tt.category)
			if err != nil {
				t.Fatalf("Failed to get occurrence count: %v", err)
			}
			if count != 1 {
				t.Errorf("OccurrenceCount = %d, want 1", count)
			}
		})
	}
}

func TestCaptureCommandError(t *testing.T) {
	tests := []struct {
		name     string
		taskID   string
		category string
		details  string
		source   string
		dbPath   string
		wantErr  string
	}{
		{
			name:     "database error",
			taskID:   "TASK-999",
			category: "test-category",
			details:  "test details",
			source:   "test-source",
			dbPath:   "/nonexistent/path/failures.db",
			wantErr:  "opening database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the capture command with bad database path
			output, err := runCaptureCommandWithConfig(tt.taskID, tt.category, tt.details, tt.source, tt.dbPath)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			// Parse JSON output
			var result CaptureOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Verify error output
			if result.Success {
				t.Errorf("Success = true, want false")
			}
			if result.Error == "" {
				t.Errorf("Error is empty, want error message")
			}
			if !strings.Contains(result.Error, tt.wantErr) {
				t.Errorf("Error = %q, want to contain %q", result.Error, tt.wantErr)
			}
		})
	}
}

func TestCaptureCommandMultipleCaptures(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-failures.db")

	// Initialize database
	store, err := failures.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	store.Close()

	// Capture first failure
	_, err = runCaptureCommandWithConfig("TASK-001", "missing-tests", "First failure", "spec-review", dbPath)
	if err != nil {
		t.Fatalf("First capture failed: %v", err)
	}

	// Capture second failure for same category
	_, err = runCaptureCommandWithConfig("TASK-002", "missing-tests", "Second failure", "quality-review", dbPath)
	if err != nil {
		t.Fatalf("Second capture failed: %v", err)
	}

	// Capture third failure for different category
	_, err = runCaptureCommandWithConfig("TASK-003", "scope-creep", "Third failure", "spec-review", dbPath)
	if err != nil {
		t.Fatalf("Third capture failed: %v", err)
	}

	// Verify database state
	store, err = failures.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	// Check missing-tests category
	missingTestsFailures, err := store.GetByCategory("missing-tests")
	if err != nil {
		t.Fatalf("Failed to get missing-tests failures: %v", err)
	}
	if len(missingTestsFailures) != 2 {
		t.Errorf("Expected 2 missing-tests failures, got %d", len(missingTestsFailures))
	}

	missingTestsCount, err := store.GetOccurrenceCount("missing-tests")
	if err != nil {
		t.Fatalf("Failed to get missing-tests count: %v", err)
	}
	if missingTestsCount != 2 {
		t.Errorf("Expected missing-tests count = 2, got %d", missingTestsCount)
	}

	// Check scope-creep category
	scopeCreepFailures, err := store.GetByCategory("scope-creep")
	if err != nil {
		t.Fatalf("Failed to get scope-creep failures: %v", err)
	}
	if len(scopeCreepFailures) != 1 {
		t.Errorf("Expected 1 scope-creep failure, got %d", len(scopeCreepFailures))
	}

	scopeCreepCount, err := store.GetOccurrenceCount("scope-creep")
	if err != nil {
		t.Fatalf("Failed to get scope-creep count: %v", err)
	}
	if scopeCreepCount != 1 {
		t.Errorf("Expected scope-creep count = 1, got %d", scopeCreepCount)
	}
}

func TestCaptureCommandIntegration(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for config
	tmpDir := t.TempDir()

	// Set up HOME to use temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create config directory structure
	configDir := filepath.Join(tmpDir, ".config", "kaizen")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	dbPath := filepath.Join(configDir, "failures.db")

	// Initialize database
	store, err := failures.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	store.Close()

	// Run capture command (should use default config path)
	output, err := runCaptureCommand("TASK-999", "missing-tests", "Integration test failure", "integration-test")
	if err != nil {
		t.Fatalf("runCaptureCommand failed: %v", err)
	}

	// Parse JSON output
	var result CaptureOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify output
	if !result.Success {
		t.Errorf("Success = false, want true")
	}
	if result.TaskID != "TASK-999" {
		t.Errorf("TaskID = %q, want %q", result.TaskID, "TASK-999")
	}
	if result.Category != "missing-tests" {
		t.Errorf("Category = %q, want %q", result.Category, "missing-tests")
	}
}
