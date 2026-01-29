package failures

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FixTaskTemplate represents the structure of a fix task template
type FixTaskTemplate struct {
	TitleTemplate       string  `yaml:"title_template"`
	Type                string  `yaml:"type"`
	DescriptionTemplate string  `yaml:"description_template"`
	EstimateHours       float64 `yaml:"estimate_hours"`
}

// Template represents the complete template file structure
type Template struct {
	Category string          `yaml:"category"`
	Prefix   string          `yaml:"prefix"`
	FixTask  FixTaskTemplate `yaml:"fix_task"`
}

// TemplateLoader handles loading templates from a directory
type TemplateLoader struct {
	templatesDir string
}

// NewTemplateLoader creates a new TemplateLoader with the specified templates directory
func NewTemplateLoader(templatesDir string) *TemplateLoader {
	return &TemplateLoader{
		templatesDir: templatesDir,
	}
}

// LoadTemplate loads a specific template by category name
func (l *TemplateLoader) LoadTemplate(category string) (*Template, error) {
	templatePath := filepath.Join(l.templatesDir, category+".yaml")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file for category %s: %w", category, err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse YAML for category %s: %w", category, err)
	}

	return &template, nil
}

// LoadAllTemplates loads all templates from the templates directory
func (l *TemplateLoader) LoadAllTemplates() (map[string]*Template, error) {
	entries, err := os.ReadDir(l.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	templates := make(map[string]*Template)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .yaml files
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		// Extract category name from filename (remove .yaml extension)
		category := strings.TrimSuffix(entry.Name(), ".yaml")

		template, err := l.LoadTemplate(category)
		if err != nil {
			return nil, fmt.Errorf("failed to load template %s: %w", entry.Name(), err)
		}

		templates[category] = template
	}

	return templates, nil
}

// RenderTitle replaces {var} placeholders in the title template with values from the vars map
func (t *Template) RenderTitle(vars map[string]string) string {
	result := t.FixTask.TitleTemplate

	for key, value := range vars {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// RenderDescription replaces {var} placeholders in the description template with values from the vars map
func (t *Template) RenderDescription(vars map[string]string) string {
	result := t.FixTask.DescriptionTemplate

	for key, value := range vars {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}
