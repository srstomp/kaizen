package metrics

// PassAtK evaluates the pass@k metric for a set of evaluation results.
//
// pass@k is a capability measure that answers the question:
// "Can the system succeed at least once in k tries?"
//
// It returns true if at least one result in the slice is true,
// indicating the system is capable of producing a successful outcome.
// Returns false if all results are false or if the slice is empty.
//
// Example use cases:
// - Measuring if a code generation system can produce correct code
// - Evaluating if a search system can find the right answer
// - Determining if a system has the capability to solve a problem
//
// Semantics:
//   - PassAtK([]bool{true, false, false}) → true (capable, succeeded once)
//   - PassAtK([]bool{false, false, false}) → false (not capable)
//   - PassAtK([]bool{}) → false (no evidence of capability)
func PassAtK(results []bool) bool {
	for _, result := range results {
		if result {
			return true
		}
	}
	return false
}

// PassCaretK evaluates the pass^k metric for a set of evaluation results.
//
// pass^k is a consistency measure that answers the question:
// "Does the system succeed every time in k tries?"
//
// It returns true only if all results in the slice are true,
// indicating the system consistently produces successful outcomes.
// Returns false if any result is false or if the slice is empty.
//
// Example use cases:
// - Measuring reliability of a production system
// - Evaluating consistency of model outputs
// - Determining if a system is stable enough for deployment
//
// Semantics:
//   - PassCaretK([]bool{true, true, true}) → true (consistent)
//   - PassCaretK([]bool{true, false, true}) → false (inconsistent)
//   - PassCaretK([]bool{}) → false (no evidence of consistency)
func PassCaretK(results []bool) bool {
	if len(results) == 0 {
		return false
	}

	for _, result := range results {
		if !result {
			return false
		}
	}
	return true
}

// PassAtKProbability calculates the pass@k probability using the formula from
// the OpenAI Codex paper: pass@k = 1 - C(n-c, k) / C(n, k)
//
// Parameters:
//   - n: total number of samples generated
//   - c: number of correct (passing) samples
//   - k: number of samples to consider
//
// Returns the probability that at least one of k randomly selected samples
// from n total samples (with c correct) will be correct.
//
// Edge cases:
//   - If k > n, returns 0 (can't select more samples than exist)
//   - If c >= k, returns 1.0 (guaranteed to get at least one correct)
//   - If c == 0, returns 0.0 (no correct samples exist)
//   - If n == 0, returns 0.0 (no samples exist)
func PassAtKProbability(n, c, k int) float64 {
	// Edge case: no samples exist
	if n == 0 {
		return 0.0
	}

	// Edge case: can't select more samples than exist
	if k > n {
		return 0.0
	}

	// Edge case: no correct samples exist
	if c == 0 {
		return 0.0
	}

	// Edge case: not enough incorrect samples to avoid all correct ones
	// If there are fewer than k incorrect samples (n-c < k), then we must select at least one correct
	if n-c < k {
		return 1.0
	}

	// Calculate pass@k using the formula: 1 - C(n-c, k) / C(n, k)
	// This represents the probability that at least one of k samples is correct
	numerator := binomialCoefficient(n-c, k)
	denominator := binomialCoefficient(n, k)

	if denominator == 0 {
		return 0.0
	}

	return 1.0 - (numerator / denominator)
}

// binomialCoefficient calculates C(n, k) = n! / (k! * (n-k)!)
// Uses multiplicative formula to avoid overflow for reasonable inputs
func binomialCoefficient(n, k int) float64 {
	// Handle edge cases
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}

	// Optimize by using the smaller of k and n-k
	if k > n-k {
		k = n - k
	}

	// Use multiplicative formula: C(n,k) = n * (n-1) * ... * (n-k+1) / (k * (k-1) * ... * 1)
	result := 1.0
	for i := 0; i < k; i++ {
		result *= float64(n - i)
		result /= float64(i + 1)
	}

	return result
}

// EstimatePassAtK estimates pass@k from observed results.
// Given a slice of boolean results (true = pass, false = fail),
// calculates the pass@k probability for selecting k samples.
func EstimatePassAtK(results []bool, k int) float64 {
	n := len(results)
	if n == 0 {
		return 0.0
	}

	// Count correct samples
	c := 0
	for _, result := range results {
		if result {
			c++
		}
	}

	return PassAtKProbability(n, c, k)
}
