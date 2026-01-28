package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// CriteriaScore represents the average score for a specific criteria across all skills
type CriteriaScore struct {
	Name    string
	Average float64
}

// GradeReport represents parsed data from a skill-clarity report
type GradeReport struct {
	FilePath         string
	GeneratedDate    string
	TotalSkills      int
	AverageScore     float64
	PassRate         float64
	PassingThreshold float64
	CriteriaScores   []CriteriaScore
}

// findGradeReports finds all skill-clarity-*.md reports in the given directory
// Returns reports sorted by date (newest first)
func findGradeReports(reportsDir string) ([]string, error) {
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("reading reports directory: %w", err)
	}

	var reports []string
	pattern := regexp.MustCompile(`^skill-clarity-\d{4}-\d{2}-\d{2}\.md$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if pattern.MatchString(entry.Name()) {
			reports = append(reports, filepath.Join(reportsDir, entry.Name()))
		}
	}

	// Sort by filename (which includes date) in descending order (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return filepath.Base(reports[i]) > filepath.Base(reports[j])
	})

	return reports, nil
}

// parseGradeReport parses a skill-clarity report and extracts key metrics
func parseGradeReport(reportPath string) (GradeReport, error) {
	content, err := os.ReadFile(reportPath)
	if err != nil {
		return GradeReport{}, fmt.Errorf("reading report: %w", err)
	}

	report := GradeReport{
		FilePath: reportPath,
	}

	lines := strings.Split(string(content), "\n")

	// Regex patterns to extract metrics
	generatedPattern := regexp.MustCompile(`Generated:\s*(.+)`)
	totalSkillsPattern := regexp.MustCompile(`\*\*Total Skills\*\*:\s*(\d+)`)
	averageScorePattern := regexp.MustCompile(`\*\*Average Score\*\*:\s*([\d.]+)/100`)
	passRatePattern := regexp.MustCompile(`\*\*Pass Rate\*\*:\s*([\d.]+)%`)
	passingThresholdPattern := regexp.MustCompile(`\*\*Passing Threshold\*\*:\s*([\d.]+)`)

	for _, line := range lines {
		// Extract GeneratedDate
		if matches := generatedPattern.FindStringSubmatch(line); matches != nil {
			report.GeneratedDate = strings.TrimSpace(matches[1])
		}

		// Extract TotalSkills
		if matches := totalSkillsPattern.FindStringSubmatch(line); matches != nil {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				report.TotalSkills = val
			}
		}

		// Extract AverageScore
		if matches := averageScorePattern.FindStringSubmatch(line); matches != nil {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
				report.AverageScore = val
			}
		}

		// Extract PassRate
		if matches := passRatePattern.FindStringSubmatch(line); matches != nil {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
				report.PassRate = val
			}
		}

		// Extract PassingThreshold
		if matches := passingThresholdPattern.FindStringSubmatch(line); matches != nil {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
				report.PassingThreshold = val
			}
		}
	}

	// Extract per-criteria scores from Detailed Breakdown section
	report.CriteriaScores = extractCriteriaScores(lines)

	return report, nil
}

// extractCriteriaScores parses the Detailed Breakdown section and aggregates per-criteria scores
func extractCriteriaScores(lines []string) []CriteriaScore {
	// Map to accumulate scores for each criteria
	criteriaMap := make(map[string][]float64)

	// Regex pattern to match criteria score lines like:
	// - **Clear Instructions** (weight: 30%): 75.0/100
	criteriaPattern := regexp.MustCompile(`^\s*-\s*\*\*([^*]+)\*\*\s*\(weight:[^)]+\):\s*([\d.]+)/100`)

	inDetailedBreakdown := false

	for _, line := range lines {
		// Check if we're in the Detailed Breakdown section
		if strings.Contains(line, "## Detailed Breakdown") {
			inDetailedBreakdown = true
			continue
		}

		// Stop if we reach another second-level section (##) after Detailed Breakdown
		// Note: Don't stop at third-level sections (###) which are skill names
		if inDetailedBreakdown && strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") && !strings.Contains(line, "Detailed Breakdown") {
			break
		}

		// Extract criteria scores
		if inDetailedBreakdown {
			if matches := criteriaPattern.FindStringSubmatch(line); matches != nil {
				criteriaName := strings.TrimSpace(matches[1])
				score, err := strconv.ParseFloat(matches[2], 64)
				if err == nil {
					criteriaMap[criteriaName] = append(criteriaMap[criteriaName], score)
				}
			}
		}
	}

	// Calculate averages and create result slice
	var result []CriteriaScore

	// Define the expected criteria order
	criteriaOrder := []string{
		"Clear Instructions",
		"Actionable Steps",
		"Good Examples",
		"Appropriate Scope",
	}

	for _, criteriaName := range criteriaOrder {
		if scores, exists := criteriaMap[criteriaName]; exists && len(scores) > 0 {
			sum := 0.0
			for _, score := range scores {
				sum += score
			}
			average := sum / float64(len(scores))

			// Round to 1 decimal place
			average = float64(int(average*10+0.5)) / 10

			result = append(result, CriteriaScore{
				Name:    criteriaName,
				Average: average,
			})
		}
	}

	return result
}

// formatReportSummaryMarkdown formats a GradeReport as markdown
func formatReportSummaryMarkdown(report GradeReport, trends *GradeTrends, enableTrends bool) string {
	var sb strings.Builder

	sb.WriteString("# Evaluation Report Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Report Type**: grade\n"))
	sb.WriteString(fmt.Sprintf("**Report**: %s\n", filepath.Base(report.FilePath)))
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", report.GeneratedDate))

	sb.WriteString("## Current Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Skills**: %d\n", report.TotalSkills))
	sb.WriteString(fmt.Sprintf("- **Average Score**: %.1f/100\n", report.AverageScore))
	sb.WriteString(fmt.Sprintf("- **Pass Rate**: %.1f%%\n", report.PassRate))
	sb.WriteString(fmt.Sprintf("- **Passing Threshold**: %.1f/100\n", report.PassingThreshold))

	// Add trend analysis if enabled and available
	if enableTrends && trends != nil {
		sb.WriteString("\n## Trend Analysis\n\n")
		sb.WriteString("| Metric | Previous | Current | Change | Status |\n")
		sb.WriteString("|--------|----------|---------|--------|--------|\n")

		// Threshold for warnings (5% regression)
		threshold := 5.0

		// Average Score trend
		exceeds := exceedsRegressionThreshold(trends.AverageScore, threshold)
		sb.WriteString(formatTrendMarkdown("Average Score", trends.AverageScore, exceeds))
		sb.WriteString("\n")

		// Pass Rate trend
		exceeds = exceedsRegressionThreshold(trends.PassRate, threshold)
		sb.WriteString(formatTrendMarkdown("Pass Rate", trends.PassRate, exceeds))
		sb.WriteString("\n")

		// Total Skills trend
		exceeds = exceedsRegressionThreshold(trends.TotalSkills, threshold)
		sb.WriteString(formatTrendMarkdown("Total Skills", trends.TotalSkills, exceeds))
		sb.WriteString("\n")

		// Per-criteria trends if available
		if len(trends.PerCriteriaTrends) > 0 {
			sb.WriteString("\n### Per-Criteria Trends\n\n")
			sb.WriteString("| Criteria | Previous | Current | Change | Status |\n")
			sb.WriteString("|----------|----------|---------|--------|--------|\n")

			// Define criteria order for consistent display
			criteriaOrder := []string{
				"Clear Instructions",
				"Actionable Steps",
				"Good Examples",
				"Appropriate Scope",
			}

			for _, name := range criteriaOrder {
				if trend, exists := trends.PerCriteriaTrends[name]; exists {
					exceeds := exceedsRegressionThreshold(trend, threshold)
					sb.WriteString(formatTrendMarkdown(name, trend, exceeds))
					sb.WriteString("\n")
				}
			}
		}
	}

	// Add per-category breakdown if available
	if len(report.CriteriaScores) > 0 {
		sb.WriteString("\n## Per-Category Breakdown\n\n")
		sb.WriteString("| Criteria | Average Score |\n")
		sb.WriteString("|----------|---------------|\n")

		for _, criteria := range report.CriteriaScores {
			sb.WriteString(fmt.Sprintf("| %s | %.1f |\n", criteria.Name, criteria.Average))
		}
	}

	return sb.String()
}

// formatReportSummaryJSON formats a GradeReport as JSON
func formatReportSummaryJSON(report GradeReport, trends *GradeTrends, enableTrends bool) (string, error) {
	// Convert CriteriaScores to JSON-friendly format
	criteriaScores := make([]map[string]interface{}, 0, len(report.CriteriaScores))
	for _, criteria := range report.CriteriaScores {
		criteriaScores = append(criteriaScores, map[string]interface{}{
			"name":    criteria.Name,
			"average": criteria.Average,
		})
	}

	data := map[string]interface{}{
		"file_path":         report.FilePath,
		"generated_date":    report.GeneratedDate,
		"total_skills":      report.TotalSkills,
		"average_score":     report.AverageScore,
		"pass_rate":         report.PassRate,
		"passing_threshold": report.PassingThreshold,
		"criteria_scores":   criteriaScores,
	}

	// Add trend data if enabled and available
	if enableTrends && trends != nil {
		trendData := map[string]interface{}{
			"average_score": trends.AverageScore,
			"pass_rate":     trends.PassRate,
			"total_skills":  trends.TotalSkills,
		}

		// Add per-criteria trends if available
		if len(trends.PerCriteriaTrends) > 0 {
			criteriaTrends := make(map[string]interface{})
			for name, trend := range trends.PerCriteriaTrends {
				criteriaTrends[name] = trend
			}
			trendData["per_criteria"] = criteriaTrends
		}

		data["trend"] = trendData
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// listGradeReports lists all available grade reports
func listGradeReports(reportsDir string) string {
	var sb strings.Builder

	sb.WriteString("# Grade Reports\n\n")

	reports, err := findGradeReports(reportsDir)
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error finding reports: %v\n", err))
		return sb.String()
	}

	if len(reports) == 0 {
		sb.WriteString("No grade reports found.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d report(s):\n\n", len(reports)))

	for i, reportPath := range reports {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, filepath.Base(reportPath)))
	}

	return sb.String()
}

// formatMetaReportMarkdown formats meta-evaluation results as markdown
func formatMetaReportMarkdown(results []ConsistencyResult, trends *MetaTrends, enableTrends bool) string {
	var sb strings.Builder

	sb.WriteString("# Meta-Evaluation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Report Type**: meta\n"))

	// Get the timestamp from the latest result
	if len(results) > 0 {
		sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", results[len(results)-1].Timestamp))
	} else {
		sb.WriteString("\n")
	}

	// Group results by agent to get the latest for each
	latestByAgent := make(map[string]ConsistencyResult)
	for _, result := range results {
		// Keep the last occurrence for each agent
		latestByAgent[result.Agent] = result
	}

	// Display current metrics
	sb.WriteString("## Current Metrics\n\n")
	sb.WriteString("| Agent | Consistency | Runs |\n")
	sb.WriteString("|-------|-------------|------|\n")

	// Sort agents for consistent output
	var agents []string
	for agent := range latestByAgent {
		agents = append(agents, agent)
	}
	sort.Strings(agents)

	for _, agent := range agents {
		result := latestByAgent[agent]
		sb.WriteString(fmt.Sprintf("| %s | %.1f%% | %d |\n",
			result.Agent,
			result.ConsistencyPercentage,
			result.TotalCount))
	}

	// Add trend analysis if enabled
	if enableTrends && trends != nil && len(trends.PerAgentTrends) > 0 {
		sb.WriteString("\n## Trend Analysis\n\n")
		sb.WriteString("| Agent | Previous | Current | Change | Status |\n")
		sb.WriteString("|-------|----------|---------|--------|--------|\n")

		for _, agent := range agents {
			if trend, exists := trends.PerAgentTrends[agent]; exists {
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

				// Format as percentage points (pp)
				sb.WriteString(fmt.Sprintf("| %s | %.1f%% | %.1f%% | %s%.1fpp | %s %s |\n",
					agent,
					trend.PreviousValue,
					trend.CurrentValue,
					deltaSign,
					trend.AbsoluteDelta,
					indicator,
					trend.Direction,
				))
			}
		}
	}

	return sb.String()
}

// formatMetaReportJSON formats meta-evaluation results as JSON
func formatMetaReportJSON(results []ConsistencyResult, trends *MetaTrends, enableTrends bool) (string, error) {
	// Group results by agent to get the latest for each
	latestByAgent := make(map[string]ConsistencyResult)
	for _, result := range results {
		latestByAgent[result.Agent] = result
	}

	// Create agents array
	var agents []map[string]interface{}
	for agent, result := range latestByAgent {
		agents = append(agents, map[string]interface{}{
			"agent":                 agent,
			"consistency_percentage": result.ConsistencyPercentage,
			"total_count":           result.TotalCount,
			"timestamp":             result.Timestamp,
		})
	}

	// Sort agents for consistent output
	sort.Slice(agents, func(i, j int) bool {
		return agents[i]["agent"].(string) < agents[j]["agent"].(string)
	})

	data := map[string]interface{}{
		"report_type": "meta",
		"agents":      agents,
	}

	// Get timestamp from latest result
	if len(results) > 0 {
		data["generated"] = results[len(results)-1].Timestamp
	}

	// Add trend data if enabled and available
	if enableTrends && trends != nil && len(trends.PerAgentTrends) > 0 {
		trendData := map[string]interface{}{
			"per_agent": trends.PerAgentTrends,
		}
		data["trend"] = trendData
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// formatEvalReportMarkdown formats eval results as markdown
func formatEvalReportMarkdown(results []GradeTaskOutput, trends *EvalTrends, enableTrends bool) string {
	var sb strings.Builder

	sb.WriteString("# Evaluation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Report Type**: eval\n"))

	// Find the latest timestamp
	latestTimestamp := ""
	for _, result := range results {
		if result.Timestamp > latestTimestamp {
			latestTimestamp = result.Timestamp
		}
	}

	if latestTimestamp != "" {
		sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", latestTimestamp))
	} else {
		sb.WriteString("\n")
	}

	// Calculate current metrics (only from latest timestamp)
	var latestResults []GradeTaskOutput
	for _, result := range results {
		if result.Timestamp == latestTimestamp {
			latestResults = append(latestResults, result)
		}
	}

	totalScore := 0.0
	passedCount := 0
	for _, result := range latestResults {
		totalScore += result.OverallScore
		if result.OverallPassed {
			passedCount++
		}
	}

	avgScore := 0.0
	if len(latestResults) > 0 {
		avgScore = totalScore / float64(len(latestResults))
	}

	passRate := 0.0
	if len(latestResults) > 0 {
		passRate = (float64(passedCount) / float64(len(latestResults))) * 100.0
	}

	// Display current metrics
	sb.WriteString("## Current Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- **Average Score**: %.1f/100\n", avgScore))
	sb.WriteString(fmt.Sprintf("- **Pass Rate**: %.1f%% (%d/%d tasks)\n", passRate, passedCount, len(latestResults)))

	// Add trend analysis if enabled
	if enableTrends && trends != nil {
		sb.WriteString("\n## Trend Analysis\n\n")

		// Aggregate metrics trends
		sb.WriteString("### Aggregate Metrics\n\n")
		sb.WriteString("| Metric | Previous | Current | Change | Status |\n")
		sb.WriteString("|--------|----------|---------|--------|--------|\n")

		// Threshold for warnings (5% regression)
		threshold := 5.0

		// Average Score trend
		exceeds := exceedsRegressionThreshold(trends.AverageScore, threshold)
		sb.WriteString(formatTrendMarkdown("Average Score", trends.AverageScore, exceeds))
		sb.WriteString("\n")

		// Pass Rate trend
		exceeds = exceedsRegressionThreshold(trends.PassRate, threshold)
		sb.WriteString(formatTrendMarkdown("Pass Rate", trends.PassRate, exceeds))
		sb.WriteString("\n")

		// Per-task trends if available
		if len(trends.PerTaskTrends) > 0 {
			sb.WriteString("\n### Per-Task Trends\n\n")
			sb.WriteString("| Task | Previous | Current | Change | Status |\n")
			sb.WriteString("|------|----------|---------|--------|--------|\n")

			// Sort task IDs for consistent display
			var taskIDs []string
			for taskID := range trends.PerTaskTrends {
				taskIDs = append(taskIDs, taskID)
			}
			sort.Strings(taskIDs)

			for _, taskID := range taskIDs {
				trend := trends.PerTaskTrends[taskID]
				exceeds := exceedsRegressionThreshold(trend, threshold)
				sb.WriteString(formatTrendMarkdown(taskID, trend, exceeds))
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// formatEvalReportJSON formats eval results as JSON
func formatEvalReportJSON(results []GradeTaskOutput, trends *EvalTrends, enableTrends bool) (string, error) {
	// Find the latest timestamp
	latestTimestamp := ""
	for _, result := range results {
		if result.Timestamp > latestTimestamp {
			latestTimestamp = result.Timestamp
		}
	}

	// Calculate current metrics (only from latest timestamp)
	var latestResults []GradeTaskOutput
	for _, result := range results {
		if result.Timestamp == latestTimestamp {
			latestResults = append(latestResults, result)
		}
	}

	totalScore := 0.0
	passedCount := 0
	for _, result := range latestResults {
		totalScore += result.OverallScore
		if result.OverallPassed {
			passedCount++
		}
	}

	avgScore := 0.0
	if len(latestResults) > 0 {
		avgScore = totalScore / float64(len(latestResults))
	}

	passRate := 0.0
	if len(latestResults) > 0 {
		passRate = (float64(passedCount) / float64(len(latestResults))) * 100.0
	}

	data := map[string]interface{}{
		"report_type":   "eval",
		"average_score": avgScore,
		"pass_rate":     passRate,
		"total_tasks":   len(latestResults),
		"passed_tasks":  passedCount,
	}

	// Get timestamp from latest result
	if latestTimestamp != "" {
		data["generated"] = latestTimestamp
	}

	// Add trend data if enabled and available
	if enableTrends && trends != nil {
		trendData := map[string]interface{}{
			"average_score": trends.AverageScore,
			"pass_rate":     trends.PassRate,
		}

		// Add per-task trends if available
		if len(trends.PerTaskTrends) > 0 {
			trendData["per_task"] = trends.PerTaskTrends
		}

		data["trend"] = trendData
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// runReportCommand executes the report CLI command
func runReportCommand(reportType, format string, listMode bool, outputPath, reportsDir string, enableTrends bool) error {
	// List mode: just list available reports
	if listMode {
		output := listGradeReports(reportsDir)

		if outputPath != "" {
			// Write to file
			if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
				return fmt.Errorf("writing output file: %w", err)
			}
			fmt.Printf("Report list written to: %s\n", outputPath)
		} else {
			// Write to stdout
			fmt.Print(output)
		}

		return nil
	}

	// Handle different report types
	var output string
	switch reportType {
	case "grade":
		// Find reports
		reports, err := findGradeReports(reportsDir)
		if err != nil {
			return fmt.Errorf("finding grade reports: %w", err)
		}

		if len(reports) == 0 {
			return fmt.Errorf("no grade reports found in %s", reportsDir)
		}

		// Get the latest report
		latestReportPath := reports[0]

		// Parse the report
		report, err := parseGradeReport(latestReportPath)
		if err != nil {
			return fmt.Errorf("parsing report: %w", err)
		}

		// Load trend data if enabled
		var trends *GradeTrends
		if enableTrends {
			trends, err = loadGradeTrends(reportsDir)
			if err != nil {
				// Don't fail if trends can't be loaded, just disable them
				fmt.Fprintf(os.Stderr, "Warning: Could not load trend data: %v\n", err)
				trends = nil
			}
		}

		// Format the output
		switch format {
		case "json":
			jsonOutput, err := formatReportSummaryJSON(report, trends, enableTrends)
			if err != nil {
				return fmt.Errorf("formatting as JSON: %w", err)
			}
			output = jsonOutput
		case "markdown":
			output = formatReportSummaryMarkdown(report, trends, enableTrends)
		default:
			return fmt.Errorf("unsupported format: %s (use 'markdown' or 'json')", format)
		}

	case "meta":
		// Load meta results from consistency-log.json
		metaLogPath := filepath.Join(reportsDir, "consistency-log.json")
		results, err := loadMetaResults(metaLogPath)
		if err != nil {
			return fmt.Errorf("loading meta results: %w", err)
		}

		if len(results) == 0 {
			return fmt.Errorf("no meta results found in %s", metaLogPath)
		}

		// Load trend data if enabled
		var metaTrends *MetaTrends
		if enableTrends {
			metaTrends, err = loadMetaTrends(metaLogPath)
			if err != nil {
				// Don't fail if trends can't be loaded, just disable them
				fmt.Fprintf(os.Stderr, "Warning: Could not load trend data: %v\n", err)
				metaTrends = nil
			}
		}

		// Format the output
		switch format {
		case "json":
			jsonOutput, err := formatMetaReportJSON(results, metaTrends, enableTrends)
			if err != nil {
				return fmt.Errorf("formatting as JSON: %w", err)
			}
			output = jsonOutput
		case "markdown":
			output = formatMetaReportMarkdown(results, metaTrends, enableTrends)
		default:
			return fmt.Errorf("unsupported format: %s (use 'markdown' or 'json')", format)
		}

	case "eval":
		// Load eval results from task-eval-log.json
		evalLogPath := filepath.Join(reportsDir, "task-eval-log.json")
		results, err := loadEvalResults(evalLogPath)
		if err != nil {
			return fmt.Errorf("loading eval results: %w", err)
		}

		if len(results) == 0 {
			return fmt.Errorf("no eval results found in %s", evalLogPath)
		}

		// Load trend data if enabled
		var evalTrends *EvalTrends
		if enableTrends {
			evalTrends, err = loadEvalTrends(evalLogPath)
			if err != nil {
				// Don't fail if trends can't be loaded, just disable them
				fmt.Fprintf(os.Stderr, "Warning: Could not load trend data: %v\n", err)
				evalTrends = nil
			}
		}

		// Format the output
		switch format {
		case "json":
			jsonOutput, err := formatEvalReportJSON(results, evalTrends, enableTrends)
			if err != nil {
				return fmt.Errorf("formatting as JSON: %w", err)
			}
			output = jsonOutput
		case "markdown":
			output = formatEvalReportMarkdown(results, evalTrends, enableTrends)
		default:
			return fmt.Errorf("unsupported format: %s (use 'markdown' or 'json')", format)
		}

	default:
		return fmt.Errorf("report type '%s' not supported (use 'grade', 'meta', or 'eval')", reportType)
	}

	// Write output
	if outputPath != "" {
		// Write to file
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Printf("Report written to: %s\n", outputPath)
	} else {
		// Write to stdout
		fmt.Print(output)
	}

	return nil
}
