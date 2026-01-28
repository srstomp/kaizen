package modelbased

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/srstomp/kaizen/internal/llm"
)

func TestSpecComplianceGrader_New(t *testing.T) {
	grader := NewSpecComplianceGrader()
	if grader == nil {
		t.Fatal("Expected NewSpecComplianceGrader to return non-nil grader")
	}
}

func TestSpecComplianceGrader_Grade(t *testing.T) {
	grader := NewSpecComplianceGrader()

	tests := []struct {
		name        string
		input       GradeInput
		expectError bool
	}{
		{
			name: "implementation matches spec - simple case",
			input: GradeInput{
				Content: `+func Add(a, b int) int {
+    return a + b
+}`,
				Context: map[string]any{
					"spec": "Create an Add function that takes two integers and returns their sum",
				},
			},
			expectError: false,
		},
		{
			name: "implementation with context but no spec",
			input: GradeInput{
				Content: `+func DoSomething() {}`,
				Context: map[string]any{
					"other_key": "value",
				},
			},
			expectError: true, // Should error when spec is missing
		},
		{
			name: "empty implementation diff",
			input: GradeInput{
				Content: "",
				Context: map[string]any{
					"spec": "Create a new function",
				},
			},
			expectError: true, // Should error with empty diff
		},
		{
			name: "empty spec",
			input: GradeInput{
				Content: `+func Test() {}`,
				Context: map[string]any{
					"spec": "",
				},
			},
			expectError: true, // Should error with empty spec
		},
		{
			name: "nil context",
			input: GradeInput{
				Content: `+func Test() {}`,
				Context: nil,
			},
			expectError: true, // Should error with nil context
		},
		{
			name: "spec not a string",
			input: GradeInput{
				Content: `+func Test() {}`,
				Context: map[string]any{
					"spec": 123, // Not a string
				},
			},
			expectError: true, // Should error when spec is not a string
		},
		{
			name: "valid complex implementation",
			input: GradeInput{
				Content: `+func Multiply(x, y int) int {
+    result := 0
+    for i := 0; i < y; i++ {
+        result += x
+    }
+    return result
+}
+
+func TestMultiply(t *testing.T) {
+    if Multiply(3, 4) != 12 {
+        t.Error("multiplication failed")
+    }
+}`,
				Context: map[string]any{
					"spec": `Create a Multiply function that takes two integers and returns their product.
Include a test that verifies the function works correctly.`,
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := grader.Grade(tt.input)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if err == nil {
				// Verify result structure
				if result.Score < 0 || result.Score > 100 {
					t.Errorf("Expected score between 0 and 100, got %f", result.Score)
				}
				if result.Message == "" {
					t.Error("Expected non-empty message")
				}
				if result.Details == nil {
					t.Error("Expected non-nil Details map")
				}
				// Verify details contains expected keys
				if _, ok := result.Details["verdict"]; !ok {
					t.Error("Expected Details to contain 'verdict' key")
				}
				if _, ok := result.Details["reasoning"]; !ok {
					t.Error("Expected Details to contain 'reasoning' key")
				}
			}
		})
	}
}

func TestSpecComplianceGrader_ParseResponse(t *testing.T) {
	grader := NewSpecComplianceGrader()

	tests := []struct {
		name            string
		response        string
		expectPass      bool
		expectScore     float64
		expectReasoning string
		expectError     bool
	}{
		{
			name: "valid PASS response",
			response: `VERDICT: PASS
SCORE: 95
REASONING: Implementation correctly adds two integers as specified`,
			expectPass:      true,
			expectScore:     95.0,
			expectReasoning: "Implementation correctly adds two integers as specified",
			expectError:     false,
		},
		{
			name: "valid FAIL response",
			response: `VERDICT: FAIL
SCORE: 30
REASONING: Function signature does not match specification
ISSUES: Missing parameter validation, incorrect return type`,
			expectPass:      false,
			expectScore:     30.0,
			expectReasoning: "Function signature does not match specification",
			expectError:     false,
		},
		{
			name: "response with extra whitespace",
			response: `  VERDICT:  PASS
  SCORE:  85
  REASONING:  Good implementation with minor style issues  `,
			expectPass:      true,
			expectScore:     85.0,
			expectReasoning: "Good implementation with minor style issues",
			expectError:     false,
		},
		{
			name: "missing verdict",
			response: `SCORE: 50
REASONING: Some reason`,
			expectError: true,
		},
		{
			name: "missing score",
			response: `VERDICT: PASS
REASONING: Some reason`,
			expectError: true,
		},
		{
			name: "missing reasoning",
			response: `VERDICT: PASS
SCORE: 50`,
			expectError: true,
		},
		{
			name: "invalid verdict",
			response: `VERDICT: MAYBE
SCORE: 50
REASONING: Unclear`,
			expectError: true,
		},
		{
			name: "invalid score - not a number",
			response: `VERDICT: PASS
SCORE: abc
REASONING: Good`,
			expectError: true,
		},
		{
			name: "invalid score - out of range",
			response: `VERDICT: PASS
SCORE: 150
REASONING: Good`,
			expectError: true,
		},
		{
			name:        "empty response",
			response:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := grader.parseResponse(tt.response)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if err == nil {
				if result.Passed != tt.expectPass {
					t.Errorf("Expected Passed=%v, got %v", tt.expectPass, result.Passed)
				}
				if result.Score != tt.expectScore {
					t.Errorf("Expected Score=%f, got %f", tt.expectScore, result.Score)
				}
				if result.Details["reasoning"] != tt.expectReasoning {
					t.Errorf("Expected reasoning='%s', got '%v'", tt.expectReasoning, result.Details["reasoning"])
				}
			}
		})
	}
}

func TestSpecComplianceGrader_BuildPrompt(t *testing.T) {
	grader := NewSpecComplianceGrader()

	spec := "Create an Add function"
	diff := "+func Add(a, b int) int { return a + b }"

	prompt := grader.buildPrompt(spec, diff)

	// Verify prompt contains key sections
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// Check for specification section
	if !contains(prompt, "Specification") && !contains(prompt, "spec") {
		t.Error("Expected prompt to mention specification")
	}

	// Check for diff section
	if !contains(prompt, "Implementation") && !contains(prompt, "diff") {
		t.Error("Expected prompt to mention implementation or diff")
	}

	// Check for evaluation instruction
	if !contains(prompt, "VERDICT") {
		t.Error("Expected prompt to request VERDICT")
	}
	if !contains(prompt, "SCORE") {
		t.Error("Expected prompt to request SCORE")
	}
	if !contains(prompt, "REASONING") {
		t.Error("Expected prompt to request REASONING")
	}

	// Verify spec and diff are included in prompt
	if !contains(prompt, spec) {
		t.Error("Expected prompt to contain the specification")
	}
	if !contains(prompt, diff) {
		t.Error("Expected prompt to contain the diff")
	}
}

func TestSpecComplianceGrader_GradeWithLLM_NilClient(t *testing.T) {
	grader := NewSpecComplianceGrader()
	// grader.llmClient is nil by default

	spec := "Create an Add function"
	diff := "+func Add(a, b int) int { return a + b }"

	result, err := grader.gradeWithLLM(spec, diff)

	if err == nil {
		t.Error("Expected error when LLM client is nil, got none")
	}

	expectedErrMsg := "LLM client not initialized"
	if err != nil && err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}

	// Verify result is empty on error
	if result.Score != 0 {
		t.Errorf("Expected zero score on error, got %f", result.Score)
	}
	if result.Message != "" {
		t.Errorf("Expected empty message on error, got '%s'", result.Message)
	}
}

func TestSpecComplianceGrader_GradeWithLLM_ContextTimeout(t *testing.T) {
	// Create a mock LLM client that simulates a timeout
	mockClient := &mockLLMClient{
		simulateTimeout: true,
	}

	grader := &SpecComplianceGrader{
		llmClient: mockClient,
		timeout:   50 * time.Millisecond, // Very short timeout
	}

	spec := "Create an Add function"
	diff := "+func Add(a, b int) int { return a + b }"

	result, err := grader.gradeWithLLM(spec, diff)

	if err == nil {
		t.Error("Expected timeout error, got none")
	}

	// Verify error message mentions timeout
	if err != nil && !contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error message, got: %s", err.Error())
	}

	// Verify result is empty on error
	if result.Score != 0 {
		t.Errorf("Expected zero score on error, got %f", result.Score)
	}
}

func TestSpecComplianceGrader_GradeWithLLM_Success(t *testing.T) {
	// Create a mock LLM client that returns a valid response
	mockClient := &mockLLMClient{
		response: `VERDICT: PASS
SCORE: 95
REASONING: Implementation correctly adds two integers as specified`,
	}

	grader := &SpecComplianceGrader{
		llmClient: mockClient,
		timeout:   30 * time.Second,
	}

	spec := "Create an Add function"
	diff := "+func Add(a, b int) int { return a + b }"

	result, err := grader.gradeWithLLM(spec, diff)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify result fields
	if !result.Passed {
		t.Error("Expected Passed to be true")
	}
	if result.Score != 95.0 {
		t.Errorf("Expected score 95.0, got %f", result.Score)
	}
	if result.Details["verdict"] != "PASS" {
		t.Errorf("Expected verdict PASS, got %v", result.Details["verdict"])
	}
	if result.Details["reasoning"] != "Implementation correctly adds two integers as specified" {
		t.Errorf("Expected specific reasoning, got %v", result.Details["reasoning"])
	}
}

func TestSpecComplianceGrader_GradeWithLLM_LLMError(t *testing.T) {
	// Create a mock LLM client that returns an error
	mockClient := &mockLLMClient{
		simulateError: true,
		errorMsg:      "API rate limit exceeded",
	}

	grader := &SpecComplianceGrader{
		llmClient: mockClient,
		timeout:   30 * time.Second,
	}

	spec := "Create an Add function"
	diff := "+func Add(a, b int) int { return a + b }"

	result, err := grader.gradeWithLLM(spec, diff)

	if err == nil {
		t.Error("Expected LLM error, got none")
	}

	// Verify error message contains "LLM request failed"
	if err != nil && !contains(err.Error(), "LLM request failed") {
		t.Errorf("Expected 'LLM request failed' in error, got: %s", err.Error())
	}

	// Verify result is empty on error
	if result.Score != 0 {
		t.Errorf("Expected zero score on error, got %f", result.Score)
	}
}

// mockLLMClient is a mock implementation of llm.Client for testing
type mockLLMClient struct {
	response        string
	simulateTimeout bool
	simulateError   bool
	errorMsg        string
}

func (m *mockLLMClient) Complete(ctx context.Context, prompt string, options ...llm.CompletionOption) (string, error) {
	if m.simulateTimeout {
		// Simulate a timeout by waiting longer than the context timeout
		time.Sleep(100 * time.Millisecond)
		return "", context.DeadlineExceeded
	}

	if m.simulateError {
		return "", errors.New(m.errorMsg)
	}

	return m.response, nil
}

// Helper function to check if a string contains a substring (case-sensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
