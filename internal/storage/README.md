# Storage Package

The storage package provides persistence for evaluation results in the yokay-evals system.

## Overview

This package handles saving and loading historical evaluation results to/from JSON files in the `yokay-evals/results/` directory.

## Usage

```go
import "github.com/srstomp/kaizen/internal/storage"

// Create a store
store := storage.NewResultsStore("yokay-evals/results")

// Save eval results
evalResult := &storage.EvalResult{
    TaskID:        "task-123",
    Timestamp:     time.Now().UTC().Format(time.RFC3339),
    Results:       []codebased.GradeResult{...},
    OverallPassed: true,
    OverallScore:  95.0,
}
err := store.SaveEvalResult(evalResult)

// Load eval results
results, err := store.LoadEvalResults()
```

## Result Types

### EvalResult
Represents output from the `grade-task` command. Contains grading results for a specific task execution.

### MetaResult
Represents consistency evaluation results from meta-evals. Tracks consistency metrics across multiple eval runs.

### TaskQualityResult
Represents task quality grading results. Evaluates the quality of task definitions themselves.

## File Format

Results are stored as JSON arrays in separate files:
- `eval-results.json` - Task execution evaluations
- `meta-results.json` - Consistency meta-evaluations
- `task-quality-results.json` - Task quality evaluations

Files are formatted with indentation for human readability.

## Features

- **Append-only**: New results are appended to existing files
- **Auto-timestamps**: Timestamps are automatically added if not present
- **Graceful failures**: Missing files return empty slices, not errors
- **Directory creation**: Automatically creates results directory if needed
- **RFC3339 timestamps**: Uses standard timestamp format for consistency
