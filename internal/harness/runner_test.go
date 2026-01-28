package harness

import (
	"testing"
)

func TestNewRunner(t *testing.T) {
	runner := NewRunner()

	if runner == nil {
		t.Fatal("NewRunner() returned nil")
	}

	if runner.registry == nil {
		t.Error("NewRunner() registry is nil")
	}

	if runner.timeout == 0 {
		t.Error("NewRunner() timeout is 0")
	}
}

func TestRunner_RunEvaluation_EmptyCriteria(t *testing.T) {
	runner := NewRunner()

	failureCase := FailureCase{
		ID:           "TEST-001",
		Category:     "test-category",
		EvalCriteria: []EvalCriterion{},
	}

	runs, err := runner.RunEvaluation(failureCase, 3)
	if err != nil {
		t.Fatalf("RunEvaluation() error = %v, want nil", err)
	}

	if len(runs) != 3 {
		t.Errorf("RunEvaluation() returned %d runs, want 3", len(runs))
	}

	// Empty criteria should pass (all criteria passed vacuously)
	for i, passed := range runs {
		if !passed {
			t.Errorf("RunEvaluation() run %d = false, want true (empty criteria should pass)", i)
		}
	}
}

func TestRunner_RunEvaluation_SingleCodeBasedCriterion(t *testing.T) {
	runner := NewRunner()

	failureCase := FailureCase{
		ID:       "TEST-002",
		Category: "test-category",
		EvalCriteria: []EvalCriterion{
			{
				Type:  "code-based",
				Check: "file-exists(test.txt)",
			},
		},
	}

	// This will fail because file doesn't exist in isolated context
	runs, err := runner.RunEvaluation(failureCase, 2)
	if err != nil {
		t.Fatalf("RunEvaluation() error = %v, want nil", err)
	}

	if len(runs) != 2 {
		t.Errorf("RunEvaluation() returned %d runs, want 2", len(runs))
	}

	// Should fail because no file exists in isolated context
	for i, passed := range runs {
		if passed {
			t.Errorf("RunEvaluation() run %d = true, want false (file doesn't exist)", i)
		}
	}
}

func TestRunner_RunEvaluation_SingleModelBasedCriterion(t *testing.T) {
	runner := NewRunner()

	failureCase := FailureCase{
		ID:       "TEST-003",
		Category: "test-category",
		Evidence: FailureEvidence{
			TaskSpec:     "Create a user authentication API",
			WhatWasBuilt: "Implemented login endpoint",
		},
		EvalCriteria: []EvalCriterion{
			{
				Type:  "model-based",
				Check: "spec_compliance",
			},
		},
	}

	runs, err := runner.RunEvaluation(failureCase, 1)
	if err != nil {
		t.Fatalf("RunEvaluation() error = %v, want nil", err)
	}

	if len(runs) != 1 {
		t.Errorf("RunEvaluation() returned %d runs, want 1", len(runs))
	}

	// Model-based grader uses stub evaluation which returns pass
	if !runs[0] {
		t.Errorf("RunEvaluation() run 0 = false, want true (stub evaluation should pass)")
	}
}

func TestRunner_RunEvaluation_MultipleCriteria(t *testing.T) {
	runner := NewRunner()

	failureCase := FailureCase{
		ID:       "TEST-004",
		Category: "test-category",
		Evidence: FailureEvidence{
			TaskSpec:     "Create a user authentication API",
			WhatWasBuilt: "Implemented login endpoint",
		},
		EvalCriteria: []EvalCriterion{
			{
				Type:  "model-based",
				Check: "spec_compliance",
			},
			{
				Type:  "code-based",
				Check: "file-exists(nonexistent.txt)",
			},
		},
	}

	runs, err := runner.RunEvaluation(failureCase, 1)
	if err != nil {
		t.Fatalf("RunEvaluation() error = %v, want nil", err)
	}

	if len(runs) != 1 {
		t.Errorf("RunEvaluation() returned %d runs, want 1", len(runs))
	}

	// Should fail because one criterion (file-exists) fails
	// All criteria must pass for the run to pass
	if runs[0] {
		t.Errorf("RunEvaluation() run 0 = true, want false (one criterion failed)")
	}
}

func TestRunner_RunEvaluation_UnknownGrader(t *testing.T) {
	runner := NewRunner()

	failureCase := FailureCase{
		ID:       "TEST-005",
		Category: "test-category",
		EvalCriteria: []EvalCriterion{
			{
				Type:  "code-based",
				Check: "unknown-grader(test)",
			},
		},
	}

	runs, err := runner.RunEvaluation(failureCase, 1)
	if err != nil {
		t.Fatalf("RunEvaluation() error = %v, want nil", err)
	}

	// Unknown grader should cause criterion to fail
	if runs[0] {
		t.Errorf("RunEvaluation() run 0 = true, want false (unknown grader should fail)")
	}
}
