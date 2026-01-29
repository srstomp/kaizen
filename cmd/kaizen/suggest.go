package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/srstomp/kaizen/internal/failures"
)

// SuggestOutput represents the JSON output from suggest command
type SuggestOutput struct {
	Category    string           `json:"category"`
	Occurrences int              `json:"occurrences"`
	Confidence  string           `json:"confidence"`
	Action      string           `json:"action"`
	FixTask     *SuggestFixTask  `json:"fix_task"`
}

// SuggestFixTask represents the fix task in the suggest output
type SuggestFixTask struct {
	Title         string  `json:"title"`
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	EstimateHours float64 `json:"estimate_hours"`
}

// runSuggestCommand executes the suggest CLI command with default config paths
func runSuggestCommand(taskID, category string) error {
	// Get home directory and build config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "kaizen")
	dbPath := filepath.Join(configDir, "failures.db")

	// Use templates from failures/templates directory in the project
	// In production, this would be configurable
	templatesDir := filepath.Join("failures", "templates")

	output, err := runSuggestCommandWithConfig(taskID, category, dbPath, templatesDir)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// runSuggestCommandWithConfig executes the suggest command with explicit config paths
// This is separated for testing purposes
func runSuggestCommandWithConfig(taskID, category, dbPath, templatesDir string) (string, error) {
	// Open the failures store
	store, err := failures.NewStore(dbPath)
	if err != nil {
		return "", fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	// Get occurrence count
	count, err := store.GetOccurrenceCount(category)
	if err != nil {
		return "", fmt.Errorf("getting occurrence count: %w", err)
	}

	// Calculate confidence
	confidence := failures.CalculateConfidence(count)

	// Build output structure
	output := SuggestOutput{
		Category:    category,
		Occurrences: count,
		Confidence:  string(confidence.Level),
		Action:      string(confidence.Action),
		FixTask:     nil,
	}

	// Try to load template and render fix task
	loader := failures.NewTemplateLoader(templatesDir)
	template, err := loader.LoadTemplate(category)
	if err != nil {
		// Template not found - return output with null fix_task
		// This is not an error, just means no template exists for this category
	} else {
		// Render title and description with variables
		// Support both new and legacy variable names
		vars := map[string]string{
			"task_id":          taskID,
			"original_task_id": taskID, // Legacy template compatibility
			"category":         category,
			"files":            "", // Placeholder for template compatibility
			"details":          "", // Placeholder for template compatibility
		}

		title := template.RenderTitle(vars)
		description := template.RenderDescription(vars)

		output.FixTask = &SuggestFixTask{
			Title:         title,
			Type:          template.FixTask.Type,
			Description:   description,
			EstimateHours: template.FixTask.EstimateHours,
		}
	}

	// Encode to JSON with pretty printing
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding JSON output: %w", err)
	}

	return string(jsonBytes), nil
}
