package main

import (
	"math"
	"os"
	"testing"
)

// TestCalculateDelta verifies delta calculation for trend analysis
func TestCalculateDelta(t *testing.T) {
	tests := []struct {
		name              string
		previous          float64
		current           float64
		expectedAbsolute  float64
		expectedPercent   float64
		expectedDirection string
	}{
		{
			name:              "improvement",
			previous:          85.0,
			current:           90.0,
			expectedAbsolute:  5.0,
			expectedPercent:   5.88,
			expectedDirection: "improvement",
		},
		{
			name:              "regression",
			previous:          90.0,
			current:           85.0,
			expectedAbsolute:  -5.0,
			expectedPercent:   -5.56,
			expectedDirection: "regression",
		},
		{
			name:              "stable",
			previous:          85.0,
			current:           85.0,
			expectedAbsolute:  0.0,
			expectedPercent:   0.0,
			expectedDirection: "stable",
		},
		{
			name:              "zero previous",
			previous:          0.0,
			current:           50.0,
			expectedAbsolute:  50.0,
			expectedPercent:   0.0, // Can't calculate percentage from zero
			expectedDirection: "improvement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := calculateDelta(tt.previous, tt.current)

			// Check absolute delta
			if math.Abs(delta.AbsoluteDelta-tt.expectedAbsolute) > 0.01 {
				t.Errorf("Expected absolute delta %.2f, got %.2f", tt.expectedAbsolute, delta.AbsoluteDelta)
			}

			// Check percentage delta (with some tolerance for floating point)
			if math.Abs(delta.PercentageDelta-tt.expectedPercent) > 0.01 {
				t.Errorf("Expected percentage delta %.2f, got %.2f", tt.expectedPercent, delta.PercentageDelta)
			}

			// Check direction
			if delta.Direction != tt.expectedDirection {
				t.Errorf("Expected direction %s, got %s", tt.expectedDirection, delta.Direction)
			}
		})
	}
}

// TestLoadEvalTrends verifies loading eval trend data from task-eval-log.json
func TestLoadEvalTrends(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/task-eval-log.json"

	// Create test log with two entries
	logContent := `[
		{
			"task_id": "task-001",
			"timestamp": "2026-01-26T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 85.0
		},
		{
			"task_id": "task-002",
			"timestamp": "2026-01-27T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 90.0
		}
	]`

	if err := writeFile(logPath, logContent); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Load trends
	trends, err := loadEvalTrends(logPath)
	if err != nil {
		t.Fatalf("loadEvalTrends failed: %v", err)
	}

	// Should have aggregate trends
	if trends.AverageScore.PreviousValue != 85.0 {
		t.Errorf("Expected previous average score 85.0, got %.1f", trends.AverageScore.PreviousValue)
	}
	if trends.AverageScore.CurrentValue != 90.0 {
		t.Errorf("Expected current average score 90.0, got %.1f", trends.AverageScore.CurrentValue)
	}

	// Check direction
	if trends.AverageScore.Direction != "improvement" {
		t.Errorf("Expected direction 'improvement', got %s", trends.AverageScore.Direction)
	}
}

// TestLoadMetaTrends verifies loading meta trend data from consistency-log.json
func TestLoadMetaTrends(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/consistency-log.json"

	// Create test log with entries for same agent
	logContent := `[
		{
			"timestamp": "2026-01-26T10:00:00Z",
			"agent": "yokay-quality-reviewer",
			"boundary_type": "epic",
			"consistency_percentage": 92.0,
			"consistent_count": 23,
			"total_count": 25
		},
		{
			"timestamp": "2026-01-27T10:00:00Z",
			"agent": "yokay-quality-reviewer",
			"boundary_type": "epic",
			"consistency_percentage": 95.0,
			"consistent_count": 19,
			"total_count": 20
		}
	]`

	if err := writeFile(logPath, logContent); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Load trends
	trends, err := loadMetaTrends(logPath)
	if err != nil {
		t.Fatalf("loadMetaTrends failed: %v", err)
	}

	// Should have consistency trend
	if trends.ConsistencyPercentage.PreviousValue != 92.0 {
		t.Errorf("Expected previous consistency 92.0, got %.1f", trends.ConsistencyPercentage.PreviousValue)
	}
	if trends.ConsistencyPercentage.CurrentValue != 95.0 {
		t.Errorf("Expected current consistency 95.0, got %.1f", trends.ConsistencyPercentage.CurrentValue)
	}
}

// TestLoadMetaTrendsWithMultipleAgents verifies per-agent trend tracking
func TestLoadMetaTrendsWithMultipleAgents(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/consistency-log.json"

	// Create test log with entries for multiple agents
	logContent := `[
		{
			"timestamp": "2026-01-26T10:00:00Z",
			"agent": "yokay-quality-reviewer",
			"boundary_type": "epic",
			"consistency_percentage": 80.0,
			"consistent_count": 16,
			"total_count": 20
		},
		{
			"timestamp": "2026-01-27T10:00:00Z",
			"agent": "yokay-quality-reviewer",
			"boundary_type": "epic",
			"consistency_percentage": 85.0,
			"consistent_count": 17,
			"total_count": 20
		},
		{
			"timestamp": "2026-01-26T10:00:00Z",
			"agent": "yokay-spec-reviewer",
			"boundary_type": "epic",
			"consistency_percentage": 90.0,
			"consistent_count": 18,
			"total_count": 20
		},
		{
			"timestamp": "2026-01-27T10:00:00Z",
			"agent": "yokay-spec-reviewer",
			"boundary_type": "epic",
			"consistency_percentage": 92.0,
			"consistent_count": 23,
			"total_count": 25
		}
	]`

	if err := writeFile(logPath, logContent); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Load trends
	trends, err := loadMetaTrends(logPath)
	if err != nil {
		t.Fatalf("loadMetaTrends failed: %v", err)
	}

	// Verify per-agent trends exist
	if len(trends.PerAgentTrends) != 2 {
		t.Errorf("Expected 2 agent trends, got %d", len(trends.PerAgentTrends))
	}

	// Verify yokay-quality-reviewer trend
	qualityTrend, exists := trends.PerAgentTrends["yokay-quality-reviewer"]
	if !exists {
		t.Fatal("Expected trend for yokay-quality-reviewer")
	}
	if qualityTrend.PreviousValue != 80.0 {
		t.Errorf("Expected previous value 80.0 for quality-reviewer, got %.1f", qualityTrend.PreviousValue)
	}
	if qualityTrend.CurrentValue != 85.0 {
		t.Errorf("Expected current value 85.0 for quality-reviewer, got %.1f", qualityTrend.CurrentValue)
	}
	if qualityTrend.Direction != "improvement" {
		t.Errorf("Expected direction 'improvement' for quality-reviewer, got %s", qualityTrend.Direction)
	}

	// Verify yokay-spec-reviewer trend
	specTrend, exists := trends.PerAgentTrends["yokay-spec-reviewer"]
	if !exists {
		t.Fatal("Expected trend for yokay-spec-reviewer")
	}
	if specTrend.PreviousValue != 90.0 {
		t.Errorf("Expected previous value 90.0 for spec-reviewer, got %.1f", specTrend.PreviousValue)
	}
	if specTrend.CurrentValue != 92.0 {
		t.Errorf("Expected current value 92.0 for spec-reviewer, got %.1f", specTrend.CurrentValue)
	}
	if specTrend.Direction != "improvement" {
		t.Errorf("Expected direction 'improvement' for spec-reviewer, got %s", specTrend.Direction)
	}
}

// TestLoadGradeTrends verifies loading grade trend data from skill-clarity reports
func TestLoadGradeTrends(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two skill-clarity reports
	oldReport := tmpDir + "/skill-clarity-2026-01-25.md"
	newReport := tmpDir + "/skill-clarity-2026-01-26.md"

	oldContent := `# Skill Clarity Report

Generated: 2026-01-25 10:00:00

## Summary

- **Total Skills**: 10
- **Average Score**: 75.0/100
- **Pass Rate**: 80.0% (8/10)
- **Passing Threshold**: 70.0
`

	newContent := `# Skill Clarity Report

Generated: 2026-01-26 10:00:00

## Summary

- **Total Skills**: 10
- **Average Score**: 80.0/100
- **Pass Rate**: 90.0% (9/10)
- **Passing Threshold**: 70.0
`

	if err := writeFile(oldReport, oldContent); err != nil {
		t.Fatalf("Failed to write old report: %v", err)
	}
	if err := writeFile(newReport, newContent); err != nil {
		t.Fatalf("Failed to write new report: %v", err)
	}

	// Load trends
	trends, err := loadGradeTrends(tmpDir)
	if err != nil {
		t.Fatalf("loadGradeTrends failed: %v", err)
	}

	// Should have average score trend
	if trends.AverageScore.PreviousValue != 75.0 {
		t.Errorf("Expected previous average score 75.0, got %.1f", trends.AverageScore.PreviousValue)
	}
	if trends.AverageScore.CurrentValue != 80.0 {
		t.Errorf("Expected current average score 80.0, got %.1f", trends.AverageScore.CurrentValue)
	}

	// Should have pass rate trend
	if trends.PassRate.PreviousValue != 80.0 {
		t.Errorf("Expected previous pass rate 80.0, got %.1f", trends.PassRate.PreviousValue)
	}
	if trends.PassRate.CurrentValue != 90.0 {
		t.Errorf("Expected current pass rate 90.0, got %.1f", trends.PassRate.CurrentValue)
	}
}

// Helper function to write file content
func writeFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// TestFormatTrendMarkdown verifies markdown formatting of trend data
func TestFormatTrendMarkdown(t *testing.T) {
	trend := TrendData{
		PreviousValue:   85.0,
		CurrentValue:    90.0,
		AbsoluteDelta:   5.0,
		PercentageDelta: 5.88,
		Direction:       "improvement",
	}

	output := formatTrendMarkdown("Average Score", trend, false)

	// Verify output contains expected elements
	expectedStrings := []string{
		"85.0",
		"90.0",
		"+5.0",
		"+5.88%",
		"↑",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// TestFormatTrendMarkdownRegression verifies regression formatting
func TestFormatTrendMarkdownRegression(t *testing.T) {
	trend := TrendData{
		PreviousValue:   90.0,
		CurrentValue:    85.0,
		AbsoluteDelta:   -5.0,
		PercentageDelta: -5.56,
		Direction:       "regression",
	}

	output := formatTrendMarkdown("Average Score", trend, false)

	// Verify output contains regression indicators
	expectedStrings := []string{
		"90.0",
		"85.0",
		"-5.0",
		"-5.56%",
		"↓",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// TestFormatTrendMarkdownWithWarning verifies warning indicators for regressions exceeding threshold
func TestFormatTrendMarkdownWithWarning(t *testing.T) {
	trend := TrendData{
		PreviousValue:   90.0,
		CurrentValue:    80.0,
		AbsoluteDelta:   -10.0,
		PercentageDelta: -11.11,
		Direction:       "regression",
	}

	// With warning (exceeds 5% threshold)
	output := formatTrendMarkdown("Average Score", trend, true)

	// Should contain warning indicator
	if !contains(output, "⚠") && !contains(output, "WARNING") {
		t.Errorf("Expected warning indicator in output: %s", output)
	}
}

// TestFormatTrendJSON verifies JSON formatting of trend data
func TestFormatTrendJSON(t *testing.T) {
	trend := TrendData{
		PreviousValue:   85.0,
		CurrentValue:    90.0,
		AbsoluteDelta:   5.0,
		PercentageDelta: 5.88,
		Direction:       "improvement",
	}

	// Verify the struct can be marshaled
	if trend.PreviousValue != 85.0 {
		t.Errorf("Expected previous value 85.0, got %.2f", trend.PreviousValue)
	}
	if trend.Direction != "improvement" {
		t.Errorf("Expected direction 'improvement', got %s", trend.Direction)
	}
}

// TestNoTrendDataAvailable verifies handling when no previous data exists
func TestNoTrendDataAvailable(t *testing.T) {
	// When there's no previous data, trend should indicate this
	output := formatNoTrendData()

	if !contains(output, "No trend data available") && !contains(output, "first run") {
		t.Errorf("Expected message about no trend data, got: %s", output)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestLoadEvalTrendsWithPerTaskTracking verifies per-task trend tracking
func TestLoadEvalTrendsWithPerTaskTracking(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/task-eval-log.json"

	// Create test log with multiple tasks across two time periods
	logContent := `[
		{
			"task_id": "task-001",
			"timestamp": "2026-01-26T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 80.0
		},
		{
			"task_id": "task-002",
			"timestamp": "2026-01-26T10:00:00Z",
			"results": [],
			"overall_passed": false,
			"overall_score": 60.0
		},
		{
			"task_id": "task-001",
			"timestamp": "2026-01-27T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 90.0
		},
		{
			"task_id": "task-002",
			"timestamp": "2026-01-27T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 85.0
		}
	]`

	if err := writeFile(logPath, logContent); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Load trends
	trends, err := loadEvalTrends(logPath)
	if err != nil {
		t.Fatalf("loadEvalTrends failed: %v", err)
	}

	// Verify aggregate average score trend
	// Previous: (80 + 60) / 2 = 70.0
	// Current: (90 + 85) / 2 = 87.5
	if trends.AverageScore.PreviousValue != 70.0 {
		t.Errorf("Expected previous average score 70.0, got %.1f", trends.AverageScore.PreviousValue)
	}
	if trends.AverageScore.CurrentValue != 87.5 {
		t.Errorf("Expected current average score 87.5, got %.1f", trends.AverageScore.CurrentValue)
	}

	// Verify aggregate pass rate trend
	// Previous: 1/2 = 50%
	// Current: 2/2 = 100%
	if trends.PassRate.PreviousValue != 50.0 {
		t.Errorf("Expected previous pass rate 50.0, got %.1f", trends.PassRate.PreviousValue)
	}
	if trends.PassRate.CurrentValue != 100.0 {
		t.Errorf("Expected current pass rate 100.0, got %.1f", trends.PassRate.CurrentValue)
	}

	// Verify per-task trends exist
	if len(trends.PerTaskTrends) != 2 {
		t.Errorf("Expected 2 task trends, got %d", len(trends.PerTaskTrends))
	}

	// Verify task-001 trend
	task1Trend, exists := trends.PerTaskTrends["task-001"]
	if !exists {
		t.Fatal("Expected trend for task-001")
	}
	if task1Trend.PreviousValue != 80.0 {
		t.Errorf("Expected previous value 80.0 for task-001, got %.1f", task1Trend.PreviousValue)
	}
	if task1Trend.CurrentValue != 90.0 {
		t.Errorf("Expected current value 90.0 for task-001, got %.1f", task1Trend.CurrentValue)
	}
	if task1Trend.Direction != "improvement" {
		t.Errorf("Expected direction 'improvement' for task-001, got %s", task1Trend.Direction)
	}

	// Verify task-002 trend
	task2Trend, exists := trends.PerTaskTrends["task-002"]
	if !exists {
		t.Fatal("Expected trend for task-002")
	}
	if task2Trend.PreviousValue != 60.0 {
		t.Errorf("Expected previous value 60.0 for task-002, got %.1f", task2Trend.PreviousValue)
	}
	if task2Trend.CurrentValue != 85.0 {
		t.Errorf("Expected current value 85.0 for task-002, got %.1f", task2Trend.CurrentValue)
	}
	if task2Trend.Direction != "improvement" {
		t.Errorf("Expected direction 'improvement' for task-002, got %s", task2Trend.Direction)
	}
}

// TestLoadEvalTrendsWithMixedResults tests trends with mixed pass/fail results
func TestLoadEvalTrendsWithMixedResults(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/task-eval-log.json"

	// Create test log with task that regresses
	logContent := `[
		{
			"task_id": "task-001",
			"timestamp": "2026-01-26T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 90.0
		},
		{
			"task_id": "task-002",
			"timestamp": "2026-01-26T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 95.0
		},
		{
			"task_id": "task-001",
			"timestamp": "2026-01-27T10:00:00Z",
			"results": [],
			"overall_passed": false,
			"overall_score": 60.0
		},
		{
			"task_id": "task-002",
			"timestamp": "2026-01-27T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 100.0
		}
	]`

	if err := writeFile(logPath, logContent); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Load trends
	trends, err := loadEvalTrends(logPath)
	if err != nil {
		t.Fatalf("loadEvalTrends failed: %v", err)
	}

	// Verify task-001 shows regression
	task1Trend := trends.PerTaskTrends["task-001"]
	if task1Trend.Direction != "regression" {
		t.Errorf("Expected regression for task-001, got %s", task1Trend.Direction)
	}
	if task1Trend.AbsoluteDelta >= 0 {
		t.Errorf("Expected negative delta for task-001, got %.1f", task1Trend.AbsoluteDelta)
	}

	// Verify task-002 shows improvement
	task2Trend := trends.PerTaskTrends["task-002"]
	if task2Trend.Direction != "improvement" {
		t.Errorf("Expected improvement for task-002, got %s", task2Trend.Direction)
	}

	// Verify aggregate shows regression
	// Previous: (90 + 95) / 2 = 92.5
	// Current: (60 + 100) / 2 = 80.0
	if trends.AverageScore.Direction != "regression" {
		t.Errorf("Expected regression in aggregate, got %s", trends.AverageScore.Direction)
	}
}

// TestLoadGradeTrendsInsufficientData tests error handling with only one report
func TestLoadGradeTrendsInsufficientData(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only one report
	reportPath := tmpDir + "/skill-clarity-2026-01-25.md"
	content := `# Skill Clarity Report

Generated: 2026-01-25 10:00:00

## Summary

- **Total Skills**: 10
- **Average Score**: 75.0/100
- **Pass Rate**: 80.0% (8/10)
- **Passing Threshold**: 70.0
`

	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test report: %v", err)
	}

	// Attempt to load trends
	trends, err := loadGradeTrends(tmpDir)

	// Should fail with insufficient data error
	if err == nil {
		t.Fatal("Expected error when loading trends with only 1 report, got nil")
	}

	if trends != nil {
		t.Fatal("Expected nil trends when insufficient data")
	}
}
