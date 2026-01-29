package failures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// FixTaskTemplate represents the structure of a fix task template
type FixTaskTemplate struct {
	TitleTemplate       string `yaml:"title_template"`
	Type                string `yaml:"type"`
	DescriptionTemplate string `yaml:"description_template"`
	EstimateHours       float64 `yaml:"estimate_hours"`
}

// Template represents the complete template file structure
type Template struct {
	Category string           `yaml:"category"`
	Prefix   string           `yaml:"prefix"`
	FixTask  FixTaskTemplate  `yaml:"fix_task"`
}

func TestTemplateFilesExist(t *testing.T) {
	// Get project root (go up from internal/failures to project root)
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	templatesDir := filepath.Join(projectRoot, "failures", "templates")

	tests := []struct {
		name     string
		filename string
	}{
		{"missing-tests template exists", "missing-tests.yaml"},
		{"scope-creep template exists", "scope-creep.yaml"},
		{"wrong-product template exists", "wrong-product.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templatePath := filepath.Join(templatesDir, tt.filename)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				t.Errorf("Template file does not exist: %s", templatePath)
			}
		})
	}
}

func TestMissingTestsTemplate(t *testing.T) {
	template := loadTemplate(t, "missing-tests.yaml")

	// Verify category
	if template.Category != "missing-tests" {
		t.Errorf("Expected category 'missing-tests', got '%s'", template.Category)
	}

	// Verify prefix
	if template.Prefix != "MT" {
		t.Errorf("Expected prefix 'MT', got '%s'", template.Prefix)
	}

	// Verify fix_task fields
	if template.FixTask.Type != "bug" {
		t.Errorf("Expected type 'bug', got '%s'", template.FixTask.Type)
	}

	if template.FixTask.EstimateHours != 1.0 {
		t.Errorf("Expected estimate 1 hour, got %f", template.FixTask.EstimateHours)
	}

	// Verify template variables are present
	if !strings.Contains(template.FixTask.TitleTemplate, "{original_task_id}") {
		t.Error("Title template missing {original_task_id} variable")
	}

	if !strings.Contains(template.FixTask.DescriptionTemplate, "{original_task_id}") {
		t.Error("Description template missing {original_task_id} variable")
	}

	// Verify acceptance criteria format
	if !strings.Contains(template.FixTask.DescriptionTemplate, "- [ ]") {
		t.Error("Description template missing acceptance criteria checkboxes")
	}
}

func TestScopeCreepTemplate(t *testing.T) {
	template := loadTemplate(t, "scope-creep.yaml")

	// Verify category
	if template.Category != "scope-creep" {
		t.Errorf("Expected category 'scope-creep', got '%s'", template.Category)
	}

	// Verify prefix
	if template.Prefix != "SC" {
		t.Errorf("Expected prefix 'SC', got '%s'", template.Prefix)
	}

	// Verify fix_task fields
	if template.FixTask.Type != "bug" {
		t.Errorf("Expected type 'bug', got '%s'", template.FixTask.Type)
	}

	if template.FixTask.EstimateHours != 0.5 {
		t.Errorf("Expected estimate 0.5 hours, got %f", template.FixTask.EstimateHours)
	}

	// Verify template variables
	if !strings.Contains(template.FixTask.TitleTemplate, "{original_task_id}") {
		t.Error("Title template missing {original_task_id} variable")
	}
}

func TestWrongProductTemplate(t *testing.T) {
	template := loadTemplate(t, "wrong-product.yaml")

	// Verify category
	if template.Category != "wrong-product" {
		t.Errorf("Expected category 'wrong-product', got '%s'", template.Category)
	}

	// Verify prefix
	if template.Prefix != "WP" {
		t.Errorf("Expected prefix 'WP', got '%s'", template.Prefix)
	}

	// Verify fix_task fields
	if template.FixTask.Type != "bug" {
		t.Errorf("Expected type 'bug', got '%s'", template.FixTask.Type)
	}

	if template.FixTask.EstimateHours != 1.5 {
		t.Errorf("Expected estimate 1.5 hours, got %f", template.FixTask.EstimateHours)
	}

	// Verify template variables
	if !strings.Contains(template.FixTask.TitleTemplate, "{original_task_id}") {
		t.Error("Title template missing {original_task_id} variable")
	}
}

func TestAllTemplatesHaveValidYAML(t *testing.T) {
	templateFiles := []string{
		"missing-tests.yaml",
		"scope-creep.yaml",
		"wrong-product.yaml",
	}

	for _, filename := range templateFiles {
		t.Run(filename, func(t *testing.T) {
			template := loadTemplate(t, filename)

			// Verify all required fields are present
			if template.Category == "" {
				t.Error("Category is empty")
			}
			if template.Prefix == "" {
				t.Error("Prefix is empty")
			}
			if template.FixTask.TitleTemplate == "" {
				t.Error("Title template is empty")
			}
			if template.FixTask.Type == "" {
				t.Error("Type is empty")
			}
			if template.FixTask.DescriptionTemplate == "" {
				t.Error("Description template is empty")
			}
			if template.FixTask.EstimateHours <= 0 {
				t.Error("Estimate hours must be positive")
			}
		})
	}
}

// Helper function to load and parse a template
func loadTemplate(t *testing.T, filename string) Template {
	t.Helper()

	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	templatePath := filepath.Join(projectRoot, "failures", "templates", filename)

	data, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template file %s: %v", filename, err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		t.Fatalf("Failed to parse YAML in %s: %v", filename, err)
	}

	return template
}

