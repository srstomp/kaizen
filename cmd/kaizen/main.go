package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/srstomp/kaizen/internal/graders/codebased"
	"github.com/srstomp/kaizen/internal/graders/modelbased"
)

type skillResult struct {
	Name    string
	Path    string
	Score   float64
	Passed  bool
	Message string
	Details map[string]any
}

func main() {
	// Define subcommands
	gradeCmd := flag.NewFlagSet("grade-skills", flag.ExitOnError)
	skillsDir := gradeCmd.String("skills-dir", "/Users/sis4m4/Projects/stevestomp/pokayokay/plugins/pokayokay/skills", "Path to skills directory")
	reportPath := gradeCmd.String("output", "", "Output report path (default: yokay-evals/reports/skill-clarity-YYYY-MM-DD.md)")

	metaCmd := flag.NewFlagSet("meta", flag.ExitOnError)
	suite := metaCmd.String("suite", "", "Suite to run: 'agents' or 'skills'")
	agent := metaCmd.String("agent", "", "Specific agent to run (e.g., 'yokay-spec-reviewer')")
	k := metaCmd.Int("k", 5, "Number of runs for pass^k (default: 5)")
	metaDirFlag := metaCmd.String("meta-dir", "", "Path to meta directory (default: yokay-evals/meta)")
	confirm := metaCmd.Bool("confirm", false, "Confirm before running suite (skips prompt)")

	evalCmd := flag.NewFlagSet("eval", flag.ExitOnError)
	failuresDirFlag := evalCmd.String("failures-dir", "", "Path to failures directory (default: yokay-evals/failures)")
	categoryFlag := evalCmd.String("category", "", "Filter to specific category (e.g., 'missing-tests')")
	kFlag := evalCmd.Int("k", 1, "Number of evaluation runs (default: 1)")
	formatFlag := evalCmd.String("format", "table", "Output format: 'table' or 'json'")

	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	reportType := reportCmd.String("type", "grade", "Report type: 'grade', 'eval', or 'all'")
	reportFormat := reportCmd.String("format", "markdown", "Output format: 'markdown' or 'json'")
	listReports := reportCmd.Bool("list", false, "List available reports without aggregating")
	outputFile := reportCmd.String("output", "", "Write output to file instead of stdout")
	reportsDirFlag := reportCmd.String("reports-dir", "", "Path to reports directory (default: reports/)")
	noTrends := reportCmd.Bool("no-trends", false, "Disable trend analysis")

	gradeTaskCmd := flag.NewFlagSet("grade-task", flag.ExitOnError)
	taskID := gradeTaskCmd.String("task-id", "", "Task ID")
	taskType := gradeTaskCmd.String("task-type", "feature", "Task type (feature, bug, test, spike, chore)")
	changedFiles := gradeTaskCmd.String("changed-files", "", "Comma-separated list of changed files")
	workDir := gradeTaskCmd.String("work-dir", ".", "Working directory")
	gradeFormat := gradeTaskCmd.String("format", "json", "Output format (json, text)")

	gradeTaskQualityCmd := flag.NewFlagSet("grade-task-quality", flag.ExitOnError)
	qualityTaskID := gradeTaskQualityCmd.String("task-id", "", "Task ID")
	qualityTaskTitle := gradeTaskQualityCmd.String("task-title", "", "Task title")
	qualityTaskType := gradeTaskQualityCmd.String("task-type", "", "Task type (feature, bug, test, spike, chore)")
	qualityDescription := gradeTaskQualityCmd.String("description", "", "Task description")
	qualityAcceptanceCriteria := gradeTaskQualityCmd.String("acceptance-criteria", "", "Acceptance criteria (comma-separated or JSON)")
	qualityMinDescLength := gradeTaskQualityCmd.Int("min-description-length", 100, "Minimum description length")
	qualityFormat := gradeTaskQualityCmd.String("format", "json", "Output format (json, text)")

	gradeSingleCmd := flag.NewFlagSet("grade", flag.ExitOnError)
	graderFlag := gradeSingleCmd.String("grader", "", "Grader name (required)")
	inputFlag := gradeSingleCmd.String("input", "", "Path to input JSON file (required)")
	specFlag := gradeSingleCmd.String("spec", "", "Specification text (optional, for model-based graders)")
	singleFormatFlag := gradeSingleCmd.String("format", "text", "Output format (text, json)")

	gateCmd := flag.NewFlagSet("gate", flag.ExitOnError)
	gateType := gateCmd.String("type", "all", "Check type: 'eval', 'meta', or 'all'")
	gateThreshold := gateCmd.Float64("threshold", 95.0, "Threshold percentage (0-100)")
	gateReportsDir := gateCmd.String("reports-dir", "", "Path to reports directory (default: reports/)")

	dashboardCmd := flag.NewFlagSet("dashboard", flag.ExitOnError)
	dashboardReportsDir := dashboardCmd.String("reports-dir", "", "Path to reports directory (default: reports/)")
	dashboardOutput := dashboardCmd.String("output", "dashboard.html", "Output file path for HTML dashboard")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initBootstrap := initCmd.Bool("bootstrap", false, "Bootstrap category stats from existing failures directory")
	initFailuresDir := initCmd.String("failures-dir", "./failures", "Path to failures directory for bootstrapping")

	suggestCmd := flag.NewFlagSet("suggest", flag.ExitOnError)
	suggestTaskID := suggestCmd.String("task-id", "", "Task ID to associate with (required)")
	suggestCategory := suggestCmd.String("category", "", "Failure category to get suggestions for (required)")

	detectCmd := flag.NewFlagSet("detect-category", flag.ExitOnError)
	detectDetails := detectCmd.String("details", "", "Text to analyze for category detection (required)")

	captureCmd := flag.NewFlagSet("capture", flag.ExitOnError)
	captureTaskID := captureCmd.String("task-id", "", "Task ID where the failure occurred (required)")
	captureCategory := captureCmd.String("category", "", "Failure category (required)")
	captureDetails := captureCmd.String("details", "", "Details about the failure (required)")
	captureSource := captureCmd.String("source", "", "Source of the failure, e.g. spec-review, quality-review (required)")

	if len(os.Args) < 2 {
		fmt.Println("Usage: kaizen <command> [options]")
		fmt.Println("\nCommands:")
		fmt.Println("  init                Initialize kaizen configuration directory")
		fmt.Println("  capture             Capture a failure record in the database")
		fmt.Println("  suggest             Generate fix task suggestions based on failure patterns")
		fmt.Println("  detect-category     Detect failure category from text details")
		fmt.Println("  grade               Run a single grader on a single input")
		fmt.Println("  grade-skills        Grade all pokayokay skills and generate report")
		fmt.Println("  grade-task          Run code-based graders on task changes")
		fmt.Println("  grade-task-quality  Evaluate task quality based on metadata")
		fmt.Println("  meta                Run meta-evaluations on agents or skills")
		fmt.Println("  eval                Run eval suite against failure cases")
		fmt.Println("  report              View and analyze evaluation reports")
		fmt.Println("  gate                Check if eval/meta results pass threshold (for CI)")
		fmt.Println("  dashboard           Generate HTML dashboard from eval/meta results")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])

		// Get home directory and create config path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		configDir := filepath.Join(homeDir, ".config", "kaizen")

		if err := runInitCommand(configDir); err != nil {
			log.Fatalf("Failed to initialize kaizen: %v", err)
		}

		fmt.Printf("Kaizen initialized successfully!\n")
		fmt.Printf("Configuration directory: %s\n", configDir)
		fmt.Printf("  - failures.db created\n")
		fmt.Printf("  - config.yaml created\n")

		// Run bootstrap if flag is set
		if *initBootstrap {
			fmt.Printf("\nBootstrapping from failures directory: %s\n", *initFailuresDir)

			dbPath := filepath.Join(configDir, "failures.db")
			stats, err := bootstrapFromFailures(dbPath, *initFailuresDir)
			if err != nil {
				log.Fatalf("Failed to bootstrap from failures: %v", err)
			}

			// Print results
			if len(stats) == 0 {
				fmt.Printf("No failure cases found in directory.\n")
			} else {
				fmt.Printf("\nBootstrap complete:\n")

				// Sort categories for consistent output
				categories := make([]string, 0, len(stats))
				for cat := range stats {
					categories = append(categories, cat)
				}
				sort.Strings(categories)

				total := 0
				for _, cat := range categories {
					count := stats[cat]
					total += count
					fmt.Printf("  %s: %d cases\n", cat, count)
				}
				fmt.Printf("Total: %d cases across %d categories\n", total, len(stats))
			}
		}

	case "suggest":
		suggestCmd.Parse(os.Args[2:])

		// Validate required flags
		if *suggestTaskID == "" {
			fmt.Println("Error: --task-id flag is required")
			suggestCmd.Usage()
			os.Exit(1)
		}
		if *suggestCategory == "" {
			fmt.Println("Error: --category flag is required")
			suggestCmd.Usage()
			os.Exit(1)
		}

		// Check if kaizen is initialized
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		configDir := filepath.Join(homeDir, ".config", "kaizen")
		dbPath := filepath.Join(configDir, "failures.db")

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Error: kaizen not initialized. Run 'kaizen init' first.")
			os.Exit(1)
		}

		if err := runSuggestCommand(*suggestTaskID, *suggestCategory); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "detect-category":
		detectCmd.Parse(os.Args[2:])

		// Validate required flags
		if *detectDetails == "" {
			fmt.Println("Error: --details flag is required")
			detectCmd.Usage()
			os.Exit(1)
		}

		output, err := runDetectCategoryCommand(*detectDetails)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(output)

	case "capture":
		captureCmd.Parse(os.Args[2:])

		// Validate required flags
		if *captureTaskID == "" {
			fmt.Println("Error: --task-id flag is required")
			captureCmd.Usage()
			os.Exit(1)
		}
		if *captureCategory == "" {
			fmt.Println("Error: --category flag is required")
			captureCmd.Usage()
			os.Exit(1)
		}
		if *captureDetails == "" {
			fmt.Println("Error: --details flag is required")
			captureCmd.Usage()
			os.Exit(1)
		}
		if *captureSource == "" {
			fmt.Println("Error: --source flag is required")
			captureCmd.Usage()
			os.Exit(1)
		}

		// Check if kaizen is initialized
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		configDir := filepath.Join(homeDir, ".config", "kaizen")
		dbPath := filepath.Join(configDir, "failures.db")

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Error: kaizen not initialized. Run 'kaizen init' first.")
			os.Exit(1)
		}

		output, err := runCaptureCommand(*captureTaskID, *captureCategory, *captureDetails, *captureSource)
		if err != nil {
			fmt.Fprintln(os.Stderr, output)
			os.Exit(1)
		}

		fmt.Println(output)

	case "grade":
		gradeSingleCmd.Parse(os.Args[2:])

		// Validate required flags
		if *graderFlag == "" {
			fmt.Println("Error: --grader flag is required")
			gradeSingleCmd.Usage()
			os.Exit(1)
		}
		if *inputFlag == "" {
			fmt.Println("Error: --input flag is required")
			gradeSingleCmd.Usage()
			os.Exit(1)
		}

		if err := runGradeCommand(*graderFlag, *inputFlag, *specFlag, *singleFormatFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "grade-skills":
		gradeCmd.Parse(os.Args[2:])

		// Set default output path if not specified
		output := *reportPath
		if output == "" {
			// Get the yokay-evals directory (parent of cmd)
			execPath, err := os.Executable()
			if err != nil {
				log.Fatalf("Failed to get executable path: %v", err)
			}
			evalsDir := filepath.Join(filepath.Dir(filepath.Dir(execPath)), "..")
			reportsDir := filepath.Join(evalsDir, "reports")

			// Create reports directory if it doesn't exist
			if err := os.MkdirAll(reportsDir, 0755); err != nil {
				log.Fatalf("Failed to create reports directory: %v", err)
			}

			today := time.Now().Format("2006-01-02")
			output = filepath.Join(reportsDir, fmt.Sprintf("skill-clarity-%s.md", today))
		}

		if err := gradeSkills(*skillsDir, output); err != nil {
			log.Fatalf("Failed to grade skills: %v", err)
		}

		fmt.Printf("Report generated: %s\n", output)

	case "meta":
		metaCmd.Parse(os.Args[2:])

		// Set default meta directory if not specified
		metaDir := *metaDirFlag
		if metaDir == "" {
			// Try to find meta directory relative to current working directory
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}

			// Check if we're in yokay-evals directory or a subdirectory
			if strings.Contains(cwd, "yokay-evals") {
				// Find the yokay-evals directory
				parts := strings.Split(cwd, "yokay-evals")
				if len(parts) > 0 {
					evalsDir := parts[0] + "yokay-evals"
					metaDir = filepath.Join(evalsDir, "meta")
				}
			} else {
				// Assume meta is relative to current directory
				metaDir = "meta"
			}
		}

		if err := runMetaCommand(*suite, *agent, *k, metaDir, *confirm); err != nil {
			log.Fatalf("Failed to run meta-evaluation: %v", err)
		}

	case "eval":
		evalCmd.Parse(os.Args[2:])

		// Set default failures directory if not specified
		failuresDir := *failuresDirFlag
		if failuresDir == "" {
			// Try to find failures directory relative to current working directory
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}

			// Check if we're in yokay-evals directory or a subdirectory
			if strings.Contains(cwd, "yokay-evals") {
				// Find the yokay-evals directory
				parts := strings.Split(cwd, "yokay-evals")
				if len(parts) > 0 {
					evalsDir := parts[0] + "yokay-evals"
					failuresDir = filepath.Join(evalsDir, "failures")
				}
			} else if strings.Contains(cwd, "pokayokay") {
				// Find the project root
				parts := strings.Split(cwd, "pokayokay")
				if len(parts) > 0 {
					projectRoot := parts[0] + "pokayokay"
					failuresDir = filepath.Join(projectRoot, "yokay-evals", "failures")
				}
			} else {
				// Assume failures is relative to current directory
				failuresDir = "failures"
			}
		}

		if err := runEvalCommand(failuresDir, *categoryFlag, *kFlag, *formatFlag); err != nil {
			log.Fatalf("Failed to run eval command: %v", err)
		}

	case "report":
		reportCmd.Parse(os.Args[2:])

		// Set default reports directory if not specified
		reportsDir := *reportsDirFlag
		if reportsDir == "" {
			// Try to find reports directory relative to current working directory
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}

			// Check if we're in the project root or a subdirectory
			if strings.Contains(cwd, "pokayokay") {
				// Find the project root
				parts := strings.Split(cwd, "pokayokay")
				if len(parts) > 0 {
					projectRoot := parts[0] + "pokayokay"
					reportsDir = filepath.Join(projectRoot, "reports")
				}
			} else {
				// Assume reports is relative to current directory
				reportsDir = "reports"
			}
		}

		if err := runReportCommand(*reportType, *reportFormat, *listReports, *outputFile, reportsDir, !*noTrends); err != nil {
			log.Fatalf("Failed to run report command: %v", err)
		}

	case "grade-task":
		gradeTaskCmd.Parse(os.Args[2:])

		// Parse changed files
		var files []string
		if *changedFiles != "" {
			files = strings.Split(*changedFiles, ",")
			// Trim whitespace
			for i := range files {
				files[i] = strings.TrimSpace(files[i])
			}
		}

		if err := runGradeTaskCommand(*taskID, *taskType, files, *workDir, *gradeFormat); err != nil {
			log.Fatalf("Failed to run grade-task command: %v", err)
		}

	case "grade-task-quality":
		gradeTaskQualityCmd.Parse(os.Args[2:])

		if err := runGradeTaskQuality(*qualityTaskID, *qualityTaskTitle, *qualityTaskType, *qualityDescription, *qualityAcceptanceCriteria, *qualityMinDescLength, *qualityFormat); err != nil {
			log.Fatalf("Failed to run grade-task-quality command: %v", err)
		}

	case "gate":
		gateCmd.Parse(os.Args[2:])

		// Set default reports directory if not specified
		reportsDir := *gateReportsDir
		if reportsDir == "" {
			// Try to find reports directory relative to current working directory
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}

			// Check if we're in the project root or a subdirectory
			if strings.Contains(cwd, "pokayokay") {
				// Find the project root
				parts := strings.Split(cwd, "pokayokay")
				if len(parts) > 0 {
					projectRoot := parts[0] + "pokayokay"
					reportsDir = filepath.Join(projectRoot, "reports")
				}
			} else {
				// Assume reports is relative to current directory
				reportsDir = "reports"
			}
		}

		if err := runGateCommand(*gateType, *gateThreshold, reportsDir); err != nil {
			log.Fatalf("Gate check failed: %v", err)
		}

	case "dashboard":
		dashboardCmd.Parse(os.Args[2:])

		// Set default reports directory if not specified
		reportsDir := *dashboardReportsDir
		if reportsDir == "" {
			// Try to find reports directory relative to current working directory
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}

			// Check if we're in the project root or a subdirectory
			if strings.Contains(cwd, "pokayokay") {
				// Find the project root
				parts := strings.Split(cwd, "pokayokay")
				if len(parts) > 0 {
					projectRoot := parts[0] + "pokayokay"
					reportsDir = filepath.Join(projectRoot, "reports")
				}
			} else {
				// Assume reports is relative to current directory
				reportsDir = "reports"
			}
		}

		if err := runDashboardCommand(reportsDir, *dashboardOutput); err != nil {
			log.Fatalf("Dashboard generation failed: %v", err)
		}

		fmt.Printf("Dashboard generated: %s\n", *dashboardOutput)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

// gradeSkills finds all skill files, grades them, and generates a report
func gradeSkills(skillsDir, reportPath string) error {
	// Find all SKILL.md files
	skillFiles, err := findSkillFiles(skillsDir)
	if err != nil {
		return fmt.Errorf("finding skill files: %w", err)
	}

	if len(skillFiles) == 0 {
		return fmt.Errorf("no skill files found in %s", skillsDir)
	}

	fmt.Printf("Found %d skills to grade...\n", len(skillFiles))

	// Grade each skill
	grader := modelbased.NewSkillClarityGrader()
	results := make([]skillResult, 0, len(skillFiles))

	for i, skillPath := range skillFiles {
		fmt.Printf("[%d/%d] Grading %s...\n", i+1, len(skillFiles), filepath.Base(filepath.Dir(skillPath)))

		// Read skill content
		content, err := os.ReadFile(skillPath)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", skillPath, err)
			continue
		}

		// Grade the skill
		result, err := grader.Grade(modelbased.GradeInput{
			Content: string(content),
			Context: map[string]any{
				"path": skillPath,
			},
		})
		if err != nil {
			log.Printf("Warning: Failed to grade %s: %v", skillPath, err)
			continue
		}

		// Extract skill name from path (directory name containing SKILL.md)
		skillName := filepath.Base(filepath.Dir(skillPath))

		results = append(results, skillResult{
			Name:    skillName,
			Path:    skillPath,
			Score:   result.Score,
			Passed:  result.Passed,
			Message: result.Message,
			Details: result.Details,
		})
	}

	if len(results) == 0 {
		return fmt.Errorf("no skills were successfully graded")
	}

	// Generate report
	if err := generateReport(results, reportPath); err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	return nil
}

// findSkillFiles recursively finds all SKILL.md files in the given directory
func findSkillFiles(rootDir string) ([]string, error) {
	var skillFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "SKILL.md" {
			skillFiles = append(skillFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return skillFiles, nil
}

// generateReport creates a markdown report from grading results
func generateReport(results []skillResult, reportPath string) error {
	// Sort results by score (highest to lowest)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Calculate summary statistics
	totalScore := 0.0
	passCount := 0
	for _, r := range results {
		totalScore += r.Score
		if r.Passed {
			passCount++
		}
	}
	avgScore := totalScore / float64(len(results))
	passRate := float64(passCount) / float64(len(results)) * 100

	// Build report content
	var sb strings.Builder

	// Header
	sb.WriteString("# Skill Clarity Report\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("This report evaluates pokayokay skills using the Skill Clarity Grader.\n")
	sb.WriteString("**Note**: Current grading uses heuristic-based evaluation (stub implementation). LLM-based grading not yet implemented.\n\n")

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Skills**: %d\n", len(results)))
	sb.WriteString(fmt.Sprintf("- **Average Score**: %.1f/100\n", avgScore))
	sb.WriteString(fmt.Sprintf("- **Pass Rate**: %.1f%% (%d/%d)\n", passRate, passCount, len(results)))
	sb.WriteString(fmt.Sprintf("- **Passing Threshold**: 70.0\n\n"))

	// Skills below threshold
	belowThreshold := []skillResult{}
	for _, r := range results {
		if r.Score < 80.0 {
			belowThreshold = append(belowThreshold, r)
		}
	}

	if len(belowThreshold) > 0 {
		sb.WriteString("## Skills Below Threshold (< 80%)\n\n")
		sb.WriteString("These skills need improvement:\n\n")
		for _, r := range belowThreshold {
			status := "Needs Improvement"
			if r.Score < 70.0 {
				status = "**FAILED**"
			}
			sb.WriteString(fmt.Sprintf("- **%s** - %.1f/100 - %s\n", r.Name, r.Score, status))
		}
		sb.WriteString("\n")
	}

	// Ranked list
	sb.WriteString("## Skills by Score\n\n")
	sb.WriteString("All skills ranked from highest to lowest:\n\n")
	sb.WriteString("| Rank | Skill | Score | Status |\n")
	sb.WriteString("|------|-------|-------|--------|\n")

	for i, r := range results {
		status := "✅ Pass"
		if !r.Passed {
			status = "❌ Fail"
		} else if r.Score < 80.0 {
			status = "⚠️  Pass (Low)"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %.1f | %s |\n", i+1, r.Name, r.Score, status))
	}
	sb.WriteString("\n")

	// Detailed breakdown
	sb.WriteString("## Detailed Breakdown\n\n")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("### %s\n\n", r.Name))
		sb.WriteString(fmt.Sprintf("**Overall Score**: %.1f/100 - %s\n\n", r.Score, r.Message))
		sb.WriteString("**Criteria Scores**:\n\n")

		// Extract and display criteria details
		criteria := []string{"clear_instructions", "actionable_steps", "good_examples", "appropriate_scope"}
		for _, criterion := range criteria {
			if details, ok := r.Details[criterion].(map[string]any); ok {
				// Safely extract fields with type checking
				score, scoreOk := details["score"].(float64)
				feedback, feedbackOk := details["feedback"].(string)
				weight, weightOk := details["weight"].(float64)

				// Skip this criterion if any field is missing or has wrong type
				if !scoreOk || !feedbackOk || !weightOk {
					continue
				}

				sb.WriteString(fmt.Sprintf("- **%s** (weight: %.0f%%): %.1f/100\n",
					formatCriterionName(criterion), weight*100, score))
				sb.WriteString(fmt.Sprintf("  - %s\n", feedback))
			}
		}
		sb.WriteString("\n")
	}

	// Write report to file
	if err := os.WriteFile(reportPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing report file: %w", err)
	}

	return nil
}

// formatCriterionName converts snake_case to Title Case
func formatCriterionName(name string) string {
	parts := strings.Split(name, "_")
	for i, part := range parts {
		if len(part) > 0 {
			// Manually title case: capitalize first letter, lowercase the rest
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, " ")
}

// GradeTaskOutput represents the JSON output from grade-task command
type GradeTaskOutput struct {
	TaskID        string                  `json:"task_id"`
	Timestamp     string                  `json:"timestamp"`
	Results       []codebased.GradeResult `json:"results"`
	OverallPassed bool                    `json:"overall_passed"`
	OverallScore  float64                 `json:"overall_score"`
}

// runGradeTaskCommand executes the grade-task CLI command
func runGradeTaskCommand(taskID, taskType string, changedFiles []string, workDir, format string) error {
	// Validate taskType
	validTaskTypes := []string{"feature", "bug", "test", "spike", "chore"}
	isValid := false
	for _, validType := range validTaskTypes {
		if taskType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid task type %q: must be one of: %s", taskType, strings.Join(validTaskTypes, ", "))
	}

	// Create input for graders
	input := codebased.GradeInput{
		TaskID:       taskID,
		TaskType:     taskType,
		ChangedFiles: changedFiles,
		WorkDir:      workDir,
	}

	// Initialize graders
	graders := []codebased.CodeGrader{
		codebased.NewFileExistsGrader(),
		codebased.NewTestExistsGrader(),
	}

	// Run graders
	var results []codebased.GradeResult
	totalScore := float64(0)
	applicableCount := 0

	for _, grader := range graders {
		result := grader.Grade(input)
		results = append(results, result)

		// Only count applicable graders in overall score
		if !result.Skipped {
			totalScore += result.Score
			applicableCount++
		}
	}

	// Calculate overall metrics
	overallScore := float64(0)
	if applicableCount > 0 {
		overallScore = totalScore / float64(applicableCount)
	}

	// Overall passes if all applicable graders pass
	overallPassed := true
	for _, r := range results {
		if !r.Skipped && !r.Passed {
			overallPassed = false
			break
		}
	}

	// Format output
	if format == "json" {
		output := GradeTaskOutput{
			TaskID:        taskID,
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Results:       results,
			OverallPassed: overallPassed,
			OverallScore:  overallScore,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("encoding JSON output: %w", err)
		}
	} else {
		// Text format
		fmt.Printf("Task Grading Results\n")
		fmt.Printf("====================\n\n")
		fmt.Printf("Task ID: %s\n", taskID)
		fmt.Printf("Task Type: %s\n", taskType)
		fmt.Printf("Changed Files: %d\n\n", len(changedFiles))

		fmt.Printf("Grader Results:\n")
		for _, r := range results {
			if r.Skipped {
				fmt.Printf("  %s: SKIPPED (%s)\n", r.GraderName, r.SkipReason)
			} else {
				status := "PASS"
				if !r.Passed {
					status = "FAIL"
				}
				fmt.Printf("  %s: %s (score: %.1f) - %s\n", r.GraderName, status, r.Score, r.Details)
			}
		}

		fmt.Printf("\nOverall Score: %.1f\n", overallScore)
		fmt.Printf("Overall Result: ")
		if overallPassed {
			fmt.Printf("PASS\n")
		} else {
			fmt.Printf("FAIL\n")
		}
	}

	return nil
}

// TaskQualityIssue represents a quality check issue
type TaskQualityIssue struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

// TaskQualityResult represents the output of task quality grading
type TaskQualityResult struct {
	TaskID     string             `json:"task_id"`
	Passed     bool               `json:"passed"`
	Score      float64            `json:"score"`
	Issues     []TaskQualityIssue `json:"issues"`
	Suggestion string             `json:"suggestion"`
}

// runGradeTaskQuality evaluates task quality based on metadata
func runGradeTaskQuality(taskID, taskTitle, taskType, description, acceptanceCriteria string, minDescLength int, format string) error {
	// Validate task type
	validTaskTypes := []string{"feature", "bug", "test", "spike", "chore"}
	isValid := false
	for _, validType := range validTaskTypes {
		if taskType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid task type %q: must be one of: %s", taskType, strings.Join(validTaskTypes, ", "))
	}

	// Initialize result
	result := TaskQualityResult{
		TaskID:     taskID,
		Passed:     true,
		Score:      100.0,
		Issues:     []TaskQualityIssue{},
		Suggestion: "",
	}

	// Track total checks and failed checks for scoring
	totalChecks := 4.0 // description_length, acceptance_criteria, ambiguous_keywords, spike_type
	failedChecks := 0.0

	// Check 1: Description length
	descLength := len(description)
	if descLength < minDescLength {
		result.Issues = append(result.Issues, TaskQualityIssue{
			Check:   "description_length",
			Message: fmt.Sprintf("Description too short (%d chars, minimum %d)", descLength, minDescLength),
		})
		failedChecks++
	}

	// Check 2: Acceptance criteria (required for feature, test, spike)
	requiresAcceptanceCriteria := taskType == "feature" || taskType == "test" || taskType == "spike"
	hasAcceptanceCriteria := strings.TrimSpace(acceptanceCriteria) != ""

	if requiresAcceptanceCriteria && !hasAcceptanceCriteria {
		result.Issues = append(result.Issues, TaskQualityIssue{
			Check:   "acceptance_criteria",
			Message: "Missing acceptance criteria",
		})
		failedChecks++
	}

	// Check 3: Ambiguous keywords in title or description
	ambiguousKeywords := []string{"investigate", "explore", "figure out", "look into", "understand"}
	combinedText := strings.ToLower(taskTitle + " " + description)

	for _, keyword := range ambiguousKeywords {
		if strings.Contains(combinedText, keyword) {
			result.Issues = append(result.Issues, TaskQualityIssue{
				Check:   "ambiguous_keywords",
				Message: fmt.Sprintf("Contains ambiguous keyword '%s' - task may be too vague", keyword),
			})
			failedChecks++
			break // Only report once
		}
	}

	// Check 4: Spike type (already covered by acceptance criteria check, but noted separately)
	// This is implicitly handled by the acceptance criteria check above

	// Calculate score (percentage of checks passed)
	result.Score = ((totalChecks - failedChecks) / totalChecks) * 100.0
	result.Passed = len(result.Issues) == 0

	// Add suggestion if failed
	if !result.Passed {
		result.Suggestion = "Run /pokayokay:brainstorm to refine task requirements"
	}

	// Format output
	if format == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			return fmt.Errorf("encoding JSON output: %w", err)
		}
	} else {
		// Text format
		fmt.Printf("Task Quality Check\n")
		fmt.Printf("==================\n\n")
		fmt.Printf("Task ID: %s\n", taskID)
		fmt.Printf("Task Title: %s\n", taskTitle)
		fmt.Printf("Task Type: %s\n", taskType)
		fmt.Printf("Score: %.1f/100\n\n", result.Score)

		if result.Passed {
			fmt.Printf("Status: PASSED\n")
		} else {
			fmt.Printf("Status: FAILED\n\n")
			fmt.Printf("Issues:\n")
			for _, issue := range result.Issues {
				fmt.Printf("  - [%s] %s\n", issue.Check, issue.Message)
			}
			fmt.Printf("\nSuggestion: %s\n", result.Suggestion)
		}
	}

	return nil
}
