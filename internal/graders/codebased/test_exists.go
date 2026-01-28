package codebased

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TestExistsGrader checks if code files have corresponding test files
type TestExistsGrader struct{}

// NewTestExistsGrader creates a new TestExistsGrader
func NewTestExistsGrader() *TestExistsGrader {
	return &TestExistsGrader{}
}

// Name returns the grader name
func (g *TestExistsGrader) Name() string {
	return "test-exists"
}

// IsApplicable returns true if there are code files (non-test) to check
func (g *TestExistsGrader) IsApplicable(input GradeInput) bool {
	// Skip for certain task types
	skipTaskTypes := map[string]bool{
		"chore": true,
		"spike": true,
	}
	if skipTaskTypes[input.TaskType] {
		return false
	}

	// Check if there are any code files (non-test, non-config)
	for _, file := range input.ChangedFiles {
		if g.isCodeFile(file) {
			return true
		}
	}

	return false
}

// Grade checks if code files have corresponding test files
func (g *TestExistsGrader) Grade(input GradeInput) GradeResult {
	// Skip if not applicable
	if !g.IsApplicable(input) {
		skipReason := "No code files to check"
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

	filesWithTests := 0
	totalCodeFiles := 0
	missingTests := []string{}

	for _, file := range input.ChangedFiles {
		// Skip non-code files
		if !g.isCodeFile(file) {
			continue
		}

		totalCodeFiles++

		// Get expected test file patterns
		testFiles := g.getTestFilePaths(file, input.WorkDir)

		// Check if any test file exists
		testExists := false
		for _, testFile := range testFiles {
			if _, err := os.Stat(testFile); err == nil {
				testExists = true
				break
			}
		}

		if testExists {
			filesWithTests++
		} else {
			missingTests = append(missingTests, filepath.Base(file))
		}
	}

	// Calculate score
	score := float64(0)
	if totalCodeFiles > 0 {
		score = float64(filesWithTests) / float64(totalCodeFiles) * 100
	}
	passed := len(missingTests) == 0

	// Build details message
	var details string
	if passed {
		details = fmt.Sprintf("All %d code files have tests", totalCodeFiles)
	} else {
		details = fmt.Sprintf("%d/%d code files have tests, missing tests for: %v", filesWithTests, totalCodeFiles, missingTests)
	}

	return GradeResult{
		GraderName: g.Name(),
		Passed:     passed,
		Score:      score,
		Details:    details,
		Skipped:    false,
		SkipReason: "",
	}
}

// isCodeFile checks if a file is a code file (not test, not config)
func (g *TestExistsGrader) isCodeFile(file string) bool {
	// Skip test files
	if g.isTestFile(file) {
		return false
	}

	// Skip config/data files
	ext := strings.ToLower(filepath.Ext(file))
	configExts := map[string]bool{
		".yaml": true, ".yml": true, ".json": true,
		".toml": true, ".ini": true, ".conf": true,
		".md": true, ".txt": true, ".xml": true,
	}
	if configExts[ext] {
		return false
	}

	// Check if it's a supported code file extension
	codeExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true,
		".jsx": true, ".tsx": true,
	}
	return codeExts[ext]
}

// isTestFile checks if a file is a test file
func (g *TestExistsGrader) isTestFile(file string) bool {
	base := filepath.Base(file)

	// Go test files
	if strings.HasSuffix(base, "_test.go") {
		return true
	}

	// Python test files
	if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
		return true
	}
	if strings.HasSuffix(base, "_test.py") {
		return true
	}

	// JavaScript/TypeScript test files
	if strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
		return true
	}

	return false
}

// getTestFilePaths returns possible test file paths for a code file
func (g *TestExistsGrader) getTestFilePaths(file string, workDir string) []string {
	var testPaths []string

	// Resolve file path
	filePath := file
	if !filepath.IsAbs(file) {
		filePath = filepath.Join(workDir, file)
	}

	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	switch ext {
	case ".go":
		// Go pattern: foo.go -> foo_test.go
		testPaths = append(testPaths, filepath.Join(dir, name+"_test.go"))

	case ".py":
		// Python patterns: bar.py -> test_bar.py or bar_test.py
		testPaths = append(testPaths, filepath.Join(dir, "test_"+base))
		testPaths = append(testPaths, filepath.Join(dir, name+"_test.py"))

	case ".js", ".jsx":
		// JavaScript patterns: baz.js -> baz.test.js or baz.spec.js
		testPaths = append(testPaths, filepath.Join(dir, name+".test"+ext))
		testPaths = append(testPaths, filepath.Join(dir, name+".spec"+ext))

	case ".ts", ".tsx":
		// TypeScript patterns: qux.ts -> qux.test.ts or qux.spec.ts
		testPaths = append(testPaths, filepath.Join(dir, name+".test"+ext))
		testPaths = append(testPaths, filepath.Join(dir, name+".spec"+ext))
	}

	return testPaths
}
