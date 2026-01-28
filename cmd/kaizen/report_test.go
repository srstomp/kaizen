package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindGradeReports verifies that findGradeReports can locate skill-clarity reports
func TestFindGradeReports(t *testing.T) {
	// Create a temporary directory with test reports
	tmpDir := t.TempDir()
	reportsDir := filepath.Join(tmpDir, "reports")
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test reports dir: %v", err)
	}

	// Create test report files
	testFiles := []string{
		"skill-clarity-2026-01-25.md",
		"skill-clarity-2026-01-26.md",
		"skill-clarity-2026-01-24.md",
		"other-report.md", // Should not be included
	}

	for _, filename := range testFiles {
		path := filepath.Join(reportsDir, filename)
		err := os.WriteFile(path, []byte("# Test Report\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test: Find grade reports
	reports, err := findGradeReports(reportsDir)
	if err != nil {
		t.Fatalf("findGradeReports failed: %v", err)
	}

	// Verify: Should find 3 skill-clarity reports, sorted by date (newest first)
	expectedCount := 3
	if len(reports) != expectedCount {
		t.Errorf("Expected %d reports, got %d", expectedCount, len(reports))
	}

	// Verify: Reports should be sorted by date (newest first)
	expectedOrder := []string{
		"skill-clarity-2026-01-26.md",
		"skill-clarity-2026-01-25.md",
		"skill-clarity-2026-01-24.md",
	}

	for i, expected := range expectedOrder {
		if i >= len(reports) {
			break
		}
		if filepath.Base(reports[i]) != expected {
			t.Errorf("Report %d: expected %s, got %s", i, expected, filepath.Base(reports[i]))
		}
	}
}

// TestParseGradeReport verifies that parseGradeReport can extract metrics from a report file
func TestParseGradeReport(t *testing.T) {
	// Create a temporary report file with known content
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "skill-clarity-2026-01-26.md")

	reportContent := `# Skill Clarity Report

Generated: 2026-01-26 21:30:43

This report evaluates pokayokay skills using the Skill Clarity Grader.
**Note**: Current grading uses heuristic-based evaluation (stub implementation). LLM-based grading not yet implemented.

## Summary

- **Total Skills**: 27
- **Average Score**: 60.3/100
- **Pass Rate**: 3.7% (1/27)
- **Passing Threshold**: 70.0

## Skills Below Threshold (< 80%)

These skills need improvement:

- **ux-design** - 75.0/100 - Needs Improvement
- **documentation** - 68.0/100 - **FAILED**
`

	err := os.WriteFile(reportPath, []byte(reportContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test report: %v", err)
	}

	// Test: Parse the report
	report, err := parseGradeReport(reportPath)
	if err != nil {
		t.Fatalf("parseGradeReport failed: %v", err)
	}

	// Verify: Extracted metrics are correct
	if report.TotalSkills != 27 {
		t.Errorf("Expected TotalSkills=27, got %d", report.TotalSkills)
	}

	if report.AverageScore != 60.3 {
		t.Errorf("Expected AverageScore=60.3, got %.1f", report.AverageScore)
	}

	expectedPassRate := 3.7
	if report.PassRate != expectedPassRate {
		t.Errorf("Expected PassRate=%.1f, got %.1f", expectedPassRate, report.PassRate)
	}

	if report.PassingThreshold != 70.0 {
		t.Errorf("Expected PassingThreshold=70.0, got %.1f", report.PassingThreshold)
	}

	// Verify: GeneratedDate is extracted
	if !strings.Contains(report.GeneratedDate, "2026-01-26") {
		t.Errorf("Expected GeneratedDate to contain '2026-01-26', got '%s'", report.GeneratedDate)
	}
}

// TestFormatReportSummaryMarkdown verifies markdown formatting of report summary
func TestFormatReportSummaryMarkdown(t *testing.T) {
	report := GradeReport{
		FilePath:         "/path/to/skill-clarity-2026-01-26.md",
		GeneratedDate:    "2026-01-26 21:30:43",
		TotalSkills:      27,
		AverageScore:     60.3,
		PassRate:         3.7,
		PassingThreshold: 70.0,
	}

	// Test: Format as markdown (without trends)
	output := formatReportSummaryMarkdown(report, nil, false)

	// Verify: Output contains key metrics
	expectedStrings := []string{
		"# Evaluation Report Summary",
		"skill-clarity-2026-01-26.md",
		"2026-01-26 21:30:43",
		"27",
		"60.3",
		"3.7%",
		"70.0",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestFormatReportSummaryJSON verifies JSON formatting of report summary
func TestFormatReportSummaryJSON(t *testing.T) {
	report := GradeReport{
		FilePath:         "/path/to/skill-clarity-2026-01-26.md",
		GeneratedDate:    "2026-01-26 21:30:43",
		TotalSkills:      27,
		AverageScore:     60.3,
		PassRate:         3.7,
		PassingThreshold: 70.0,
	}

	// Test: Format as JSON (without trends)
	output, err := formatReportSummaryJSON(report, nil, false)
	if err != nil {
		t.Fatalf("formatReportSummaryJSON failed: %v", err)
	}

	// Verify: Output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify: JSON contains expected fields
	if parsed["file_path"] != report.FilePath {
		t.Errorf("Expected file_path=%s, got %v", report.FilePath, parsed["file_path"])
	}

	if parsed["total_skills"] != float64(27) {
		t.Errorf("Expected total_skills=27, got %v", parsed["total_skills"])
	}
}

// TestListGradeReports verifies that listGradeReports outputs correct format
func TestListGradeReports(t *testing.T) {
	// Create temporary directory with test reports
	tmpDir := t.TempDir()
	reportsDir := filepath.Join(tmpDir, "reports")
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test reports dir: %v", err)
	}

	// Create test report files
	testFiles := []string{
		"skill-clarity-2026-01-26.md",
		"skill-clarity-2026-01-25.md",
	}

	for _, filename := range testFiles {
		path := filepath.Join(reportsDir, filename)
		err := os.WriteFile(path, []byte("# Test Report\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test: List grade reports
	output := listGradeReports(reportsDir)

	// Verify: Output contains both reports
	for _, filename := range testFiles {
		if !strings.Contains(output, filename) {
			t.Errorf("Expected output to contain '%s'", filename)
		}
	}

	// Verify: Output has header
	if !strings.Contains(output, "Grade Reports") {
		t.Error("Expected output to have 'Grade Reports' header")
	}
}

// TestRunReportCommand verifies the main report command execution
func TestRunReportCommand(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	reportsDir := filepath.Join(tmpDir, "reports")
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test reports dir: %v", err)
	}

	// Create a test report with valid content
	reportPath := filepath.Join(reportsDir, "skill-clarity-2026-01-26.md")
	reportContent := `# Skill Clarity Report

Generated: 2026-01-26 21:30:43

## Summary

- **Total Skills**: 10
- **Average Score**: 75.5/100
- **Pass Rate**: 80.0% (8/10)
- **Passing Threshold**: 70.0
`

	err = os.WriteFile(reportPath, []byte(reportContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test report: %v", err)
	}

	// Test: Run report command with grade type (without trends)
	err = runReportCommand("grade", "markdown", false, "", reportsDir, false)
	if err != nil {
		t.Fatalf("runReportCommand failed: %v", err)
	}

	// Note: Since runReportCommand outputs to stdout, we can't easily capture
	// the output in this test. In a real implementation, we'd refactor to
	// accept an io.Writer for testability.
}

// TestRunReportCommandListMode verifies list mode
func TestRunReportCommandListMode(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	reportsDir := filepath.Join(tmpDir, "reports")
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test reports dir: %v", err)
	}

	// Create test reports
	testFiles := []string{
		"skill-clarity-2026-01-26.md",
		"skill-clarity-2026-01-25.md",
	}

	for _, filename := range testFiles {
		path := filepath.Join(reportsDir, filename)
		err := os.WriteFile(path, []byte("# Test Report\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test: Run report command in list mode
	err = runReportCommand("grade", "markdown", true, "", reportsDir, false)
	if err != nil {
		t.Fatalf("runReportCommand in list mode failed: %v", err)
	}
}

// TestParseGradeReportWithCriteriaScores verifies that parseGradeReport extracts per-criteria scores
func TestParseGradeReportWithCriteriaScores(t *testing.T) {
	// Create a temporary report file with Detailed Breakdown section
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "skill-clarity-2026-01-26.md")

	reportContent := `# Skill Clarity Report

Generated: 2026-01-26 21:30:43

## Summary

- **Total Skills**: 3
- **Average Score**: 67.5/100
- **Pass Rate**: 33.3% (1/3)
- **Passing Threshold**: 70.0

## Detailed Breakdown

### skill-one

**Overall Score**: 75.0/100

**Criteria Scores**:

- **Clear Instructions** (weight: 30%): 80.0/100
  - Evaluation note
- **Actionable Steps** (weight: 25%): 75.0/100
  - Evaluation note
- **Good Examples** (weight: 25%): 70.0/100
  - Evaluation note
- **Appropriate Scope** (weight: 20%): 75.0/100
  - Evaluation note

### skill-two

**Overall Score**: 65.0/100

**Criteria Scores**:

- **Clear Instructions** (weight: 30%): 70.0/100
  - Evaluation note
- **Actionable Steps** (weight: 25%): 65.0/100
  - Evaluation note
- **Good Examples** (weight: 25%): 60.0/100
  - Evaluation note
- **Appropriate Scope** (weight: 20%): 65.0/100
  - Evaluation note

### skill-three

**Overall Score**: 62.5/100

**Criteria Scores**:

- **Clear Instructions** (weight: 30%): 60.0/100
  - Evaluation note
- **Actionable Steps** (weight: 25%): 60.0/100
  - Evaluation note
- **Good Examples** (weight: 25%): 65.0/100
  - Evaluation note
- **Appropriate Scope** (weight: 20%): 65.0/100
  - Evaluation note
`

	err := os.WriteFile(reportPath, []byte(reportContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test report: %v", err)
	}

	// Test: Parse the report
	report, err := parseGradeReport(reportPath)
	if err != nil {
		t.Fatalf("parseGradeReport failed: %v", err)
	}

	// Verify: CriteriaScores are extracted
	if len(report.CriteriaScores) != 4 {
		t.Errorf("Expected 4 criteria scores, got %d", len(report.CriteriaScores))
	}

	// Expected averages:
	// Clear Instructions: (80 + 70 + 60) / 3 = 70.0
	// Actionable Steps: (75 + 65 + 60) / 3 = 66.67 (rounded to 66.7)
	// Good Examples: (70 + 60 + 65) / 3 = 65.0
	// Appropriate Scope: (75 + 65 + 65) / 3 = 68.33 (rounded to 68.3)
	expectedScores := map[string]float64{
		"Clear Instructions": 70.0,
		"Actionable Steps":   66.7,
		"Good Examples":      65.0,
		"Appropriate Scope":  68.3,
	}

	for _, criteria := range report.CriteriaScores {
		expected, exists := expectedScores[criteria.Name]
		if !exists {
			t.Errorf("Unexpected criteria name: %s", criteria.Name)
			continue
		}

		// Allow small floating point difference
		if diff := criteria.Average - expected; diff > 0.1 || diff < -0.1 {
			t.Errorf("Criteria '%s': expected average %.1f, got %.1f", criteria.Name, expected, criteria.Average)
		}
	}
}

// TestFormatReportSummaryMarkdownWithCriteria verifies markdown includes per-category breakdown
func TestFormatReportSummaryMarkdownWithCriteria(t *testing.T) {
	report := GradeReport{
		FilePath:         "/path/to/skill-clarity-2026-01-26.md",
		GeneratedDate:    "2026-01-26 21:30:43",
		TotalSkills:      3,
		AverageScore:     67.5,
		PassRate:         33.3,
		PassingThreshold: 70.0,
		CriteriaScores: []CriteriaScore{
			{Name: "Clear Instructions", Average: 70.0},
			{Name: "Actionable Steps", Average: 66.7},
			{Name: "Good Examples", Average: 65.0},
			{Name: "Appropriate Scope", Average: 68.3},
		},
	}

	// Test: Format as markdown (without trends)
	output := formatReportSummaryMarkdown(report, nil, false)

	// Verify: Output contains per-category breakdown section
	if !strings.Contains(output, "## Per-Category Breakdown") {
		t.Error("Expected output to contain '## Per-Category Breakdown' section")
	}

	// Verify: Output contains criteria names and scores
	expectedStrings := []string{
		"Clear Instructions",
		"Actionable Steps",
		"Good Examples",
		"Appropriate Scope",
		"70.0",
		"66.7",
		"65.0",
		"68.3",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}

	// Verify: Output contains table structure
	if !strings.Contains(output, "| Criteria | Average Score |") {
		t.Error("Expected output to contain table header")
	}
}

// TestFormatReportSummaryJSONWithCriteria verifies JSON includes criteria_scores array
func TestFormatReportSummaryJSONWithCriteria(t *testing.T) {
	report := GradeReport{
		FilePath:         "/path/to/skill-clarity-2026-01-26.md",
		GeneratedDate:    "2026-01-26 21:30:43",
		TotalSkills:      3,
		AverageScore:     67.5,
		PassRate:         33.3,
		PassingThreshold: 70.0,
		CriteriaScores: []CriteriaScore{
			{Name: "Clear Instructions", Average: 70.0},
			{Name: "Actionable Steps", Average: 66.7},
			{Name: "Good Examples", Average: 65.0},
			{Name: "Appropriate Scope", Average: 68.3},
		},
	}

	// Test: Format as JSON (without trends)
	output, err := formatReportSummaryJSON(report, nil, false)
	if err != nil {
		t.Fatalf("formatReportSummaryJSON failed: %v", err)
	}

	// Verify: Output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify: JSON contains criteria_scores array
	criteriaScoresRaw, exists := parsed["criteria_scores"]
	if !exists {
		t.Fatal("Expected JSON to contain 'criteria_scores' field")
	}

	criteriaScores, ok := criteriaScoresRaw.([]interface{})
	if !ok {
		t.Fatalf("Expected criteria_scores to be an array, got %T", criteriaScoresRaw)
	}

	if len(criteriaScores) != 4 {
		t.Errorf("Expected 4 criteria scores in JSON, got %d", len(criteriaScores))
	}

	// Verify: First criteria score has correct structure
	firstScore, ok := criteriaScores[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected criteria score to be an object, got %T", criteriaScores[0])
	}

	if _, exists := firstScore["name"]; !exists {
		t.Error("Expected criteria score to have 'name' field")
	}

	if _, exists := firstScore["average"]; !exists {
		t.Error("Expected criteria score to have 'average' field")
	}
}

// TestFormatReportSummaryMarkdownWithTrends verifies markdown includes trend analysis
func TestFormatReportSummaryMarkdownWithTrends(t *testing.T) {
	report := GradeReport{
		FilePath:         "/path/to/skill-clarity-2026-01-26.md",
		GeneratedDate:    "2026-01-26 21:30:43",
		TotalSkills:      10,
		AverageScore:     80.0,
		PassRate:         90.0,
		PassingThreshold: 70.0,
	}

	trends := &GradeTrends{
		TotalSkills:  calculateDelta(10.0, 10.0),
		AverageScore: calculateDelta(75.0, 80.0),
		PassRate:     calculateDelta(80.0, 90.0),
	}

	// Test: Format as markdown with trends
	output := formatReportSummaryMarkdown(report, trends, true)

	// Verify: Output contains trend section
	if !strings.Contains(output, "## Trend Analysis") {
		t.Error("Expected output to contain '## Trend Analysis' section")
	}

	// Verify: Output contains trend data
	expectedStrings := []string{
		"Average Score",
		"75.0",
		"80.0",
		"↑",
		"improvement",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestFormatReportSummaryJSONWithTrends verifies JSON includes trend data
func TestFormatReportSummaryJSONWithTrends(t *testing.T) {
	report := GradeReport{
		FilePath:         "/path/to/skill-clarity-2026-01-26.md",
		GeneratedDate:    "2026-01-26 21:30:43",
		TotalSkills:      10,
		AverageScore:     80.0,
		PassRate:         90.0,
		PassingThreshold: 70.0,
	}

	trends := &GradeTrends{
		TotalSkills:  calculateDelta(10.0, 10.0),
		AverageScore: calculateDelta(75.0, 80.0),
		PassRate:     calculateDelta(80.0, 90.0),
	}

	// Test: Format as JSON with trends
	output, err := formatReportSummaryJSON(report, trends, true)
	if err != nil {
		t.Fatalf("formatReportSummaryJSON failed: %v", err)
	}

	// Verify: Output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify: JSON contains trend field
	trendRaw, exists := parsed["trend"]
	if !exists {
		t.Fatal("Expected JSON to contain 'trend' field")
	}

	trend, ok := trendRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected trend to be an object, got %T", trendRaw)
	}

	// Verify: Trend contains average_score
	if _, exists := trend["average_score"]; !exists {
		t.Error("Expected trend to have 'average_score' field")
	}

	// Verify: average_score trend has expected structure
	avgScoreTrend, ok := trend["average_score"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected average_score trend to be an object")
	}

	if avgScoreTrend["previous_value"] != 75.0 {
		t.Errorf("Expected previous_value=75.0, got %v", avgScoreTrend["previous_value"])
	}

	if avgScoreTrend["current_value"] != 80.0 {
		t.Errorf("Expected current_value=80.0, got %v", avgScoreTrend["current_value"])
	}

	if avgScoreTrend["direction"] != "improvement" {
		t.Errorf("Expected direction='improvement', got %v", avgScoreTrend["direction"])
	}
}

// TestReportWithTrendsEndToEnd tests the full integration of trends in report command
func TestReportWithTrendsEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	reportsDir := tmpDir + "/reports"
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	// Create two reports with different metrics
	oldReport := reportsDir + "/skill-clarity-2026-01-25.md"
	newReport := reportsDir + "/skill-clarity-2026-01-26.md"

	oldContent := `# Skill Clarity Report

Generated: 2026-01-25 10:00:00

## Summary

- **Total Skills**: 10
- **Average Score**: 70.0/100
- **Pass Rate**: 60.0% (6/10)
- **Passing Threshold**: 70.0

## Detailed Breakdown

### skill-one

**Overall Score**: 70.0/100

**Criteria Scores**:

- **Clear Instructions** (weight: 30%): 65.0/100
- **Actionable Steps** (weight: 25%): 70.0/100
- **Good Examples** (weight: 25%): 72.0/100
- **Appropriate Scope** (weight: 20%): 73.0/100
`

	newContent := `# Skill Clarity Report

Generated: 2026-01-26 10:00:00

## Summary

- **Total Skills**: 10
- **Average Score**: 80.0/100
- **Pass Rate**: 80.0% (8/10)
- **Passing Threshold**: 70.0

## Detailed Breakdown

### skill-one

**Overall Score**: 80.0/100

**Criteria Scores**:

- **Clear Instructions** (weight: 30%): 75.0/100
- **Actionable Steps** (weight: 25%): 80.0/100
- **Good Examples** (weight: 25%): 82.0/100
- **Appropriate Scope** (weight: 20%): 83.0/100
`

	if err := os.WriteFile(oldReport, []byte(oldContent), 0644); err != nil {
		t.Fatalf("Failed to write old report: %v", err)
	}
	if err := os.WriteFile(newReport, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to write new report: %v", err)
	}

	// Test with trends enabled
	outputPath := tmpDir + "/output-with-trends.md"
	err = runReportCommand("grade", "markdown", false, outputPath, reportsDir, true)
	if err != nil {
		t.Fatalf("runReportCommand with trends failed: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	output := string(content)

	// Verify trend section exists
	if !strings.Contains(output, "## Trend Analysis") {
		t.Error("Expected trend analysis section in output")
	}

	// Verify improvements are shown
	if !strings.Contains(output, "↑") {
		t.Error("Expected improvement indicators in output")
	}

	// Verify specific metrics
	expectedStrings := []string{
		"70.0", // previous average score
		"80.0", // current average score
		"+10.0", // delta
		"improvement",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}

	// Test with trends disabled
	outputPathNoTrends := tmpDir + "/output-no-trends.md"
	err = runReportCommand("grade", "markdown", false, outputPathNoTrends, reportsDir, false)
	if err != nil {
		t.Fatalf("runReportCommand without trends failed: %v", err)
	}

	// Read and verify output without trends
	contentNoTrends, err := os.ReadFile(outputPathNoTrends)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	outputNoTrends := string(contentNoTrends)

	// Verify trend section does NOT exist
	if strings.Contains(outputNoTrends, "## Trend Analysis") {
		t.Error("Expected NO trend analysis section when trends disabled")
	}
}

// TestRunReportCommandMetaType tests the report command with --type meta
func TestRunReportCommandMetaType(t *testing.T) {
	tmpDir := t.TempDir()
	reportsDir := tmpDir + "/reports"
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	// Create consistency-log.json with multiple agent entries
	logPath := filepath.Join(reportsDir, "consistency-log.json")
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

	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to write consistency log: %v", err)
	}

	// Test: Run report command with meta type (without trends)
	outputPath := tmpDir + "/meta-report.md"
	err = runReportCommand("meta", "markdown", false, outputPath, reportsDir, false)
	if err != nil {
		t.Fatalf("runReportCommand with meta type failed: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	output := string(content)

	// Verify output contains expected content
	expectedStrings := []string{
		"# Meta-Evaluation Report",
		"**Report Type**: meta",
		"yokay-quality-reviewer",
		"yokay-spec-reviewer",
		"85.0%",
		"92.0%",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestRunReportCommandMetaTypeWithTrends tests meta report with trend analysis
func TestRunReportCommandMetaTypeWithTrends(t *testing.T) {
	tmpDir := t.TempDir()
	reportsDir := tmpDir + "/reports"
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	// Create consistency-log.json with multiple entries per agent
	logPath := filepath.Join(reportsDir, "consistency-log.json")
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

	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to write consistency log: %v", err)
	}

	// Test: Run report command with meta type and trends enabled
	outputPath := tmpDir + "/meta-report-trends.md"
	err = runReportCommand("meta", "markdown", false, outputPath, reportsDir, true)
	if err != nil {
		t.Fatalf("runReportCommand with meta type and trends failed: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	output := string(content)

	// Verify trend section exists
	if !strings.Contains(output, "## Trend Analysis") {
		t.Error("Expected trend analysis section in output")
	}

	// Verify per-agent trends are shown
	expectedStrings := []string{
		"yokay-quality-reviewer",
		"yokay-spec-reviewer",
		"80.0",  // previous value for quality-reviewer
		"85.0",  // current value for quality-reviewer
		"90.0",  // previous value for spec-reviewer
		"92.0",  // current value for spec-reviewer
		"+5.0pp", // delta for quality-reviewer
		"+2.0pp", // delta for spec-reviewer
		"↑",     // improvement indicator
		"improvement",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestFormatMetaReportMarkdown tests markdown formatting for meta reports
func TestFormatMetaReportMarkdown(t *testing.T) {
	// Create test data with multiple agents
	results := []ConsistencyResult{
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-quality-reviewer",
			ConsistencyPercentage: 85.0,
			ConsistentCount:       17,
			TotalCount:            20,
		},
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-spec-reviewer",
			ConsistencyPercentage: 92.0,
			ConsistentCount:       23,
			TotalCount:            25,
		},
	}

	// Test: Format as markdown without trends
	output := formatMetaReportMarkdown(results, nil, false)

	// Verify output structure
	expectedStrings := []string{
		"# Meta-Evaluation Report",
		"**Report Type**: meta",
		"## Current Metrics",
		"| Agent | Consistency | Runs |",
		"yokay-quality-reviewer",
		"yokay-spec-reviewer",
		"85.0%",
		"92.0%",
		"20",
		"25",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestFormatMetaReportMarkdownWithTrends tests markdown formatting with trends
func TestFormatMetaReportMarkdownWithTrends(t *testing.T) {
	// Create test data
	results := []ConsistencyResult{
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-quality-reviewer",
			ConsistencyPercentage: 85.0,
			ConsistentCount:       17,
			TotalCount:            20,
		},
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-spec-reviewer",
			ConsistencyPercentage: 92.0,
			ConsistentCount:       23,
			TotalCount:            25,
		},
	}

	// Create trends data
	trends := &MetaTrends{
		PerAgentTrends: map[string]TrendData{
			"yokay-quality-reviewer": calculateDelta(80.0, 85.0),
			"yokay-spec-reviewer":    calculateDelta(90.0, 92.0),
		},
	}

	// Test: Format as markdown with trends
	output := formatMetaReportMarkdown(results, trends, true)

	// Verify trend section exists
	if !strings.Contains(output, "## Trend Analysis") {
		t.Error("Expected trend analysis section")
	}

	// Verify trend data is present
	expectedStrings := []string{
		"yokay-quality-reviewer",
		"yokay-spec-reviewer",
		"80.0",
		"85.0",
		"90.0",
		"92.0",
		"+5.0pp",
		"+2.0pp",
		"↑",
		"improvement",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestFormatMetaReportJSON tests JSON formatting for meta reports
func TestFormatMetaReportJSON(t *testing.T) {
	// Create test data
	results := []ConsistencyResult{
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-quality-reviewer",
			ConsistencyPercentage: 85.0,
			ConsistentCount:       17,
			TotalCount:            20,
		},
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-spec-reviewer",
			ConsistencyPercentage: 92.0,
			ConsistentCount:       23,
			TotalCount:            25,
		},
	}

	// Test: Format as JSON without trends
	output, err := formatMetaReportJSON(results, nil, false)
	if err != nil {
		t.Fatalf("formatMetaReportJSON failed: %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify JSON structure
	if parsed["report_type"] != "meta" {
		t.Errorf("Expected report_type='meta', got %v", parsed["report_type"])
	}

	agents, ok := parsed["agents"].([]interface{})
	if !ok {
		t.Fatalf("Expected agents to be an array")
	}

	if len(agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents))
	}
}

// TestFormatMetaReportJSONWithTrends tests JSON formatting with trends
func TestFormatMetaReportJSONWithTrends(t *testing.T) {
	// Create test data
	results := []ConsistencyResult{
		{
			Timestamp:             "2026-01-27T10:00:00Z",
			Agent:                 "yokay-quality-reviewer",
			ConsistencyPercentage: 85.0,
			ConsistentCount:       17,
			TotalCount:            20,
		},
	}

	// Create trends data
	trends := &MetaTrends{
		PerAgentTrends: map[string]TrendData{
			"yokay-quality-reviewer": calculateDelta(80.0, 85.0),
		},
	}

	// Test: Format as JSON with trends
	output, err := formatMetaReportJSON(results, trends, true)
	if err != nil {
		t.Fatalf("formatMetaReportJSON failed: %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify trend field exists
	trendRaw, exists := parsed["trend"]
	if !exists {
		t.Fatal("Expected JSON to contain 'trend' field")
	}

	trend, ok := trendRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected trend to be an object, got %T", trendRaw)
	}

	// Verify per-agent trends
	perAgent, exists := trend["per_agent"]
	if !exists {
		t.Fatal("Expected trend to contain 'per_agent' field")
	}

	perAgentMap, ok := perAgent.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected per_agent to be an object, got %T", perAgent)
	}

	if _, exists := perAgentMap["yokay-quality-reviewer"]; !exists {
		t.Error("Expected per_agent to contain 'yokay-quality-reviewer'")
	}
}

// TestRunReportCommandEvalType tests the report command with --type eval
func TestRunReportCommandEvalType(t *testing.T) {
	tmpDir := t.TempDir()
	reportsDir := tmpDir + "/reports"
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	// Create task-eval-log.json with sample eval results at same timestamp
	logPath := filepath.Join(reportsDir, "task-eval-log.json")
	logContent := `[
		{
			"task_id": "task-001",
			"timestamp": "2026-01-27T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 85.0
		},
		{
			"task_id": "task-002",
			"timestamp": "2026-01-27T10:00:00Z",
			"results": [],
			"overall_passed": true,
			"overall_score": 95.0
		}
	]`

	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to write eval log: %v", err)
	}

	// Test: Run report command with eval type (without trends)
	outputPath := tmpDir + "/eval-report.md"
	err = runReportCommand("eval", "markdown", false, outputPath, reportsDir, false)
	if err != nil {
		t.Fatalf("runReportCommand with eval type failed: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	output := string(content)

	// Verify output contains expected content
	expectedStrings := []string{
		"# Evaluation Report",
		"**Report Type**: eval",
		"## Current Metrics",
		"Average Score",
		"90.0", // (85.0 + 95.0) / 2
		"Pass Rate",
		"100.0%", // 2/2 passed
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestRunReportCommandEvalTypeWithTrends tests eval report with trend analysis
func TestRunReportCommandEvalTypeWithTrends(t *testing.T) {
	tmpDir := t.TempDir()
	reportsDir := tmpDir + "/reports"
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	// Create task-eval-log.json with multiple entries for trend analysis
	logPath := filepath.Join(reportsDir, "task-eval-log.json")
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

	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to write eval log: %v", err)
	}

	// Test: Run report command with eval type and trends enabled
	outputPath := tmpDir + "/eval-report-trends.md"
	err = runReportCommand("eval", "markdown", false, outputPath, reportsDir, true)
	if err != nil {
		t.Fatalf("runReportCommand with eval type and trends failed: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	output := string(content)

	// Verify trend section exists
	if !strings.Contains(output, "## Trend Analysis") {
		t.Error("Expected trend analysis section in output")
	}

	// Verify aggregate trends are shown
	expectedStrings := []string{
		"### Aggregate Metrics",
		"Average Score",
		"70.0", // (80 + 60) / 2 = 70 previous
		"87.5", // (90 + 85) / 2 = 87.5 current
		"+17.5", // delta
		"↑",
		"improvement",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}

	// Verify per-task trends are shown
	if !strings.Contains(output, "### Per-Task Trends") {
		t.Error("Expected per-task trends section in output")
	}

	perTaskExpected := []string{
		"task-001",
		"task-002",
		"80.0", // task-001 previous
		"90.0", // task-001 current
		"60.0", // task-002 previous
		"85.0", // task-002 current
	}

	for _, expected := range perTaskExpected {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain per-task value '%s'", expected)
		}
	}
}

// TestFormatEvalReportMarkdown tests markdown formatting for eval reports
func TestFormatEvalReportMarkdown(t *testing.T) {
	// Create test eval results with same timestamp
	results := []GradeTaskOutput{
		{
			TaskID:        "task-001",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  90.0,
		},
		{
			TaskID:        "task-002",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  85.0,
		},
	}

	// Test: Format as markdown without trends
	output := formatEvalReportMarkdown(results, nil, false)

	// Verify output structure
	expectedStrings := []string{
		"# Evaluation Report",
		"**Report Type**: eval",
		"## Current Metrics",
		"Average Score",
		"87.5", // (90 + 85) / 2
		"Pass Rate",
		"100.0%",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestFormatEvalReportMarkdownWithTrends tests markdown formatting with trends
func TestFormatEvalReportMarkdownWithTrends(t *testing.T) {
	// Create test eval results
	results := []GradeTaskOutput{
		{
			TaskID:        "task-001",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  90.0,
		},
		{
			TaskID:        "task-002",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  85.0,
		},
	}

	// Create trends data
	trends := &EvalTrends{
		AverageScore: calculateDelta(70.0, 87.5),
		PassRate:     calculateDelta(50.0, 100.0),
		PerTaskTrends: map[string]TrendData{
			"task-001": calculateDelta(80.0, 90.0),
			"task-002": calculateDelta(60.0, 85.0),
		},
	}

	// Test: Format as markdown with trends
	output := formatEvalReportMarkdown(results, trends, true)

	// Verify trend section exists
	if !strings.Contains(output, "## Trend Analysis") {
		t.Error("Expected trend analysis section")
	}

	// Verify aggregate trends
	expectedStrings := []string{
		"### Aggregate Metrics",
		"70.0",
		"87.5",
		"+17.5",
		"↑",
		"improvement",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}

	// Verify per-task trends
	if !strings.Contains(output, "### Per-Task Trends") {
		t.Error("Expected per-task trends section")
	}

	perTaskExpected := []string{
		"task-001",
		"task-002",
		"80.0",
		"90.0",
		"+10.0",
	}

	for _, expected := range perTaskExpected {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestFormatEvalReportJSON tests JSON formatting for eval reports
func TestFormatEvalReportJSON(t *testing.T) {
	// Create test eval results with same timestamp
	results := []GradeTaskOutput{
		{
			TaskID:        "task-001",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  90.0,
		},
		{
			TaskID:        "task-002",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  85.0,
		},
	}

	// Test: Format as JSON without trends
	output, err := formatEvalReportJSON(results, nil, false)
	if err != nil {
		t.Fatalf("formatEvalReportJSON failed: %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify JSON structure
	if parsed["report_type"] != "eval" {
		t.Errorf("Expected report_type='eval', got %v", parsed["report_type"])
	}

	if parsed["average_score"] != 87.5 {
		t.Errorf("Expected average_score=87.5, got %v", parsed["average_score"])
	}

	if parsed["pass_rate"] != 100.0 {
		t.Errorf("Expected pass_rate=100.0, got %v", parsed["pass_rate"])
	}
}

// TestFormatEvalReportJSONWithTrends tests JSON formatting with trends
func TestFormatEvalReportJSONWithTrends(t *testing.T) {
	// Create test eval results
	results := []GradeTaskOutput{
		{
			TaskID:        "task-001",
			Timestamp:     "2026-01-27T10:00:00Z",
			OverallPassed: true,
			OverallScore:  90.0,
		},
	}

	// Create trends data
	trends := &EvalTrends{
		AverageScore: calculateDelta(80.0, 90.0),
		PassRate:     calculateDelta(100.0, 100.0),
		PerTaskTrends: map[string]TrendData{
			"task-001": calculateDelta(80.0, 90.0),
		},
	}

	// Test: Format as JSON with trends
	output, err := formatEvalReportJSON(results, trends, true)
	if err != nil {
		t.Fatalf("formatEvalReportJSON failed: %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify trend field exists
	trendRaw, exists := parsed["trend"]
	if !exists {
		t.Fatal("Expected JSON to contain 'trend' field")
	}

	trend, ok := trendRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected trend to be an object, got %T", trendRaw)
	}

	// Verify aggregate trends
	if _, exists := trend["average_score"]; !exists {
		t.Error("Expected trend to have 'average_score' field")
	}

	if _, exists := trend["pass_rate"]; !exists {
		t.Error("Expected trend to have 'pass_rate' field")
	}

	// Verify per-task trends
	perTask, exists := trend["per_task"]
	if !exists {
		t.Fatal("Expected trend to contain 'per_task' field")
	}

	perTaskMap, ok := perTask.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected per_task to be an object, got %T", perTask)
	}

	if _, exists := perTaskMap["task-001"]; !exists {
		t.Error("Expected per_task to contain 'task-001'")
	}
}

// TestRunReportCommandWithTrendsButInsufficientData tests graceful handling when trends can't be loaded
func TestRunReportCommandWithTrendsButInsufficientData(t *testing.T) {
	tmpDir := t.TempDir()
	reportsDir := filepath.Join(tmpDir, "reports")
	err := os.MkdirAll(reportsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	// Create only one report
	reportPath := filepath.Join(reportsDir, "skill-clarity-2026-01-26.md")
	content := `# Skill Clarity Report

Generated: 2026-01-26 10:00:00

## Summary

- **Total Skills**: 10
- **Average Score**: 75.5/100
- **Pass Rate**: 80.0% (8/10)
- **Passing Threshold**: 70.0
`

	err = os.WriteFile(reportPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test report: %v", err)
	}

	// Run report command with trends enabled but insufficient data available
	// Should NOT fail, should gracefully handle the missing trends
	outputPath := filepath.Join(tmpDir, "output.md")
	err = runReportCommand("grade", "markdown", false, outputPath, reportsDir, true)
	if err != nil {
		t.Fatalf("runReportCommand should not fail with insufficient trend data, got: %v", err)
	}

	// Verify output was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Expected output file to be created")
	}
}
