package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/srstomp/kaizen/internal/graders/codebased"
	"github.com/srstomp/kaizen/internal/graders/modelbased"
	"github.com/srstomp/kaizen/internal/harness"
)

// GradeOutput represents the JSON output from the grade command
type GradeOutput struct {
	Grader  string  `json:"grader"`
	Passed  bool    `json:"passed"`
	Score   float64 `json:"score"`
	Message string  `json:"message"`
}

// runGradeCommand executes a single grader on a single input
func runGradeCommand(grader, inputPath, spec, format string) error {
	// Support both hyphen and underscore variants
	normalizedGraderUnderscore := strings.ReplaceAll(grader, "-", "_")
	normalizedGraderHyphen := strings.ReplaceAll(grader, "_", "-")

	// Read input JSON file
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create grader registry
	registry := harness.NewGraderRegistry()

	// Try to find code grader (check both underscore and hyphen variants)
	codeGrader := registry.GetCodeGrader(grader)
	if codeGrader == nil {
		codeGrader = registry.GetCodeGrader(normalizedGraderHyphen)
	}
	if codeGrader == nil {
		codeGrader = registry.GetCodeGrader(normalizedGraderUnderscore)
	}

	// Try to find model grader (uses underscores)
	modelGrader := registry.GetModelGrader(normalizedGraderUnderscore)

	// Check if grader exists
	if codeGrader == nil && modelGrader == nil {
		return fmt.Errorf("unknown grader: %s", grader)
	}

	// Execute grader based on type
	if codeGrader != nil {
		return runCodeBasedGrader(codeGrader, inputData, format)
	}

	return runModelBasedGrader(modelGrader, inputData, spec, format, normalizedGraderUnderscore)
}

// runCodeBasedGrader executes a code-based grader
func runCodeBasedGrader(grader codebased.CodeGrader, inputData []byte, format string) error {
	// Parse code-based input
	var input codebased.GradeInput
	if err := json.Unmarshal(inputData, &input); err != nil {
		return fmt.Errorf("failed to parse input JSON for code-based grader: %w", err)
	}

	// Run grader
	result := grader.Grade(input)

	// Format output
	if format == "json" {
		output := GradeOutput{
			Grader:  grader.Name(),
			Passed:  result.Passed,
			Score:   result.Score,
			Message: result.Details,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		// Text format
		fmt.Printf("Grader: %s\n", grader.Name())

		status := "PASS"
		if result.Skipped {
			status = "SKIP"
		} else if !result.Passed {
			status = "FAIL"
		}
		fmt.Printf("Result: %s\n", status)

		fmt.Printf("Score: %.1f\n", result.Score)

		if result.Skipped {
			fmt.Printf("Skip Reason: %s\n", result.SkipReason)
		} else {
			fmt.Printf("Message: %s\n", result.Details)
		}
	}

	return nil
}

// runModelBasedGrader executes a model-based grader
func runModelBasedGrader(grader modelbased.Grader, inputData []byte, spec, format, graderName string) error {
	// Parse model-based input
	var input modelbased.GradeInput
	if err := json.Unmarshal(inputData, &input); err != nil {
		return fmt.Errorf("failed to parse input JSON for model-based grader: %w", err)
	}

	// Add spec to context if provided
	if spec != "" {
		if input.Context == nil {
			input.Context = make(map[string]any)
		}
		input.Context["spec"] = spec
	}

	// Run grader
	result, err := grader.Grade(input)
	if err != nil {
		return fmt.Errorf("grader failed: %w", err)
	}

	// Format output
	if format == "json" {
		output := GradeOutput{
			Grader:  graderName,
			Passed:  result.Passed,
			Score:   result.Score,
			Message: result.Message,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		// Text format
		fmt.Printf("Grader: %s\n", graderName)

		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		fmt.Printf("Result: %s\n", status)

		fmt.Printf("Score: %.1f\n", result.Score)
		fmt.Printf("Message: %s\n", result.Message)
	}

	return nil
}
