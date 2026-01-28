package modelbased

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/srstomp/kaizen/internal/llm"
)

// SpecComplianceGrader evaluates if an implementation matches its specification
type SpecComplianceGrader struct {
	llmClient llm.Client
	timeout   time.Duration
}

// NewSpecComplianceGrader creates a new spec compliance grader
// Note: This is a stub implementation that doesn't use LLM yet
func NewSpecComplianceGrader() *SpecComplianceGrader {
	return &SpecComplianceGrader{
		llmClient: nil, // Will be set when LLM integration is needed
		timeout:   60 * time.Second,
	}
}

// Grade evaluates if the implementation matches the specification
func (g *SpecComplianceGrader) Grade(input GradeInput) (Result, error) {
	// Validate inputs
	if err := g.validateInput(input); err != nil {
		return Result{}, err
	}

	// Extract spec from context
	spec := input.Context["spec"].(string)

	// For now, use stub evaluation since LLM integration is not ready
	return g.stubEvaluate(spec, input.Content)
}

// validateInput checks that the input has required fields
func (g *SpecComplianceGrader) validateInput(input GradeInput) error {
	// Check for empty content
	if strings.TrimSpace(input.Content) == "" {
		return errors.New("implementation diff cannot be empty")
	}

	// Check for nil context
	if input.Context == nil {
		return errors.New("context cannot be nil")
	}

	// Check for spec in context
	specValue, exists := input.Context["spec"]
	if !exists {
		return errors.New("spec not found in context")
	}

	// Check that spec is a string
	spec, ok := specValue.(string)
	if !ok {
		return errors.New("spec must be a string")
	}

	// Check for empty spec
	if strings.TrimSpace(spec) == "" {
		return errors.New("spec cannot be empty")
	}

	return nil
}

// stubEvaluate performs a basic heuristic evaluation
// TODO: Replace with LLM-based evaluation
func (g *SpecComplianceGrader) stubEvaluate(spec, diff string) (Result, error) {
	// Simple heuristic: if diff is non-empty and spec is non-empty, assume pass
	// This is a placeholder until LLM integration is complete

	score := 75.0
	passed := true
	reasoning := "Stub evaluation: Implementation appears to have content matching spec requirements"
	verdict := "PASS"

	// Basic heuristic: longer diff with more content likely more complete
	if len(diff) < 50 {
		score = 60.0
		reasoning = "Stub evaluation: Implementation seems minimal, may not fully address spec"
	} else if len(diff) > 200 {
		score = 85.0
		reasoning = "Stub evaluation: Implementation has substantial content"
	}

	message := fmt.Sprintf("Spec compliance evaluation: %s (score: %.1f/100). Note: Using stub evaluation; LLM-based grading not yet implemented.", verdict, score)

	return Result{
		Passed:  passed,
		Score:   score,
		Message: message,
		Details: map[string]any{
			"verdict":   verdict,
			"reasoning": reasoning,
			"note":      "Stub evaluation - LLM integration pending",
		},
	}, nil
}

// buildPrompt constructs the LLM prompt for spec compliance evaluation
func (g *SpecComplianceGrader) buildPrompt(spec, diff string) string {
	return fmt.Sprintf(`You are evaluating if an implementation matches its specification.

## Specification
%s

## Implementation Diff
%s

Evaluate if the implementation correctly fulfills the specification.
Respond with:
VERDICT: PASS or FAIL
SCORE: 0-100
REASONING: Brief explanation
ISSUES: List any issues found (if FAIL)`, spec, diff)
}

// parseResponse parses the LLM response to extract verdict, score, and reasoning
func (g *SpecComplianceGrader) parseResponse(response string) (Result, error) {
	if strings.TrimSpace(response) == "" {
		return Result{}, errors.New("LLM returned empty response")
	}

	lines := strings.Split(response, "\n")

	var verdict string
	var scoreStr string
	var reasoning string
	var issues string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "VERDICT:") {
			verdict = strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
		} else if strings.HasPrefix(line, "SCORE:") {
			scoreStr = strings.TrimSpace(strings.TrimPrefix(line, "SCORE:"))
		} else if strings.HasPrefix(line, "REASONING:") {
			reasoning = strings.TrimSpace(strings.TrimPrefix(line, "REASONING:"))
		} else if strings.HasPrefix(line, "ISSUES:") {
			issues = strings.TrimSpace(strings.TrimPrefix(line, "ISSUES:"))
		}
	}

	// Validate required fields
	if verdict == "" {
		return Result{}, errors.New("missing VERDICT in LLM response")
	}
	if scoreStr == "" {
		return Result{}, errors.New("missing SCORE in LLM response")
	}
	if reasoning == "" {
		return Result{}, errors.New("missing REASONING in LLM response")
	}

	// Validate verdict
	verdict = strings.ToUpper(verdict)
	if verdict != "PASS" && verdict != "FAIL" {
		return Result{}, fmt.Errorf("invalid VERDICT: %s (expected PASS or FAIL)", verdict)
	}

	// Parse score
	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return Result{}, fmt.Errorf("invalid SCORE: %s (must be a number)", scoreStr)
	}

	// Validate score range
	if score < 0 || score > 100 {
		return Result{}, fmt.Errorf("invalid SCORE: %.1f (must be between 0 and 100)", score)
	}

	passed := verdict == "PASS"
	message := fmt.Sprintf("Spec compliance evaluation: %s (score: %.1f/100). %s", verdict, score, reasoning)

	details := map[string]any{
		"verdict":   verdict,
		"reasoning": reasoning,
	}
	if issues != "" {
		details["issues"] = issues
	}

	return Result{
		Passed:  passed,
		Score:   score,
		Message: message,
		Details: details,
	}, nil
}

// gradeWithLLM performs LLM-based evaluation
// TODO: Implement this when LLM client is integrated
func (g *SpecComplianceGrader) gradeWithLLM(spec, diff string) (Result, error) {
	if g.llmClient == nil {
		return Result{}, errors.New("LLM client not initialized")
	}

	// Build prompt
	prompt := g.buildPrompt(spec, diff)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Call LLM
	response, err := g.llmClient.Complete(ctx, prompt, llm.WithModel("claude-haiku-4"))
	if err != nil {
		// Handle timeout specifically
		if errors.Is(err, context.DeadlineExceeded) {
			return Result{}, fmt.Errorf("LLM request timed out after %v: %w", g.timeout, err)
		}
		return Result{}, fmt.Errorf("LLM request failed: %w", err)
	}

	// Parse response
	return g.parseResponse(response)
}
