package metrics

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCalculateAccuracy tests accuracy calculation with majority vote
func TestCalculateAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		results  []TestResult
		expected float64
	}{
		{
			name: "all correct with unanimous agreement",
			results: []TestResult{
				{TestID: "T1", Expected: "PASS", Runs: []string{"PASS", "PASS", "PASS"}},
				{TestID: "T2", Expected: "FAIL", Runs: []string{"FAIL", "FAIL", "FAIL"}},
			},
			expected: 1.0,
		},
		{
			name: "all correct with majority vote",
			results: []TestResult{
				{TestID: "T1", Expected: "PASS", Runs: []string{"PASS", "PASS", "FAIL"}},
				{TestID: "T2", Expected: "FAIL", Runs: []string{"FAIL", "FAIL", "PASS"}},
			},
			expected: 1.0,
		},
		{
			name: "50% correct",
			results: []TestResult{
				{TestID: "T1", Expected: "PASS", Runs: []string{"PASS", "PASS", "PASS"}},
				{TestID: "T2", Expected: "PASS", Runs: []string{"FAIL", "FAIL", "FAIL"}},
			},
			expected: 0.5,
		},
		{
			name: "all incorrect",
			results: []TestResult{
				{TestID: "T1", Expected: "PASS", Runs: []string{"FAIL", "FAIL", "FAIL"}},
				{TestID: "T2", Expected: "FAIL", Runs: []string{"PASS", "PASS", "PASS"}},
			},
			expected: 0.0,
		},
		{
			name:     "empty results",
			results:  []TestResult{},
			expected: 0.0,
		},
		{
			name: "single result correct",
			results: []TestResult{
				{TestID: "T1", Expected: "PASS", Runs: []string{"PASS"}},
			},
			expected: 1.0,
		},
		{
			name: "tie breaks alphabetically",
			results: []TestResult{
				{TestID: "T1", Expected: "FAIL", Runs: []string{"PASS", "FAIL"}}, // Tie: FAIL wins alphabetically
			},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accuracy := CalculateAccuracy(tt.results)
			if accuracy != tt.expected {
				t.Errorf("CalculateAccuracy() = %v, expected %v", accuracy, tt.expected)
			}
		})
	}
}

// TestCalculateConsistency tests consistency calculation
func TestCalculateConsistency(t *testing.T) {
	tests := []struct {
		name     string
		results  []TestResult
		expected float64
	}{
		{
			name: "all consistent",
			results: []TestResult{
				{TestID: "T1", Runs: []string{"PASS", "PASS", "PASS"}},
				{TestID: "T2", Runs: []string{"FAIL", "FAIL", "FAIL"}},
			},
			expected: 1.0,
		},
		{
			name: "none consistent",
			results: []TestResult{
				{TestID: "T1", Runs: []string{"PASS", "FAIL", "PASS"}},
				{TestID: "T2", Runs: []string{"FAIL", "PASS", "FAIL"}},
			},
			expected: 0.0,
		},
		{
			name: "50% consistent",
			results: []TestResult{
				{TestID: "T1", Runs: []string{"PASS", "PASS", "PASS"}},
				{TestID: "T2", Runs: []string{"PASS", "FAIL", "PASS"}},
			},
			expected: 0.5,
		},
		{
			name:     "empty results",
			results:  []TestResult{},
			expected: 0.0,
		},
		{
			name: "single run is consistent",
			results: []TestResult{
				{TestID: "T1", Runs: []string{"PASS"}},
			},
			expected: 1.0,
		},
		{
			name: "empty runs is consistent",
			results: []TestResult{
				{TestID: "T1", Runs: []string{}},
			},
			expected: 1.0, // Empty or single run is considered consistent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consistency := CalculateConsistency(tt.results)
			if consistency != tt.expected {
				t.Errorf("CalculateConsistency() = %v, expected %v", consistency, tt.expected)
			}
		})
	}
}

// TestCalculatePassRate tests pass rate calculation
func TestCalculatePassRate(t *testing.T) {
	tests := []struct {
		name     string
		results  []EvalResult
		expected float64
	}{
		{
			name: "all pass",
			results: []EvalResult{
				{CaseID: "C1", Runs: []bool{true, true, true}},
				{CaseID: "C2", Runs: []bool{true, true, true}},
			},
			expected: 1.0,
		},
		{
			name: "all fail",
			results: []EvalResult{
				{CaseID: "C1", Runs: []bool{false, false, false}},
				{CaseID: "C2", Runs: []bool{false, false, false}},
			},
			expected: 0.0,
		},
		{
			name: "50% pass with majority vote",
			results: []EvalResult{
				{CaseID: "C1", Runs: []bool{true, true, false}},  // Majority pass
				{CaseID: "C2", Runs: []bool{false, false, true}}, // Majority fail
			},
			expected: 0.5,
		},
		{
			name: "tie breaks to fail",
			results: []EvalResult{
				{CaseID: "C1", Runs: []bool{true, false}}, // Tie = fail
			},
			expected: 0.0,
		},
		{
			name:     "empty results",
			results:  []EvalResult{},
			expected: 0.0,
		},
		{
			name: "single pass",
			results: []EvalResult{
				{CaseID: "C1", Runs: []bool{true}},
			},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passRate := CalculatePassRate(tt.results)
			if passRate != tt.expected {
				t.Errorf("CalculatePassRate() = %v, expected %v", passRate, tt.expected)
			}
		})
	}
}

// TestAggregateByCategory tests category aggregation
func TestAggregateByCategory(t *testing.T) {
	tests := []struct {
		name     string
		results  []EvalResult
		expected map[string]CategoryMetrics
	}{
		{
			name: "single category all pass",
			results: []EvalResult{
				{CaseID: "C1", Category: "task-decomp", Runs: []bool{true, true, true}},
				{CaseID: "C2", Category: "task-decomp", Runs: []bool{true, true, false}},
			},
			expected: map[string]CategoryMetrics{
				"task-decomp": {Total: 2, Pass: 2, Fail: 0, PassRate: 1.0},
			},
		},
		{
			name: "multiple categories",
			results: []EvalResult{
				{CaseID: "C1", Category: "task-decomp", Runs: []bool{true, true, true}},
				{CaseID: "C2", Category: "task-decomp", Runs: []bool{false, false, false}},
				{CaseID: "C3", Category: "code-gen", Runs: []bool{true, true, false}},
			},
			expected: map[string]CategoryMetrics{
				"task-decomp": {Total: 2, Pass: 1, Fail: 1, PassRate: 0.5},
				"code-gen":    {Total: 1, Pass: 1, Fail: 0, PassRate: 1.0},
			},
		},
		{
			name:     "empty results",
			results:  []EvalResult{},
			expected: map[string]CategoryMetrics{},
		},
		{
			name: "all fail in category",
			results: []EvalResult{
				{CaseID: "C1", Category: "task-decomp", Runs: []bool{false, false, false}},
			},
			expected: map[string]CategoryMetrics{
				"task-decomp": {Total: 1, Pass: 0, Fail: 1, PassRate: 0.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateByCategory(tt.results)

			// Check all expected categories
			for cat, expectedMetrics := range tt.expected {
				actualMetrics, exists := result[cat]
				if !exists {
					t.Errorf("Category %s not found in result", cat)
					continue
				}

				if actualMetrics.Total != expectedMetrics.Total {
					t.Errorf("Category %s: Total = %d, expected %d", cat, actualMetrics.Total, expectedMetrics.Total)
				}
				if actualMetrics.Pass != expectedMetrics.Pass {
					t.Errorf("Category %s: Pass = %d, expected %d", cat, actualMetrics.Pass, expectedMetrics.Pass)
				}
				if actualMetrics.Fail != expectedMetrics.Fail {
					t.Errorf("Category %s: Fail = %d, expected %d", cat, actualMetrics.Fail, expectedMetrics.Fail)
				}
				if actualMetrics.PassRate != expectedMetrics.PassRate {
					t.Errorf("Category %s: PassRate = %v, expected %v", cat, actualMetrics.PassRate, expectedMetrics.PassRate)
				}
			}

			// Check no unexpected categories
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d categories, got %d", len(tt.expected), len(result))
			}
		})
	}
}

// TestFormatAsJSON tests JSON formatting
func TestFormatAsJSON(t *testing.T) {
	tests := []struct {
		name    string
		metrics AggregatedMetrics
		wantErr bool
	}{
		{
			name: "complete metrics",
			metrics: AggregatedMetrics{
				TotalItems:  10,
				Accuracy:    0.8,
				Consistency: 0.6,
				PassRate:    0.75,
				ByCategory: map[string]CategoryMetrics{
					"task-decomp": {Total: 5, Pass: 4, Fail: 1, PassRate: 0.8},
					"code-gen":    {Total: 5, Pass: 3, Fail: 2, PassRate: 0.6},
				},
			},
			wantErr: false,
		},
		{
			name: "empty metrics",
			metrics: AggregatedMetrics{
				TotalItems:  0,
				Accuracy:    0.0,
				Consistency: 0.0,
				PassRate:    0.0,
				ByCategory:  map[string]CategoryMetrics{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr, err := FormatAsJSON(tt.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatAsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify it's valid JSON
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
					t.Errorf("FormatAsJSON() produced invalid JSON: %v", err)
				}

				// Verify key fields exist
				if _, ok := parsed["total_items"]; !ok {
					t.Error("JSON missing 'total_items' field")
				}
				if _, ok := parsed["accuracy"]; !ok {
					t.Error("JSON missing 'accuracy' field")
				}
				if _, ok := parsed["consistency"]; !ok {
					t.Error("JSON missing 'consistency' field")
				}
				if _, ok := parsed["pass_rate"]; !ok {
					t.Error("JSON missing 'pass_rate' field")
				}
				if _, ok := parsed["by_category"]; !ok {
					t.Error("JSON missing 'by_category' field")
				}
			}
		})
	}
}

// TestFormatAsTable tests table formatting
func TestFormatAsTable(t *testing.T) {
	tests := []struct {
		name          string
		metrics       AggregatedMetrics
		expectInTable []string
	}{
		{
			name: "complete metrics",
			metrics: AggregatedMetrics{
				TotalItems:  10,
				Accuracy:    0.8,
				Consistency: 0.6,
				PassRate:    0.75,
				ByCategory: map[string]CategoryMetrics{
					"task-decomp": {Total: 5, Pass: 4, Fail: 1, PassRate: 0.8},
					"code-gen":    {Total: 5, Pass: 3, Fail: 2, PassRate: 0.6},
				},
			},
			expectInTable: []string{
				"Aggregated Metrics",
				"Total Items:",
				"Accuracy:",
				"Consistency:",
				"Pass Rate:",
				"Category",
				"task-decomp",
				"code-gen",
			},
		},
		{
			name: "empty metrics",
			metrics: AggregatedMetrics{
				TotalItems:  0,
				Accuracy:    0.0,
				Consistency: 0.0,
				PassRate:    0.0,
				ByCategory:  map[string]CategoryMetrics{},
			},
			expectInTable: []string{
				"Aggregated Metrics",
				"Total Items:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := FormatAsTable(tt.metrics)

			for _, expected := range tt.expectInTable {
				if !strings.Contains(table, expected) {
					t.Errorf("FormatAsTable() missing expected string: %s", expected)
				}
			}
		})
	}
}

// TestEdgeCases tests edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("accuracy with no test results", func(t *testing.T) {
		result := CalculateAccuracy([]TestResult{})
		if result != 0.0 {
			t.Errorf("Expected 0.0 for empty results, got %v", result)
		}
	})

	t.Run("consistency with no test results", func(t *testing.T) {
		result := CalculateConsistency([]TestResult{})
		if result != 0.0 {
			t.Errorf("Expected 0.0 for empty results, got %v", result)
		}
	})

	t.Run("pass rate with no eval results", func(t *testing.T) {
		result := CalculatePassRate([]EvalResult{})
		if result != 0.0 {
			t.Errorf("Expected 0.0 for empty results, got %v", result)
		}
	})

	t.Run("aggregate by category with no results", func(t *testing.T) {
		result := AggregateByCategory([]EvalResult{})
		if len(result) != 0 {
			t.Errorf("Expected empty map for empty results, got %v", result)
		}
	})

	t.Run("test result with empty runs", func(t *testing.T) {
		results := []TestResult{
			{TestID: "T1", Expected: "PASS", Runs: []string{}},
		}
		accuracy := CalculateAccuracy(results)
		// Empty runs means no verdict, which can't match expected
		if accuracy != 0.0 {
			t.Errorf("Expected 0.0 for empty runs, got %v", accuracy)
		}
	})

	t.Run("eval result with empty runs", func(t *testing.T) {
		results := []EvalResult{
			{CaseID: "C1", Runs: []bool{}},
		}
		passRate := CalculatePassRate(results)
		// Empty runs means no passes
		if passRate != 0.0 {
			t.Errorf("Expected 0.0 for empty runs, got %v", passRate)
		}
	})
}
