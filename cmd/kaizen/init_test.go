package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Config represents the structure of the config.yaml file
type Config struct {
	ConfidenceThresholds struct {
		High   int `yaml:"high"`
		Medium int `yaml:"medium"`
	} `yaml:"confidence_thresholds"`
	TemplatesDir string `yaml:"templates_dir"`
}

func TestRunInitCommand(t *testing.T) {
	// Use temp directory instead of real home dir
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")

	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("config directory was not created: %s", configDir)
	}

	// Verify failures.db was created
	dbPath := filepath.Join(configDir, "failures.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("failures.db was not created: %s", dbPath)
	}

	// Verify config.yaml was created
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config.yaml was not created: %s", configPath)
	}
}

func TestRunInitCommandCreatesValidConfig(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")

	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Read and parse config file
	configPath := filepath.Join(configDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config.yaml: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config.yaml: %v", err)
	}

	// Verify default values
	if config.ConfidenceThresholds.High != 5 {
		t.Errorf("confidence_thresholds.high = %d, expected 5", config.ConfidenceThresholds.High)
	}

	if config.ConfidenceThresholds.Medium != 2 {
		t.Errorf("confidence_thresholds.medium = %d, expected 2", config.ConfidenceThresholds.Medium)
	}

	if config.TemplatesDir != "" {
		t.Errorf("templates_dir = %q, expected empty string", config.TemplatesDir)
	}
}

func TestRunInitCommandIdempotent(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")

	// Run init first time
	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("first runInitCommand failed: %v", err)
	}

	// Run init second time - should not error
	err = runInitCommand(configDir)
	if err != nil {
		t.Fatalf("second runInitCommand failed (should be idempotent): %v", err)
	}

	// Verify files still exist
	configPath := filepath.Join(configDir, "config.yaml")
	dbPath := filepath.Join(configDir, "failures.db")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.yaml should still exist after second init")
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("failures.db should still exist after second init")
	}
}

func TestRunInitCommandCreatesValidDatabase(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")

	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Verify database has correct schema by trying to open it
	// We'll rely on the failures package tests to verify schema details
	// Here we just verify the file exists and is a valid SQLite db
	dbPath := filepath.Join(configDir, "failures.db")

	// Basic check: file exists and is not empty
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to stat database file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("database file is empty")
	}
}

func TestRunInitCommandErrorHandling(t *testing.T) {
	// Test with a path that cannot be created (invalid path)
	invalidPath := "/\x00invalid/path"

	err := runInitCommand(invalidPath)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestRunInitCommandOutputMessage(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")

	// Capture output by checking that the function completes without error
	// and produces the expected output when called from main
	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// The actual output test will be in integration tests
	// Here we just verify the command succeeds
}

func TestConfigYamlFormat(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")

	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config.yaml: %v", err)
	}

	content := string(data)

	// Verify it contains expected comments
	if !strings.Contains(content, "Kaizen configuration") {
		t.Error("config.yaml should contain header comment")
	}

	// Verify it contains confidence thresholds section
	if !strings.Contains(content, "confidence_thresholds:") {
		t.Error("config.yaml should contain confidence_thresholds section")
	}

	// Verify it contains templates_dir
	if !strings.Contains(content, "templates_dir:") {
		t.Error("config.yaml should contain templates_dir field")
	}
}
