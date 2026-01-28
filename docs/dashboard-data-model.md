# Dashboard Data Model

This document describes the data model for the Yokay Evals dashboard. The dashboard displays evaluation and meta-evaluation metrics over time.

## Overview

The dashboard data model consists of four main sections:

1. **Metadata** - Information about when and how the data was generated
2. **Time Series** - Metrics aggregated by day (daily granularity)
3. **Agents** - Per-agent performance metrics
4. **Failure Categories** - Counts and pass rates for each failure category

## Schema

The complete JSON Schema is defined in `config/dashboard-schema.json` (JSON Schema Draft 7).

## Data Structure

### Metadata

Contains information about the dashboard data generation:

- `generated_at` (string, required): ISO 8601 timestamp when data was generated
- `version` (string, required): Schema version (semantic versioning)
- `date_range` (object, required): Time range for the data
  - `start` (string, required): ISO 8601 timestamp for range start
  - `end` (string, required): ISO 8601 timestamp for range end

### Time Series

Array of daily data points. Each point contains:

- `timestamp` (string, required): ISO 8601 timestamp for the day
- `eval_metrics` (object, required): Evaluation metrics
  - `pass_rate` (number, 0-100): Percentage of passing evaluations
  - `avg_score` (number, 0-100): Average score across evaluations
  - `total_count` (integer): Total number of evaluations
  - `pass_count` (integer): Number of passing evaluations
  - `fail_count` (integer): Number of failing evaluations
- `meta_metrics` (object, required): Meta-evaluation metrics
  - `avg_consistency` (number, 0-100): Average consistency across agents
  - `total_agents` (integer): Number of agents evaluated

### Agents

Array of per-agent metrics. Each entry contains:

- `agent_name` (string, required): Name of the agent
- `metrics` (object, required): Agent performance metrics
  - `consistency_percentage` (number, 0-100): Consistency percentage
  - `run_count` (integer): Total number of runs
  - `last_run` (string): ISO 8601 timestamp of last run

### Failure Categories

Array of failure category metrics. Each entry contains:

- `category` (string, required): Failure category name (must be one of the 12 valid categories)
- `count` (integer): Number of failures in this category
- `pass_rate` (number, 0-100): Pass rate for this category

#### Valid Failure Categories

The following 12 failure categories are supported (from `failures/schema.yaml`):

1. `missed-tasks` - Agent skipped required functionality
2. `missing-tests` - Agent didn't write tests
3. `wrong-product` - Built something different than requested
4. `regression` - Agent broke previously working code
5. `premature-completion` - Agent claimed done but wasn't
6. `scope-creep` - Agent built more than asked
7. `integration-failure` - Failed to integrate with existing code
8. `session-amnesia` - Agent forgot earlier context
9. `hallucinated-deps` - Made up non-existent dependencies
10. `security-flaw` - Introduced security vulnerability
11. `tool-misuse` - Misused tools or APIs
12. `task-quality` - Poor code quality or design

## Example JSON Payload

```json
{
  "metadata": {
    "generated_at": "2026-01-28T12:00:00Z",
    "version": "1.0.0",
    "date_range": {
      "start": "2026-01-01T00:00:00Z",
      "end": "2026-01-28T23:59:59Z"
    }
  },
  "time_series": [
    {
      "timestamp": "2026-01-27T00:00:00Z",
      "eval_metrics": {
        "pass_rate": 85.5,
        "avg_score": 87.3,
        "total_count": 10,
        "pass_count": 9,
        "fail_count": 1
      },
      "meta_metrics": {
        "avg_consistency": 92.0,
        "total_agents": 3
      }
    },
    {
      "timestamp": "2026-01-26T00:00:00Z",
      "eval_metrics": {
        "pass_rate": 90.0,
        "avg_score": 88.5,
        "total_count": 10,
        "pass_count": 9,
        "fail_count": 1
      },
      "meta_metrics": {
        "avg_consistency": 88.0,
        "total_agents": 3
      }
    }
  ],
  "agents": [
    {
      "agent_name": "yokay-quality-reviewer",
      "metrics": {
        "consistency_percentage": 85.0,
        "run_count": 20,
        "last_run": "2026-01-27T18:13:50Z"
      }
    },
    {
      "agent_name": "yokay-spec-reviewer",
      "metrics": {
        "consistency_percentage": 92.5,
        "run_count": 15,
        "last_run": "2026-01-27T16:22:33Z"
      }
    },
    {
      "agent_name": "yokay-task-reviewer",
      "metrics": {
        "consistency_percentage": 88.0,
        "run_count": 18,
        "last_run": "2026-01-27T17:45:12Z"
      }
    }
  ],
  "failure_categories": [
    {
      "category": "missing-tests",
      "count": 15,
      "pass_rate": 80.0
    },
    {
      "category": "missed-tasks",
      "count": 8,
      "pass_rate": 75.0
    },
    {
      "category": "wrong-product",
      "count": 3,
      "pass_rate": 85.0
    },
    {
      "category": "regression",
      "count": 5,
      "pass_rate": 90.0
    },
    {
      "category": "premature-completion",
      "count": 2,
      "pass_rate": 95.0
    },
    {
      "category": "scope-creep",
      "count": 1,
      "pass_rate": 98.0
    },
    {
      "category": "integration-failure",
      "count": 4,
      "pass_rate": 82.0
    },
    {
      "category": "session-amnesia",
      "count": 0,
      "pass_rate": 100.0
    },
    {
      "category": "hallucinated-deps",
      "count": 1,
      "pass_rate": 97.0
    },
    {
      "category": "security-flaw",
      "count": 2,
      "pass_rate": 88.0
    },
    {
      "category": "tool-misuse",
      "count": 3,
      "pass_rate": 85.0
    },
    {
      "category": "task-quality",
      "count": 6,
      "pass_rate": 78.0
    }
  ]
}
```

## Go Type Definitions

The Go types are defined in `cmd/yokay-evals/dashboard_model.go`:

```go
type DashboardData struct {
    Metadata          DashboardMetadata        `json:"metadata"`
    TimeSeries        []TimeSeriesPoint        `json:"time_series"`
    Agents            []AgentMetrics           `json:"agents"`
    FailureCategories []FailureCategoryMetrics `json:"failure_categories"`
}
```

### Validation

The `DashboardData` type includes a `Validate() error` method that performs the following checks:

- Required fields are present
- Timestamps are valid RFC3339 format
- Percentages are between 0 and 100
- Counts are non-negative
- Failure categories are valid (from the 12 defined categories)

Example usage:

```go
data := DashboardData{
    // ... populate fields
}

if err := data.Validate(); err != nil {
    log.Fatalf("Invalid dashboard data: %v", err)
}
```

## Date Range Filtering

To filter data by date range, populate the `metadata.date_range` fields:

```json
{
  "metadata": {
    "date_range": {
      "start": "2026-01-01T00:00:00Z",
      "end": "2026-01-31T23:59:59Z"
    }
  }
}
```

The aggregation logic (to be implemented separately) should:
1. Filter evaluation and meta-evaluation results by timestamp
2. Only include data within the specified date range
3. Aggregate metrics for each day in the range

## Notes

- All timestamps use RFC3339 format (ISO 8601)
- Time series uses daily granularity (one data point per day)
- Pass rates and percentages are stored as floats (0-100 range)
- Empty arrays are valid (e.g., when no data exists for a category)
- This is a data model definition only - aggregation logic is separate

## Future Enhancements

Potential future additions to the data model:

- Hourly granularity option for time series
- Additional agent metrics (average score, failure categories)
- Trend indicators (week-over-week, month-over-month changes)
- Confidence intervals for percentages
- Detailed breakdown of failure reasons per category
