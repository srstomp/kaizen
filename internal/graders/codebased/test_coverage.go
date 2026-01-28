package codebased

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TestCoverageGrader runs Go test coverage and compares against a minimum threshold
type TestCoverageGrader struct {
	threshold float64
}

// NewTestCoverageGrader creates a new TestCoverageGrader
func NewTestCoverageGrader() *TestCoverageGrader {
	return &TestCoverageGrader{
		threshold: 80.0,
	}
}

// Name returns the grader name
func (g *TestCoverageGrader) Name() string {
	return "test-coverage"
}

// IsApplicable returns true if there are Go files in ChangedFiles and task type is not chore/spike
func (g *TestCoverageGrader) IsApplicable(input GradeInput) bool {
	// Skip for certain task types
	skipTaskTypes := map[string]bool{
		"chore": true,
		"spike": true,
	}
	if skipTaskTypes[input.TaskType] {
		return false
	}

	// Check if there are any Go files
	for _, file := range input.ChangedFiles {
		if g.isGoFile(file) {
			return true
		}
	}

	return false
}

// Grade runs test coverage and evaluates against threshold
func (g *TestCoverageGrader) Grade(input GradeInput) GradeResult {
	// Skip if not applicable
	if !g.IsApplicable(input) {
		skipReason := "No Go files to check"
		if input.TaskType == "chore" || input.TaskType == "spike" {
			skipReason = fmt.Sprintf("Not applicable for %s tasks", input.TaskType)
		}
		return GradeResult{
			GraderName: g.Name(),
			Passed:     false,
			Score:      0,
			Details:    "",
			Skipped:    true,
			SkipReason: skipReason,
		}
	}

	// Execute go test with coverage
	coverage, err := g.runCoverageTests(input.WorkDir)
	if err != nil {
		// Check if it's a "no test files" error
		if strings.Contains(err.Error(), "no test files") {
			return GradeResult{
				GraderName: g.Name(),
				Passed:     false,
				Score:      0,
				Details:    "No test files found",
				Skipped:    false,
				SkipReason: "",
			}
		}

		// Check if it's a parse error
		if strings.Contains(err.Error(), "Failed to parse coverage") {
			return GradeResult{
				GraderName: g.Name(),
				Passed:     false,
				Score:      0,
				Details:    err.Error(),
				Skipped:    false,
				SkipReason: "",
			}
		}

		// General test execution failure
		return GradeResult{
			GraderName: g.Name(),
			Passed:     false,
			Score:      0,
			Details:    fmt.Sprintf("Test execution failed: %v", err),
			Skipped:    false,
			SkipReason: "",
		}
	}

	// Evaluate against threshold
	passed := g.evaluateCoverage(coverage, g.threshold)
	details := fmt.Sprintf("Coverage: %.1f%% (threshold: %.1f%%)", coverage, g.threshold)

	return GradeResult{
		GraderName: g.Name(),
		Passed:     passed,
		Score:      coverage,
		Details:    details,
		Skipped:    false,
		SkipReason: "",
	}
}

// isGoFile checks if a file has a .go extension
func (g *TestCoverageGrader) isGoFile(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	return ext == ".go"
}

// runCoverageTests executes go test with coverage and returns the coverage percentage
func (g *TestCoverageGrader) runCoverageTests(workDir string) (float64, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Execute go test -cover -v ./... to get verbose output
	cmd := exec.CommandContext(ctx, "go", "test", "-cover", "-v", "./...")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Check for "no test files" case
	if strings.Contains(outputStr, "no test files") || strings.Contains(outputStr, "[no test files]") {
		return 0, fmt.Errorf("no test files")
	}

	// If command failed, check if it's because of test failures (still might have coverage)
	// or because of compilation/other errors
	if err != nil {
		// Try to parse coverage even if tests failed
		coverage, parseErr := g.parseCoverage(outputStr)
		if parseErr == nil {
			// We got coverage even though tests failed - return the coverage
			return coverage, nil
		}
		// Could not parse coverage and command failed
		return 0, fmt.Errorf("command failed: %v, output: %s", err, outputStr)
	}

	// Parse coverage from output
	coverage, err := g.parseCoverage(outputStr)
	if err != nil {
		return 0, err
	}

	// Check if there were actually any tests run
	// When there are no test files, go test still reports "coverage: 0.0%"
	// but doesn't have any test execution output (RUN, PASS, FAIL, ok, etc.)
	if coverage == 0.0 && !g.hasTestExecution(outputStr) {
		return 0, fmt.Errorf("no test files")
	}

	return coverage, nil
}

// hasTestExecution checks if the output indicates any tests were actually run
func (g *TestCoverageGrader) hasTestExecution(output string) bool {
	// Look for indicators that tests were executed
	indicators := []string{
		"=== RUN",   // Test execution
		"--- PASS",  // Test passed
		"--- FAIL",  // Test failed
		"ok  ",      // Package test summary
		"FAIL\t",    // Failed package
	}

	for _, indicator := range indicators {
		if strings.Contains(output, indicator) {
			return true
		}
	}

	return false
}

// parseCoverage extracts coverage percentage from go test output
// Expected format: "coverage: 94.9% of statements" or similar
func (g *TestCoverageGrader) parseCoverage(output string) (float64, error) {
	// Regex pattern to match coverage percentage
	pattern := regexp.MustCompile(`coverage:\s+(\d+\.?\d*)%\s+of\s+statements`)
	matches := pattern.FindStringSubmatch(output)

	if len(matches) < 2 {
		return 0, fmt.Errorf("Failed to parse coverage output")
	}

	coverage, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse coverage value: %v", err)
	}

	return coverage, nil
}

// evaluateCoverage checks if coverage meets or exceeds the threshold
func (g *TestCoverageGrader) evaluateCoverage(coverage, threshold float64) bool {
	return coverage >= threshold
}
