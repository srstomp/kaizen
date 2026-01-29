package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/srstomp/kaizen/internal/failures"
)

func TestSuggestCommand(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-failures.db")

	// Initialize database
	store, err := failures.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	// Add some test data
	now := time.Now()
	err = store.UpsertCategoryStats("missing-tests", 7, now.Add(-7*24*time.Hour), now)
	if err != nil {
		t.Fatalf("Failed to upsert category stats: %v", err)
	}
	store.Close()

	// Create templates directory with test template
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	templateYAML := `category: missing-tests
prefix: MT
fix_task:
  title_template: "Add missing tests for {original_task_id}"
  type: test
  description_template: "Task {original_task_id} is missing tests. Category: {category}"
  estimate_hours: 1.5
`
	templatePath := filepath.Join(templatesDir, "missing-tests.yaml")
	if err := os.WriteFile(templatePath, []byte(templateYAML), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	tests := []struct {
		name           string
		taskID         string
		category       string
		wantCategory   string
		wantOccur      int
		wantConfidence string
		wantAction     string
		wantTitle      string
		wantType       string
		wantEstimate   float64
	}{
		{
			name:           "high confidence auto-create",
			taskID:         "TASK-123",
			category:       "missing-tests",
			wantCategory:   "missing-tests",
			wantOccur:      7,
			wantConfidence: "high",
			wantAction:     "auto-create",
			wantTitle:      "Add missing tests for TASK-123",
			wantType:       "test",
			wantEstimate:   1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the suggest command
			output, err := runSuggestCommandWithConfig(tt.taskID, tt.category, dbPath, templatesDir)
			if err != nil {
				t.Fatalf("runSuggestCommand failed: %v", err)
			}

			// Parse JSON output
			var result SuggestOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Verify output
			if result.Category != tt.wantCategory {
				t.Errorf("Category = %q, want %q", result.Category, tt.wantCategory)
			}
			if result.Occurrences != tt.wantOccur {
				t.Errorf("Occurrences = %d, want %d", result.Occurrences, tt.wantOccur)
			}
			if result.Confidence != tt.wantConfidence {
				t.Errorf("Confidence = %q, want %q", result.Confidence, tt.wantConfidence)
			}
			if result.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", result.Action, tt.wantAction)
			}
			if result.FixTask == nil {
				t.Fatal("FixTask is nil")
			}
			if result.FixTask.Title != tt.wantTitle {
				t.Errorf("FixTask.Title = %q, want %q", result.FixTask.Title, tt.wantTitle)
			}
			if result.FixTask.Type != tt.wantType {
				t.Errorf("FixTask.Type = %q, want %q", result.FixTask.Type, tt.wantType)
			}
			if result.FixTask.EstimateHours != tt.wantEstimate {
				t.Errorf("FixTask.EstimateHours = %f, want %f", result.FixTask.EstimateHours, tt.wantEstimate)
			}
			// Check that description contains task_id and category
			if !strings.Contains(result.FixTask.Description, tt.taskID) {
				t.Errorf("FixTask.Description does not contain task_id %q: %s", tt.taskID, result.FixTask.Description)
			}
			if !strings.Contains(result.FixTask.Description, tt.category) {
				t.Errorf("FixTask.Description does not contain category %q: %s", tt.category, result.FixTask.Description)
			}
		})
	}
}

func TestSuggestCommandMissingTemplate(t *testing.T) {
	// Create a temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-failures.db")

	// Initialize database
	store, err := failures.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	// Add some test data for a category without a template
	now := time.Now()
	err = store.UpsertCategoryStats("unknown-category", 3, now.Add(-3*24*time.Hour), now)
	if err != nil {
		t.Fatalf("Failed to upsert category stats: %v", err)
	}
	store.Close()

	// Create empty templates directory
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Run the suggest command
	output, err := runSuggestCommandWithConfig("TASK-456", "unknown-category", dbPath, templatesDir)
	if err != nil {
		t.Fatalf("runSuggestCommand failed: %v", err)
	}

	// Parse JSON output
	var result SuggestOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify output has null fix_task when template is missing
	if result.Category != "unknown-category" {
		t.Errorf("Category = %q, want %q", result.Category, "unknown-category")
	}
	if result.Occurrences != 3 {
		t.Errorf("Occurrences = %d, want %d", result.Occurrences, 3)
	}
	if result.Confidence != "medium" {
		t.Errorf("Confidence = %q, want %q", result.Confidence, "medium")
	}
	if result.Action != "suggest" {
		t.Errorf("Action = %q, want %q", result.Action, "suggest")
	}
	if result.FixTask != nil {
		t.Errorf("FixTask should be nil when template is missing, got: %+v", result.FixTask)
	}
}

func TestSuggestCommandDatabaseNotExist(t *testing.T) {
	// Use a non-existent database path
	dbPath := "/nonexistent/path/failures.db"
	templatesDir := t.TempDir()

	_, err := runSuggestCommandWithConfig("TASK-789", "missing-tests", dbPath, templatesDir)
	if err == nil {
		t.Fatal("Expected error for non-existent database, got nil")
	}

	// Error should mention database not existing
	if !strings.Contains(err.Error(), "opening database") && !strings.Contains(err.Error(), "no such file") {
		t.Errorf("Expected database error, got: %v", err)
	}
}
