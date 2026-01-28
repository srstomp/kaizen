package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadEvalResults tests loading eval results from JSON log
func TestLoadEvalResults(t *testing.T) {
	tmpDir := t.TempDir()
	evalLog := filepath.Join(tmpDir, "task-eval-log.json")

	// Create sample eval log
	evalData := []GradeTaskOutput{
		{
			TaskID:        "test-task-001",
			Timestamp:     "2026-01-27T18:29:07Z",
			OverallPassed: true,
			OverallScore:  100.0,
		},
		{
			TaskID:        "test-task-002",
			Timestamp:     "2026-01-27T19:00:00Z",
			OverallPassed: true,
			OverallScore:  95.0,
		},
	}

	data, _ := json.Marshal(evalData)
	err := os.WriteFile(evalLog, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write eval log: %v", err)
	}

	// Execute
	results, err := loadEvalResults(evalLog)
	if err != nil {
		t.Fatalf("loadEvalResults failed: %v", err)
	}

	// Verify
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].OverallScore != 100.0 {
		t.Errorf("Expected score 100.0, got %.1f", results[0].OverallScore)
	}
}

// TestLoadMetaResults tests loading meta-eval results from JSON log
func TestLoadMetaResults(t *testing.T) {
	tmpDir := t.TempDir()
	metaLog := filepath.Join(tmpDir, "consistency-log.json")

	// Create sample meta log
	metaData := []ConsistencyResult{
		{
			Timestamp:              "2026-01-27T18:13:50Z",
			Agent:                  "yokay-quality-reviewer",
			BoundaryType:           "epic",
			ConsistencyPercentage:  95.0,
			ConsistentCount:        19,
			TotalCount:             20,
		},
		{
			Timestamp:              "2026-01-27T18:13:50Z",
			Agent:                  "yokay-spec-reviewer",
			BoundaryType:           "epic",
			ConsistencyPercentage:  88.0,
			ConsistentCount:        22,
			TotalCount:             25,
		},
	}

	data, _ := json.Marshal(metaData)
	err := os.WriteFile(metaLog, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write meta log: %v", err)
	}

	// Execute
	results, err := loadMetaResults(metaLog)
	if err != nil {
		t.Fatalf("loadMetaResults failed: %v", err)
	}

	// Verify
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].ConsistencyPercentage != 95.0 {
		t.Errorf("Expected consistency 95.0, got %.1f", results[0].ConsistencyPercentage)
	}
}

// TestCalculateEvalGateScore tests calculating average score from eval results
func TestCalculateEvalGateScore(t *testing.T) {
	tests := []struct {
		name          string
		results       []GradeTaskOutput
		expectedScore float64
	}{
		{
			name: "All pass",
			results: []GradeTaskOutput{
				{OverallScore: 100.0},
				{OverallScore: 100.0},
				{OverallScore: 100.0},
			},
			expectedScore: 100.0,
		},
		{
			name: "Mixed scores",
			results: []GradeTaskOutput{
				{OverallScore: 100.0},
				{OverallScore: 90.0},
				{OverallScore: 80.0},
			},
			expectedScore: 90.0,
		},
		{
			name: "Some failures",
			results: []GradeTaskOutput{
				{OverallScore: 100.0},
				{OverallScore: 50.0},
				{OverallScore: 0.0},
			},
			expectedScore: 50.0,
		},
		{
			name:          "Empty results",
			results:       []GradeTaskOutput{},
			expectedScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateEvalGateScore(tt.results)
			if score != tt.expectedScore {
				t.Errorf("Expected score %.1f, got %.1f", tt.expectedScore, score)
			}
		})
	}
}

// TestCalculateMetaGateScore tests calculating average consistency from meta results
func TestCalculateMetaGateScore(t *testing.T) {
	tests := []struct {
		name          string
		results       []ConsistencyResult
		expectedScore float64
	}{
		{
			name: "High consistency",
			results: []ConsistencyResult{
				{ConsistencyPercentage: 95.0},
				{ConsistencyPercentage: 98.0},
				{ConsistencyPercentage: 92.0},
			},
			expectedScore: 95.0,
		},
		{
			name: "Mixed consistency",
			results: []ConsistencyResult{
				{ConsistencyPercentage: 100.0},
				{ConsistencyPercentage: 80.0},
				{ConsistencyPercentage: 70.0},
			},
			expectedScore: 83.33, // Round to 2 decimals
		},
		{
			name:          "Empty results",
			results:       []ConsistencyResult{},
			expectedScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateMetaGateScore(tt.results)
			// Round to 2 decimal places for comparison
			scoreRounded := float64(int(score*100+0.5)) / 100
			if scoreRounded != tt.expectedScore {
				t.Errorf("Expected score %.2f, got %.2f", tt.expectedScore, scoreRounded)
			}
		})
	}
}

// TestCheckGate tests the gate check logic
func TestCheckGate(t *testing.T) {
	tests := []struct {
		name       string
		score      float64
		threshold  float64
		expectPass bool
	}{
		{
			name:       "Pass threshold exactly",
			score:      95.0,
			threshold:  95.0,
			expectPass: true,
		},
		{
			name:       "Pass above threshold",
			score:      98.0,
			threshold:  95.0,
			expectPass: true,
		},
		{
			name:       "Fail below threshold",
			score:      92.0,
			threshold:  95.0,
			expectPass: false,
		},
		{
			name:       "Perfect score",
			score:      100.0,
			threshold:  95.0,
			expectPass: true,
		},
		{
			name:       "Zero score",
			score:      0.0,
			threshold:  95.0,
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass := checkGate(tt.score, tt.threshold)
			if pass != tt.expectPass {
				t.Errorf("Expected pass=%v, got %v", tt.expectPass, pass)
			}
		})
	}
}

// TestRunGateCommand tests the gate command with eval type
func TestRunGateCommand_Eval(t *testing.T) {
	tmpDir := t.TempDir()
	evalLog := filepath.Join(tmpDir, "task-eval-log.json")

	// Create eval log with passing scores
	evalData := []GradeTaskOutput{
		{TaskID: "test-1", OverallScore: 100.0},
		{TaskID: "test-2", OverallScore: 96.0},
		{TaskID: "test-3", OverallScore: 98.0},
	}
	data, _ := json.Marshal(evalData)
	os.WriteFile(evalLog, data, 0644)

	// Execute - should pass
	err := runGateCommand("eval", 95.0, tmpDir)
	if err != nil {
		t.Errorf("Expected gate to pass, got error: %v", err)
	}
}

// TestRunGateCommand_EvalFail tests gate command failing
func TestRunGateCommand_EvalFail(t *testing.T) {
	tmpDir := t.TempDir()
	evalLog := filepath.Join(tmpDir, "task-eval-log.json")

	// Create eval log with failing scores
	evalData := []GradeTaskOutput{
		{TaskID: "test-1", OverallScore: 90.0},
		{TaskID: "test-2", OverallScore: 85.0},
	}
	data, _ := json.Marshal(evalData)
	os.WriteFile(evalLog, data, 0644)

	// Execute - should fail
	err := runGateCommand("eval", 95.0, tmpDir)
	if err == nil {
		t.Error("Expected gate to fail, but it passed")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("Expected error to contain 'failed', got: %v", err)
	}
}

// TestRunGateCommand_Meta tests gate command with meta type
func TestRunGateCommand_Meta(t *testing.T) {
	tmpDir := t.TempDir()
	metaLog := filepath.Join(tmpDir, "consistency-log.json")

	// Create meta log with passing consistency
	metaData := []ConsistencyResult{
		{Agent: "agent-1", ConsistencyPercentage: 96.0},
		{Agent: "agent-2", ConsistencyPercentage: 98.0},
	}
	data, _ := json.Marshal(metaData)
	os.WriteFile(metaLog, data, 0644)

	// Execute - should pass
	err := runGateCommand("meta", 95.0, tmpDir)
	if err != nil {
		t.Errorf("Expected gate to pass, got error: %v", err)
	}
}

// TestRunGateCommand_All tests gate command with both eval and meta
func TestRunGateCommand_All(t *testing.T) {
	tmpDir := t.TempDir()
	evalLog := filepath.Join(tmpDir, "task-eval-log.json")
	metaLog := filepath.Join(tmpDir, "consistency-log.json")

	// Create logs with passing scores
	evalData := []GradeTaskOutput{
		{TaskID: "test-1", OverallScore: 98.0},
	}
	metaData := []ConsistencyResult{
		{Agent: "agent-1", ConsistencyPercentage: 97.0},
	}

	evalJSON, _ := json.Marshal(evalData)
	metaJSON, _ := json.Marshal(metaData)
	os.WriteFile(evalLog, evalJSON, 0644)
	os.WriteFile(metaLog, metaJSON, 0644)

	// Execute - should pass both
	err := runGateCommand("all", 95.0, tmpDir)
	if err != nil {
		t.Errorf("Expected gate to pass, got error: %v", err)
	}
}

// TestRunGateCommand_AllPartialFail tests gate with one type failing
func TestRunGateCommand_AllPartialFail(t *testing.T) {
	tmpDir := t.TempDir()
	evalLog := filepath.Join(tmpDir, "task-eval-log.json")
	metaLog := filepath.Join(tmpDir, "consistency-log.json")

	// Eval passes, meta fails
	evalData := []GradeTaskOutput{
		{TaskID: "test-1", OverallScore: 98.0},
	}
	metaData := []ConsistencyResult{
		{Agent: "agent-1", ConsistencyPercentage: 85.0},
	}

	evalJSON, _ := json.Marshal(evalData)
	metaJSON, _ := json.Marshal(metaData)
	os.WriteFile(evalLog, evalJSON, 0644)
	os.WriteFile(metaLog, metaJSON, 0644)

	// Execute - should fail overall
	err := runGateCommand("all", 95.0, tmpDir)
	if err == nil {
		t.Error("Expected gate to fail, but it passed")
	}
}

// TestRunGateCommand_MissingFile tests error handling for missing files
func TestRunGateCommand_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Execute without creating log files - should error
	err := runGateCommand("eval", 95.0, tmpDir)
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

// TestRunGateCommand_InvalidType tests error handling for invalid type
func TestRunGateCommand_InvalidType(t *testing.T) {
	tmpDir := t.TempDir()

	// Execute with invalid type
	err := runGateCommand("invalid", 95.0, tmpDir)
	if err == nil {
		t.Error("Expected error for invalid type, got nil")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("Expected error about invalid type, got: %v", err)
	}
}

// TestRunGateCommand_InvalidThreshold tests error handling for invalid threshold
func TestRunGateCommand_InvalidThreshold(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		threshold float64
	}{
		{"Negative threshold", -5.0},
		{"Threshold over 100", 105.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runGateCommand("eval", tt.threshold, tmpDir)
			if err == nil {
				t.Error("Expected error for invalid threshold, got nil")
			}
			if !strings.Contains(err.Error(), "threshold") {
				t.Errorf("Expected error about threshold, got: %v", err)
			}
		})
	}
}

// TestFormatGateResult tests formatting gate check results
func TestFormatGateResult(t *testing.T) {
	tests := []struct {
		name     string
		checkType string
		score    float64
		threshold float64
		pass     bool
		contains []string
	}{
		{
			name:      "Pass eval",
			checkType: "eval",
			score:     98.0,
			threshold: 95.0,
			pass:      true,
			contains:  []string{"PASS", "eval", "98.0", "95.0"},
		},
		{
			name:      "Fail meta",
			checkType: "meta",
			score:     92.0,
			threshold: 95.0,
			pass:      false,
			contains:  []string{"FAIL", "meta", "92.0", "95.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGateResult(tt.checkType, tt.score, tt.threshold, tt.pass)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}
