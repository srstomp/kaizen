package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// extractSnippet extracts a snippet of HTML around a search term for debugging
func extractSnippet(html, searchTerm string) string {
	idx := strings.Index(html, searchTerm)
	if idx == -1 {
		return "not found"
	}
	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := idx + len(searchTerm) + 50
	if end > len(html) {
		end = len(html)
	}
	return html[start:end]
}

func TestRunDashboardCommand(t *testing.T) {
	// Create temp directory for test reports
	tempDir := t.TempDir()
	reportsDir := filepath.Join(tempDir, "reports")
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test reports dir: %v", err)
	}

	// Create test eval log
	evalLog := `[{
  "task_id": "test-task-001",
  "timestamp": "2026-01-27T18:29:07Z",
  "results": [
    {
      "grader_name": "file-exists",
      "passed": true,
      "score": 100,
      "details": "All 4 files exist",
      "skipped": false,
      "skip_reason": ""
    }
  ],
  "overall_passed": true,
  "overall_score": 100
}]`
	evalLogPath := filepath.Join(reportsDir, "task-eval-log.json")
	if err := os.WriteFile(evalLogPath, []byte(evalLog), 0644); err != nil {
		t.Fatalf("Failed to write eval log: %v", err)
	}

	// Create test meta log
	metaLog := `[{
  "timestamp": "2026-01-27T18:13:50Z",
  "agent": "yokay-quality-reviewer",
  "boundary_type": "epic",
  "consistency_percentage": 85.0,
  "consistent_count": 17,
  "total_count": 20
}]`
	metaLogPath := filepath.Join(reportsDir, "consistency-log.json")
	if err := os.WriteFile(metaLogPath, []byte(metaLog), 0644); err != nil {
		t.Fatalf("Failed to write meta log: %v", err)
	}

	// Run dashboard command
	outputPath := filepath.Join(tempDir, "dashboard.html")
	err = runDashboardCommand(reportsDir, outputPath)
	if err != nil {
		t.Fatalf("runDashboardCommand failed: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Dashboard file was not created at %s", outputPath)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read dashboard file: %v", err)
	}

	html := string(content)

	// Verify HTML structure
	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head>",
		"<title>Yokay Evals Dashboard</title>",
		"<body>",
		"<h1>Yokay Evals Dashboard</h1>",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Dashboard HTML missing required element: %s", elem)
		}
	}

	// Verify sections exist
	sections := []string{
		"Summary",
		"Eval Results",
		"Meta Results",
	}

	for _, section := range sections {
		if !strings.Contains(html, section) {
			t.Errorf("Dashboard missing section: %s", section)
		}
	}

	// Verify data is present
	if !strings.Contains(html, "test-task-001") {
		t.Error("Dashboard does not contain eval task ID")
	}
	if !strings.Contains(html, "yokay-quality-reviewer") {
		t.Error("Dashboard does not contain agent name")
	}
	if !strings.Contains(html, "85.0") {
		t.Error("Dashboard does not contain consistency percentage")
	}
}

func TestGenerateDashboardHTML(t *testing.T) {
	// Create test data
	evalResults := []GradeTaskOutput{
		{
			TaskID:        "task-001",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  95.0,
		},
		{
			TaskID:        "task-002",
			Timestamp:     "2026-01-27T11:00:00Z",
			OverallPassed: false,
			OverallScore:  60.0,
		},
	}

	metaResults := []ConsistencyResult{
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-quality-reviewer",
			BoundaryType:          "epic",
			ConsistencyPercentage: 85.0,
			ConsistentCount:       17,
			TotalCount:            20,
		},
		{
			Timestamp:             "2026-01-27T11:00:00Z",
			Agent:                 "yokay-spec-reviewer",
			BoundaryType:          "epic",
			ConsistencyPercentage: 90.0,
			ConsistentCount:       18,
			TotalCount:            20,
		},
	}

	// Generate HTML
	html := generateDashboardHTML(evalResults, metaResults)

	// Verify HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML missing DOCTYPE")
	}
	if !strings.Contains(html, "<html") {
		t.Error("HTML missing html tag")
	}

	// Verify task IDs present
	if !strings.Contains(html, "task-001") {
		t.Error("HTML missing task-001")
	}
	if !strings.Contains(html, "task-002") {
		t.Error("HTML missing task-002")
	}

	// Verify agent names present
	if !strings.Contains(html, "yokay-quality-reviewer") {
		t.Error("HTML missing yokay-quality-reviewer")
	}
	if !strings.Contains(html, "yokay-spec-reviewer") {
		t.Error("HTML missing yokay-spec-reviewer")
	}

	// Verify scores present
	if !strings.Contains(html, "95.0") {
		t.Error("HTML missing score 95.0")
	}
	if !strings.Contains(html, "85.0") {
		t.Error("HTML missing consistency 85.0")
	}
}

func TestCalculateDashboardSummary(t *testing.T) {
	evalResults := []GradeTaskOutput{
		{OverallPassed: true, OverallScore: 100.0},
		{OverallPassed: true, OverallScore: 90.0},
		{OverallPassed: false, OverallScore: 60.0},
	}

	metaResults := []ConsistencyResult{
		{ConsistencyPercentage: 85.0},
		{ConsistencyPercentage: 90.0},
	}

	summary := calculateDashboardSummary(evalResults, metaResults)

	// Check eval summary
	if summary.EvalTotalCount != 3 {
		t.Errorf("Expected 3 eval total count, got %d", summary.EvalTotalCount)
	}
	if summary.EvalPassCount != 2 {
		t.Errorf("Expected 2 eval pass count, got %d", summary.EvalPassCount)
	}
	expectedAvgScore := (100.0 + 90.0 + 60.0) / 3.0
	if summary.EvalAvgScore != expectedAvgScore {
		t.Errorf("Expected avg score %.2f, got %.2f", expectedAvgScore, summary.EvalAvgScore)
	}

	// Check meta summary
	if summary.MetaTotalCount != 2 {
		t.Errorf("Expected 2 meta total count, got %d", summary.MetaTotalCount)
	}
	expectedMetaAvg := (85.0 + 90.0) / 2.0
	if summary.MetaAvgConsistency != expectedMetaAvg {
		t.Errorf("Expected meta avg %.2f, got %.2f", expectedMetaAvg, summary.MetaAvgConsistency)
	}
}

func TestRunDashboardCommand_MissingReportsDir(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	outputPath := filepath.Join(tempDir, "dashboard.html")

	err := runDashboardCommand(nonExistentDir, outputPath)
	if err == nil {
		t.Error("Expected error for non-existent reports directory, got nil")
	}
}

func TestRunDashboardCommand_EmptyLogs(t *testing.T) {
	// Create temp directory with empty log files
	tempDir := t.TempDir()
	reportsDir := filepath.Join(tempDir, "reports")
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test reports dir: %v", err)
	}

	// Create empty log files
	evalLogPath := filepath.Join(reportsDir, "task-eval-log.json")
	if err := os.WriteFile(evalLogPath, []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to write eval log: %v", err)
	}

	metaLogPath := filepath.Join(reportsDir, "consistency-log.json")
	if err := os.WriteFile(metaLogPath, []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to write meta log: %v", err)
	}

	// Run dashboard command - should succeed but show no data message
	outputPath := filepath.Join(tempDir, "dashboard.html")
	err = runDashboardCommand(reportsDir, outputPath)
	if err != nil {
		t.Fatalf("runDashboardCommand failed: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Dashboard file was not created at %s", outputPath)
	}

	// Read content and verify it handles empty data gracefully
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read dashboard file: %v", err)
	}

	html := string(content)
	if !strings.Contains(html, "No data") && !strings.Contains(html, "0") {
		t.Error("Dashboard should indicate no data available")
	}
}

func TestGenerateDashboardHTML_XSSPrevention(t *testing.T) {
	// Create test data with XSS attack vectors
	evalResults := []GradeTaskOutput{
		{
			TaskID:        "<script>alert('xss')</script>",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  95.0,
		},
		{
			TaskID:        "<img src=x onerror=alert('xss')>",
			Timestamp:     "<script>alert('timestamp')</script>",
			OverallPassed: false,
			OverallScore:  60.0,
		},
	}

	metaResults := []ConsistencyResult{
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "<script>alert('agent')</script>",
			BoundaryType:          "<img src=x onerror=alert('boundary')>",
			ConsistencyPercentage: 85.0,
			ConsistentCount:       17,
			TotalCount:            20,
		},
	}

	// Generate HTML
	html := generateDashboardHTML(evalResults, metaResults)

	// Verify that dangerous HTML tags are escaped (not executable)
	// Check that < and > are escaped, making tags non-functional
	if strings.Contains(html, "<script>alert") || strings.Contains(html, "<img src=x onerror=") {
		t.Error("HTML tags should be escaped to prevent XSS execution")
	}

	// Verify escaped content is present
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Expected escaped script tag '&lt;script&gt;' not found")
	}

	if !strings.Contains(html, "&lt;img") {
		t.Error("Expected escaped img tag '&lt;img' not found")
	}

	// Verify that attribute quotes are also escaped
	if !strings.Contains(html, "&#39;") || !strings.Contains(html, "&lt;") {
		t.Error("Expected HTML entities for escaping special characters")
	}

	// Verify generated timestamp is also escaped
	if strings.Contains(html, "Generated: <script>") {
		t.Error("Generated timestamp should not contain unescaped script tags")
	}
}
