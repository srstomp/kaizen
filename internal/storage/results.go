package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/srstomp/kaizen/internal/graders/codebased"
)

// EvalResult represents the output from grade-task command
type EvalResult struct {
	TaskID        string                  `json:"task_id"`
	Timestamp     string                  `json:"timestamp"`
	Results       []codebased.GradeResult `json:"results"`
	OverallPassed bool                    `json:"overall_passed"`
	OverallScore  float64                 `json:"overall_score"`
}

// MetaResult represents a consistency evaluation result from meta-evals
type MetaResult struct {
	Timestamp             string  `json:"timestamp"`
	Agent                 string  `json:"agent"`
	BoundaryType          string  `json:"boundary_type"`
	ConsistencyPercentage float64 `json:"consistency_percentage"`
	ConsistentCount       int     `json:"consistent_count"`
	TotalCount            int     `json:"total_count"`
}

// TaskQualityIssue represents a quality check issue
type TaskQualityIssue struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

// TaskQualityResult represents the output of task quality grading
type TaskQualityResult struct {
	TaskID     string             `json:"task_id"`
	Passed     bool               `json:"passed"`
	Score      float64            `json:"score"`
	Issues     []TaskQualityIssue `json:"issues"`
	Suggestion string             `json:"suggestion"`
}

// ResultsStore handles persistence of evaluation results
type ResultsStore struct {
	resultsDir string
}

// NewResultsStore creates a new ResultsStore with the specified directory
func NewResultsStore(resultsDir string) *ResultsStore {
	return &ResultsStore{
		resultsDir: resultsDir,
	}
}

// SaveEvalResult saves an eval result to the results directory
func (s *ResultsStore) SaveEvalResult(result *EvalResult) error {
	// Add timestamp if not present
	if result.Timestamp == "" {
		result.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	return s.appendToFile("eval-results.json", result, func() (interface{}, error) {
		return s.LoadEvalResults()
	})
}

// LoadEvalResults loads all eval results from the results directory
func (s *ResultsStore) LoadEvalResults() ([]EvalResult, error) {
	filePath := filepath.Join(s.resultsDir, "eval-results.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice if file doesn't exist
			return []EvalResult{}, nil
		}
		return nil, fmt.Errorf("reading eval results: %w", err)
	}

	var results []EvalResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("parsing eval results: %w", err)
	}

	return results, nil
}

// SaveMetaResult saves a meta result to the results directory
func (s *ResultsStore) SaveMetaResult(result *MetaResult) error {
	// Add timestamp if not present
	if result.Timestamp == "" {
		result.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	return s.appendToFile("meta-results.json", result, func() (interface{}, error) {
		return s.LoadMetaResults()
	})
}

// LoadMetaResults loads all meta results from the results directory
func (s *ResultsStore) LoadMetaResults() ([]MetaResult, error) {
	filePath := filepath.Join(s.resultsDir, "meta-results.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice if file doesn't exist
			return []MetaResult{}, nil
		}
		return nil, fmt.Errorf("reading meta results: %w", err)
	}

	var results []MetaResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("parsing meta results: %w", err)
	}

	return results, nil
}

// SaveTaskQualityResult saves a task quality result to the results directory
func (s *ResultsStore) SaveTaskQualityResult(result *TaskQualityResult) error {
	return s.appendToFile("task-quality-results.json", result, func() (interface{}, error) {
		return s.LoadTaskQualityResults()
	})
}

// LoadTaskQualityResults loads all task quality results from the results directory
func (s *ResultsStore) LoadTaskQualityResults() ([]TaskQualityResult, error) {
	filePath := filepath.Join(s.resultsDir, "task-quality-results.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice if file doesn't exist
			return []TaskQualityResult{}, nil
		}
		return nil, fmt.Errorf("reading task quality results: %w", err)
	}

	var results []TaskQualityResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("parsing task quality results: %w", err)
	}

	return results, nil
}

// appendToFile is a generic helper to append a result to a JSON array file
func (s *ResultsStore) appendToFile(filename string, newResult interface{}, loadFunc func() (interface{}, error)) error {
	// Ensure directory exists
	if err := os.MkdirAll(s.resultsDir, 0755); err != nil {
		return fmt.Errorf("creating results directory: %w", err)
	}

	// Load existing results
	existingResults, err := loadFunc()
	if err != nil {
		return err
	}

	// Append new result using reflection-like approach
	var allResults interface{}
	switch existing := existingResults.(type) {
	case []EvalResult:
		allResults = append(existing, *newResult.(*EvalResult))
	case []MetaResult:
		allResults = append(existing, *newResult.(*MetaResult))
	case []TaskQualityResult:
		allResults = append(existing, *newResult.(*TaskQualityResult))
	default:
		return fmt.Errorf("unsupported result type")
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(allResults, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling results: %w", err)
	}

	// Write to file
	filePath := filepath.Join(s.resultsDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("writing results file: %w", err)
	}

	return nil
}
