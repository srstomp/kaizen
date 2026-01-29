package main

import (
	"encoding/json"
	"fmt"

	"github.com/srstomp/kaizen/internal/failures"
)

// DetectOutput represents the JSON output from detect-category command
type DetectOutput struct {
	DetectedCategory string   `json:"detected_category"`
	AllCategories    []string `json:"all_categories"`
	Matched          bool     `json:"matched"`
}

// runDetectCategoryCommand executes the detect-category command logic
func runDetectCategoryCommand(details string) (string, error) {
	// Detect the primary category
	primaryCategory, matched := failures.DetectCategory(details)

	// Detect all matching categories
	allCategories := failures.DetectAllCategories(details)

	// Build output structure
	output := DetectOutput{
		DetectedCategory: string(primaryCategory),
		AllCategories:    make([]string, len(allCategories)),
		Matched:          matched,
	}

	// Convert Category types to strings
	for i, cat := range allCategories {
		output.AllCategories[i] = string(cat)
	}

	// If no match, set detected_category to "unknown"
	if !matched {
		output.DetectedCategory = "unknown"
	}

	// Encode to JSON with pretty printing
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding JSON output: %w", err)
	}

	return string(jsonBytes), nil
}
