package codebased

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTestExistsGraderInterface verifies TestExistsGrader implements CodeGrader
func TestTestExistsGraderInterface(t *testing.T) {
	var _ CodeGrader = (*TestExistsGrader)(nil)
}

// TestTestExistsGraderName verifies the grader name
func TestTestExistsGraderName(t *testing.T) {
	grader := NewTestExistsGrader()
	if grader.Name() != "test-exists" {
		t.Errorf("Expected name 'test-exists', got %s", grader.Name())
	}
}

// TestTestExistsGraderIsApplicable verifies applicability logic
func TestTestExistsGraderIsApplicable(t *testing.T) {
	grader := NewTestExistsGrader()

	tests := []struct {
		name     string
		input    GradeInput
		expected bool
	}{
		{
			name: "applicable when code files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"foo.go"},
				WorkDir:      "/tmp",
			},
			expected: true,
		},
		{
			name: "not applicable when only test files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"foo_test.go"},
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

// TestTestExistsGraderGoFiles verifies grading for Go files
func TestTestExistsGraderGoFiles(t *testing.T) {
	// Create temp directory with Go files and tests
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "foo.go")
	testFile := filepath.Join(tmpDir, "foo_test.go")

	if err := os.WriteFile(sourceFile, []byte("package main\n\nfunc Foo() {}"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("package main\n\nfunc TestFoo(t *testing.T) {}"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{sourceFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.GraderName != "test-exists" {
		t.Errorf("Expected GraderName 'test-exists', got %s", result.GraderName)
	}
	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
	if result.Skipped {
		t.Error("Expected Skipped to be false")
	}
}

// TestTestExistsGraderGoFilesMissingTest verifies grading when test missing
func TestTestExistsGraderGoFilesMissingTest(t *testing.T) {
	// Create temp directory with only source file
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "foo.go")

	if err := os.WriteFile(sourceFile, []byte("package main\n\nfunc Foo() {}"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{sourceFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.Passed {
		t.Error("Expected Passed to be false when test missing")
	}
	if result.Score != 0 {
		t.Errorf("Expected Score 0, got %f", result.Score)
	}
	if result.Details == "" {
		t.Error("Expected Details to contain information about missing test")
	}
}

// TestTestExistsGraderPythonFiles verifies grading for Python files
func TestTestExistsGraderPythonFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "bar.py")
	testFile := filepath.Join(tmpDir, "test_bar.py")

	if err := os.WriteFile(sourceFile, []byte("def bar():\n    pass"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("def test_bar():\n    pass"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{sourceFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
}

// TestTestExistsGraderPythonFilesAlternatePattern verifies alternate test pattern
func TestTestExistsGraderPythonFilesAlternatePattern(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "bar.py")
	testFile := filepath.Join(tmpDir, "bar_test.py")

	if err := os.WriteFile(sourceFile, []byte("def bar():\n    pass"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("def test_bar():\n    pass"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{sourceFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
}

// TestTestExistsGraderTypeScriptFiles verifies grading for TypeScript files
func TestTestExistsGraderTypeScriptFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "baz.ts")
	testFile := filepath.Join(tmpDir, "baz.test.ts")

	if err := os.WriteFile(sourceFile, []byte("export function baz() {}"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("test('baz', () => {})"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{sourceFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
}

// TestTestExistsGraderJavaScriptFiles verifies grading for JavaScript files
func TestTestExistsGraderJavaScriptFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "qux.js")
	testFile := filepath.Join(tmpDir, "qux.test.js")

	if err := os.WriteFile(sourceFile, []byte("export function qux() {}"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("test('qux', () => {})"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{sourceFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
}

// TestTestExistsGraderMixedFiles verifies grading with multiple files
func TestTestExistsGraderMixedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 2 files with tests and 1 without
	file1 := filepath.Join(tmpDir, "foo.go")
	test1 := filepath.Join(tmpDir, "foo_test.go")
	file2 := filepath.Join(tmpDir, "bar.py")
	test2 := filepath.Join(tmpDir, "test_bar.py")
	file3 := filepath.Join(tmpDir, "baz.ts")

	os.WriteFile(file1, []byte("package main"), 0644)
	os.WriteFile(test1, []byte("package main"), 0644)
	os.WriteFile(file2, []byte("def bar(): pass"), 0644)
	os.WriteFile(test2, []byte("def test_bar(): pass"), 0644)
	os.WriteFile(file3, []byte("export function baz() {}"), 0644)

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{file1, file2, file3},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.Passed {
		t.Error("Expected Passed to be false when some tests missing")
	}
	// Score should be 2/3 = 66.67%
	if result.Score < 66 || result.Score > 67 {
		t.Errorf("Expected Score around 66.67 (2/3 files with tests), got %f", result.Score)
	}
}

// TestTestExistsGraderSkipTestFiles verifies test files are skipped
func TestTestExistsGraderSkipTestFiles(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "foo_test.go")

	os.WriteFile(testFile, []byte("package main"), 0644)

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{testFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Skipped {
		t.Error("Expected Skipped to be true when only test files changed")
	}
	if result.SkipReason == "" {
		t.Error("Expected SkipReason to be provided")
	}
}

// TestTestExistsGraderSkipConfigFiles verifies config files are skipped
func TestTestExistsGraderSkipConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	os.WriteFile(configFile, []byte("key: value"), 0644)

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{configFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Skipped {
		t.Error("Expected Skipped to be true for config files")
	}
}

// TestTestExistsGraderSkipChores verifies chore tasks are skipped
func TestTestExistsGraderSkipChores(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "foo.go")

	os.WriteFile(sourceFile, []byte("package main"), 0644)

	grader := NewTestExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "chore",
		ChangedFiles: []string{sourceFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Skipped {
		t.Error("Expected Skipped to be true for chore tasks")
	}
	if result.SkipReason == "" {
		t.Error("Expected SkipReason to be provided")
	}
}
