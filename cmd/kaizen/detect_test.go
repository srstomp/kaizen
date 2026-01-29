package main

import (
	"encoding/json"
	"testing"
)

func TestDetectCategoryCommand(t *testing.T) {
	tests := []struct {
		name              string
		details           string
		wantCategory      string
		wantAllCategories []string
		wantMatched       bool
	}{
		{
			name:              "single match - missing tests",
			details:           "This task is missing test coverage",
			wantCategory:      "missing-tests",
			wantAllCategories: []string{"missing-tests"},
			wantMatched:       true,
		},
		{
			name:              "single match - scope creep",
			details:           "This feature is out of scope",
			wantCategory:      "scope-creep",
			wantAllCategories: []string{"scope-creep"},
			wantMatched:       true,
		},
		{
			name:              "single match - wrong product",
			details:           "Wrong file was modified",
			wantCategory:      "wrong-product",
			wantAllCategories: []string{"wrong-product"},
			wantMatched:       true,
		},
		{
			name:              "multiple matches - returns first",
			details:           "This task is missing test and has extra features out of scope",
			wantCategory:      "missing-tests",
			wantAllCategories: []string{"missing-tests", "scope-creep"},
			wantMatched:       true,
		},
		{
			name:              "no match - returns unknown",
			details:           "Everything looks good",
			wantCategory:      "unknown",
			wantAllCategories: []string{},
			wantMatched:       false,
		},
		{
			name:              "case insensitive match",
			details:           "MISSING TEST for this feature",
			wantCategory:      "missing-tests",
			wantAllCategories: []string{"missing-tests"},
			wantMatched:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the detect command
			output, err := runDetectCategoryCommand(tt.details)
			if err != nil {
				t.Fatalf("runDetectCategoryCommand failed: %v", err)
			}

			// Parse JSON output
			var result DetectOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Verify output
			if result.DetectedCategory != tt.wantCategory {
				t.Errorf("DetectedCategory = %q, want %q", result.DetectedCategory, tt.wantCategory)
			}
			if result.Matched != tt.wantMatched {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatched)
			}

			// Verify all categories
			if len(result.AllCategories) != len(tt.wantAllCategories) {
				t.Errorf("AllCategories length = %d, want %d", len(result.AllCategories), len(tt.wantAllCategories))
			} else {
				for i, cat := range tt.wantAllCategories {
					if result.AllCategories[i] != cat {
						t.Errorf("AllCategories[%d] = %q, want %q", i, result.AllCategories[i], cat)
					}
				}
			}
		})
	}
}

func TestDetectCategoryCommandEmptyDetails(t *testing.T) {
	// Run with empty details
	output, err := runDetectCategoryCommand("")
	if err != nil {
		t.Fatalf("runDetectCategoryCommand failed: %v", err)
	}

	// Parse JSON output
	var result DetectOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Should return no match
	if result.DetectedCategory != "unknown" {
		t.Errorf("DetectedCategory = %q, want %q", result.DetectedCategory, "unknown")
	}
	if result.Matched != false {
		t.Errorf("Matched = %v, want %v", result.Matched, false)
	}
	if len(result.AllCategories) != 0 {
		t.Errorf("AllCategories length = %d, want %d", len(result.AllCategories), 0)
	}
}
