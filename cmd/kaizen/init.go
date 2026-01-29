package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/srstomp/kaizen/internal/failures"
)

const defaultConfigYAML = `# Kaizen configuration
confidence_thresholds:
  high: 5      # 5+ occurrences = auto-create
  medium: 2    # 2-4 occurrences = suggest
  # 1 occurrence = log-only (implicit)

templates_dir: ""  # Empty means use built-in templates
`

// runInitCommand initializes the kaizen configuration directory and database.
// It creates the config directory, failures database, and config file with defaults.
// This function is idempotent - safe to run multiple times.
func runInitCommand(configDir string) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Create failures database
	dbPath := filepath.Join(configDir, "failures.db")
	store, err := failures.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("creating failures database: %w", err)
	}
	// Close the database connection (we just needed to initialize it)
	defer store.Close()

	// Create config.yaml with defaults
	configPath := filepath.Join(configDir, "config.yaml")
	// Only create config if it doesn't exist (idempotent)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0644); err != nil {
			return fmt.Errorf("creating config file: %w", err)
		}
	}

	return nil
}
