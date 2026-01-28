package metrics

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// TestResult represents the result of running a test case k times
type TestResult struct {
	TestID   string
	Name     string
	Expected string
	Runs     []string // Each run's verdict
}

// EvalResult represents the result of evaluating a failure case
type EvalResult struct {
	CaseID   string
	Category string
	Runs     []bool // Each run's pass/fail status
}

// CategoryMetrics represents evaluation metrics for a category
type CategoryMetrics struct {
	Total    int     `json:"total"`
	Pass     int     `json:"pass"`
	Fail     int     `json:"fail"`
	PassRate float64 `json:"pass_rate"`
}

// AggregatedMetrics contains aggregated evaluation metrics
type AggregatedMetrics struct {
	TotalItems  int                        `json:"total_items"`
	Accuracy    float64                    `json:"accuracy"`
	Consistency float64                    `json:"consistency"`
	PassRate    float64                    `json:"pass_rate"`
	ByCategory  map[string]CategoryMetrics `json:"by_category"`
}

// CalculateAccuracy calculates accuracy using majority vote
// Returns the percentage of test results where the majority verdict matches the expected verdict
func CalculateAccuracy(results []TestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	correctCount := 0
	for _, result := range results {
		verdict := getMajorityVerdict(result.Runs)
		if verdict == result.Expected {
			correctCount++
		}
	}

	return float64(correctCount) / float64(len(results))
}

// CalculateConsistency calculates consistency using pass^k logic
// Returns the percentage of test results where all runs agree
func CalculateConsistency(results []TestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	consistentCount := 0
	for _, result := range results {
		if areAllRunsConsistent(result.Runs) {
			consistentCount++
		}
	}

	return float64(consistentCount) / float64(len(results))
}

// CalculatePassRate calculates pass rate using majority vote
// Returns the percentage of eval results that passed (majority of runs are true)
func CalculatePassRate(results []EvalResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	passCount := 0
	for _, result := range results {
		if isMajorityPass(result.Runs) {
			passCount++
		}
	}

	return float64(passCount) / float64(len(results))
}

// AggregateByCategory aggregates eval results by category
// Returns a map of category names to their metrics
func AggregateByCategory(results []EvalResult) map[string]CategoryMetrics {
	metrics := make(map[string]CategoryMetrics)

	for _, result := range results {
		cat := result.Category
		m := metrics[cat]
		m.Total++

		if isMajorityPass(result.Runs) {
			m.Pass++
		} else {
			m.Fail++
		}

		metrics[cat] = m
	}

	// Calculate pass rates
	for cat, m := range metrics {
		if m.Total > 0 {
			m.PassRate = float64(m.Pass) / float64(m.Total)
		}
		metrics[cat] = m
	}

	return metrics
}

// FormatAsJSON formats aggregated metrics as JSON
func FormatAsJSON(metrics AggregatedMetrics) (string, error) {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling metrics to JSON: %w", err)
	}
	return string(data), nil
}

// FormatAsTable formats aggregated metrics as a human-readable table
func FormatAsTable(metrics AggregatedMetrics) string {
	var sb strings.Builder

	sb.WriteString("Aggregated Metrics\n")
	sb.WriteString("==================\n\n")

	sb.WriteString(fmt.Sprintf("Total Items: %d\n", metrics.TotalItems))
	sb.WriteString(fmt.Sprintf("Accuracy: %.1f%%\n", metrics.Accuracy*100))
	sb.WriteString(fmt.Sprintf("Consistency: %.1f%%\n", metrics.Consistency*100))
	sb.WriteString(fmt.Sprintf("Pass Rate: %.1f%%\n\n", metrics.PassRate*100))

	if len(metrics.ByCategory) > 0 {
		sb.WriteString("By Category:\n")
		sb.WriteString(fmt.Sprintf("%-20s | %-6s | %-6s | %-6s | %-10s\n",
			"Category", "Total", "Pass", "Fail", "Pass Rate"))
		sb.WriteString(strings.Repeat("-", 70) + "\n")

		// Sort categories for consistent output
		categories := make([]string, 0, len(metrics.ByCategory))
		for cat := range metrics.ByCategory {
			categories = append(categories, cat)
		}
		sort.Strings(categories)

		for _, cat := range categories {
			m := metrics.ByCategory[cat]
			sb.WriteString(fmt.Sprintf("%-20s | %-6d | %-6d | %-6d | %9.1f%%\n",
				cat, m.Total, m.Pass, m.Fail, m.PassRate*100))
		}
	}

	return sb.String()
}

// getMajorityVerdict returns the most common verdict from runs
// In case of a tie, returns the alphabetically first verdict for determinism
func getMajorityVerdict(runs []string) string {
	if len(runs) == 0 {
		return ""
	}

	counts := make(map[string]int)
	for _, verdict := range runs {
		counts[verdict]++
	}

	// Find the verdict with highest count
	// In case of tie, collect all tied verdicts and return alphabetically first
	maxCount := 0
	var tiedVerdicts []string

	for verdict, count := range counts {
		if count > maxCount {
			maxCount = count
			tiedVerdicts = []string{verdict}
		} else if count == maxCount {
			tiedVerdicts = append(tiedVerdicts, verdict)
		}
	}

	// If multiple verdicts are tied, sort and return first (deterministic)
	if len(tiedVerdicts) > 1 {
		sort.Strings(tiedVerdicts)
	}

	return tiedVerdicts[0]
}

// areAllRunsConsistent checks if all runs returned the same verdict
func areAllRunsConsistent(runs []string) bool {
	if len(runs) <= 1 {
		return true
	}

	first := runs[0]
	for _, verdict := range runs[1:] {
		if verdict != first {
			return false
		}
	}

	return true
}

// isMajorityPass checks if majority of runs are true (passed)
// Ties result in false (fail)
func isMajorityPass(runs []bool) bool {
	if len(runs) == 0 {
		return false
	}

	passCount := 0
	for _, run := range runs {
		if run {
			passCount++
		}
	}

	// Strict majority required (more than half)
	return passCount > len(runs)/2
}
