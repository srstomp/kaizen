package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/srstomp/kaizen/internal/failures"
	"gopkg.in/yaml.v3"
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

// bootstrapFromFailures scans a failures directory and populates category_stats.
// Returns a map of category -> count for reporting purposes.
func bootstrapFromFailures(dbPath, failuresDir string) (map[string]int, error) {
	// Open the database
	store, err := failures.NewStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	// Track category stats
	categoryStats := make(map[string]int)
	categoryFirstSeen := make(map[string]time.Time)
	categoryLastSeen := make(map[string]time.Time)

	// Walk the failures directory
	err = filepath.Walk(failuresDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: failed to access %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			// Skip examples directory
			if info.Name() == "examples" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip non-YAML files
		if !strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml") {
			return nil
		}

		// Skip schema.yaml and template.yaml
		if info.Name() == "schema.yaml" || info.Name() == "template.yaml" {
			return nil
		}

		// Read and parse the YAML file
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", path, err)
			return nil // Continue walking
		}

		var failureCase FailureCase
		if err := yaml.Unmarshal(data, &failureCase); err != nil {
			log.Printf("Warning: failed to parse %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip if no category found
		if failureCase.Category == "" {
			log.Printf("Warning: no category in %s", path)
			return nil // Continue walking
		}

		// Update stats
		categoryStats[failureCase.Category]++

		// Parse discovered date for timestamps
		var discoveredTime time.Time
		if failureCase.Discovered != "" {
			discoveredTime, err = time.Parse("2006-01-02", failureCase.Discovered)
			if err != nil {
				// If parsing fails, use file modification time
				discoveredTime = info.ModTime()
			}
		} else {
			// Use file modification time if no discovered date
			discoveredTime = info.ModTime()
		}

		// Track first and last seen
		if firstSeen, ok := categoryFirstSeen[failureCase.Category]; !ok || discoveredTime.Before(firstSeen) {
			categoryFirstSeen[failureCase.Category] = discoveredTime
		}
		if lastSeen, ok := categoryLastSeen[failureCase.Category]; !ok || discoveredTime.After(lastSeen) {
			categoryLastSeen[failureCase.Category] = discoveredTime
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking failures directory: %w", err)
	}

	// Persist stats to database
	for category, count := range categoryStats {
		firstSeen := categoryFirstSeen[category]
		lastSeen := categoryLastSeen[category]

		if err := store.UpsertCategoryStats(category, count, firstSeen, lastSeen); err != nil {
			return nil, fmt.Errorf("upserting stats for %s: %w", category, err)
		}
	}

	return categoryStats, nil
}
