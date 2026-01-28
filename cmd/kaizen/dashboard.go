package main

import (
	"fmt"
	htmlpkg "html"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DashboardSummary contains overall summary statistics
type DashboardSummary struct {
	EvalTotalCount     int
	EvalPassCount      int
	EvalFailCount      int
	EvalAvgScore       float64
	MetaTotalCount     int
	MetaAvgConsistency float64
	GeneratedAt        string
}

// calculateDashboardSummary computes summary statistics from eval and meta results
func calculateDashboardSummary(evalResults []GradeTaskOutput, metaResults []ConsistencyResult) DashboardSummary {
	summary := DashboardSummary{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Calculate eval summary
	summary.EvalTotalCount = len(evalResults)
	totalScore := 0.0
	for _, result := range evalResults {
		if result.OverallPassed {
			summary.EvalPassCount++
		} else {
			summary.EvalFailCount++
		}
		totalScore += result.OverallScore
	}
	if summary.EvalTotalCount > 0 {
		summary.EvalAvgScore = totalScore / float64(summary.EvalTotalCount)
	}

	// Calculate meta summary
	summary.MetaTotalCount = len(metaResults)
	totalConsistency := 0.0
	for _, result := range metaResults {
		totalConsistency += result.ConsistencyPercentage
	}
	if summary.MetaTotalCount > 0 {
		summary.MetaAvgConsistency = totalConsistency / float64(summary.MetaTotalCount)
	}

	return summary
}

// generateDashboardHTML creates the HTML content for the dashboard
func generateDashboardHTML(evalResults []GradeTaskOutput, metaResults []ConsistencyResult) string {
	summary := calculateDashboardSummary(evalResults, metaResults)

	var html strings.Builder

	// HTML header
	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Yokay Evals Dashboard</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
      line-height: 1.6;
      color: #333;
      background: #f5f5f5;
      padding: 20px;
    }
    .container {
      max-width: 1200px;
      margin: 0 auto;
      background: white;
      padding: 30px;
      border-radius: 8px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    h1 {
      color: #2c3e50;
      margin-bottom: 10px;
      font-size: 2.5em;
    }
    .timestamp {
      color: #7f8c8d;
      font-size: 0.9em;
      margin-bottom: 30px;
    }
    h2 {
      color: #34495e;
      margin-top: 30px;
      margin-bottom: 15px;
      padding-bottom: 10px;
      border-bottom: 2px solid #3498db;
    }
    .summary {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
      gap: 20px;
      margin-bottom: 30px;
    }
    .summary-card {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      padding: 20px;
      border-radius: 8px;
      box-shadow: 0 4px 6px rgba(0,0,0,0.1);
    }
    .summary-card.success {
      background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
    }
    .summary-card.warning {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    }
    .summary-card h3 {
      font-size: 0.9em;
      opacity: 0.9;
      margin-bottom: 10px;
    }
    .summary-card .value {
      font-size: 2.5em;
      font-weight: bold;
      margin-bottom: 5px;
    }
    .summary-card .label {
      font-size: 0.85em;
      opacity: 0.8;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      margin-top: 15px;
      background: white;
    }
    thead {
      background: #34495e;
      color: white;
    }
    th, td {
      padding: 12px;
      text-align: left;
      border-bottom: 1px solid #ecf0f1;
    }
    th {
      font-weight: 600;
      text-transform: uppercase;
      font-size: 0.85em;
      letter-spacing: 0.5px;
    }
    tbody tr:hover {
      background: #f8f9fa;
    }
    .status {
      display: inline-block;
      padding: 4px 12px;
      border-radius: 12px;
      font-size: 0.85em;
      font-weight: 600;
    }
    .status.pass {
      background: #d4edda;
      color: #155724;
    }
    .status.fail {
      background: #f8d7da;
      color: #721c24;
    }
    .score {
      font-weight: 600;
      font-size: 1.1em;
    }
    .score.high {
      color: #27ae60;
    }
    .score.medium {
      color: #f39c12;
    }
    .score.low {
      color: #e74c3c;
    }
    .no-data {
      padding: 40px;
      text-align: center;
      color: #7f8c8d;
      font-style: italic;
    }
    .chart-container {
      margin-top: 20px;
      padding: 20px;
      background: #f8f9fa;
      border-radius: 8px;
    }
    .bar-chart {
      display: flex;
      align-items: flex-end;
      height: 200px;
      gap: 10px;
      margin-top: 20px;
    }
    .bar {
      flex: 1;
      background: linear-gradient(180deg, #3498db 0%, #2980b9 100%);
      border-radius: 4px 4px 0 0;
      position: relative;
      min-height: 20px;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      align-items: center;
      padding: 10px 5px;
      color: white;
      font-weight: 600;
      font-size: 0.85em;
    }
    .bar-label {
      text-align: center;
      margin-top: 8px;
      font-size: 0.85em;
      color: #555;
      font-weight: 500;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Yokay Evals Dashboard</h1>
    <div class="timestamp">Generated: ` + htmlpkg.EscapeString(summary.GeneratedAt) + `</div>
`)

	// Summary section
	html.WriteString(`    <h2>Summary</h2>
    <div class="summary">
`)

	// Eval pass rate card
	evalPassRate := 0.0
	if summary.EvalTotalCount > 0 {
		evalPassRate = float64(summary.EvalPassCount) / float64(summary.EvalTotalCount) * 100
	}
	evalCardClass := "summary-card"
	if evalPassRate >= 90 {
		evalCardClass += " success"
	} else if evalPassRate < 70 {
		evalCardClass += " warning"
	}

	html.WriteString(fmt.Sprintf(`      <div class="%s">
        <h3>Eval Pass Rate</h3>
        <div class="value">%.1f%%</div>
        <div class="label">%d of %d tasks passed</div>
      </div>
`, evalCardClass, evalPassRate, summary.EvalPassCount, summary.EvalTotalCount))

	// Eval avg score card
	scoreCardClass := "summary-card"
	if summary.EvalAvgScore >= 90 {
		scoreCardClass += " success"
	} else if summary.EvalAvgScore < 70 {
		scoreCardClass += " warning"
	}

	html.WriteString(fmt.Sprintf(`      <div class="%s">
        <h3>Avg Eval Score</h3>
        <div class="value">%.1f</div>
        <div class="label">Average task score</div>
      </div>
`, scoreCardClass, summary.EvalAvgScore))

	// Meta consistency card
	metaCardClass := "summary-card"
	if summary.MetaAvgConsistency >= 90 {
		metaCardClass += " success"
	} else if summary.MetaAvgConsistency < 70 {
		metaCardClass += " warning"
	}

	html.WriteString(fmt.Sprintf(`      <div class="%s">
        <h3>Meta Consistency</h3>
        <div class="value">%.1f%%</div>
        <div class="label">Average agent consistency</div>
      </div>
`, metaCardClass, summary.MetaAvgConsistency))

	html.WriteString(`    </div>
`)

	// Eval Results section
	html.WriteString(`    <h2>Eval Results</h2>
`)

	if len(evalResults) == 0 {
		html.WriteString(`    <div class="no-data">No eval results available</div>
`)
	} else {
		html.WriteString(`    <table>
      <thead>
        <tr>
          <th>Task ID</th>
          <th>Timestamp</th>
          <th>Status</th>
          <th>Score</th>
        </tr>
      </thead>
      <tbody>
`)

		for _, result := range evalResults {
			status := "pass"
			statusText := "PASS"
			if !result.OverallPassed {
				status = "fail"
				statusText = "FAIL"
			}

			scoreClass := "score"
			if result.OverallScore >= 90 {
				scoreClass += " high"
			} else if result.OverallScore >= 70 {
				scoreClass += " medium"
			} else {
				scoreClass += " low"
			}

			// Parse and format timestamp
			timestamp := result.Timestamp
			if t, err := time.Parse(time.RFC3339, result.Timestamp); err == nil {
				timestamp = t.Format("2006-01-02 15:04")
			}

			html.WriteString(fmt.Sprintf(`        <tr>
          <td>%s</td>
          <td>%s</td>
          <td><span class="status %s">%s</span></td>
          <td><span class="%s">%.1f</span></td>
        </tr>
`, htmlpkg.EscapeString(result.TaskID), htmlpkg.EscapeString(timestamp), status, statusText, scoreClass, result.OverallScore))
		}

		html.WriteString(`      </tbody>
    </table>
`)
	}

	// Meta Results section
	html.WriteString(`    <h2>Meta Results</h2>
`)

	if len(metaResults) == 0 {
		html.WriteString(`    <div class="no-data">No meta results available</div>
`)
	} else {
		html.WriteString(`    <table>
      <thead>
        <tr>
          <th>Agent</th>
          <th>Boundary Type</th>
          <th>Timestamp</th>
          <th>Consistency</th>
          <th>Runs</th>
        </tr>
      </thead>
      <tbody>
`)

		for _, result := range metaResults {
			scoreClass := "score"
			if result.ConsistencyPercentage >= 90 {
				scoreClass += " high"
			} else if result.ConsistencyPercentage >= 70 {
				scoreClass += " medium"
			} else {
				scoreClass += " low"
			}

			// Parse and format timestamp
			timestamp := result.Timestamp
			if t, err := time.Parse(time.RFC3339, result.Timestamp); err == nil {
				timestamp = t.Format("2006-01-02 15:04")
			}

			html.WriteString(fmt.Sprintf(`        <tr>
          <td>%s</td>
          <td>%s</td>
          <td>%s</td>
          <td><span class="%s">%.1f%%</span></td>
          <td>%d/%d</td>
        </tr>
`, htmlpkg.EscapeString(result.Agent), htmlpkg.EscapeString(result.BoundaryType), htmlpkg.EscapeString(timestamp), scoreClass, result.ConsistencyPercentage, result.ConsistentCount, result.TotalCount))
		}

		html.WriteString(`      </tbody>
    </table>
`)
	}

	// Trends section (simple bar chart)
	html.WriteString(`    <h2>Trends</h2>
    <div class="chart-container">
      <h3>Score Distribution</h3>
      <div class="bar-chart">
`)

	// Create score distribution chart for eval results
	if len(evalResults) > 0 {
		// Count scores by ranges
		scoreRanges := map[string]int{
			"90-100": 0,
			"70-89":  0,
			"50-69":  0,
			"0-49":   0,
		}

		for _, result := range evalResults {
			if result.OverallScore >= 90 {
				scoreRanges["90-100"]++
			} else if result.OverallScore >= 70 {
				scoreRanges["70-89"]++
			} else if result.OverallScore >= 50 {
				scoreRanges["50-69"]++
			} else {
				scoreRanges["0-49"]++
			}
		}

		// Calculate max for scaling
		maxCount := 0
		for _, count := range scoreRanges {
			if count > maxCount {
				maxCount = count
			}
		}
		if maxCount == 0 {
			maxCount = 1
		}

		// Render bars
		ranges := []string{"90-100", "70-89", "50-69", "0-49"}
		for _, r := range ranges {
			count := scoreRanges[r]
			height := float64(count) / float64(maxCount) * 100
			if count > 0 && height < 10 {
				height = 10 // Minimum visible height
			}

			html.WriteString(fmt.Sprintf(`        <div style="display: flex; flex-direction: column; flex: 1; align-items: center;">
          <div class="bar" style="height: %.1f%%; width: 80%%;">
            <span>%d</span>
          </div>
          <div class="bar-label">%s</div>
        </div>
`, height, count, r))
		}
	} else {
		html.WriteString(`        <div class="no-data">No data to display</div>
`)
	}

	html.WriteString(`      </div>
    </div>
`)

	// Close HTML
	html.WriteString(`  </div>
</body>
</html>
`)

	return html.String()
}

// runDashboardCommand executes the dashboard generation command
func runDashboardCommand(reportsDir, outputPath string) error {
	// Validate reports directory exists
	if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
		return fmt.Errorf("reports directory not found: %s", reportsDir)
	}

	// Load eval results
	evalLogPath := filepath.Join(reportsDir, "task-eval-log.json")
	var evalResults []GradeTaskOutput
	if _, err := os.Stat(evalLogPath); err == nil {
		results, err := loadEvalResults(evalLogPath)
		if err != nil {
			return fmt.Errorf("loading eval results: %w", err)
		}
		evalResults = results
	}

	// Load meta results
	metaLogPath := filepath.Join(reportsDir, "consistency-log.json")
	var metaResults []ConsistencyResult
	if _, err := os.Stat(metaLogPath); err == nil {
		results, err := loadMetaResults(metaLogPath)
		if err != nil {
			return fmt.Errorf("loading meta results: %w", err)
		}
		metaResults = results
	}

	// Generate HTML
	html := generateDashboardHTML(evalResults, metaResults)

	// Write to output file
	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("writing dashboard file: %w", err)
	}

	return nil
}
