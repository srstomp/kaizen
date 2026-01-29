package failures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

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

func TestNewTemplateLoader(t *testing.T) {
	loader := NewTemplateLoader("/path/to/templates")
	if loader == nil {
		t.Fatal("NewTemplateLoader returned nil")
	}
}

func TestLoadTemplate(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	templatesDir := filepath.Join(projectRoot, "failures", "templates")
	loader := NewTemplateLoader(templatesDir)

	tests := []struct {
		name         string
		category     string
		wantCategory string
		wantPrefix   string
		wantErr      bool
	}{
		{
			name:         "load missing-tests template",
			category:     "missing-tests",
			wantCategory: "missing-tests",
			wantPrefix:   "MT",
			wantErr:      false,
		},
		{
			name:         "load scope-creep template",
			category:     "scope-creep",
			wantCategory: "scope-creep",
			wantPrefix:   "SC",
			wantErr:      false,
		},
		{
			name:         "load wrong-product template",
			category:     "wrong-product",
			wantCategory: "wrong-product",
			wantPrefix:   "WP",
			wantErr:      false,
		},
		{
			name:     "non-existent template",
			category: "does-not-exist",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := loader.LoadTemplate(tt.category)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if template.Category != tt.wantCategory {
				t.Errorf("Expected category %s, got %s", tt.wantCategory, template.Category)
			}

			if template.Prefix != tt.wantPrefix {
				t.Errorf("Expected prefix %s, got %s", tt.wantPrefix, template.Prefix)
			}
		})
	}
}

func TestLoadAllTemplates(t *testing.T) {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	templatesDir := filepath.Join(projectRoot, "failures", "templates")
	loader := NewTemplateLoader(templatesDir)

	templates, err := loader.LoadAllTemplates()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedCategories := []string{"missing-tests", "scope-creep", "wrong-product"}
	for _, category := range expectedCategories {
		if _, ok := templates[category]; !ok {
			t.Errorf("Expected template for category %s", category)
		}
	}

	if len(templates) != len(expectedCategories) {
		t.Errorf("Expected %d templates, got %d", len(expectedCategories), len(templates))
	}
}

func TestRenderTitle(t *testing.T) {
	template := &Template{
		Category: "missing-tests",
		Prefix:   "MT",
		FixTask: FixTaskTemplate{
			TitleTemplate:       "Fix: Add tests for {original_task_id}",
			Type:                "bug",
			DescriptionTemplate: "Description",
			EstimateHours:       1.0,
		},
	}

	tests := []struct {
		name string
		vars map[string]string
		want string
	}{
		{
			name: "replace single variable",
			vars: map[string]string{
				"original_task_id": "TASK-123",
			},
			want: "Fix: Add tests for TASK-123",
		},
		{
			name: "missing variable leaves placeholder",
			vars: map[string]string{},
			want: "Fix: Add tests for {original_task_id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := template.RenderTitle(tt.vars)
			if got != tt.want {
				t.Errorf("RenderTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderDescription(t *testing.T) {
	template := &Template{
		Category: "missing-tests",
		Prefix:   "MT",
		FixTask: FixTaskTemplate{
			TitleTemplate: "Title",
			Type:          "bug",
			DescriptionTemplate: `Quality review failed on {original_task_id}.

Issue: Missing test coverage
Files: {files}
Details: {details}

Acceptance criteria:
- [ ] Add unit tests for each file
- [ ] Tests pass locally
- [ ] Coverage meets threshold`,
			EstimateHours: 1.0,
		},
	}

	tests := []struct {
		name string
		vars map[string]string
		want string
	}{
		{
			name: "replace all variables",
			vars: map[string]string{
				"original_task_id": "TASK-123",
				"files":            "auth.go, login.go",
				"details":          "Missing unit tests",
			},
			want: `Quality review failed on TASK-123.

Issue: Missing test coverage
Files: auth.go, login.go
Details: Missing unit tests

Acceptance criteria:
- [ ] Add unit tests for each file
- [ ] Tests pass locally
- [ ] Coverage meets threshold`,
		},
		{
			name: "partial replacement",
			vars: map[string]string{
				"original_task_id": "TASK-456",
			},
			want: `Quality review failed on TASK-456.

Issue: Missing test coverage
Files: {files}
Details: {details}

Acceptance criteria:
- [ ] Add unit tests for each file
- [ ] Tests pass locally
- [ ] Coverage meets threshold`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := template.RenderDescription(tt.vars)
			if got != tt.want {
				t.Errorf("RenderDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

