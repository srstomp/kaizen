package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunGradeCommand_CodeBasedGrader tests running a code-based grader
func TestRunGradeCommand_CodeBasedGrader(t *testing.T) {
	tmpDir := t.TempDir()

	// Create input JSON for code-based grader
	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := map[string]interface{}{
		"task_id":       "test-123",
		"task_type":     "feature",
		"changed_files": []string{"test.go", "test_test.go"},
		"work_dir":      tmpDir,
	}
	inputJSON, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Create the files referenced in changed_files
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

	err := runGradeCommand("file-exists", inputFile, "", "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify fields
	if result["grader"] != "file-exists" {
		t.Errorf("Expected grader 'file-exists', got %v", result["grader"])
	}
	if _, ok := result["passed"]; !ok {
		t.Error("Missing 'passed' field in output")
	}
	if _, ok := result["score"]; !ok {
		t.Error("Missing 'score' field in output")
	}
}

// TestRunGradeCommand_ModelBasedGrader tests running a model-based grader
func TestRunGradeCommand_ModelBasedGrader(t *testing.T) {
	tmpDir := t.TempDir()

	// Create input JSON for model-based grader
	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := map[string]interface{}{
		"content": "# Test Skill\n\nThis is a test skill with clear instructions.",
		"context": map[string]interface{}{
			"path": "test/SKILL.md",
		},
	}
	inputJSON, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeCommand("skill_clarity", inputFile, "", "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify fields
	if result["grader"] != "skill_clarity" {
		t.Errorf("Expected grader 'skill_clarity', got %v", result["grader"])
	}
	if _, ok := result["passed"]; !ok {
		t.Error("Missing 'passed' field in output")
	}
	if _, ok := result["score"]; !ok {
		t.Error("Missing 'score' field in output")
	}
	if _, ok := result["message"]; !ok {
		t.Error("Missing 'message' field in output")
	}
}

// TestRunGradeCommand_UnknownGrader tests error handling for unknown grader
func TestRunGradeCommand_UnknownGrader(t *testing.T) {
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := map[string]interface{}{
		"content": "test",
	}
	inputJSON, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := runGradeCommand("unknown-grader", inputFile, "", "json")

	if err == nil {
		t.Error("Expected error for unknown grader, got nil")
	}
	if !strings.Contains(err.Error(), "unknown grader") {
		t.Errorf("Expected 'unknown grader' error, got: %v", err)
	}
}

// TestRunGradeCommand_MissingInputFile tests error handling for missing input file
func TestRunGradeCommand_MissingInputFile(t *testing.T) {
	err := runGradeCommand("file-exists", "/nonexistent/file.json", "", "json")

	if err == nil {
		t.Error("Expected error for missing input file, got nil")
	}
}

// TestRunGradeCommand_MalformedJSON tests error handling for malformed JSON
func TestRunGradeCommand_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(inputFile, []byte("not valid json{"), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := runGradeCommand("file-exists", inputFile, "", "json")

	if err == nil {
		t.Error("Expected error for malformed JSON, got nil")
	}
}

// TestRunGradeCommand_TextFormat tests text output format
func TestRunGradeCommand_TextFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create input JSON
	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := map[string]interface{}{
		"task_id":       "test-123",
		"task_type":     "feature",
		"changed_files": []string{"test.go"},
		"work_dir":      tmpDir,
	}
	inputJSON, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Create the file
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeCommand("file-exists", inputFile, "", "text")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify text output contains expected sections
	expectedStrings := []string{
		"Grader:",
		"Result:",
		"Score:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Text output missing expected string: %s\nOutput:\n%s", expected, output)
		}
	}
}

// TestRunGradeCommand_SpecFlag tests --spec flag for model-based graders
func TestRunGradeCommand_SpecFlag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create input JSON for spec_compliance grader
	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := map[string]interface{}{
		"content": "Implementation details here",
		"context": map[string]interface{}{},
	}
	inputJSON, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeCommand("spec_compliance", inputFile, "Add user authentication", "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeCommand failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify grader ran
	if result["grader"] != "spec_compliance" {
		t.Errorf("Expected grader 'spec_compliance', got %v", result["grader"])
	}
}

// TestRunGradeCommand_HyphenUnderscore tests both hyphen and underscore variants
func TestRunGradeCommand_HyphenUnderscore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create input JSON
	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := map[string]interface{}{
		"task_id":       "test-123",
		"task_type":     "feature",
		"changed_files": []string{},
		"work_dir":      tmpDir,
	}
	inputJSON, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	testCases := []struct {
		name        string
		graderName  string
		shouldError bool
	}{
		{"hyphen_variant", "file-exists", false},
		{"underscore_variant", "file_exists", false},
		{"task_quality_hyphen", "task-quality", false},
		{"task_quality_underscore", "task_quality", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeCommand(tc.graderName, inputFile, "", "json")

			w.Close()
			os.Stdout = oldStdout

			// Drain output
			var buf bytes.Buffer
			buf.ReadFrom(r)

			if tc.shouldError && err == nil {
				t.Errorf("Expected error for grader %s, got nil", tc.graderName)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error for grader %s, got: %v", tc.graderName, err)
			}
		})
	}
}

// TestRunGradeCommand_FailingGrader tests that failures are properly reported
func TestRunGradeCommand_FailingGrader(t *testing.T) {
	tmpDir := t.TempDir()

	// Create input with no test files - should fail test-exists grader
	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := map[string]interface{}{
		"task_id":       "test-123",
		"task_type":     "feature",
		"changed_files": []string{"main.go"}, // No test file
		"work_dir":      tmpDir,
	}
	inputJSON, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Create the main file but no test
	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeCommand("test-exists", inputFile, "", "json")

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeCommand should not error on failing grade: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify it failed
	passed, ok := result["passed"].(bool)
	if !ok {
		t.Fatal("'passed' field should be boolean")
	}
	if passed {
		t.Error("Expected grader to fail (no test file), but it passed")
	}
}
