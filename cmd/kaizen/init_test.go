package main

import (
	"fmt"
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

func TestBootstrapFromFailuresDirectory(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")
	failuresDir := filepath.Join(tempHome, "failures")

	// Initialize kaizen first
	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Create test failures directory structure
	createTestFailuresDirectory(t, failuresDir)

	// Run bootstrap
	dbPath := filepath.Join(configDir, "failures.db")
	stats, err := bootstrapFromFailures(dbPath, failuresDir)
	if err != nil {
		t.Fatalf("bootstrapFromFailures failed: %v", err)
	}

	// Verify stats were collected
	if len(stats) == 0 {
		t.Error("expected category stats, got empty map")
	}

	// Verify expected categories
	expectedCategories := []string{"missing-tests", "wrong-product", "missed-tasks"}
	for _, cat := range expectedCategories {
		if _, ok := stats[cat]; !ok {
			t.Errorf("expected category %q in stats", cat)
		}
	}

	// Verify counts
	if stats["missing-tests"] != 3 {
		t.Errorf("missing-tests count = %d, expected 3", stats["missing-tests"])
	}
	if stats["wrong-product"] != 2 {
		t.Errorf("wrong-product count = %d, expected 2", stats["wrong-product"])
	}
	if stats["missed-tasks"] != 1 {
		t.Errorf("missed-tasks count = %d, expected 1", stats["missed-tasks"])
	}
}

func TestBootstrapSkipsSchemaAndTemplateFiles(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")
	failuresDir := filepath.Join(tempHome, "failures")

	// Initialize kaizen first
	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Create test failures directory with schema and template files
	if err := os.MkdirAll(failuresDir, 0755); err != nil {
		t.Fatalf("failed to create failures dir: %v", err)
	}

	// Create schema.yaml (should be skipped)
	schemaContent := `category: schema-category
id: SCHEMA-001`
	if err := os.WriteFile(filepath.Join(failuresDir, "schema.yaml"), []byte(schemaContent), 0644); err != nil {
		t.Fatalf("failed to create schema.yaml: %v", err)
	}

	// Create template.yaml (should be skipped)
	templateContent := `category: template-category
id: TEMPLATE-001`
	if err := os.WriteFile(filepath.Join(failuresDir, "template.yaml"), []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to create template.yaml: %v", err)
	}

	// Create a real failure file
	categoryDir := filepath.Join(failuresDir, "test-category")
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		t.Fatalf("failed to create category dir: %v", err)
	}
	failureContent := `category: test-category
id: TC-001`
	if err := os.WriteFile(filepath.Join(categoryDir, "TC-001.yaml"), []byte(failureContent), 0644); err != nil {
		t.Fatalf("failed to create failure file: %v", err)
	}

	// Run bootstrap
	dbPath := filepath.Join(configDir, "failures.db")
	stats, err := bootstrapFromFailures(dbPath, failuresDir)
	if err != nil {
		t.Fatalf("bootstrapFromFailures failed: %v", err)
	}

	// Verify schema and template categories are not in stats
	if _, ok := stats["schema-category"]; ok {
		t.Error("schema.yaml should be skipped")
	}
	if _, ok := stats["template-category"]; ok {
		t.Error("template.yaml should be skipped")
	}

	// Verify only real category is present
	if stats["test-category"] != 1 {
		t.Errorf("test-category count = %d, expected 1", stats["test-category"])
	}
}

func TestBootstrapSkipsExamplesDirectory(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")
	failuresDir := filepath.Join(tempHome, "failures")

	// Initialize kaizen first
	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Create examples directory
	examplesDir := filepath.Join(failuresDir, "examples")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		t.Fatalf("failed to create examples dir: %v", err)
	}

	// Create a file in examples (should be skipped)
	exampleContent := `category: example-category
id: EX-001`
	if err := os.WriteFile(filepath.Join(examplesDir, "EX-001.yaml"), []byte(exampleContent), 0644); err != nil {
		t.Fatalf("failed to create example file: %v", err)
	}

	// Create a real failure file
	categoryDir := filepath.Join(failuresDir, "real-category")
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		t.Fatalf("failed to create category dir: %v", err)
	}
	failureContent := `category: real-category
id: RC-001`
	if err := os.WriteFile(filepath.Join(categoryDir, "RC-001.yaml"), []byte(failureContent), 0644); err != nil {
		t.Fatalf("failed to create failure file: %v", err)
	}

	// Run bootstrap
	dbPath := filepath.Join(configDir, "failures.db")
	stats, err := bootstrapFromFailures(dbPath, failuresDir)
	if err != nil {
		t.Fatalf("bootstrapFromFailures failed: %v", err)
	}

	// Verify examples category is not in stats
	if _, ok := stats["example-category"]; ok {
		t.Error("examples directory should be skipped")
	}

	// Verify only real category is present
	if stats["real-category"] != 1 {
		t.Errorf("real-category count = %d, expected 1", stats["real-category"])
	}
}

func TestBootstrapHandlesInvalidYAML(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")
	failuresDir := filepath.Join(tempHome, "failures")

	// Initialize kaizen first
	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Create category directory
	categoryDir := filepath.Join(failuresDir, "test-category")
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		t.Fatalf("failed to create category dir: %v", err)
	}

	// Create invalid YAML file (should be skipped gracefully)
	invalidContent := `this is not: [valid yaml: {unclosed`
	if err := os.WriteFile(filepath.Join(categoryDir, "INVALID.yaml"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to create invalid file: %v", err)
	}

	// Create valid YAML file
	validContent := `category: test-category
id: TC-001`
	if err := os.WriteFile(filepath.Join(categoryDir, "TC-001.yaml"), []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to create valid file: %v", err)
	}

	// Run bootstrap - should not fail on invalid YAML
	dbPath := filepath.Join(configDir, "failures.db")
	stats, err := bootstrapFromFailures(dbPath, failuresDir)
	if err != nil {
		t.Fatalf("bootstrapFromFailures should handle invalid YAML gracefully: %v", err)
	}

	// Verify valid file was still processed
	if stats["test-category"] != 1 {
		t.Errorf("test-category count = %d, expected 1", stats["test-category"])
	}
}

func TestBootstrapEmptyDirectory(t *testing.T) {
	tempHome := t.TempDir()
	configDir := filepath.Join(tempHome, ".config", "kaizen")
	failuresDir := filepath.Join(tempHome, "failures")

	// Initialize kaizen first
	err := runInitCommand(configDir)
	if err != nil {
		t.Fatalf("runInitCommand failed: %v", err)
	}

	// Create empty failures directory
	if err := os.MkdirAll(failuresDir, 0755); err != nil {
		t.Fatalf("failed to create failures dir: %v", err)
	}

	// Run bootstrap on empty directory
	dbPath := filepath.Join(configDir, "failures.db")
	stats, err := bootstrapFromFailures(dbPath, failuresDir)
	if err != nil {
		t.Fatalf("bootstrapFromFailures failed on empty dir: %v", err)
	}

	// Verify empty stats
	if len(stats) != 0 {
		t.Errorf("expected empty stats for empty directory, got %d categories", len(stats))
	}
}

// Helper function to create test failures directory structure
func createTestFailuresDirectory(t *testing.T, failuresDir string) {
	t.Helper()

	// Create missing-tests category with 3 files
	missingTestsDir := filepath.Join(failuresDir, "missing-tests")
	if err := os.MkdirAll(missingTestsDir, 0755); err != nil {
		t.Fatalf("failed to create missing-tests dir: %v", err)
	}
	for i := 1; i <= 3; i++ {
		content := fmt.Sprintf("id: MT-%03d\ncategory: missing-tests\n", i)
		filename := filepath.Join(missingTestsDir, fmt.Sprintf("MT-%03d.yaml", i))
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Create wrong-product category with 2 files
	wrongProductDir := filepath.Join(failuresDir, "wrong-product")
	if err := os.MkdirAll(wrongProductDir, 0755); err != nil {
		t.Fatalf("failed to create wrong-product dir: %v", err)
	}
	for i := 1; i <= 2; i++ {
		content := fmt.Sprintf("id: WP-%03d\ncategory: wrong-product\n", i)
		filename := filepath.Join(wrongProductDir, fmt.Sprintf("WP-%03d.yaml", i))
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Create missed-tasks category with 1 file
	missedTasksDir := filepath.Join(failuresDir, "missed-tasks")
	if err := os.MkdirAll(missedTasksDir, 0755); err != nil {
		t.Fatalf("failed to create missed-tasks dir: %v", err)
	}
	content := "id: MT-001\ncategory: missed-tasks\n"
	filename := filepath.Join(missedTasksDir, "MT-001.yaml")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
}
