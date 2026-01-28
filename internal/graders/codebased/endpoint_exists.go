package codebased

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// EndpointExistsGrader discovers and reports API endpoints from changed files
type EndpointExistsGrader struct{}

// NewEndpointExistsGrader creates a new EndpointExistsGrader
func NewEndpointExistsGrader() *EndpointExistsGrader {
	return &EndpointExistsGrader{}
}

// Name returns the grader name
func (g *EndpointExistsGrader) Name() string {
	return "endpoint-exists"
}

// IsApplicable returns true if there are JS/TS files to check and task type is not chore/spike
func (g *EndpointExistsGrader) IsApplicable(input GradeInput) bool {
	// Skip for certain task types
	skipTaskTypes := map[string]bool{
		"chore": true,
		"spike": true,
	}
	if skipTaskTypes[input.TaskType] {
		return false
	}

	// Skip if no files changed
	if len(input.ChangedFiles) == 0 {
		return false
	}

	// Check if there are any JS/TS files
	for _, file := range input.ChangedFiles {
		if g.isJSFile(file) {
			return true
		}
	}

	return false
}

// Grade discovers and reports API endpoints from changed files
func (g *EndpointExistsGrader) Grade(input GradeInput) GradeResult {
	// Skip if not applicable
	if !g.IsApplicable(input) {
		skipReason := "No changed files"
		if len(input.ChangedFiles) > 0 {
			if input.TaskType == "chore" || input.TaskType == "spike" {
				skipReason = fmt.Sprintf("Not applicable for %s tasks", input.TaskType)
			} else if !g.hasJSFiles(input.ChangedFiles) {
				skipReason = "No JS/TS files to check"
			}
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

	// Extract endpoints from all JS/TS files
	endpoints := g.extractEndpoints(input)

	// Calculate score and build result
	score := float64(0)
	passed := false
	var details string

	if len(endpoints) > 0 {
		score = 100
		passed = true
		details = fmt.Sprintf("Discovered endpoints: %s", strings.Join(endpoints, ", "))
	} else {
		details = "No endpoints discovered"
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

// isJSFile checks if a file is a JavaScript/TypeScript file
func (g *EndpointExistsGrader) isJSFile(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	jsExts := map[string]bool{
		".js":  true,
		".ts":  true,
		".jsx": true,
		".tsx": true,
	}
	return jsExts[ext]
}

// hasJSFiles checks if there are any JS/TS files in the list
func (g *EndpointExistsGrader) hasJSFiles(files []string) bool {
	for _, file := range files {
		if g.isJSFile(file) {
			return true
		}
	}
	return false
}

// extractEndpoints extracts API endpoints from changed files
func (g *EndpointExistsGrader) extractEndpoints(input GradeInput) []string {
	var endpoints []string
	seenEndpoints := make(map[string]bool)

	// Regex pattern to match Express route definitions
	// Pattern: (app|router).(get|post|put|patch|delete)('path' or "path"
	// Note: This pattern does not enforce matching quotes (e.g., '/path" will match).
	// This is acceptable for typical code patterns and simplifies the regex.
	pattern := regexp.MustCompile(`(app|router)\.(get|post|put|patch|delete)\s*\(\s*['"]([^'"]+)['"]`)

	for _, file := range input.ChangedFiles {
		// Skip non-JS/TS files
		if !g.isJSFile(file) {
			continue
		}

		// Resolve path relative to WorkDir if not absolute
		filePath := file
		if !filepath.IsAbs(file) {
			filePath = filepath.Join(input.WorkDir, file)
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		// Find all matches in the file
		matches := pattern.FindAllSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 4 {
				method := strings.ToUpper(string(match[2]))
				path := string(match[3])
				endpoint := fmt.Sprintf("%s %s", method, path)

				// Only add unique endpoints
				if !seenEndpoints[endpoint] {
					seenEndpoints[endpoint] = true
					endpoints = append(endpoints, endpoint)
				}
			}
		}
	}

	return endpoints
}
