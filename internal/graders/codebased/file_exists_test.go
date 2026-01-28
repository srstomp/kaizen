package codebased

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFileExistsGraderInterface verifies FileExistsGrader implements CodeGrader
func TestFileExistsGraderInterface(t *testing.T) {
	var _ CodeGrader = (*FileExistsGrader)(nil)
}

// TestFileExistsGraderName verifies the grader name
func TestFileExistsGraderName(t *testing.T) {
	grader := NewFileExistsGrader()
	if grader.Name() != "file-exists" {
		t.Errorf("Expected name 'file-exists', got %s", grader.Name())
	}
}

// TestFileExistsGraderIsApplicable verifies applicability logic
func TestFileExistsGraderIsApplicable(t *testing.T) {
	grader := NewFileExistsGrader()

	tests := []struct {
		name     string
		input    GradeInput
		expected bool
	}{
		{
			name: "applicable when files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"foo.go"},
				WorkDir:      "/tmp",
			},
			expected: true,
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

// TestFileExistsGraderAllFilesExist verifies grading when all files exist
func TestFileExistsGraderAllFilesExist(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.py")

	if err := os.WriteFile(file1, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("print('hello')"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	grader := NewFileExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{file1, file2},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.GraderName != "file-exists" {
		t.Errorf("Expected GraderName 'file-exists', got %s", result.GraderName)
	}
	if !result.Passed {
		t.Error("Expected Passed to be true")
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
	if result.Skipped {
		t.Error("Expected Skipped to be false")
	}
	if result.Details == "" {
		t.Error("Expected Details to contain information")
	}
}

// TestFileExistsGraderSomeFilesMissing verifies grading when some files missing
func TestFileExistsGraderSomeFilesMissing(t *testing.T) {
	// Create temp directory with only one test file
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.py")

	if err := os.WriteFile(file1, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	// file2 deliberately not created

	grader := NewFileExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{file1, file2},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.GraderName != "file-exists" {
		t.Errorf("Expected GraderName 'file-exists', got %s", result.GraderName)
	}
	if result.Passed {
		t.Error("Expected Passed to be false when files missing")
	}
	if result.Score != 50 {
		t.Errorf("Expected Score 50 (1/2 files exist), got %f", result.Score)
	}
	if result.Skipped {
		t.Error("Expected Skipped to be false")
	}
	if result.Details == "" {
		t.Error("Expected Details to contain information about missing files")
	}
}

// TestFileExistsGraderAllFilesMissing verifies grading when all files missing
func TestFileExistsGraderAllFilesMissing(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "missing1.go")
	file2 := filepath.Join(tmpDir, "missing2.py")

	grader := NewFileExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{file1, file2},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.GraderName != "file-exists" {
		t.Errorf("Expected GraderName 'file-exists', got %s", result.GraderName)
	}
	if result.Passed {
		t.Error("Expected Passed to be false when all files missing")
	}
	if result.Score != 0 {
		t.Errorf("Expected Score 0 (0/2 files exist), got %f", result.Score)
	}
	if result.Skipped {
		t.Error("Expected Skipped to be false")
	}
}

// TestFileExistsGraderSkipWhenNotApplicable verifies grader skips appropriately
func TestFileExistsGraderSkipWhenNotApplicable(t *testing.T) {
	grader := NewFileExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{},
		WorkDir:      "/tmp",
	}

	result := grader.Grade(input)

	if !result.Skipped {
		t.Error("Expected Skipped to be true when no files changed")
	}
	if result.SkipReason == "" {
		t.Error("Expected SkipReason to be provided")
	}
}

// TestFileExistsGraderRelativePaths verifies handling of relative paths
func TestFileExistsGraderRelativePaths(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.go")

	if err := os.WriteFile(file1, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	grader := NewFileExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{"file1.go"}, // Relative path
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	// Should resolve relative path against WorkDir
	if !result.Passed {
		t.Error("Expected Passed to be true for relative path that exists")
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
}
