package modelbased

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/srstomp/kaizen/internal/llm"
)

// TaskQualityGrader evaluates task specifications against quality criteria
type TaskQualityGrader struct {
	// Criteria weights for evaluation
	weights map[string]float64
	// Passing threshold (0-100)
	passingScore float64
	// LLM client for evaluation (optional, uses stub evaluation if nil)
	llmClient llm.Client
	// Timeout for LLM requests
	timeout time.Duration
}

// NewTaskQualityGrader creates a new task quality grader with default weights
func NewTaskQualityGrader() *TaskQualityGrader {
	return &TaskQualityGrader{
		weights: map[string]float64{
			"clarity":       0.30, // 30% - Is the task description clear and understandable?
			"acceptance":    0.30, // 30% - Are acceptance criteria well-defined?
			"scope":         0.25, // 25% - Is the scope unambiguous and achievable?
			"actionability": 0.15, // 15% - Can work begin immediately?
		},
		passingScore: 70.0,          // Default passing threshold
		llmClient:    nil,            // Will be set when LLM integration is needed
		timeout:      60 * time.Second, // Default timeout
	}
}

// Grade evaluates task content against quality criteria
func (g *TaskQualityGrader) Grade(input GradeInput) (Result, error) {
	// Stub implementation - will be replaced with LLM-based evaluation
	criteria := g.evaluateCriteria(input.Content)

	// Calculate weighted score
	totalScore := 0.0
	for criterionName, criterionResult := range criteria {
		weight := g.weights[criterionName]
		totalScore += criterionResult.Score * weight
	}

	// Build detailed feedback
	details := make(map[string]any)
	for name, criterion := range criteria {
		details[name] = map[string]any{
			"score":    criterion.Score,
			"feedback": criterion.Feedback,
			"weight":   g.weights[name],
		}
	}

	// Generate summary message
	message := g.generateMessage(totalScore, criteria)

	return Result{
		Passed:  totalScore >= g.passingScore,
		Score:   totalScore,
		Message: message,
		Details: details,
	}, nil
}

// evaluateCriteria performs stub evaluation of each criterion
// TODO: Replace with LLM-based evaluation
func (g *TaskQualityGrader) evaluateCriteria(content string) map[string]Criterion {
	// Stub implementation using basic heuristics
	// This will be replaced with LLM calls in the future

	criteria := make(map[string]Criterion)
	contentLower := strings.ToLower(content)

	// Handle empty content
	if content == "" {
		criteria["clarity"] = Criterion{
			Score:    0.0,
			Feedback: "Stub evaluation: Empty task content",
		}
		criteria["acceptance"] = Criterion{
			Score:    0.0,
			Feedback: "Stub evaluation: Empty task content",
		}
		criteria["scope"] = Criterion{
			Score:    0.0,
			Feedback: "Stub evaluation: Empty task content",
		}
		criteria["actionability"] = Criterion{
			Score:    0.0,
			Feedback: "Stub evaluation: Empty task content",
		}
		return criteria
	}

	// Clarity - check for descriptive content
	clarityScore := 50.0 // default neutral score
	clarityFeedback := "Stub evaluation: Task clarity not yet evaluated by LLM"

	// Check for vague keywords that indicate poor clarity
	vagueKeywords := []string{"investigate", "explore", "look at", "check out", "figure out"}
	hasVagueKeywords := false
	for _, keyword := range vagueKeywords {
		if strings.Contains(contentLower, keyword) {
			hasVagueKeywords = true
			break
		}
	}

	if hasVagueKeywords {
		clarityScore = 30.0
		clarityFeedback = "Stub evaluation: Task contains vague keywords suggesting unclear requirements"
	} else if strings.Contains(contentLower, "description") || strings.Contains(contentLower, "##") {
		clarityScore = 75.0
		clarityFeedback = "Stub evaluation: Task appears to have structured description"
	}

	criteria["clarity"] = Criterion{
		Score:    clarityScore,
		Feedback: clarityFeedback,
	}

	// Acceptance Criteria - check for explicit acceptance criteria
	acceptanceScore := 50.0
	acceptanceFeedback := "Stub evaluation: Acceptance criteria not yet evaluated by LLM"

	hasAcceptanceCriteria := strings.Contains(contentLower, "acceptance criteria") ||
		strings.Contains(contentLower, "success criteria") ||
		strings.Contains(contentLower, "requirements:")

	// Count numbered items (potential acceptance criteria)
	numberedItems := strings.Count(content, "1.") + strings.Count(content, "2.") + strings.Count(content, "3.")

	if hasAcceptanceCriteria && numberedItems >= 3 {
		acceptanceScore = 85.0
		acceptanceFeedback = "Stub evaluation: Found explicit acceptance criteria with multiple items"
	} else if hasAcceptanceCriteria {
		acceptanceScore = 70.0
		acceptanceFeedback = "Stub evaluation: Found acceptance criteria section"
	} else if numberedItems >= 3 {
		acceptanceScore = 60.0
		acceptanceFeedback = "Stub evaluation: Found numbered items that may be criteria"
	} else {
		acceptanceScore = 25.0
		acceptanceFeedback = "Stub evaluation: No clear acceptance criteria found"
	}

	criteria["acceptance"] = Criterion{
		Score:    acceptanceScore,
		Feedback: acceptanceFeedback,
	}

	// Scope - check for clear, bounded scope
	scopeScore := 50.0
	scopeFeedback := "Stub evaluation: Scope clarity not yet evaluated by LLM"

	// Check for scope-related sections
	hasScope := strings.Contains(contentLower, "scope") ||
		strings.Contains(contentLower, "out of scope") ||
		strings.Contains(contentLower, "technical details")

	// Check for vague, open-ended language
	hasVagueScope := strings.Contains(contentLower, "various") ||
		strings.Contains(contentLower, "several") ||
		strings.Contains(contentLower, "improve") ||
		strings.Contains(contentLower, "better") ||
		strings.Contains(contentLower, "make it")

	contentLength := len(content)

	if hasVagueScope {
		scopeScore = 30.0
		scopeFeedback = "Stub evaluation: Task contains vague language suggesting unclear scope"
	} else if hasScope && contentLength > 200 {
		scopeScore = 80.0
		scopeFeedback = "Stub evaluation: Task has explicit scope definition"
	} else if contentLength > 100 && contentLength < 2000 {
		scopeScore = 65.0
		scopeFeedback = "Stub evaluation: Task has reasonable content length"
	} else if contentLength < 100 {
		scopeScore = 35.0
		scopeFeedback = "Stub evaluation: Task may be too brief to have clear scope"
	}

	criteria["scope"] = Criterion{
		Score:    scopeScore,
		Feedback: scopeFeedback,
	}

	// Actionability - can work begin immediately?
	actionabilityScore := 50.0
	actionabilityFeedback := "Stub evaluation: Actionability not yet evaluated by LLM"

	// Check for technical details, implementation notes, or specific instructions
	hasTechnicalDetails := strings.Contains(contentLower, "technical details") ||
		strings.Contains(contentLower, "implementation") ||
		strings.Contains(contentLower, "use") ||
		strings.Contains(contentLower, "create") ||
		strings.Contains(contentLower, "add") ||
		strings.Contains(contentLower, "build")

	// Tasks with vague keywords are not actionable
	if hasVagueKeywords {
		actionabilityScore = 20.0
		actionabilityFeedback = "Stub evaluation: Task has vague keywords, not immediately actionable"
	} else if hasTechnicalDetails && hasAcceptanceCriteria {
		actionabilityScore = 85.0
		actionabilityFeedback = "Stub evaluation: Task has technical details and clear criteria"
	} else if hasTechnicalDetails {
		actionabilityScore = 70.0
		actionabilityFeedback = "Stub evaluation: Task has some technical details"
	} else {
		actionabilityScore = 45.0
		actionabilityFeedback = "Stub evaluation: Task may need more details to be actionable"
	}

	criteria["actionability"] = Criterion{
		Score:    actionabilityScore,
		Feedback: actionabilityFeedback,
	}

	return criteria
}

// generateMessage creates a human-readable summary message
func (g *TaskQualityGrader) generateMessage(score float64, criteria map[string]Criterion) string {
	if score >= g.passingScore {
		return fmt.Sprintf("Task quality evaluation passed with score %.1f/100. Note: Using stub evaluation; LLM-based grading not yet implemented.", score)
	}

	// Find weakest criterion
	weakestName := ""
	weakestScore := 100.0
	for name, criterion := range criteria {
		if criterion.Score < weakestScore {
			weakestScore = criterion.Score
			weakestName = name
		}
	}

	// Provide actionable feedback based on weakest criterion
	suggestions := map[string]string{
		"clarity":       "Add a clear description explaining what needs to be done and why",
		"acceptance":    "Define specific, testable acceptance criteria",
		"scope":         "Clarify the scope boundaries - what's included and what's not",
		"actionability": "Provide technical details or implementation guidance",
	}

	suggestion := suggestions[weakestName]
	if suggestion == "" {
		suggestion = "Review and improve task specification"
	}

	return fmt.Sprintf("Task quality evaluation failed with score %.1f/100. Weakest area: %s (%.1f). Suggestion: %s. Note: Using stub evaluation; LLM-based grading not yet implemented.",
		score, weakestName, weakestScore, suggestion)
}

// buildPrompt constructs the LLM prompt for task quality evaluation
func (g *TaskQualityGrader) buildPrompt(taskContent string) string {
	return fmt.Sprintf(`You are evaluating the quality of a task specification.

## Task to Evaluate
%s

Evaluate the task against these four criteria:

1. CLARITY (30%% weight): Is the task description clear and understandable?
2. ACCEPTANCE (30%% weight): Are acceptance criteria well-defined and testable?
3. SCOPE (25%% weight): Is the scope unambiguous and achievable?
4. ACTIONABILITY (15%% weight): Can work begin immediately with the information provided?

For each criterion, provide:
- A score from 0-100
- Brief feedback explaining the score

Respond in this exact format:
CLARITY: <score 0-100>
CLARITY_FEEDBACK: <brief explanation>
ACCEPTANCE: <score 0-100>
ACCEPTANCE_FEEDBACK: <brief explanation>
SCOPE: <score 0-100>
SCOPE_FEEDBACK: <brief explanation>
ACTIONABILITY: <score 0-100>
ACTIONABILITY_FEEDBACK: <brief explanation>`, taskContent)
}

// parseResponse parses the LLM response to extract scores and feedback for each criterion
func (g *TaskQualityGrader) parseResponse(response string) (Result, error) {
	if strings.TrimSpace(response) == "" {
		return Result{}, errors.New("LLM returned empty response")
	}

	lines := strings.Split(response, "\n")

	// Maps to store parsed values
	scores := make(map[string]float64)
	feedbacks := make(map[string]string)

	criteriaNames := []string{"clarity", "acceptance", "scope", "actionability"}
	upperCriteriaNames := map[string]string{
		"CLARITY":       "clarity",
		"ACCEPTANCE":    "acceptance",
		"SCOPE":         "scope",
		"ACTIONABILITY": "actionability",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse scores
		for upper, lower := range upperCriteriaNames {
			if strings.HasPrefix(line, upper+":") {
				scoreStr := strings.TrimSpace(strings.TrimPrefix(line, upper+":"))
				score, err := strconv.ParseFloat(scoreStr, 64)
				if err != nil {
					return Result{}, fmt.Errorf("invalid %s score: %s (must be a number)", upper, scoreStr)
				}
				if score < 0 || score > 100 {
					return Result{}, fmt.Errorf("invalid %s score: %.1f (must be between 0 and 100)", upper, score)
				}
				scores[lower] = score
			}

			// Parse feedback
			feedbackPrefix := upper + "_FEEDBACK:"
			if strings.HasPrefix(line, feedbackPrefix) {
				feedback := strings.TrimSpace(strings.TrimPrefix(line, feedbackPrefix))
				feedbacks[lower] = feedback
			}
		}
	}

	// Validate all required fields are present
	for _, criterion := range criteriaNames {
		if _, exists := scores[criterion]; !exists {
			return Result{}, fmt.Errorf("missing %s score in LLM response", strings.ToUpper(criterion))
		}
		if _, exists := feedbacks[criterion]; !exists {
			return Result{}, fmt.Errorf("missing %s feedback in LLM response", strings.ToUpper(criterion))
		}
	}

	// Calculate weighted total score
	totalScore := 0.0
	for criterion, score := range scores {
		weight := g.weights[criterion]
		totalScore += score * weight
	}

	// Build details map
	details := make(map[string]any)
	for _, criterion := range criteriaNames {
		details[criterion] = map[string]any{
			"score":    scores[criterion],
			"feedback": feedbacks[criterion],
			"weight":   g.weights[criterion],
		}
	}

	passed := totalScore >= g.passingScore
	message := fmt.Sprintf("Task quality evaluation: %s (score: %.1f/100).",
		map[bool]string{true: "PASS", false: "FAIL"}[passed], totalScore)

	return Result{
		Passed:  passed,
		Score:   totalScore,
		Message: message,
		Details: details,
	}, nil
}

// gradeWithLLM performs LLM-based evaluation
func (g *TaskQualityGrader) gradeWithLLM(taskContent string) (Result, error) {
	if g.llmClient == nil {
		return Result{}, errors.New("LLM client not initialized")
	}

	// Build prompt
	prompt := g.buildPrompt(taskContent)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Call LLM
	response, err := g.llmClient.Complete(ctx, prompt, llm.WithModel("claude-haiku-4"))
	if err != nil {
		// Handle timeout specifically
		if errors.Is(err, context.DeadlineExceeded) {
			return Result{}, fmt.Errorf("LLM request timed out after %v: %w", g.timeout, err)
		}
		return Result{}, fmt.Errorf("LLM request failed: %w", err)
	}

	// Parse response
	return g.parseResponse(response)
}
