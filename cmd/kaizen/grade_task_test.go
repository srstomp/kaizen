package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunGradeTaskCommand_InvalidTaskType tests that invalid task types are rejected
func TestRunGradeTaskCommand_InvalidTaskType(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		taskType string
		wantErr  bool
	}{
		{"valid_feature", "feature", false},
		{"valid_bug", "bug", false},
		{"valid_test", "test", false},
		{"valid_spike", "spike", false},
		{"valid_chore", "chore", false},
		{"invalid_lowercase", "invalid", true},
		{"invalid_uppercase", "FEATURE", true},
		{"invalid_empty", "", true},
		{"invalid_random", "foo-bar", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := runGradeTaskCommand("test-task", tc.taskType, []string{}, tmpDir, "json")
			if tc.wantErr && err == nil {
				t.Errorf("Expected error for task type %q, got nil", tc.taskType)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Expected no error for task type %q, got: %v", tc.taskType, err)
			}
		})
	}
}

// TestRunGradeTaskCommand_MultipleGraders tests that multiple graders are run
func TestRunGradeTaskCommand_MultipleGraders(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testTestFile := filepath.Join(tmpDir, "test_test.go")
	if err := os.WriteFile(testTestFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskCommand("test-task", "feature", []string{testFile, testTestFile}, tmpDir, "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result GradeTaskOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify multiple graders ran (at least FileExistsGrader and TestExistsGrader)
	if len(result.Results) < 2 {
		t.Errorf("Expected at least 2 grader results, got %d", len(result.Results))
	}

	// Verify grader names are set
	for _, r := range result.Results {
		if r.GraderName == "" {
			t.Error("Grader name should not be empty")
		}
	}
}

// TestRunGradeTaskCommand_OverallScoreCalculation tests overall score calculation
func TestRunGradeTaskCommand_OverallScoreCalculation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files to ensure graders are applicable
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testTestFile := filepath.Join(tmpDir, "test_test.go")
	if err := os.WriteFile(testTestFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskCommand("test-task", "feature", []string{testFile, testTestFile}, tmpDir, "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result GradeTaskOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Calculate expected overall score (average of applicable graders)
	var totalScore float64
	applicableCount := 0
	for _, r := range result.Results {
		if !r.Skipped {
			totalScore += r.Score
			applicableCount++
		}
	}

	expectedScore := float64(0)
	if applicableCount > 0 {
		expectedScore = totalScore / float64(applicableCount)
	}

	// Verify overall score is the average of applicable graders
	if result.OverallScore != expectedScore {
		t.Errorf("Expected overall score %.2f, got %.2f", expectedScore, result.OverallScore)
	}
}

// TestRunGradeTaskCommand_SkippedGradersIgnored tests that skipped graders don't affect score
func TestRunGradeTaskCommand_SkippedGradersIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-feature task type where some graders might be skipped
	// For spike tasks, TestExistsGrader is typically skipped
	testFile := filepath.Join(tmpDir, "doc.md")
	if err := os.WriteFile(testFile, []byte("# Documentation\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskCommand("test-task", "spike", []string{testFile}, tmpDir, "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result GradeTaskOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Count applicable vs skipped graders
	applicableCount := 0
	skippedCount := 0
	var totalScore float64

	for _, r := range result.Results {
		if r.Skipped {
			skippedCount++
		} else {
			applicableCount++
			totalScore += r.Score
		}
	}

	// Verify we have at least one skipped grader (spike shouldn't require tests)
	if skippedCount == 0 {
		t.Log("Warning: Expected at least one skipped grader for spike task")
	}

	// Verify overall score only includes applicable graders
	expectedScore := float64(0)
	if applicableCount > 0 {
		expectedScore = totalScore / float64(applicableCount)
	}

	if result.OverallScore != expectedScore {
		t.Errorf("Expected overall score %.2f (from %d applicable graders), got %.2f",
			expectedScore, applicableCount, result.OverallScore)
	}
}

// TestRunGradeTaskCommand_OverallPassedLogic tests overall passed calculation
func TestRunGradeTaskCommand_OverallPassedLogic(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name         string
		files        []string
		taskType     string
		expectPassed bool
	}{
		{
			name: "all_pass",
			files: []string{
				"test.go",
				"test_test.go",
			},
			taskType:     "feature",
			expectPassed: true,
		},
		{
			name: "missing_tests",
			files: []string{
				"test.go",
			},
			taskType:     "feature",
			expectPassed: false, // Should fail because no test file
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test files
			var filePaths []string
			for _, fname := range tc.files {
				fpath := filepath.Join(tmpDir, tc.name, fname)
				if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
					t.Fatalf("Failed to create dir: %v", err)
				}
				if err := os.WriteFile(fpath, []byte("package main\n"), 0644); err != nil {
					t.Fatalf("Failed to create file %s: %v", fname, err)
				}
				filePaths = append(filePaths, fpath)
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskCommand("test-task", tc.taskType, filePaths, tmpDir, "json")

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runGradeTaskCommand failed: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result GradeTaskOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v", err)
			}

			if result.OverallPassed != tc.expectPassed {
				t.Errorf("Expected overall_passed=%v, got %v", tc.expectPassed, result.OverallPassed)
				t.Logf("Results: %+v", result.Results)
			}
		})
	}
}

// TestRunGradeTaskCommand_JSONOutput tests JSON output format
func TestRunGradeTaskCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskCommand("test-123", "feature", []string{testFile}, tmpDir, "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output is valid JSON
	var result GradeTaskOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify required fields are present
	if result.TaskID != "test-123" {
		t.Errorf("Expected task_id 'test-123', got %s", result.TaskID)
	}
	if result.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
	if len(result.Results) == 0 {
		t.Error("Results should not be empty")
	}

	// Verify results have expected structure
	for _, r := range result.Results {
		if r.GraderName == "" {
			t.Error("GraderName should not be empty")
		}
		// Score should be between 0 and 100
		if r.Score < 0 || r.Score > 100 {
			t.Errorf("Score should be 0-100, got %.2f", r.Score)
		}
	}
}

// TestRunGradeTaskCommand_TextOutput tests text output format
func TestRunGradeTaskCommand_TextOutput(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskCommand("test-456", "bug", []string{testFile}, tmpDir, "text")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify text output contains expected sections
	expectedStrings := []string{
		"Task Grading Results",
		"Task ID: test-456",
		"Task Type: bug",
		"Grader Results:",
		"Overall Score:",
		"Overall Result:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Text output missing expected string: %s\nOutput:\n%s", expected, output)
		}
	}
}

// TestRunGradeTaskCommand_NoFiles tests behavior with no changed files
func TestRunGradeTaskCommand_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskCommand("test-task", "feature", []string{}, tmpDir, "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskCommand should not error with no files, got: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should still produce valid output
	var result GradeTaskOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Some graders should be skipped
	hasSkipped := false
	for _, r := range result.Results {
		if r.Skipped {
			hasSkipped = true
			if r.SkipReason == "" {
				t.Error("Skipped grader should have skip reason")
			}
		}
	}

	if !hasSkipped {
		t.Error("Expected at least one grader to be skipped with no files")
	}
}

// TestRunGradeTaskCommand_InvalidInput tests handling of invalid input
func TestRunGradeTaskCommand_InvalidInput(t *testing.T) {
	testCases := []struct {
		name      string
		taskID    string
		taskType  string
		files     []string
		workDir   string
		format    string
		wantError bool
	}{
		{
			name:      "empty_task_id",
			taskID:    "",
			taskType:  "feature",
			files:     []string{},
			workDir:   ".",
			format:    "json",
			wantError: false, // Empty task ID is allowed
		},
		{
			name:      "invalid_format",
			taskID:    "test",
			taskType:  "feature",
			files:     []string{},
			workDir:   ".",
			format:    "xml",
			wantError: false, // Unknown format defaults to text
		},
		{
			name:      "nonexistent_workdir",
			taskID:    "test",
			taskType:  "feature",
			files:     []string{},
			workDir:   "/nonexistent/path/that/does/not/exist",
			format:    "json",
			wantError: false, // Graders handle this
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout to prevent test pollution
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskCommand(tc.taskID, tc.taskType, tc.files, tc.workDir, tc.format)

			w.Close()
			os.Stdout = oldStdout

			// Drain output
			var buf bytes.Buffer
			buf.ReadFrom(r)

			if tc.wantError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
