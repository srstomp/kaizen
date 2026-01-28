package harness

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/srstomp/kaizen/internal/graders/codebased"
	"github.com/srstomp/kaizen/internal/graders/modelbased"
)

// FailureCase represents a documented agent failure case
type FailureCase struct {
	ID           string
	Category     string
	Evidence     FailureEvidence
	EvalCriteria []EvalCriterion
}

// FailureEvidence contains the evidence of the failure
type FailureEvidence struct {
	TaskSpec     string
	WhatWasBuilt string
}

// EvalCriterion represents a single evaluation check
type EvalCriterion struct {
	Type  string
	Check string
}

// Runner orchestrates the execution of evaluation criteria against failure cases
type Runner struct {
	registry *GraderRegistry
	timeout  time.Duration
}

// NewRunner creates a new evaluation runner with default configuration
func NewRunner() *Runner {
	return &Runner{
		registry: NewGraderRegistry(),
		timeout:  60 * time.Second,
	}
}

// RunEvaluation runs evaluation on a failure case k times
// Each run executes all criteria in an isolated context
// Returns a slice of pass/fail results, one per run
func (r *Runner) RunEvaluation(failureCase FailureCase, k int) ([]bool, error) {
	results := make([]bool, k)

	// Run evaluation k times
	for i := 0; i < k; i++ {
		// Create isolated context
		ctx, err := NewIsolatedContext()
		if err != nil {
			return results, fmt.Errorf("creating isolated context for run %d: %w", i+1, err)
		}
		defer ctx.Cleanup()

		// Execute all criteria
		runPassed := true
		for _, criterion := range failureCase.EvalCriteria {
			passed, err := r.executeCriterion(criterion, failureCase, ctx.WorkingDir())
			if err != nil {
				log.Printf("Warning: criterion execution failed: %v", err)
				passed = false
			}
			if !passed {
				runPassed = false
			}
		}

		results[i] = runPassed
	}

	return results, nil
}

// executeCriterion executes a single evaluation criterion
// Returns true if the criterion passes, false otherwise
func (r *Runner) executeCriterion(criterion EvalCriterion, failureCase FailureCase, workDir string) (bool, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Execute based on criterion type
	switch criterion.Type {
	case "code-based":
		return r.executeCodeBasedCriterion(ctx, criterion.Check, failureCase, workDir)
	case "model-based":
		return r.executeModelBasedCriterion(ctx, criterion.Check, failureCase)
	default:
		return false, fmt.Errorf("unknown criterion type: %s", criterion.Type)
	}
}

// executeCodeBasedCriterion executes a code-based evaluation criterion
func (r *Runner) executeCodeBasedCriterion(ctx context.Context, check string, failureCase FailureCase, workDir string) (bool, error) {
	// Extract grader name from check string (e.g., "file-exists(test.txt)" -> "file-exists")
	graderName := extractGraderName(check)

	// Get grader from registry
	grader := r.registry.GetCodeGrader(graderName)
	if grader == nil {
		return false, fmt.Errorf("code-based grader not found: %s", graderName)
	}

	// Build GradeInput
	input := codebased.GradeInput{
		TaskID:       failureCase.ID,
		TaskType:     "",         // Not available in failure case for MVP
		ChangedFiles: []string{}, // Empty for MVP as noted in requirements
		WorkDir:      workDir,
	}

	// Execute grader
	// Note: Using a channel to make the grader call respect context timeout
	resultChan := make(chan codebased.GradeResult, 1)
	go func() {
		result := grader.Grade(input)
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result.Passed, nil
	case <-ctx.Done():
		return false, fmt.Errorf("criterion execution timed out")
	}
}

// executeModelBasedCriterion executes a model-based evaluation criterion
func (r *Runner) executeModelBasedCriterion(ctx context.Context, check string, failureCase FailureCase) (bool, error) {
	// Extract grader name from check string
	graderName := extractGraderName(check)

	// Get grader from registry
	grader := r.registry.GetModelGrader(graderName)
	if grader == nil {
		return false, fmt.Errorf("model-based grader not found: %s", graderName)
	}

	// Build GradeInput
	input := modelbased.GradeInput{
		Content: failureCase.Evidence.WhatWasBuilt,
		Context: map[string]any{
			"spec": failureCase.Evidence.TaskSpec,
		},
	}

	// Execute grader
	// Note: Using a channel to make the grader call respect context timeout
	resultChan := make(chan modelbased.Result, 1)
	errChan := make(chan error, 1)
	go func() {
		result, err := grader.Grade(input)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result.Passed, nil
	case err := <-errChan:
		return false, fmt.Errorf("grader execution failed: %w", err)
	case <-ctx.Done():
		return false, fmt.Errorf("criterion execution timed out")
	}
}

// extractGraderName extracts the grader name from a check string
// Examples:
//   - "file-exists(test.txt)" -> "file-exists"
//   - "spec_compliance" -> "spec_compliance"
func extractGraderName(check string) string {
	// Simple extraction: take everything before the first '(' or the whole string
	if idx := strings.Index(check, "("); idx != -1 {
		return strings.TrimSpace(check[:idx])
	}
	return strings.TrimSpace(check)
}
