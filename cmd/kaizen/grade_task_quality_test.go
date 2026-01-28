package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TaskQualityOutput represents the JSON output from grade-task-quality command
// Uses TaskQualityResult and TaskQualityIssue from main.go
type TaskQualityOutput = TaskQualityResult

// TestRunGradeTaskQualityCommand_DescriptionLength tests description length check
func TestRunGradeTaskQualityCommand_DescriptionLength(t *testing.T) {
	testCases := []struct {
		name               string
		description        string
		minLength          int
		expectIssue        bool
		expectedCheckName  string
	}{
		{
			name:               "description_too_short",
			description:        "Short desc",
			minLength:          100,
			expectIssue:        true,
			expectedCheckName:  "description_length",
		},
		{
			name:               "description_meets_minimum",
			description:        strings.Repeat("a", 100),
			minLength:          100,
			expectIssue:        false,
		},
		{
			name:               "description_exceeds_minimum",
			description:        strings.Repeat("a", 150),
			minLength:          100,
			expectIssue:        false,
		},
		{
			name:               "empty_description",
			description:        "",
			minLength:          100,
			expectIssue:        true,
			expectedCheckName:  "description_length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskQuality(
				"test-task",
				"Test task",
				"feature",
				tc.description,
				"",
				tc.minLength,
				"json",
			)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runGradeTaskQuality failed: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result TaskQualityOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Check if issue exists as expected
			foundIssue := false
			for _, issue := range result.Issues {
				if issue.Check == tc.expectedCheckName {
					foundIssue = true
					if !strings.Contains(issue.Message, "short") && !strings.Contains(issue.Message, "length") {
						t.Errorf("Issue message should mention length: %s", issue.Message)
					}
				}
			}

			if tc.expectIssue && !foundIssue {
				t.Errorf("Expected issue with check name %q, but not found", tc.expectedCheckName)
			}
			if !tc.expectIssue && foundIssue {
				t.Errorf("Did not expect issue, but found one")
			}
		})
	}
}

// TestRunGradeTaskQualityCommand_AcceptanceCriteria tests acceptance criteria check
func TestRunGradeTaskQualityCommand_AcceptanceCriteria(t *testing.T) {
	longDescription := strings.Repeat("a", 100)

	testCases := []struct {
		name               string
		taskType           string
		acceptanceCriteria string
		expectIssue        bool
	}{
		{
			name:               "feature_missing_criteria",
			taskType:           "feature",
			acceptanceCriteria: "",
			expectIssue:        true,
		},
		{
			name:               "feature_with_criteria",
			taskType:           "feature",
			acceptanceCriteria: "User can login, User sees dashboard",
			expectIssue:        false,
		},
		{
			name:               "bug_missing_criteria",
			taskType:           "bug",
			acceptanceCriteria: "",
			expectIssue:        false, // Bugs don't require acceptance criteria
		},
		{
			name:               "chore_missing_criteria",
			taskType:           "chore",
			acceptanceCriteria: "",
			expectIssue:        false, // Chores don't require acceptance criteria
		},
		{
			name:               "test_missing_criteria",
			taskType:           "test",
			acceptanceCriteria: "",
			expectIssue:        true,
		},
		{
			name:               "spike_missing_criteria",
			taskType:           "spike",
			acceptanceCriteria: "",
			expectIssue:        true, // Spikes ALWAYS require acceptance criteria
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskQuality(
				"test-task",
				"Test task",
				tc.taskType,
				longDescription,
				tc.acceptanceCriteria,
				100,
				"json",
			)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runGradeTaskQuality failed: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result TaskQualityOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Check if acceptance_criteria issue exists
			foundIssue := false
			for _, issue := range result.Issues {
				if issue.Check == "acceptance_criteria" {
					foundIssue = true
				}
			}

			if tc.expectIssue && !foundIssue {
				t.Errorf("Expected acceptance_criteria issue for task type %q, but not found", tc.taskType)
			}
			if !tc.expectIssue && foundIssue {
				t.Errorf("Did not expect acceptance_criteria issue for task type %q, but found one", tc.taskType)
			}
		})
	}
}

// TestRunGradeTaskQualityCommand_AmbiguousKeywords tests ambiguous keyword detection
func TestRunGradeTaskQualityCommand_AmbiguousKeywords(t *testing.T) {
	longDescription := strings.Repeat("a", 100)

	testCases := []struct {
		name            string
		title           string
		description     string
		expectIssue     bool
		expectedKeyword string
	}{
		{
			name:            "title_with_investigate",
			title:           "Investigate performance issue",
			description:     longDescription,
			expectIssue:     true,
			expectedKeyword: "investigate",
		},
		{
			name:            "description_with_explore",
			title:           "Fix bug",
			description:     "We need to explore different solutions to this problem",
			expectIssue:     true,
			expectedKeyword: "explore",
		},
		{
			name:            "title_with_figure_out",
			title:           "Figure out why tests fail",
			description:     longDescription,
			expectIssue:     true,
			expectedKeyword: "figure out",
		},
		{
			name:            "description_with_look_into",
			title:           "Fix bug",
			description:     "Look into the root cause of this issue and fix it",
			expectIssue:     true,
			expectedKeyword: "look into",
		},
		{
			name:            "title_with_understand",
			title:           "Understand the codebase",
			description:     longDescription,
			expectIssue:     true,
			expectedKeyword: "understand",
		},
		{
			name:            "clean_title_and_description",
			title:           "Implement user authentication",
			description:     "Add JWT-based authentication to the API. User should be able to login and receive a token.",
			expectIssue:     false,
		},
		{
			name:            "case_insensitive_match",
			title:           "INVESTIGATE database performance",
			description:     longDescription,
			expectIssue:     true,
			expectedKeyword: "investigate",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskQuality(
				"test-task",
				tc.title,
				"feature",
				tc.description,
				"User can login",
				100,
				"json",
			)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runGradeTaskQuality failed: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result TaskQualityOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Check if ambiguous_keywords issue exists
			foundIssue := false
			for _, issue := range result.Issues {
				if issue.Check == "ambiguous_keywords" {
					foundIssue = true
					if tc.expectedKeyword != "" && !strings.Contains(strings.ToLower(issue.Message), strings.ToLower(tc.expectedKeyword)) {
						t.Errorf("Issue message should mention keyword %q: %s", tc.expectedKeyword, issue.Message)
					}
				}
			}

			if tc.expectIssue && !foundIssue {
				t.Errorf("Expected ambiguous_keywords issue, but not found")
			}
			if !tc.expectIssue && foundIssue {
				t.Errorf("Did not expect ambiguous_keywords issue, but found one: %+v", result.Issues)
			}
		})
	}
}

// TestRunGradeTaskQualityCommand_SpikeType tests spike type always requires acceptance criteria
func TestRunGradeTaskQualityCommand_SpikeType(t *testing.T) {
	longDescription := strings.Repeat("a", 100)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskQuality(
		"test-task",
		"Research new framework",
		"spike",
		longDescription,
		"", // No acceptance criteria
		100,
		"json",
	)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskQuality failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result TaskQualityOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Spike without acceptance criteria should ALWAYS have an issue
	foundIssue := false
	for _, issue := range result.Issues {
		if issue.Check == "acceptance_criteria" {
			foundIssue = true
		}
	}

	if !foundIssue {
		t.Errorf("Spike type without acceptance criteria should always fail, but no issue found")
	}
}

// TestRunGradeTaskQualityCommand_ScoreCalculation tests score calculation logic
func TestRunGradeTaskQualityCommand_ScoreCalculation(t *testing.T) {
	longDescription := strings.Repeat("a", 150)

	testCases := []struct {
		name               string
		taskType           string
		title              string
		description        string
		acceptanceCriteria string
		expectPass         bool
		expectedScore      float64 // Approximate expected score
	}{
		{
			name:               "perfect_task",
			taskType:           "feature",
			title:              "Implement user authentication",
			description:        longDescription,
			acceptanceCriteria: "User can login, User can logout, Token is stored",
			expectPass:         true,
			expectedScore:      100.0,
		},
		{
			name:               "one_issue",
			taskType:           "feature",
			title:              "Implement feature",
			description:        "Short",
			acceptanceCriteria: "Done",
			expectPass:         false,
			expectedScore:      75.0, // 1 out of 4 checks failed
		},
		{
			name:               "two_issues",
			taskType:           "feature",
			title:              "Investigate issue",
			description:        "Short",
			acceptanceCriteria: "Done",
			expectPass:         false,
			expectedScore:      50.0, // 2 out of 4 checks failed
		},
		{
			name:               "three_issues",
			taskType:           "feature",
			title:              "Investigate issue",
			description:        "Short",
			acceptanceCriteria: "",
			expectPass:         false,
			expectedScore:      25.0, // 3 out of 4 checks failed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskQuality(
				"test-task",
				tc.title,
				tc.taskType,
				tc.description,
				tc.acceptanceCriteria,
				100,
				"json",
			)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runGradeTaskQuality failed: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result TaskQualityOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Verify passed status
			if result.Passed != tc.expectPass {
				t.Errorf("Expected passed=%v, got %v", tc.expectPass, result.Passed)
			}

			// Verify score is within reasonable range
			scoreTolerance := 5.0
			if result.Score < tc.expectedScore-scoreTolerance || result.Score > tc.expectedScore+scoreTolerance {
				t.Errorf("Expected score ~%.1f, got %.1f", tc.expectedScore, result.Score)
			}
		})
	}
}

// TestRunGradeTaskQualityCommand_JSONOutput tests JSON output format
func TestRunGradeTaskQualityCommand_JSONOutput(t *testing.T) {
	longDescription := strings.Repeat("a", 100)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskQuality(
		"task-123",
		"Test task",
		"feature",
		longDescription,
		"Done",
		100,
		"json",
	)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskQuality failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output is valid JSON
	var result TaskQualityOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify required fields are present
	if result.TaskID != "task-123" {
		t.Errorf("Expected task_id 'task-123', got %s", result.TaskID)
	}

	// Score should be 0-100
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Score should be 0-100, got %.2f", result.Score)
	}

	// Issues should be an array (even if empty)
	if result.Issues == nil {
		t.Error("Issues should not be nil")
	}

	// If not passed, should have suggestion
	if !result.Passed && result.Suggestion == "" {
		t.Error("Failed task should have suggestion")
	}
}

// TestRunGradeTaskQualityCommand_TextOutput tests text output format
func TestRunGradeTaskQualityCommand_TextOutput(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskQuality(
		"task-456",
		"Test task",
		"feature",
		"Short",
		"",
		100,
		"text",
	)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskQuality failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify text output contains expected sections
	expectedStrings := []string{
		"Task Quality Check",
		"task-456",
		"Issues:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Text output missing expected string: %s\nOutput:\n%s", expected, output)
		}
	}
}

// TestRunGradeTaskQualityCommand_InvalidTaskType tests invalid task type handling
func TestRunGradeTaskQualityCommand_InvalidTaskType(t *testing.T) {
	longDescription := strings.Repeat("a", 100)

	testCases := []struct {
		name     string
		taskType string
		wantErr  bool
	}{
		{"valid_feature", "feature", false},
		{"valid_bug", "bug", false},
		{"valid_test", "test", false},
		{"valid_spike", "spike", false},
		{"valid_chore", "chore", false},
		{"invalid_type", "invalid", true},
		{"empty_type", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskQuality(
				"test-task",
				"Test task",
				tc.taskType,
				longDescription,
				"Done",
				100,
				"json",
			)

			w.Close()
			os.Stdout = oldStdout

			// Drain output
			var buf bytes.Buffer
			buf.ReadFrom(r)

			if tc.wantErr && err == nil {
				t.Errorf("Expected error for task type %q, got nil", tc.taskType)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Expected no error for task type %q, got: %v", tc.taskType, err)
			}
		})
	}
}

// TestRunGradeTaskQualityCommand_MinDescriptionLength tests custom min description length
func TestRunGradeTaskQualityCommand_MinDescriptionLength(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		minLength   int
		expectIssue bool
	}{
		{
			name:        "meets_custom_length",
			description: strings.Repeat("a", 50),
			minLength:   50,
			expectIssue: false,
		},
		{
			name:        "below_custom_length",
			description: strings.Repeat("a", 49),
			minLength:   50,
			expectIssue: true,
		},
		{
			name:        "exceeds_custom_length",
			description: strings.Repeat("a", 200),
			minLength:   50,
			expectIssue: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runGradeTaskQuality(
				"test-task",
				"Test task",
				"feature",
				tc.description,
				"Done",
				tc.minLength,
				"json",
			)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runGradeTaskQuality failed: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result TaskQualityOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			foundIssue := false
			for _, issue := range result.Issues {
				if issue.Check == "description_length" {
					foundIssue = true
				}
			}

			if tc.expectIssue && !foundIssue {
				t.Errorf("Expected description_length issue, but not found")
			}
			if !tc.expectIssue && foundIssue {
				t.Errorf("Did not expect description_length issue, but found one")
			}
		})
	}
}

// TestRunGradeTaskQualityCommand_Suggestion tests suggestion is provided when task fails
func TestRunGradeTaskQualityCommand_Suggestion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGradeTaskQuality(
		"test-task",
		"Test task",
		"feature",
		"Short description",
		"",
		100,
		"json",
	)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGradeTaskQuality failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result TaskQualityOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Task should fail (short description and missing acceptance criteria)
	if result.Passed {
		t.Error("Expected task to fail quality check")
	}

	// Should have suggestion
	if result.Suggestion == "" {
		t.Error("Failed task should have suggestion")
	}

	// Suggestion should mention brainstorm
	if !strings.Contains(result.Suggestion, "brainstorm") {
		t.Errorf("Suggestion should mention brainstorm: %s", result.Suggestion)
	}
}
