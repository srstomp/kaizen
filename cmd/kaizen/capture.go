package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/srstomp/kaizen/internal/failures"
)

// CaptureOutput represents the JSON output from capture command
type CaptureOutput struct {
	Success  bool   `json:"success"`
	TaskID   string `json:"task_id,omitempty"`
	Category string `json:"category,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// runCaptureCommand executes the capture CLI command with default config paths
func runCaptureCommand(taskID, category, details, source string) (string, error) {
	// Get home directory and build config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return buildErrorOutput(fmt.Errorf("failed to get home directory: %w", err))
	}

	configDir := filepath.Join(homeDir, ".config", "kaizen")
	dbPath := filepath.Join(configDir, "failures.db")

	return runCaptureCommandWithConfig(taskID, category, details, source, dbPath)
}

// runCaptureCommandWithConfig executes the capture command with explicit config paths
// This is separated for testing purposes
func runCaptureCommandWithConfig(taskID, category, details, source, dbPath string) (string, error) {
	// Open the failures store
	store, err := failures.NewStore(dbPath)
	if err != nil {
		return buildErrorOutput(fmt.Errorf("opening database: %w", err))
	}
	defer store.Close()

	// Create failure record
	failure := failures.Failure{
		TaskID:   taskID,
		Category: category,
		Details:  details,
		Source:   source,
	}

	// Insert failure into database
	if err := store.Insert(failure); err != nil {
		return buildErrorOutput(fmt.Errorf("inserting failure: %w", err))
	}

	// Increment category occurrence count
	if err := store.IncrementCount(category); err != nil {
		return buildErrorOutput(fmt.Errorf("incrementing category count: %w", err))
	}

	// Build success output
	output := CaptureOutput{
		Success:  true,
		TaskID:   taskID,
		Category: category,
		Message:  "Failure captured successfully",
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return buildErrorOutput(fmt.Errorf("encoding JSON output: %w", err))
	}

	return string(jsonBytes), nil
}

// buildErrorOutput creates a JSON error response
func buildErrorOutput(err error) (string, error) {
	output := CaptureOutput{
		Success: false,
		Error:   err.Error(),
	}

	jsonBytes, jsonErr := json.MarshalIndent(output, "", "  ")
	if jsonErr != nil {
		// If we can't even marshal the error, return a simple JSON string
		return `{"success": false, "error": "failed to marshal error output"}`, err
	}

	return string(jsonBytes), err
}
