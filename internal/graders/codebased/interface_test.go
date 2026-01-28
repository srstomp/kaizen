package codebased

import (
	"testing"
)

// TestGradeInputStructure verifies GradeInput struct has expected fields
func TestGradeInputStructure(t *testing.T) {
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{"foo.go", "bar.py"},
		WorkDir:      "/tmp/test",
	}

	if input.TaskID != "task-123" {
		t.Errorf("Expected TaskID to be 'task-123', got %s", input.TaskID)
	}
	if input.TaskType != "feature" {
		t.Errorf("Expected TaskType to be 'feature', got %s", input.TaskType)
	}
	if len(input.ChangedFiles) != 2 {
		t.Errorf("Expected 2 changed files, got %d", len(input.ChangedFiles))
	}
	if input.WorkDir != "/tmp/test" {
		t.Errorf("Expected WorkDir to be '/tmp/test', got %s", input.WorkDir)
	}
}

// TestGradeResultStructure verifies GradeResult struct has expected fields
func TestGradeResultStructure(t *testing.T) {
	result := GradeResult{
		GraderName: "test-grader",
		Passed:     true,
		Score:      85.5,
		Details:    "All checks passed",
		Skipped:    false,
		SkipReason: "",
	}

	if result.GraderName != "test-grader" {
		t.Errorf("Expected GraderName to be 'test-grader', got %s", result.GraderName)
	}
	if !result.Passed {
		t.Error("Expected Passed to be true")
	}
	if result.Score != 85.5 {
		t.Errorf("Expected Score to be 85.5, got %f", result.Score)
	}
	if result.Details != "All checks passed" {
		t.Errorf("Expected Details to be 'All checks passed', got %s", result.Details)
	}
	if result.Skipped {
		t.Error("Expected Skipped to be false")
	}
}

// TestGradeResultSkipped verifies skipped result structure
func TestGradeResultSkipped(t *testing.T) {
	result := GradeResult{
		GraderName: "test-grader",
		Passed:     false,
		Score:      0,
		Details:    "",
		Skipped:    true,
		SkipReason: "Not applicable for this task type",
	}

	if !result.Skipped {
		t.Error("Expected Skipped to be true")
	}
	if result.SkipReason != "Not applicable for this task type" {
		t.Errorf("Expected SkipReason, got %s", result.SkipReason)
	}
}

// MockCodeGrader implements CodeGrader for testing
type MockCodeGrader struct {
	NameValue         string
	GradeResult       GradeResult
	IsApplicableValue bool
}

func (m *MockCodeGrader) Name() string {
	return m.NameValue
}

func (m *MockCodeGrader) Grade(input GradeInput) GradeResult {
	return m.GradeResult
}

func (m *MockCodeGrader) IsApplicable(input GradeInput) bool {
	return m.IsApplicableValue
}

// TestCodeGraderInterface verifies that MockCodeGrader implements CodeGrader
func TestCodeGraderInterface(t *testing.T) {
	var _ CodeGrader = (*MockCodeGrader)(nil)
}

// TestCodeGraderMethods verifies CodeGrader methods work as expected
func TestCodeGraderMethods(t *testing.T) {
	grader := &MockCodeGrader{
		NameValue: "mock-grader",
		GradeResult: GradeResult{
			GraderName: "mock-grader",
			Passed:     true,
			Score:      100,
			Details:    "Mock grading successful",
		},
		IsApplicableValue: true,
	}

	// Test Name method
	if grader.Name() != "mock-grader" {
		t.Errorf("Expected Name to return 'mock-grader', got %s", grader.Name())
	}

	// Test IsApplicable method
	input := GradeInput{
		TaskID:   "task-123",
		TaskType: "feature",
	}
	if !grader.IsApplicable(input) {
		t.Error("Expected IsApplicable to return true")
	}

	// Test Grade method
	result := grader.Grade(input)
	if result.GraderName != "mock-grader" {
		t.Errorf("Expected GraderName to be 'mock-grader', got %s", result.GraderName)
	}
	if !result.Passed {
		t.Error("Expected Passed to be true")
	}
	if result.Score != 100 {
		t.Errorf("Expected Score to be 100, got %f", result.Score)
	}
}
