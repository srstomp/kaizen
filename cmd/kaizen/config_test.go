package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// Thresholds defines the quality thresholds for release gates
type Thresholds struct {
	Eval       EvalThresholds       `yaml:"eval"`
	Meta       MetaThresholds       `yaml:"meta"`
	Regression RegressionThresholds `yaml:"regression"`
}

// EvalThresholds defines thresholds for task evaluation metrics
type EvalThresholds struct {
	Accuracy float64 `yaml:"accuracy"`
}

// MetaThresholds defines thresholds for meta-evaluation metrics
type MetaThresholds struct {
	Consistency float64 `yaml:"consistency"`
}

// RegressionThresholds defines acceptable regression limits
type RegressionThresholds struct {
	MaxDropPercent float64 `yaml:"max_drop_percent"`
}

// loadThresholds loads threshold configuration from YAML file
func loadThresholds(configPath string) (*Thresholds, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var thresholds Thresholds
	if err := yaml.Unmarshal(data, &thresholds); err != nil {
		return nil, err
	}

	return &thresholds, nil
}

func TestLoadThresholds(t *testing.T) {
	// Find the project root by looking for go.mod
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Navigate up to find yokay-evals directory
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("could not find project root (go.mod)")
		}
		projectRoot = parent
	}

	configPath := filepath.Join(projectRoot, "config", "thresholds.yaml")

	thresholds, err := loadThresholds(configPath)
	if err != nil {
		t.Fatalf("failed to load thresholds: %v", err)
	}

	// Test eval accuracy threshold
	if thresholds.Eval.Accuracy <= 0 {
		t.Errorf("eval accuracy threshold must be positive, got: %.1f", thresholds.Eval.Accuracy)
	}
	if thresholds.Eval.Accuracy != 90.0 {
		t.Errorf("eval accuracy threshold should be 90.0, got: %.1f", thresholds.Eval.Accuracy)
	}

	// Test meta consistency threshold
	if thresholds.Meta.Consistency <= 0 {
		t.Errorf("meta consistency threshold must be positive, got: %.1f", thresholds.Meta.Consistency)
	}
	if thresholds.Meta.Consistency != 95.0 {
		t.Errorf("meta consistency threshold should be 95.0, got: %.1f", thresholds.Meta.Consistency)
	}

	// Test regression threshold
	if thresholds.Regression.MaxDropPercent <= 0 {
		t.Errorf("regression max_drop_percent must be positive, got: %.1f", thresholds.Regression.MaxDropPercent)
	}
	if thresholds.Regression.MaxDropPercent != 5.0 {
		t.Errorf("regression max_drop_percent should be 5.0, got: %.1f", thresholds.Regression.MaxDropPercent)
	}
}

func TestThresholdsWithinBounds(t *testing.T) {
	// Find the project root by looking for go.mod
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("could not find project root (go.mod)")
		}
		projectRoot = parent
	}

	configPath := filepath.Join(projectRoot, "config", "thresholds.yaml")

	thresholds, err := loadThresholds(configPath)
	if err != nil {
		t.Fatalf("failed to load thresholds: %v", err)
	}

	// Verify thresholds are percentages (0-100)
	if thresholds.Eval.Accuracy < 0 || thresholds.Eval.Accuracy > 100 {
		t.Errorf("eval accuracy must be between 0-100, got: %.1f", thresholds.Eval.Accuracy)
	}

	if thresholds.Meta.Consistency < 0 || thresholds.Meta.Consistency > 100 {
		t.Errorf("meta consistency must be between 0-100, got: %.1f", thresholds.Meta.Consistency)
	}

	if thresholds.Regression.MaxDropPercent < 0 || thresholds.Regression.MaxDropPercent > 100 {
		t.Errorf("regression max_drop_percent must be between 0-100, got: %.1f", thresholds.Regression.MaxDropPercent)
	}
}

func TestThresholdsFileExists(t *testing.T) {
	// Find the project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("could not find project root (go.mod)")
		}
		projectRoot = parent
	}

	configPath := filepath.Join(projectRoot, "config", "thresholds.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("thresholds.yaml does not exist at: %s", configPath)
	}
}
