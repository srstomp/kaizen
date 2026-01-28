package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConsistencyResult represents a consistency evaluation result from meta-evals
type ConsistencyResult struct {
	Timestamp             string  `json:"timestamp"`
	Agent                 string  `json:"agent"`
	BoundaryType          string  `json:"boundary_type"`
	ConsistencyPercentage float64 `json:"consistency_percentage"`
	ConsistentCount       int     `json:"consistent_count"`
	TotalCount            int     `json:"total_count"`
}

// loadEvalResults loads eval results from task-eval-log.json
func loadEvalResults(logPath string) ([]GradeTaskOutput, error) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("reading eval log: %w", err)
	}

	var results []GradeTaskOutput
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("parsing eval log: %w", err)
	}

	return results, nil
}

// loadMetaResults loads meta-eval results from consistency-log.json
func loadMetaResults(logPath string) ([]ConsistencyResult, error) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("reading meta log: %w", err)
	}

	var results []ConsistencyResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("parsing meta log: %w", err)
	}

	return results, nil
}

// calculateEvalGateScore calculates the average overall score from eval results
func calculateEvalGateScore(results []GradeTaskOutput) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, result := range results {
		totalScore += result.OverallScore
	}

	return totalScore / float64(len(results))
}

// calculateMetaGateScore calculates the average consistency percentage from meta results
func calculateMetaGateScore(results []ConsistencyResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalConsistency := 0.0
	for _, result := range results {
		totalConsistency += result.ConsistencyPercentage
	}

	return totalConsistency / float64(len(results))
}

// checkGate checks if a score meets the threshold
func checkGate(score, threshold float64) bool {
	return score >= threshold
}

// formatGateResult formats a gate check result for display
func formatGateResult(checkType string, score, threshold float64, pass bool) string {
	status := "PASS"
	symbol := "✓"
	if !pass {
		status = "FAIL"
		symbol = "✗"
	}

	return fmt.Sprintf("[%s] %s gate: %.1f%% (threshold: %.1f%%) - %s",
		symbol, checkType, score, threshold, status)
}

// runGateCommand executes the gate check command
func runGateCommand(checkType string, threshold float64, reportsDir string) error {
	// Validate threshold
	if threshold < 0.0 || threshold > 100.0 {
		return fmt.Errorf("threshold must be between 0 and 100, got: %.1f", threshold)
	}

	// Validate check type
	validTypes := []string{"eval", "meta", "all"}
	isValid := false
	for _, validType := range validTypes {
		if checkType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid type: %s (must be 'eval', 'meta', or 'all')", checkType)
	}

	fmt.Printf("Release Gate Check\n")
	fmt.Printf("==================\n\n")
	fmt.Printf("Threshold: %.1f%%\n", threshold)
	fmt.Printf("Reports directory: %s\n\n", reportsDir)

	allPass := true
	results := []string{}

	// Check eval results if requested
	if checkType == "eval" || checkType == "all" {
		evalLogPath := filepath.Join(reportsDir, "task-eval-log.json")
		evalResults, err := loadEvalResults(evalLogPath)
		if err != nil {
			return fmt.Errorf("loading eval results: %w", err)
		}

		if len(evalResults) == 0 {
			return fmt.Errorf("no eval results found in %s", evalLogPath)
		}

		evalScore := calculateEvalGateScore(evalResults)
		evalPass := checkGate(evalScore, threshold)

		result := formatGateResult("eval", evalScore, threshold, evalPass)
		results = append(results, result)
		fmt.Println(result)

		if !evalPass {
			allPass = false
		}
	}

	// Check meta results if requested
	if checkType == "meta" || checkType == "all" {
		metaLogPath := filepath.Join(reportsDir, "consistency-log.json")
		metaResults, err := loadMetaResults(metaLogPath)
		if err != nil {
			return fmt.Errorf("loading meta results: %w", err)
		}

		if len(metaResults) == 0 {
			return fmt.Errorf("no meta results found in %s", metaLogPath)
		}

		metaScore := calculateMetaGateScore(metaResults)
		metaPass := checkGate(metaScore, threshold)

		result := formatGateResult("meta", metaScore, threshold, metaPass)
		results = append(results, result)
		fmt.Println(result)

		if !metaPass {
			allPass = false
		}
	}

	// Print overall result
	fmt.Printf("\n")
	if allPass {
		fmt.Printf("Overall: PASS - Release gate checks passed\n")
		return nil
	} else {
		fmt.Printf("Overall: FAIL - Release gate checks failed\n")
		return fmt.Errorf("release gate check failed: one or more checks did not meet threshold")
	}
}
