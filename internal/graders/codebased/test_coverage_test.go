package codebased

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTestCoverageGraderInterface verifies TestCoverageGrader implements CodeGrader
func TestTestCoverageGraderInterface(t *testing.T) {
	var _ CodeGrader = (*TestCoverageGrader)(nil)
}

// TestTestCoverageGraderName verifies the grader name
func TestTestCoverageGraderName(t *testing.T) {
	grader := NewTestCoverageGrader()
	if grader.Name() != "test-coverage" {
		t.Errorf("Expected name 'test-coverage', got %s", grader.Name())
	}
}

// TestTestCoverageGraderIsApplicable verifies applicability logic
func TestTestCoverageGraderIsApplicable(t *testing.T) {
	grader := NewTestCoverageGrader()

	tests := []struct {
		name     string
		input    GradeInput
		expected bool
	}{
		{
			name: "applicable when Go files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"foo.go"},
				WorkDir:      "/tmp",
			},
			expected: true,
		},
		{
			name: "applicable when Go test files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"foo_test.go"},
				WorkDir:      "/tmp",
			},
			expected: true,
		},
		{
			name: "not applicable when no Go files",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"foo.py", "bar.js"},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
		{
			name: "not applicable for chore tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "chore",
				ChangedFiles: []string{"foo.go"},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
		{
			name: "not applicable for spike tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "spike",
				ChangedFiles: []string{"foo.go"},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
		{
			name: "not applicable when no files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grader.IsApplicable(tt.input)
			if result != tt.expected {
				t.Errorf("Expected IsApplicable to be %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestTestCoverageGraderSkipReasons verifies skip reasons are correct
func TestTestCoverageGraderSkipReasons(t *testing.T) {
	grader := NewTestCoverageGrader()

	tests := []struct {
		name               string
		input              GradeInput
		expectedSkipped    bool
		expectedSkipReason string
	}{
		{
			name: "skip chore tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "chore",
				ChangedFiles: []string{"foo.go"},
				WorkDir:      "/tmp",
			},
			expectedSkipped:    true,
			expectedSkipReason: "Not applicable for chore tasks",
		},
		{
			name: "skip spike tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "spike",
				ChangedFiles: []string{"foo.go"},
				WorkDir:      "/tmp",
			},
			expectedSkipped:    true,
			expectedSkipReason: "Not applicable for spike tasks",
		},
		{
			name: "skip non-Go files",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"foo.py", "bar.js"},
				WorkDir:      "/tmp",
			},
			expectedSkipped:    true,
			expectedSkipReason: "No Go files to check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grader.Grade(tt.input)
			if !result.Skipped {
				t.Errorf("Expected Skipped to be true")
			}
			if result.SkipReason != tt.expectedSkipReason {
				t.Errorf("Expected SkipReason '%s', got '%s'", tt.expectedSkipReason, result.SkipReason)
			}
		})
	}
}

// TestTestCoverageGraderWithRealGoProject verifies grading with an actual Go project
func TestTestCoverageGraderWithRealGoProject(t *testing.T) {
	// Create a temporary Go project
	tmpDir := t.TempDir()

	// Create a simple Go module
	goModContent := `module example.com/test
go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file
	sourceContent := `package main

func Add(a, b int) int {
	return a + b
}

func Subtract(a, b int) int {
	return a - b
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create a test file with good coverage
	testContent := `package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Error("Add failed")
	}
}

func TestSubtract(t *testing.T) {
	if Subtract(5, 3) != 2 {
		t.Error("Subtract failed")
	}
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create main_test.go: %v", err)
	}

	grader := NewTestCoverageGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{"main.go"},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	// Verify the result structure
	if result.GraderName != "test-coverage" {
		t.Errorf("Expected GraderName 'test-coverage', got %s", result.GraderName)
	}
	if result.Skipped {
		t.Errorf("Expected Skipped to be false, got skip reason: %s", result.SkipReason)
	}

	// Should pass with 100% coverage (both functions covered)
	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100.0 {
		t.Errorf("Expected Score 100.0, got %f", result.Score)
	}
	if result.Details == "" {
		t.Error("Expected Details to contain coverage information")
	}
}

// TestTestCoverageGraderWithPartialCoverage verifies grading with partial coverage
func TestTestCoverageGraderWithPartialCoverage(t *testing.T) {
	// Create a temporary Go project
	tmpDir := t.TempDir()

	// Create a simple Go module
	goModContent := `module example.com/test
go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file with multiple functions
	sourceContent := `package main

func Add(a, b int) int {
	return a + b
}

func Subtract(a, b int) int {
	return a - b
}

func Multiply(a, b int) int {
	return a * b
}

func Divide(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create a test file with partial coverage (only testing Add)
	testContent := `package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Error("Add failed")
	}
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create main_test.go: %v", err)
	}

	grader := NewTestCoverageGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{"main.go"},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	// Verify the result structure
	if result.GraderName != "test-coverage" {
		t.Errorf("Expected GraderName 'test-coverage', got %s", result.GraderName)
	}
	if result.Skipped {
		t.Errorf("Expected Skipped to be false, got skip reason: %s", result.SkipReason)
	}

	// Coverage should be less than 80% (only 1 out of 4 functions tested)
	// Should fail the 80% threshold
	if result.Passed {
		t.Errorf("Expected Passed to be false for partial coverage, details: %s", result.Details)
	}
	if result.Score >= 80.0 {
		t.Errorf("Expected Score < 80.0, got %f", result.Score)
	}
	if result.Score == 0 {
		t.Error("Expected Score > 0 for partial coverage")
	}
}

// TestTestCoverageGraderNoTestFiles verifies behavior when no test files exist
func TestTestCoverageGraderNoTestFiles(t *testing.T) {
	// Create a temporary Go project
	tmpDir := t.TempDir()

	// Create a simple Go module
	goModContent := `module example.com/test
go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file without tests
	sourceContent := `package main

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	grader := NewTestCoverageGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{"main.go"},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	// Should fail with 0 score
	if result.Passed {
		t.Error("Expected Passed to be false when no test files")
	}
	if result.Score != 0 {
		t.Errorf("Expected Score 0, got %f", result.Score)
	}
	if result.Details != "No test files found" {
		t.Errorf("Expected Details 'No test files found', got '%s'", result.Details)
	}
}

// TestTestCoverageGraderInvalidWorkDir verifies error handling for invalid work directory
func TestTestCoverageGraderInvalidWorkDir(t *testing.T) {
	grader := NewTestCoverageGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{"main.go"},
		WorkDir:      "/nonexistent/directory/path",
	}

	result := grader.Grade(input)

	// Should handle error gracefully
	if result.Passed {
		t.Error("Expected Passed to be false for invalid work directory")
	}
	if result.Score != 0 {
		t.Errorf("Expected Score 0, got %f", result.Score)
	}
	if result.Details == "" {
		t.Error("Expected Details to contain error information")
	}
}

// TestTestCoverageGraderThresholdBoundary verifies behavior at threshold boundaries
func TestTestCoverageGraderThresholdBoundary(t *testing.T) {
	// This test verifies the scoring logic by checking the threshold evaluation
	// We'll create a project with known coverage and verify the pass/fail logic

	grader := NewTestCoverageGrader()

	// Test the threshold value is 80.0
	threshold := 80.0

	// Above threshold should pass
	if !grader.evaluateCoverage(85.5, threshold) {
		t.Error("Expected coverage 85.5% to pass with threshold 80.0%")
	}

	// At threshold should pass
	if !grader.evaluateCoverage(80.0, threshold) {
		t.Error("Expected coverage 80.0% to pass with threshold 80.0%")
	}

	// Below threshold should fail
	if grader.evaluateCoverage(79.9, threshold) {
		t.Error("Expected coverage 79.9% to fail with threshold 80.0%")
	}
}
