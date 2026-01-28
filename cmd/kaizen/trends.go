package main

import (
	"fmt"
	"math"
	"sort"
)

// TrendData represents comparison data between previous and current metrics
type TrendData struct {
	PreviousValue   float64 `json:"previous_value"`
	CurrentValue    float64 `json:"current_value"`
	AbsoluteDelta   float64 `json:"absolute_delta"`
	PercentageDelta float64 `json:"percentage_delta"`
	Direction       string  `json:"direction"` // "improvement", "regression", "stable"
}

// calculateDelta calculates the delta between previous and current values
func calculateDelta(previous, current float64) TrendData {
	absoluteDelta := current - previous

	// Calculate percentage change
	var percentageDelta float64
	if previous != 0 {
		percentageDelta = (absoluteDelta / previous) * 100
		// Round to 2 decimal places
		percentageDelta = math.Round(percentageDelta*100) / 100
	}

	// Determine direction
	direction := "stable"
	if absoluteDelta > 0.01 {
		direction = "improvement"
	} else if absoluteDelta < -0.01 {
		direction = "regression"
	}

	return TrendData{
		PreviousValue:   previous,
		CurrentValue:    current,
		AbsoluteDelta:   absoluteDelta,
		PercentageDelta: percentageDelta,
		Direction:       direction,
	}
}

// formatTrendMarkdown formats a trend comparison as a markdown table row
func formatTrendMarkdown(metricName string, trend TrendData, exceedsThreshold bool) string {
	// Format the direction indicator
	indicator := "→"
	if trend.Direction == "improvement" {
		indicator = "↑"
	} else if trend.Direction == "regression" {
		indicator = "↓"
	}

	// Format the delta values
	deltaSign := ""
	if trend.AbsoluteDelta > 0 {
		deltaSign = "+"
	}

	// Add warning indicator if regression exceeds threshold
	warning := ""
	if exceedsThreshold {
		warning = " ⚠"
	}

	// Format percentage delta
	percentStr := fmt.Sprintf("%s%.2f%%", deltaSign, trend.PercentageDelta)

	// Return formatted row
	return fmt.Sprintf("| %s | %.1f | %.1f | %s%.1f (%s) | %s %s%s |",
		metricName,
		trend.PreviousValue,
		trend.CurrentValue,
		deltaSign,
		trend.AbsoluteDelta,
		percentStr,
		indicator,
		trend.Direction,
		warning,
	)
}

// formatNoTrendData returns a message indicating no trend data is available
func formatNoTrendData() string {
	return "No trend data available (first run or insufficient historical data)"
}

// EvalTrends represents trend data for eval report
type EvalTrends struct {
	AverageScore TrendData
	PassRate     TrendData
	PerTaskTrends map[string]TrendData // task_id -> trend
}

// MetaTrends represents trend data for meta report
type MetaTrends struct {
	ConsistencyPercentage TrendData
	RunCount              TrendData
	PerAgentTrends        map[string]TrendData // agent -> trend
}

// GradeTrends represents trend data for grade report
type GradeTrends struct {
	TotalSkills      TrendData
	AverageScore     TrendData
	PassRate         TrendData
	PerCriteriaTrends map[string]TrendData // criteria name -> trend
}

// loadEvalTrends loads trend data from task-eval-log.json
func loadEvalTrends(logPath string) (*EvalTrends, error) {
	results, err := loadEvalResults(logPath)
	if err != nil {
		return nil, fmt.Errorf("loading eval results: %w", err)
	}

	if len(results) < 2 {
		return nil, fmt.Errorf("insufficient data for trend analysis (need at least 2 entries)")
	}

	// Group results by task_id and find the two most recent timestamps
	taskGroups := make(map[string][]GradeTaskOutput)
	timestamps := make(map[string]bool)

	for _, result := range results {
		taskGroups[result.TaskID] = append(taskGroups[result.TaskID], result)
		timestamps[result.Timestamp] = true
	}

	// Get unique timestamps and sort them
	var timeList []string
	for ts := range timestamps {
		timeList = append(timeList, ts)
	}
	sort.Strings(timeList)

	if len(timeList) < 2 {
		return nil, fmt.Errorf("insufficient data for trend analysis (need at least 2 unique timestamps)")
	}

	// Get the two most recent timestamps
	prevTimestamp := timeList[len(timeList)-2]
	currTimestamp := timeList[len(timeList)-1]

	// Calculate aggregate metrics for each timestamp
	var prevScores, currScores []float64
	var prevPassed, currPassed int
	var prevTotal, currTotal int

	trends := &EvalTrends{
		PerTaskTrends: make(map[string]TrendData),
	}

	// Process each task
	for taskID, taskResults := range taskGroups {
		var prevResult, currResult *GradeTaskOutput

		// Find results for the two timestamps
		for i := range taskResults {
			if taskResults[i].Timestamp == prevTimestamp {
				prevResult = &taskResults[i]
			}
			if taskResults[i].Timestamp == currTimestamp {
				currResult = &taskResults[i]
			}
		}

		// Calculate per-task trend if both timestamps exist
		if prevResult != nil && currResult != nil {
			trends.PerTaskTrends[taskID] = calculateDelta(prevResult.OverallScore, currResult.OverallScore)
		}

		// Aggregate for overall metrics
		if prevResult != nil {
			prevScores = append(prevScores, prevResult.OverallScore)
			prevTotal++
			if prevResult.OverallPassed {
				prevPassed++
			}
		}
		if currResult != nil {
			currScores = append(currScores, currResult.OverallScore)
			currTotal++
			if currResult.OverallPassed {
				currPassed++
			}
		}
	}

	// Calculate aggregate average score
	prevAvg := 0.0
	if len(prevScores) > 0 {
		sum := 0.0
		for _, score := range prevScores {
			sum += score
		}
		prevAvg = sum / float64(len(prevScores))
	}

	currAvg := 0.0
	if len(currScores) > 0 {
		sum := 0.0
		for _, score := range currScores {
			sum += score
		}
		currAvg = sum / float64(len(currScores))
	}

	trends.AverageScore = calculateDelta(prevAvg, currAvg)

	// Calculate aggregate pass rate
	prevPassRate := 0.0
	if prevTotal > 0 {
		prevPassRate = (float64(prevPassed) / float64(prevTotal)) * 100.0
	}

	currPassRate := 0.0
	if currTotal > 0 {
		currPassRate = (float64(currPassed) / float64(currTotal)) * 100.0
	}

	trends.PassRate = calculateDelta(prevPassRate, currPassRate)

	return trends, nil
}

// loadMetaTrends loads trend data from consistency-log.json
func loadMetaTrends(logPath string) (*MetaTrends, error) {
	results, err := loadMetaResults(logPath)
	if err != nil {
		return nil, fmt.Errorf("loading meta results: %w", err)
	}

	if len(results) < 2 {
		return nil, fmt.Errorf("insufficient data for trend analysis (need at least 2 entries)")
	}

	// Get last two entries for overall trend
	previous := results[len(results)-2]
	current := results[len(results)-1]

	trends := &MetaTrends{
		ConsistencyPercentage: calculateDelta(previous.ConsistencyPercentage, current.ConsistencyPercentage),
		RunCount: calculateDelta(float64(previous.TotalCount), float64(current.TotalCount)),
		PerAgentTrends: make(map[string]TrendData),
	}

	// Group results by agent and calculate per-agent trends
	agentResults := make(map[string][]ConsistencyResult)
	for _, result := range results {
		agentResults[result.Agent] = append(agentResults[result.Agent], result)
	}

	// Calculate trend for each agent (compare last two entries)
	for agent, agentData := range agentResults {
		if len(agentData) >= 2 {
			// Get the last two entries for this agent
			prevIdx := len(agentData) - 2
			currIdx := len(agentData) - 1
			prevResult := agentData[prevIdx]
			currResult := agentData[currIdx]

			trends.PerAgentTrends[agent] = calculateDelta(
				prevResult.ConsistencyPercentage,
				currResult.ConsistencyPercentage,
			)
		}
	}

	return trends, nil
}

// loadGradeTrends loads trend data from skill-clarity reports
func loadGradeTrends(reportsDir string) (*GradeTrends, error) {
	// Find all grade reports
	reports, err := findGradeReports(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("finding grade reports: %w", err)
	}

	if len(reports) < 2 {
		return nil, fmt.Errorf("insufficient data for trend analysis (need at least 2 reports)")
	}

	// Get the two most recent reports
	currentReportPath := reports[0]
	previousReportPath := reports[1]

	// Parse both reports
	currentReport, err := parseGradeReport(currentReportPath)
	if err != nil {
		return nil, fmt.Errorf("parsing current report: %w", err)
	}

	previousReport, err := parseGradeReport(previousReportPath)
	if err != nil {
		return nil, fmt.Errorf("parsing previous report: %w", err)
	}

	// Calculate trends
	trends := &GradeTrends{
		TotalSkills:  calculateDelta(float64(previousReport.TotalSkills), float64(currentReport.TotalSkills)),
		AverageScore: calculateDelta(previousReport.AverageScore, currentReport.AverageScore),
		PassRate:     calculateDelta(previousReport.PassRate, currentReport.PassRate),
		PerCriteriaTrends: make(map[string]TrendData),
	}

	// Calculate per-criteria trends
	previousCriteriaMap := make(map[string]float64)
	for _, criteria := range previousReport.CriteriaScores {
		previousCriteriaMap[criteria.Name] = criteria.Average
	}

	for _, criteria := range currentReport.CriteriaScores {
		if previousScore, exists := previousCriteriaMap[criteria.Name]; exists {
			trends.PerCriteriaTrends[criteria.Name] = calculateDelta(previousScore, criteria.Average)
		}
	}

	return trends, nil
}

// loadThresholdConfig loads regression threshold from config
type ThresholdConfig struct {
	Regression struct {
		MaxDropPercent float64 `json:"max_drop_percent"`
	} `json:"regression"`
}

func loadThresholdConfig(configPath string) (float64, error) {
	// For now, return default threshold
	// TODO: Implement YAML parsing when needed
	return 5.0, nil
}

// exceedsRegressionThreshold checks if a regression exceeds the threshold
func exceedsRegressionThreshold(trend TrendData, threshold float64) bool {
	if trend.Direction != "regression" {
		return false
	}
	// Check if absolute percentage drop exceeds threshold
	return math.Abs(trend.PercentageDelta) > threshold
}
