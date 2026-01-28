package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// EvalConfig represents the structure of an eval.yaml file
type EvalConfig struct {
	Agent                string     `yaml:"agent"`
	ConsistencyThreshold float64    `yaml:"consistency_threshold"`
	TestCases            []TestCase `yaml:"test_cases"`
}

// TestCase represents a single test case in the eval.yaml
type TestCase struct {
	ID        string    `yaml:"id"`
	Name      string    `yaml:"name"`
	Input     TaskInput `yaml:"input"`
	Expected  string    `yaml:"expected"`
	K         int       `yaml:"k"`
	Rationale string    `yaml:"rationale"`
}

// TaskInput represents the input to the agent being tested
type TaskInput struct {
	TaskTitle          string   `yaml:"task_title"`
	TaskDescription    string   `yaml:"task_description"`
	AcceptanceCriteria []string `yaml:"acceptance_criteria"`
	Implementation     string   `yaml:"implementation"`
}

// TestResult represents the result of running a test case k times
type TestResult struct {
	TestID   string
	Name     string
	Expected string
	Runs     []string // Each run's verdict
}

// EvaluationResult represents the complete evaluation result for an agent
type EvaluationResult struct {
	Agent       string
	TestResults []TestResult
}

// Metrics represents calculated metrics for the evaluation
type Metrics struct {
	Accuracy        float64
	Consistency     float64
	TotalTests      int
	CorrectCount    int
	ConsistentCount int
}

// loadEvalYAML loads and parses an eval.yaml file
func loadEvalYAML(path string) (*EvalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading eval.yaml: %w", err)
	}

	var config EvalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing eval.yaml: %w", err)
	}

	// Validate the config
	if err := ValidateEvalConfig(&config); err != nil {
		return nil, fmt.Errorf("validating eval.yaml: %w", err)
	}

	return &config, nil
}

// ValidateEvalConfig validates an EvalConfig against the schema rules
func ValidateEvalConfig(config *EvalConfig) error {
	// Validate agent name
	if config.Agent == "" {
		return fmt.Errorf("agent is required")
	}
	agentPattern := regexp.MustCompile(`^yokay-[a-z-]+$`)
	if !agentPattern.MatchString(config.Agent) {
		return fmt.Errorf("agent name must match pattern ^yokay-[a-z-]+$, got: %s", config.Agent)
	}

	// Validate consistency threshold
	if config.ConsistencyThreshold < 0.0 || config.ConsistencyThreshold > 1.0 {
		return fmt.Errorf("consistency_threshold must be between 0.0 and 1.0, got: %f", config.ConsistencyThreshold)
	}

	// Validate test cases
	if len(config.TestCases) == 0 {
		return fmt.Errorf("test_cases must contain at least 1 test case")
	}

	// Validate each test case
	testIDPattern := regexp.MustCompile(`^[A-Z]{2,3}-\d{3}$`)
	for _, tc := range config.TestCases {
		// Validate ID
		if !testIDPattern.MatchString(tc.ID) {
			return fmt.Errorf("test case ID '%s' must match pattern ^[A-Z]{2,3}-\\d{3}$", tc.ID)
		}

		// Validate name
		if tc.Name == "" {
			return fmt.Errorf("test case %s: name is required", tc.ID)
		}

		// Validate expected (must be non-empty, common values: PASS, FAIL, REFINED, NEEDS_INPUT, SKIP)
		if tc.Expected == "" {
			return fmt.Errorf("test case %s: expected is required", tc.ID)
		}

		// Validate k (optional, but if set must be in valid range)
		if tc.K > 100 {
			return fmt.Errorf("test case %s: k must be between 1 and 100 (or 0 for default), got: %d", tc.ID, tc.K)
		}

		// Validate rationale
		if tc.Rationale == "" {
			return fmt.Errorf("test case %s: rationale is required", tc.ID)
		}

		// Validate input
		if err := validateTaskInput(tc.ID, &tc.Input); err != nil {
			return err
		}
	}

	return nil
}

// validateTaskInput validates a TaskInput structure
func validateTaskInput(testID string, input *TaskInput) error {
	if input.TaskTitle == "" {
		return fmt.Errorf("test case %s: input.task_title is required", testID)
	}
	// TaskDescription and Implementation are optional - different agents need different fields:
	// - Brainstormer needs: task_title, task_description, acceptance_criteria (optional)
	// - Quality-reviewer needs: task_title, implementation
	// At least one of task_description or implementation should be provided
	if input.TaskDescription == "" && input.Implementation == "" {
		return fmt.Errorf("test case %s: at least one of input.task_description or input.implementation is required", testID)
	}
	return nil
}

// findEvalFiles finds all eval.yaml files in the given directory
func findEvalFiles(dir string) ([]string, error) {
	var evalFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "eval.yaml" {
			evalFiles = append(evalFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return evalFiles, nil
}

// findAgentEvalFiles finds all eval.yaml files in the agents directory
// DEPRECATED: Use findEvalFiles instead
func findAgentEvalFiles(agentsDir string) ([]string, error) {
	return findEvalFiles(agentsDir)
}

// findSkillEvalFiles finds all eval.yaml files in the skills directory
// DEPRECATED: Use findEvalFiles instead
func findSkillEvalFiles(skillsDir string) ([]string, error) {
	return findEvalFiles(skillsDir)
}

// runMetaEvaluation runs meta-evaluation on a single eval.yaml file
// kOverride: if > 0, overrides the k value from YAML test cases
func runMetaEvaluation(evalPath string, kOverride int) (EvaluationResult, error) {
	config, err := loadEvalYAML(evalPath)
	if err != nil {
		return EvaluationResult{}, err
	}

	result := EvaluationResult{
		Agent:       config.Agent,
		TestResults: make([]TestResult, 0, len(config.TestCases)),
	}

	// For each test case, run k times
	for tcIdx, tc := range config.TestCases {
		// Determine k: CLI override takes precedence over YAML, with 5 as default
		k := kOverride
		if k <= 0 {
			k = tc.K
			if k <= 0 {
				k = 5 // default
			}
		}

		testResult := TestResult{
			TestID:   tc.ID,
			Name:     tc.Name,
			Expected: tc.Expected,
			Runs:     make([]string, k),
		}

		fmt.Printf("  [%d/%d] Running test %s (k=%d)...\n", tcIdx+1, len(config.TestCases), tc.ID, k)

		// Run the test k times
		for i := 0; i < k; i++ {
			// Execute the agent and get the verdict
			verdict, err := executeAgent(config.Agent, tc.Input)
			if err != nil {
				// Log error but continue - mark as ERROR verdict
				fmt.Fprintf(os.Stderr, "Warning: Agent execution failed for %s (run %d/%d): %v\n", tc.ID, i+1, k, err)
				verdict = "ERROR"
			}
			testResult.Runs[i] = verdict
			fmt.Printf("    Run %d/%d: %s\n", i+1, k, verdict)
		}

		result.TestResults = append(result.TestResults, testResult)
	}

	return result, nil
}

// formatAgentPrompt formats a test case input into a structured prompt for the agent
func formatAgentPrompt(agentName string, input TaskInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Task Title: %s\n\n", input.TaskTitle))

	if input.TaskDescription != "" {
		sb.WriteString(fmt.Sprintf("Task Description: %s\n\n", input.TaskDescription))
	}

	if len(input.AcceptanceCriteria) > 0 {
		sb.WriteString("Acceptance Criteria:\n")
		for _, criterion := range input.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("- %s\n", criterion))
		}
		sb.WriteString("\n")
	}

	if input.Implementation != "" {
		sb.WriteString("Implementation:\n")
		sb.WriteString(input.Implementation)
		sb.WriteString("\n")
	}

	return sb.String()
}

// extractVerdict extracts the verdict from agent output using flexible keyword matching
func extractVerdict(output string) string {
	output = strings.ToUpper(output)

	// Valid verdicts to search for
	verdicts := []string{"PASS", "FAIL", "REFINED", "NEEDS_INPUT", "SKIP"}

	// Keywords that might precede the verdict
	keywords := []string{"VERDICT:", "STATUS:", "RESULT:"}

	// First, try to find verdict after keywords
	for _, keyword := range keywords {
		if idx := strings.Index(output, keyword); idx >= 0 {
			// Look for verdict in the text following the keyword
			remaining := output[idx+len(keyword):]
			for _, verdict := range verdicts {
				// Look for verdict within first 100 chars after keyword
				if len(remaining) > 100 {
					remaining = remaining[:100]
				}
				if strings.Contains(remaining, verdict) {
					return verdict
				}
			}
		}
	}

	// If no keyword found, search for any verdict in the entire output
	for _, verdict := range verdicts {
		if strings.Contains(output, verdict) {
			return verdict
		}
	}

	// If no verdict found, return ERROR
	return "ERROR"
}

// executeAgent executes an agent via Claude CLI and returns the verdict
func executeAgent(agentName string, input TaskInput) (string, error) {
	// Format the prompt
	prompt := formatAgentPrompt(agentName, input)

	// Create context with 5-minute timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build command: claude --agent <name> --print
	cmd := exec.CommandContext(ctx, "claude", "--agent", agentName, "--print")

	// Pipe the prompt as stdin
	cmd.Stdin = bytes.NewBufferString(prompt)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "ERROR", fmt.Errorf("agent execution timed out after 5 minutes")
		}
		// Return ERROR verdict with message for other failures
		return "ERROR", fmt.Errorf("agent execution failed: %w (output: %s)", err, string(output))
	}

	// Extract verdict from output
	verdict := extractVerdict(string(output))

	return verdict, nil
}

// calculateMetrics calculates accuracy and consistency metrics from test results
func calculateMetrics(results []TestResult) Metrics {
	metrics := Metrics{
		TotalTests: len(results),
	}

	for _, tr := range results {
		// Check if correct (majority vote matches expected)
		verdict := getMajorityVerdict(tr.Runs)
		if verdict == tr.Expected {
			metrics.CorrectCount++
		}

		// Check if consistent (all runs agree)
		if areAllRunsConsistent(tr.Runs) {
			metrics.ConsistentCount++
		}
	}

	// Calculate percentages
	if metrics.TotalTests > 0 {
		metrics.Accuracy = float64(metrics.CorrectCount) / float64(metrics.TotalTests)
		metrics.Consistency = float64(metrics.ConsistentCount) / float64(metrics.TotalTests)
	}

	return metrics
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

// formatMetaReport formats the evaluation result into a readable report
func formatMetaReport(result EvaluationResult) string {
	var sb strings.Builder

	sb.WriteString("Meta-Evaluation Report\n")
	sb.WriteString("======================\n\n")

	sb.WriteString(fmt.Sprintf("Agent: %s\n", result.Agent))
	sb.WriteString(fmt.Sprintf("Test Cases: %d\n\n", len(result.TestResults)))

	// Calculate metrics once
	metrics := calculateMetrics(result.TestResults)

	sb.WriteString("Results:\n")
	for _, tr := range result.TestResults {
		// Get majority verdict (calculate once)
		verdict := getMajorityVerdict(tr.Runs)

		// Calculate consistent count
		consistentCount := 0
		if areAllRunsConsistent(tr.Runs) {
			consistentCount = len(tr.Runs)
		} else {
			// Count how many agree with majority
			for _, v := range tr.Runs {
				if v == verdict {
					consistentCount++
				}
			}
		}

		status := "PASS"
		if verdict != tr.Expected {
			status = fmt.Sprintf("FAIL (expected %s, got %s)", tr.Expected, verdict)
		}

		sb.WriteString(fmt.Sprintf("  %s: %s (%d/%d consistent)\n",
			tr.TestID, status, consistentCount, len(tr.Runs)))
	}

	sb.WriteString("\nMetrics:\n")
	sb.WriteString(fmt.Sprintf("  Accuracy: %.1f%% (%d/%d correct)\n",
		metrics.Accuracy*100, metrics.CorrectCount, metrics.TotalTests))
	sb.WriteString(fmt.Sprintf("  Consistency (pass^k): %.1f%% (%d/%d all runs agree)\n",
		metrics.Consistency*100, metrics.ConsistentCount, metrics.TotalTests))

	return sb.String()
}

// confirmMetaExecution estimates API calls and prompts for confirmation if needed
func confirmMetaExecution(evalFiles []string, k int, confirm bool) error {
	// Calculate total API calls estimate
	totalTests := 0
	for _, evalPath := range evalFiles {
		config, err := loadEvalYAML(evalPath)
		if err != nil {
			return fmt.Errorf("loading eval file %s: %w", evalPath, err)
		}

		for _, tc := range config.TestCases {
			// Determine k for this test case
			testK := k
			if testK <= 0 {
				testK = tc.K
				if testK <= 0 {
					testK = 5 // default
				}
			}
			totalTests += testK
		}
	}

	// Show estimate
	fmt.Printf("\nMeta-Evaluation Estimate:\n")
	fmt.Printf("  Eval files: %d\n", len(evalFiles))
	fmt.Printf("  Total API calls: %d\n", totalTests)
	fmt.Printf("  Estimated cost: ~$%.2f (assuming $0.015 per call)\n\n", float64(totalTests)*0.015)

	// If --confirm flag is set, skip prompt
	if confirm {
		return nil
	}

	// Prompt user for confirmation
	fmt.Print("Proceed with meta-evaluation? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		return fmt.Errorf("meta-evaluation cancelled by user")
	}

	return nil
}

// runMetaCommand executes the meta CLI command
func runMetaCommand(suite, agent string, k int, metaDir string, confirm bool) error {
	var evalFiles []string
	var err error

	if agent != "" {
		// Run specific agent
		evalPath := filepath.Join(metaDir, "agents", agent, "eval.yaml")
		if _, err := os.Stat(evalPath); os.IsNotExist(err) {
			return fmt.Errorf("eval.yaml not found for agent: %s", agent)
		}
		evalFiles = []string{evalPath}
	} else if suite != "" {
		// Run entire suite
		suiteDir := filepath.Join(metaDir, suite)
		if _, err := os.Stat(suiteDir); os.IsNotExist(err) {
			return fmt.Errorf("suite directory not found: %s", suite)
		}

		if suite == "agents" {
			evalFiles, err = findAgentEvalFiles(suiteDir)
		} else if suite == "skills" {
			evalFiles, err = findSkillEvalFiles(suiteDir)
		} else {
			return fmt.Errorf("invalid suite: %s (must be 'agents' or 'skills')", suite)
		}

		if err != nil {
			return fmt.Errorf("finding eval files: %w", err)
		}

		if len(evalFiles) == 0 {
			return fmt.Errorf("no eval.yaml files found in %s suite", suite)
		}
	} else {
		return fmt.Errorf("must specify either --suite or --agent")
	}

	// Cost safeguard: estimate API calls and prompt for confirmation
	if err := confirmMetaExecution(evalFiles, k, confirm); err != nil {
		return err
	}

	// Run evaluation for each file
	for _, evalPath := range evalFiles {
		fmt.Printf("\nRunning evaluation: %s\n", evalPath)
		fmt.Println(strings.Repeat("=", 60))

		result, err := runMetaEvaluation(evalPath, k)
		if err != nil {
			return fmt.Errorf("running evaluation for %s: %w", evalPath, err)
		}

		report := formatMetaReport(result)
		fmt.Println(report)
	}

	return nil
}
