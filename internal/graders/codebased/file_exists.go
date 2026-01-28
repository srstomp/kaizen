package codebased

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileExistsGrader checks that changed files actually exist in the working directory
type FileExistsGrader struct{}

// NewFileExistsGrader creates a new FileExistsGrader
func NewFileExistsGrader() *FileExistsGrader {
	return &FileExistsGrader{}
}

// Name returns the grader name
func (g *FileExistsGrader) Name() string {
	return "file-exists"
}

// IsApplicable returns true if there are changed files to check
func (g *FileExistsGrader) IsApplicable(input GradeInput) bool {
	return len(input.ChangedFiles) > 0
}

// Grade checks if all changed files exist in the working directory
func (g *FileExistsGrader) Grade(input GradeInput) GradeResult {
	// Skip if not applicable
	if !g.IsApplicable(input) {
		return GradeResult{
			GraderName: g.Name(),
			Passed:     false,
			Score:      0,
			Details:    "",
			Skipped:    true,
			SkipReason: "No changed files to check",
		}
	}

	existingFiles := 0
	missingFiles := []string{}

	for _, file := range input.ChangedFiles {
		// Resolve path relative to WorkDir if not absolute
		filePath := file
		if !filepath.IsAbs(file) {
			filePath = filepath.Join(input.WorkDir, file)
		}

		// Check if file exists
		if _, err := os.Stat(filePath); err == nil {
			existingFiles++
		} else if os.IsNotExist(err) {
			missingFiles = append(missingFiles, file)
		} else {
			// Other error (permission, etc.) - treat as missing
			missingFiles = append(missingFiles, file)
		}
	}

	totalFiles := len(input.ChangedFiles)
	score := float64(existingFiles) / float64(totalFiles) * 100
	passed := len(missingFiles) == 0

	// Build details message
	var details string
	if passed {
		details = fmt.Sprintf("All %d files exist", totalFiles)
	} else {
		details = fmt.Sprintf("%d/%d files exist, missing: %v", existingFiles, totalFiles, missingFiles)
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
