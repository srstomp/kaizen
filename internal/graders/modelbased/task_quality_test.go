package modelbased

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/srstomp/kaizen/internal/llm"
)

func TestTaskQualityGrader_New(t *testing.T) {
	grader := NewTaskQualityGrader()
	if grader == nil {
		t.Fatal("Expected NewTaskQualityGrader to return non-nil grader")
	}
}

func TestTaskQualityGrader_Grade(t *testing.T) {
	grader := NewTaskQualityGrader()

	tests := []struct {
		name        string
		input       GradeInput
		expectError bool
	}{
		{
			name: "high quality task with clear acceptance criteria",
			input: GradeInput{
				Content: `# Implement User Authentication

## Description
Create a user authentication system using JWT tokens for the API.

## Acceptance Criteria
1. Users can register with email and password
2. Users can login and receive a JWT token
3. JWT tokens expire after 24 hours
4. Protected endpoints validate JWT tokens
5. Invalid tokens return 401 Unauthorized

## Technical Details
- Use bcrypt for password hashing
- Use RS256 algorithm for JWT signing
- Store tokens in Redis with TTL

## Out of Scope
- OAuth integration
- Social login
- Password reset functionality
`,
				Context: map[string]any{
					"task_title": "Implement User Authentication",
					"task_type":  "feature",
				},
			},
			expectError: false,
		},
		{
			name: "vague task with unclear scope",
			input: GradeInput{
				Content: `Investigate the authentication issues and explore possible solutions.`,
				Context: map[string]any{
					"task_title": "Investigate Auth",
					"task_type":  "task",
				},
			},
			expectError: false, // Should handle gracefully, but fail the grading
		},
		{
			name: "task missing acceptance criteria",
			input: GradeInput{
				Content: `Build a new dashboard. Make it look good and work well.`,
				Context: map[string]any{
					"task_title": "Build Dashboard",
					"task_type":  "feature",
				},
			},
			expectError: false, // Should handle gracefully, but fail the grading
		},
		{
			name: "empty content",
			input: GradeInput{
				Content: "",
				Context: map[string]any{},
			},
			expectError: false, // Should handle gracefully with low score
		},
		{
			name: "task with ambiguous scope",
			input: GradeInput{
				Content: `Improve the performance of the system by looking at various areas and making it faster.`,
				Context: map[string]any{
					"task_title": "Performance Improvements",
					"task_type":  "task",
				},
			},
			expectError: false, // Should handle gracefully, but fail the grading
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
			}
		})
	}
}

func TestTaskQualityGrader_CriteriaWeights(t *testing.T) {
	// Verify the grader has the expected criteria with correct weights
	expectedCriteria := map[string]float64{
		"clarity":        0.30, // 30% - Is the task description clear?
		"acceptance":     0.30, // 30% - Are acceptance criteria well-defined?
		"scope":          0.25, // 25% - Is scope unambiguous?
		"actionability":  0.15, // 15% - Can work begin immediately?
	}

	// Test that weights sum to 1.0
	totalWeight := 0.0
	for _, weight := range expectedCriteria {
		totalWeight += weight
	}

	if totalWeight != 1.0 {
		t.Errorf("Expected weights to sum to 1.0, got %f", totalWeight)
	}
}

func TestTaskQualityGrader_DetailedFeedback(t *testing.T) {
	grader := NewTaskQualityGrader()

	input := GradeInput{
		Content: `# Create API Endpoint

## Description
Create a REST API endpoint to fetch user profiles.

## Acceptance Criteria
1. GET /api/users/:id returns user profile JSON
2. Returns 404 if user not found
3. Returns 200 with user data if found

## Technical Details
- Use existing User model
- Add proper error handling
`,
		Context: map[string]any{
			"task_title": "Create API Endpoint",
			"task_type":  "feature",
		},
	}

	result, err := grader.Grade(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify detailed feedback is provided for each criterion
	expectedKeys := []string{
		"clarity",
		"acceptance",
		"scope",
		"actionability",
	}

	for _, key := range expectedKeys {
		if _, exists := result.Details[key]; !exists {
			t.Errorf("Expected Details to contain key '%s'", key)
		}
	}
}

func TestTaskQualityGrader_PassingThreshold(t *testing.T) {
	grader := NewTaskQualityGrader()

	// Test with a well-defined task that should pass
	goodTask := GradeInput{
		Content: `# Add User Profile Validation

## Description
Add input validation for user profile updates to prevent invalid data.

## Acceptance Criteria
1. Email must be valid format
2. Age must be between 13 and 120
3. Username must be alphanumeric, 3-20 characters
4. Returns 400 Bad Request with validation errors
5. Returns 200 OK with updated profile on success

## Implementation Notes
- Use existing validation library
- Add unit tests for each validation rule
`,
		Context: map[string]any{
			"task_title": "Add User Profile Validation",
			"task_type":  "feature",
		},
	}

	result, err := grader.Grade(goodTask)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Well-defined task should pass (score >= 70)
	if !result.Passed {
		t.Errorf("Expected well-defined task to pass, got score: %f", result.Score)
	}
	if result.Score < 70.0 {
		t.Errorf("Expected passing score >= 70.0, got %f", result.Score)
	}
}

func TestTaskQualityGrader_VagueTaskFails(t *testing.T) {
	grader := NewTaskQualityGrader()

	// Test with vague task containing problematic keywords
	vagueTask := GradeInput{
		Content: `Investigate the login system and explore ways to make it better.`,
		Context: map[string]any{
			"task_title": "Investigate Login",
			"task_type":  "task",
		},
	}

	result, err := grader.Grade(vagueTask)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Vague task should fail
	if result.Passed {
		t.Errorf("Expected vague task to fail, got score: %f", result.Score)
	}
	if result.Score >= 70.0 {
		t.Errorf("Expected failing score < 70.0, got %f", result.Score)
	}
}

func TestTaskQualityGrader_EmptyContentFails(t *testing.T) {
	grader := NewTaskQualityGrader()

	emptyTask := GradeInput{
		Content: "",
		Context: map[string]any{},
	}

	result, err := grader.Grade(emptyTask)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Empty task should fail with very low score
	if result.Passed {
		t.Errorf("Expected empty task to fail, got score: %f", result.Score)
	}
	if result.Score > 10.0 {
		t.Errorf("Expected very low score for empty task, got %f", result.Score)
	}
}

// TestTaskQualityGrader_BuildPrompt tests prompt construction for LLM evaluation
func TestTaskQualityGrader_BuildPrompt(t *testing.T) {
	grader := NewTaskQualityGrader()

	taskContent := `# Implement User Authentication

## Description
Create a user authentication system using JWT tokens.

## Acceptance Criteria
1. Users can register with email and password
2. Users can login and receive a JWT token
3. JWT tokens expire after 24 hours`

	prompt := grader.buildPrompt(taskContent)

	// Verify prompt is not empty
	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// Verify prompt contains task content
	if !strings.Contains(prompt, "User Authentication") {
		t.Error("Expected prompt to contain task content")
	}

	// Verify prompt mentions evaluation criteria (case-insensitive)
	promptLower := strings.ToLower(prompt)
	if !strings.Contains(promptLower, "clarity") {
		t.Error("Expected prompt to mention clarity criterion")
	}
	if !strings.Contains(promptLower, "acceptance") {
		t.Error("Expected prompt to mention acceptance criteria evaluation")
	}

	// Verify prompt requests structured output
	if !strings.Contains(prompt, "CLARITY:") {
		t.Error("Expected prompt to request CLARITY score")
	}
	if !strings.Contains(prompt, "ACCEPTANCE:") {
		t.Error("Expected prompt to request ACCEPTANCE score")
	}
	if !strings.Contains(prompt, "SCOPE:") {
		t.Error("Expected prompt to request SCOPE score")
	}
	if !strings.Contains(prompt, "ACTIONABILITY:") {
		t.Error("Expected prompt to request ACTIONABILITY score")
	}
}

// TestTaskQualityGrader_ParseResponse tests parsing of LLM responses
func TestTaskQualityGrader_ParseResponse(t *testing.T) {
	grader := NewTaskQualityGrader()

	tests := []struct {
		name            string
		response        string
		expectError     bool
		expectPass      bool
		expectMinScore  float64
		expectMaxScore  float64
	}{
		{
			name: "valid high-quality response",
			response: `CLARITY: 85
CLARITY_FEEDBACK: Task description is clear and well-structured
ACCEPTANCE: 90
ACCEPTANCE_FEEDBACK: Acceptance criteria are well-defined and testable
SCOPE: 80
SCOPE_FEEDBACK: Scope is bounded with clear out-of-scope items
ACTIONABILITY: 85
ACTIONABILITY_FEEDBACK: Task has sufficient technical details to begin work`,
			expectError:    false,
			expectPass:     true,
			expectMinScore: 70.0,
			expectMaxScore: 100.0,
		},
		{
			name: "valid low-quality response",
			response: `CLARITY: 30
CLARITY_FEEDBACK: Task description is vague and unclear
ACCEPTANCE: 25
ACCEPTANCE_FEEDBACK: No clear acceptance criteria provided
SCOPE: 40
SCOPE_FEEDBACK: Scope is ambiguous and open-ended
ACTIONABILITY: 20
ACTIONABILITY_FEEDBACK: Insufficient details to begin implementation`,
			expectError:    false,
			expectPass:     false,
			expectMinScore: 0.0,
			expectMaxScore: 70.0,
		},
		{
			name: "response with extra whitespace",
			response: `  CLARITY:  75
  CLARITY_FEEDBACK:  Good clarity
  ACCEPTANCE:  80
  ACCEPTANCE_FEEDBACK:  Well defined
  SCOPE:  70
  SCOPE_FEEDBACK:  Clear scope
  ACTIONABILITY:  75
  ACTIONABILITY_FEEDBACK:  Actionable  `,
			expectError:    false,
			expectPass:     true,
			expectMinScore: 70.0,
			expectMaxScore: 100.0,
		},
		{
			name: "missing clarity score",
			response: `CLARITY_FEEDBACK: Good
ACCEPTANCE: 80
ACCEPTANCE_FEEDBACK: Well defined
SCOPE: 70
SCOPE_FEEDBACK: Clear
ACTIONABILITY: 75
ACTIONABILITY_FEEDBACK: Actionable`,
			expectError: true,
		},
		{
			name: "missing acceptance score",
			response: `CLARITY: 75
CLARITY_FEEDBACK: Good
SCOPE: 70
SCOPE_FEEDBACK: Clear
ACTIONABILITY: 75
ACTIONABILITY_FEEDBACK: Actionable`,
			expectError: true,
		},
		{
			name: "invalid score - not a number",
			response: `CLARITY: abc
CLARITY_FEEDBACK: Good
ACCEPTANCE: 80
ACCEPTANCE_FEEDBACK: Well defined
SCOPE: 70
SCOPE_FEEDBACK: Clear
ACTIONABILITY: 75
ACTIONABILITY_FEEDBACK: Actionable`,
			expectError: true,
		},
		{
			name: "invalid score - out of range (too high)",
			response: `CLARITY: 150
CLARITY_FEEDBACK: Good
ACCEPTANCE: 80
ACCEPTANCE_FEEDBACK: Well defined
SCOPE: 70
SCOPE_FEEDBACK: Clear
ACTIONABILITY: 75
ACTIONABILITY_FEEDBACK: Actionable`,
			expectError: true,
		},
		{
			name: "invalid score - negative",
			response: `CLARITY: -10
CLARITY_FEEDBACK: Poor
ACCEPTANCE: 80
ACCEPTANCE_FEEDBACK: Well defined
SCOPE: 70
SCOPE_FEEDBACK: Clear
ACTIONABILITY: 75
ACTIONABILITY_FEEDBACK: Actionable`,
			expectError: true,
		},
		{
			name:        "empty response",
			response:    "",
			expectError: true,
		},
		{
			name:        "whitespace only response",
			response:    "   \n\t  \n  ",
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

				// Verify pass/fail expectation
				if result.Passed != tt.expectPass {
					t.Errorf("Expected Passed=%v, got %v (score: %f)", tt.expectPass, result.Passed, result.Score)
				}

				// Verify score range
				if result.Score < tt.expectMinScore || result.Score > tt.expectMaxScore {
					t.Errorf("Expected score between %f and %f, got %f", tt.expectMinScore, tt.expectMaxScore, result.Score)
				}

				// Verify details contain all criteria
				expectedKeys := []string{"clarity", "acceptance", "scope", "actionability"}
				for _, key := range expectedKeys {
					if _, exists := result.Details[key]; !exists {
						t.Errorf("Expected Details to contain key '%s'", key)
					}
				}
			}
		})
	}
}

// TestTaskQualityGrader_GradeWithLLM_NilClient tests error when LLM client is not initialized
func TestTaskQualityGrader_GradeWithLLM_NilClient(t *testing.T) {
	grader := NewTaskQualityGrader()
	// grader.llmClient is nil by default

	taskContent := `# Test Task
Description: A test task`

	result, err := grader.gradeWithLLM(taskContent)

	if err == nil {
		t.Error("Expected error when LLM client is nil, got none")
	}

	expectedErrMsg := "LLM client not initialized"
	if err != nil && !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message containing '%s', got '%s'", expectedErrMsg, err.Error())
	}

	// Verify result is empty on error
	if result.Score != 0 {
		t.Errorf("Expected zero score on error, got %f", result.Score)
	}
	if result.Message != "" {
		t.Errorf("Expected empty message on error, got '%s'", result.Message)
	}
}

// TestTaskQualityGrader_GradeWithLLM_Success tests successful LLM-based grading
func TestTaskQualityGrader_GradeWithLLM_Success(t *testing.T) {
	// Create a mock LLM client that returns a valid response
	mockClient := &mockTaskQualityLLMClient{
		response: `CLARITY: 85
CLARITY_FEEDBACK: Task description is clear and well-structured
ACCEPTANCE: 90
ACCEPTANCE_FEEDBACK: Acceptance criteria are well-defined and testable
SCOPE: 80
SCOPE_FEEDBACK: Scope is bounded with clear out-of-scope items
ACTIONABILITY: 85
ACTIONABILITY_FEEDBACK: Task has sufficient technical details to begin work`,
	}

	grader := &TaskQualityGrader{
		weights: map[string]float64{
			"clarity":       0.30,
			"acceptance":    0.30,
			"scope":         0.25,
			"actionability": 0.15,
		},
		passingScore: 70.0,
		llmClient:    mockClient,
		timeout:      30 * time.Second,
	}

	taskContent := `# Implement User Authentication
Description: Create user auth system with JWT`

	result, err := grader.gradeWithLLM(taskContent)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify result fields
	if !result.Passed {
		t.Error("Expected Passed to be true")
	}
	if result.Score < 70.0 {
		t.Errorf("Expected passing score >= 70.0, got %f", result.Score)
	}
	if result.Details == nil {
		t.Error("Expected non-nil Details")
	}

	// Verify all criteria are in details
	expectedKeys := []string{"clarity", "acceptance", "scope", "actionability"}
	for _, key := range expectedKeys {
		if _, exists := result.Details[key]; !exists {
			t.Errorf("Expected Details to contain key '%s'", key)
		}
	}
}

// TestTaskQualityGrader_GradeWithLLM_Timeout tests timeout handling
func TestTaskQualityGrader_GradeWithLLM_Timeout(t *testing.T) {
	// Create a mock LLM client that simulates a timeout
	mockClient := &mockTaskQualityLLMClient{
		simulateTimeout: true,
	}

	grader := &TaskQualityGrader{
		weights: map[string]float64{
			"clarity":       0.30,
			"acceptance":    0.30,
			"scope":         0.25,
			"actionability": 0.15,
		},
		passingScore: 70.0,
		llmClient:    mockClient,
		timeout:      50 * time.Millisecond, // Very short timeout
	}

	taskContent := `# Test Task
Description: A test task`

	result, err := grader.gradeWithLLM(taskContent)

	if err == nil {
		t.Error("Expected timeout error, got none")
	}

	// Verify error message mentions timeout
	if err != nil && !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error message, got: %s", err.Error())
	}

	// Verify result is empty on error
	if result.Score != 0 {
		t.Errorf("Expected zero score on error, got %f", result.Score)
	}
}

// TestTaskQualityGrader_GradeWithLLM_APIError tests API error handling
func TestTaskQualityGrader_GradeWithLLM_APIError(t *testing.T) {
	// Create a mock LLM client that returns an error
	mockClient := &mockTaskQualityLLMClient{
		simulateError: true,
		errorMsg:      "API rate limit exceeded",
	}

	grader := &TaskQualityGrader{
		weights: map[string]float64{
			"clarity":       0.30,
			"acceptance":    0.30,
			"scope":         0.25,
			"actionability": 0.15,
		},
		passingScore: 70.0,
		llmClient:    mockClient,
		timeout:      30 * time.Second,
	}

	taskContent := `# Test Task
Description: A test task`

	result, err := grader.gradeWithLLM(taskContent)

	if err == nil {
		t.Error("Expected LLM error, got none")
	}

	// Verify error message contains "LLM request failed"
	if err != nil && !strings.Contains(err.Error(), "LLM request failed") {
		t.Errorf("Expected 'LLM request failed' in error, got: %s", err.Error())
	}

	// Verify result is empty on error
	if result.Score != 0 {
		t.Errorf("Expected zero score on error, got %f", result.Score)
	}
}

// TestTaskQualityGrader_GradeWithLLM_InvalidResponse tests handling of malformed LLM responses
func TestTaskQualityGrader_GradeWithLLM_InvalidResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{
			name:     "missing criterion",
			response: `CLARITY: 85
CLARITY_FEEDBACK: Good
ACCEPTANCE: 90
ACCEPTANCE_FEEDBACK: Good`,
			// Missing SCOPE and ACTIONABILITY
		},
		{
			name:     "invalid score format",
			response: `CLARITY: not-a-number
CLARITY_FEEDBACK: Good
ACCEPTANCE: 90
ACCEPTANCE_FEEDBACK: Good
SCOPE: 80
SCOPE_FEEDBACK: Good
ACTIONABILITY: 85
ACTIONABILITY_FEEDBACK: Good`,
		},
		{
			name:     "empty response",
			response: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockTaskQualityLLMClient{
				response: tt.response,
			}

			grader := &TaskQualityGrader{
				weights: map[string]float64{
					"clarity":       0.30,
					"acceptance":    0.30,
					"scope":         0.25,
					"actionability": 0.15,
				},
				passingScore: 70.0,
				llmClient:    mockClient,
				timeout:      30 * time.Second,
			}

			taskContent := `# Test Task
Description: A test task`

			result, err := grader.gradeWithLLM(taskContent)

			if err == nil {
				t.Error("Expected error for invalid response, got none")
			}

			// Verify result is empty on error
			if result.Score != 0 {
				t.Errorf("Expected zero score on error, got %f", result.Score)
			}
		})
	}
}

// mockTaskQualityLLMClient is a mock implementation of llm.Client for testing TaskQualityGrader
type mockTaskQualityLLMClient struct {
	response        string
	simulateTimeout bool
	simulateError   bool
	errorMsg        string
}

func (m *mockTaskQualityLLMClient) Complete(ctx context.Context, prompt string, options ...llm.CompletionOption) (string, error) {
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
