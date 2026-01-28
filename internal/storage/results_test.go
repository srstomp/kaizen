package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/srstomp/kaizen/internal/graders/codebased"
)

func TestNewResultsStore(t *testing.T) {
	tempDir := t.TempDir()

	store := NewResultsStore(tempDir)

	if store == nil {
		t.Fatal("NewResultsStore returned nil")
	}

	if store.resultsDir != tempDir {
		t.Errorf("resultsDir = %s, expected %s", store.resultsDir, tempDir)
	}
}

func TestSaveAndLoadEvalResults(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	// Create test eval result
	result := &EvalResult{
		TaskID:    "task-123",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Results: []codebased.GradeResult{
			{
				GraderName: "test_grader",
				Passed:     true,
				Score:      95.0,
				Details:    "All tests passed",
				Skipped:    false,
			},
		},
		OverallPassed: true,
		OverallScore:  95.0,
	}

	// Save the result
	err := store.SaveEvalResult(result)
	if err != nil {
		t.Fatalf("SaveEvalResult failed: %v", err)
	}

	// Load results back
	results, err := store.LoadEvalResults()
	if err != nil {
		t.Fatalf("LoadEvalResults failed: %v", err)
	}

	// Verify
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].TaskID != result.TaskID {
		t.Errorf("TaskID = %s, expected %s", results[0].TaskID, result.TaskID)
	}

	if results[0].OverallPassed != result.OverallPassed {
		t.Errorf("OverallPassed = %v, expected %v", results[0].OverallPassed, result.OverallPassed)
	}

	if results[0].OverallScore != result.OverallScore {
		t.Errorf("OverallScore = %f, expected %f", results[0].OverallScore, result.OverallScore)
	}
}

func TestSaveEvalResultAppends(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	// Save first result
	result1 := &EvalResult{
		TaskID:        "task-1",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Results:       []codebased.GradeResult{},
		OverallPassed: true,
		OverallScore:  90.0,
	}

	err := store.SaveEvalResult(result1)
	if err != nil {
		t.Fatalf("SaveEvalResult failed: %v", err)
	}

	// Save second result
	result2 := &EvalResult{
		TaskID:        "task-2",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Results:       []codebased.GradeResult{},
		OverallPassed: false,
		OverallScore:  50.0,
	}

	err = store.SaveEvalResult(result2)
	if err != nil {
		t.Fatalf("SaveEvalResult failed: %v", err)
	}

	// Load and verify both results exist
	results, err := store.LoadEvalResults()
	if err != nil {
		t.Fatalf("LoadEvalResults failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify order (should be chronological)
	if results[0].TaskID != "task-1" {
		t.Errorf("first result TaskID = %s, expected task-1", results[0].TaskID)
	}

	if results[1].TaskID != "task-2" {
		t.Errorf("second result TaskID = %s, expected task-2", results[1].TaskID)
	}
}

func TestLoadEvalResultsEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	// Load from non-existent file should return empty slice, not error
	results, err := store.LoadEvalResults()
	if err != nil {
		t.Fatalf("LoadEvalResults failed on non-existent file: %v", err)
	}

	if results == nil {
		t.Fatal("LoadEvalResults returned nil slice")
	}

	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d results", len(results))
	}
}

func TestSaveAndLoadMetaResults(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	// Create test meta result
	result := &MetaResult{
		Timestamp:             time.Now().UTC().Format(time.RFC3339),
		Agent:                 "claude-4",
		BoundaryType:          "capability",
		ConsistencyPercentage: 85.5,
		ConsistentCount:       17,
		TotalCount:            20,
	}

	// Save the result
	err := store.SaveMetaResult(result)
	if err != nil {
		t.Fatalf("SaveMetaResult failed: %v", err)
	}

	// Load results back
	results, err := store.LoadMetaResults()
	if err != nil {
		t.Fatalf("LoadMetaResults failed: %v", err)
	}

	// Verify
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Agent != result.Agent {
		t.Errorf("Agent = %s, expected %s", results[0].Agent, result.Agent)
	}

	if results[0].ConsistencyPercentage != result.ConsistencyPercentage {
		t.Errorf("ConsistencyPercentage = %f, expected %f", results[0].ConsistencyPercentage, result.ConsistencyPercentage)
	}
}

func TestSaveAndLoadTaskQualityResults(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	// Create test task quality result
	result := &TaskQualityResult{
		TaskID: "task-456",
		Passed: true,
		Score:  88.0,
		Issues: []TaskQualityIssue{
			{
				Check:   "description_length",
				Message: "Description is clear",
			},
		},
		Suggestion: "Consider adding more acceptance criteria",
	}

	// Save the result
	err := store.SaveTaskQualityResult(result)
	if err != nil {
		t.Fatalf("SaveTaskQualityResult failed: %v", err)
	}

	// Load results back
	results, err := store.LoadTaskQualityResults()
	if err != nil {
		t.Fatalf("LoadTaskQualityResults failed: %v", err)
	}

	// Verify
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].TaskID != result.TaskID {
		t.Errorf("TaskID = %s, expected %s", results[0].TaskID, result.TaskID)
	}

	if results[0].Score != result.Score {
		t.Errorf("Score = %f, expected %f", results[0].Score, result.Score)
	}

	if len(results[0].Issues) != len(result.Issues) {
		t.Errorf("Issues length = %d, expected %d", len(results[0].Issues), len(result.Issues))
	}
}

func TestSaveEvalResultCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	resultsDir := filepath.Join(tempDir, "subdir", "results")
	store := NewResultsStore(resultsDir)

	result := &EvalResult{
		TaskID:        "task-789",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Results:       []codebased.GradeResult{},
		OverallPassed: true,
		OverallScore:  100.0,
	}

	// Save should create directory if it doesn't exist
	err := store.SaveEvalResult(result)
	if err != nil {
		t.Fatalf("SaveEvalResult failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		t.Errorf("results directory was not created: %s", resultsDir)
	}

	// Verify file exists
	evalFile := filepath.Join(resultsDir, "eval-results.json")
	if _, err := os.Stat(evalFile); os.IsNotExist(err) {
		t.Errorf("eval results file was not created: %s", evalFile)
	}
}

func TestLoadResultsInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	// Create invalid JSON file
	evalFile := filepath.Join(tempDir, "eval-results.json")
	err := os.WriteFile(evalFile, []byte("invalid json{]"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Load should return error for invalid JSON
	_, err = store.LoadEvalResults()
	if err == nil {
		t.Error("LoadEvalResults should return error for invalid JSON")
	}
}

func TestTimestampAutoGeneration(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	// Create result without timestamp
	result := &EvalResult{
		TaskID:        "task-no-timestamp",
		Timestamp:     "", // Empty timestamp
		Results:       []codebased.GradeResult{},
		OverallPassed: true,
		OverallScore:  75.0,
	}

	beforeSave := time.Now().UTC()

	err := store.SaveEvalResult(result)
	if err != nil {
		t.Fatalf("SaveEvalResult failed: %v", err)
	}

	afterSave := time.Now().UTC()

	// Load and verify timestamp was added
	results, err := store.LoadEvalResults()
	if err != nil {
		t.Fatalf("LoadEvalResults failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Timestamp == "" {
		t.Error("Timestamp should have been auto-generated")
	}

	// Parse and verify timestamp is reasonable
	savedTime, err := time.Parse(time.RFC3339, results[0].Timestamp)
	if err != nil {
		t.Errorf("Timestamp is not valid RFC3339: %v", err)
	}

	// Allow 1 second buffer for timing differences
	beforeSaveWithBuffer := beforeSave.Add(-1 * time.Second)
	afterSaveWithBuffer := afterSave.Add(1 * time.Second)

	if savedTime.Before(beforeSaveWithBuffer) || savedTime.After(afterSaveWithBuffer) {
		t.Errorf("Timestamp %v is outside expected range [%v, %v]", savedTime, beforeSaveWithBuffer, afterSaveWithBuffer)
	}
}

func TestJSONFormatting(t *testing.T) {
	tempDir := t.TempDir()
	store := NewResultsStore(tempDir)

	result := &EvalResult{
		TaskID:        "task-format",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Results:       []codebased.GradeResult{},
		OverallPassed: true,
		OverallScore:  80.0,
	}

	err := store.SaveEvalResult(result)
	if err != nil {
		t.Fatalf("SaveEvalResult failed: %v", err)
	}

	// Read raw file and check it's indented
	evalFile := filepath.Join(tempDir, "eval-results.json")
	content, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Verify it's valid JSON and indented
	var results []EvalResult
	err = json.Unmarshal(content, &results)
	if err != nil {
		t.Errorf("file content is not valid JSON: %v", err)
	}

	// Check for indentation (presence of newlines and spaces)
	contentStr := string(content)
	if !containsIndentation(contentStr) {
		t.Error("JSON should be indented for readability")
	}
}

// Helper function to check if JSON is indented
func containsIndentation(s string) bool {
	// Indented JSON will have newlines and multiple spaces
	hasNewlines := false
	hasSpaces := false

	for i, c := range s {
		if c == '\n' {
			hasNewlines = true
		}
		if i > 0 && c == ' ' && s[i-1] == '\n' {
			hasSpaces = true
		}
	}

	return hasNewlines && hasSpaces
}
