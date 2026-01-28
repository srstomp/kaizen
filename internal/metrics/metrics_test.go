package metrics

import (
	"math"
	"testing"
)

func TestPassAtK(t *testing.T) {
	tests := []struct {
		name     string
		results  []bool
		expected bool
	}{
		{
			name:     "k=1, single pass",
			results:  []bool{true},
			expected: true,
		},
		{
			name:     "k=1, single fail",
			results:  []bool{false},
			expected: false,
		},
		{
			name:     "all pass",
			results:  []bool{true, true, true},
			expected: true,
		},
		{
			name:     "all fail",
			results:  []bool{false, false, false},
			expected: false,
		},
		{
			name:     "mixed with success - majority pass",
			results:  []bool{true, true, false},
			expected: true,
		},
		{
			name:     "mixed with success - single success at end",
			results:  []bool{false, false, true},
			expected: true,
		},
		{
			name:     "empty slice",
			results:  []bool{},
			expected: false,
		},
		{
			name:     "single success in middle",
			results:  []bool{false, true, false, false},
			expected: true,
		},
		{
			name:     "large k with one success",
			results:  []bool{false, false, false, false, false, false, false, false, false, true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PassAtK(tt.results)
			if result != tt.expected {
				t.Errorf("PassAtK(%v) = %v, expected %v", tt.results, result, tt.expected)
			}
		})
	}
}

func TestPassCaretK(t *testing.T) {
	tests := []struct {
		name     string
		results  []bool
		expected bool
	}{
		{
			name:     "k=1, single pass",
			results:  []bool{true},
			expected: true,
		},
		{
			name:     "k=1, single fail",
			results:  []bool{false},
			expected: false,
		},
		{
			name:     "all pass",
			results:  []bool{true, true, true},
			expected: true,
		},
		{
			name:     "all fail",
			results:  []bool{false, false, false},
			expected: false,
		},
		{
			name:     "mixed - majority pass but not all",
			results:  []bool{true, true, false},
			expected: false,
		},
		{
			name:     "mixed - single success not enough",
			results:  []bool{false, false, true},
			expected: false,
		},
		{
			name:     "empty slice",
			results:  []bool{},
			expected: false,
		},
		{
			name:     "single failure ruins consistency",
			results:  []bool{true, true, true, false, true},
			expected: false,
		},
		{
			name:     "large k all pass",
			results:  []bool{true, true, true, true, true, true, true, true, true, true},
			expected: true,
		},
		{
			name:     "large k with one fail",
			results:  []bool{true, true, true, true, true, true, true, true, true, false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PassCaretK(tt.results)
			if result != tt.expected {
				t.Errorf("PassCaretK(%v) = %v, expected %v", tt.results, result, tt.expected)
			}
		})
	}
}

// TestPassAtKSemantics verifies the capability measure semantics
func TestPassAtKSemantics(t *testing.T) {
	t.Run("pass@k measures capability - succeeds at least once", func(t *testing.T) {
		// Even with mostly failures, one success means capable
		results := []bool{false, false, false, false, false, false, false, false, false, true}
		if !PassAtK(results) {
			t.Error("pass@k should return true if capable of succeeding at least once")
		}
	})

	t.Run("pass@k returns false only when never succeeds", func(t *testing.T) {
		results := []bool{false, false, false, false, false}
		if PassAtK(results) {
			t.Error("pass@k should return false when never succeeds")
		}
	})
}

// TestPassCaretKSemantics verifies the consistency measure semantics
func TestPassCaretKSemantics(t *testing.T) {
	t.Run("pass^k measures consistency - must succeed every time", func(t *testing.T) {
		// All successes required for consistency
		results := []bool{true, true, true, true, true}
		if !PassCaretK(results) {
			t.Error("pass^k should return true when succeeds every time")
		}
	})

	t.Run("pass^k returns false with any failure", func(t *testing.T) {
		// Single failure breaks consistency
		results := []bool{true, true, true, true, false}
		if PassCaretK(results) {
			t.Error("pass^k should return false with any failure")
		}
	})
}

func TestPassAtKProbability(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		c        int
		k        int
		expected float64
		epsilon  float64 // tolerance for float comparison
	}{
		{
			name:     "n=10, c=5, k=1 should be 0.5",
			n:        10,
			c:        5,
			k:        1,
			expected: 0.5,
			epsilon:  0.0001,
		},
		{
			name:     "n=10, c=5, k=5 should be high probability",
			n:        10,
			c:        5,
			k:        5,
			expected: 0.9960,
			epsilon:  0.0001,
		},
		{
			name:     "n=10, c=1, k=1 should be 0.1",
			n:        10,
			c:        1,
			k:        1,
			expected: 0.1,
			epsilon:  0.0001,
		},
		{
			name:     "n=10, c=10, k=5 should be 1.0",
			n:        10,
			c:        10,
			k:        5,
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "n=10, c=0, k=5 should be 0.0",
			n:        10,
			c:        0,
			k:        5,
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "k > n should return 0",
			n:        10,
			c:        5,
			k:        15,
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "n=0 should return 0",
			n:        0,
			c:        0,
			k:        5,
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "c >= k should return 1.0",
			n:        10,
			c:        7,
			k:        5,
			expected: 1.0,
			epsilon:  0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PassAtKProbability(tt.n, tt.c, tt.k)
			if math.Abs(result-tt.expected) > tt.epsilon {
				t.Errorf("PassAtKProbability(%d, %d, %d) = %f, expected %f (tolerance %f)",
					tt.n, tt.c, tt.k, result, tt.expected, tt.epsilon)
			}
		})
	}
}

func TestEstimatePassAtK(t *testing.T) {
	tests := []struct {
		name     string
		results  []bool
		k        int
		expected float64
		epsilon  float64
	}{
		{
			name:     "all true results",
			results:  []bool{true, true, true, true, true},
			k:        3,
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "all false results",
			results:  []bool{false, false, false, false, false},
			k:        3,
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "mixed results - 5 out of 10 correct, k=1",
			results:  []bool{true, false, true, false, true, false, true, false, true, false},
			k:        1,
			expected: 0.5,
			epsilon:  0.0001,
		},
		{
			name:     "mixed results - 5 out of 10 correct, k=5",
			results:  []bool{true, false, true, false, true, false, true, false, true, false},
			k:        5,
			expected: 0.9960,
			epsilon:  0.001,
		},
		{
			name:     "single result true",
			results:  []bool{true},
			k:        1,
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "single result false",
			results:  []bool{false},
			k:        1,
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "k larger than results should return 0",
			results:  []bool{true, true, false},
			k:        10,
			expected: 0.0,
			epsilon:  0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimatePassAtK(tt.results, tt.k)
			if math.Abs(result-tt.expected) > tt.epsilon {
				t.Errorf("EstimatePassAtK(%v, %d) = %f, expected %f (tolerance %f)",
					tt.results, tt.k, result, tt.expected, tt.epsilon)
			}
		})
	}
}
