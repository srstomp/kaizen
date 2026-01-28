package main

import (
	"fmt"
	"strings"
	"time"
)

// DashboardData represents the complete dashboard data model
type DashboardData struct {
	Metadata          DashboardMetadata        `json:"metadata"`
	TimeSeries        []TimeSeriesPoint        `json:"time_series"`
	Agents            []AgentMetrics           `json:"agents"`
	FailureCategories []FailureCategoryMetrics `json:"failure_categories"`
}

// DashboardMetadata contains metadata about the dashboard data
type DashboardMetadata struct {
	GeneratedAt string    `json:"generated_at"`
	Version     string    `json:"version"`
	DateRange   DateRange `json:"date_range"`
}

// DateRange represents a time range for filtering data
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TimeSeriesPoint represents a single point in the time series
type TimeSeriesPoint struct {
	Timestamp   string      `json:"timestamp"`
	EvalMetrics EvalMetrics `json:"eval_metrics"`
	MetaMetrics MetaMetrics `json:"meta_metrics"`
}

// EvalMetrics contains evaluation metrics for a time point
type EvalMetrics struct {
	PassRate   float64 `json:"pass_rate"`
	AvgScore   float64 `json:"avg_score"`
	TotalCount int     `json:"total_count"`
	PassCount  int     `json:"pass_count"`
	FailCount  int     `json:"fail_count"`
}

// MetaMetrics contains meta-evaluation metrics for a time point
type MetaMetrics struct {
	AvgConsistency float64 `json:"avg_consistency"`
	TotalAgents    int     `json:"total_agents"`
}

// AgentMetrics contains metrics for a single agent
type AgentMetrics struct {
	AgentName string           `json:"agent_name"`
	Metrics   AgentMetricsData `json:"metrics"`
}

// AgentMetricsData contains the actual metric values for an agent
type AgentMetricsData struct {
	ConsistencyPercentage float64 `json:"consistency_percentage"`
	RunCount              int     `json:"run_count"`
	LastRun               string  `json:"last_run"`
}

// FailureCategoryMetrics contains metrics for a failure category
type FailureCategoryMetrics struct {
	Category string  `json:"category"`
	Count    int     `json:"count"`
	PassRate float64 `json:"pass_rate"`
}

// validFailureCategories is the list of valid failure categories from failures/schema.yaml
var validFailureCategories = []string{
	"missed-tasks",
	"missing-tests",
	"wrong-product",
	"regression",
	"premature-completion",
	"scope-creep",
	"integration-failure",
	"session-amnesia",
	"hallucinated-deps",
	"security-flaw",
	"tool-misuse",
	"task-quality",
}

// Validate checks if the DashboardData is valid
func (d *DashboardData) Validate() error {
	// Validate metadata
	if d.Metadata.GeneratedAt == "" {
		return fmt.Errorf("metadata.generated_at is required")
	}
	if d.Metadata.Version == "" {
		return fmt.Errorf("metadata.version is required")
	}

	// Validate generated_at timestamp
	if _, err := time.Parse(time.RFC3339, d.Metadata.GeneratedAt); err != nil {
		return fmt.Errorf("metadata.generated_at must be a valid RFC3339 timestamp: %w", err)
	}

	// Validate date range timestamps
	if d.Metadata.DateRange.Start == "" {
		return fmt.Errorf("metadata.date_range.start is required")
	}
	if _, err := time.Parse(time.RFC3339, d.Metadata.DateRange.Start); err != nil {
		return fmt.Errorf("metadata.date_range.start must be a valid RFC3339 timestamp: %w", err)
	}

	if d.Metadata.DateRange.End == "" {
		return fmt.Errorf("metadata.date_range.end is required")
	}
	if _, err := time.Parse(time.RFC3339, d.Metadata.DateRange.End); err != nil {
		return fmt.Errorf("metadata.date_range.end must be a valid RFC3339 timestamp: %w", err)
	}

	// Validate time series points
	for i, point := range d.TimeSeries {
		if err := validateTimeSeriesPoint(point, i); err != nil {
			return err
		}
	}

	// Validate agents
	for i, agent := range d.Agents {
		if err := validateAgentMetrics(agent, i); err != nil {
			return err
		}
	}

	// Validate failure categories
	for i, fc := range d.FailureCategories {
		if err := validateFailureCategoryMetrics(fc, i); err != nil {
			return err
		}
	}

	return nil
}

// validateTimeSeriesPoint validates a single time series point
func validateTimeSeriesPoint(point TimeSeriesPoint, index int) error {
	// Validate timestamp
	if point.Timestamp == "" {
		return fmt.Errorf("time_series[%d].timestamp is required", index)
	}
	if _, err := time.Parse(time.RFC3339, point.Timestamp); err != nil {
		return fmt.Errorf("time_series[%d].timestamp must be a valid RFC3339 timestamp: %w", index, err)
	}

	// Validate eval metrics
	if point.EvalMetrics.PassRate < 0 || point.EvalMetrics.PassRate > 100 {
		return fmt.Errorf("time_series[%d].eval_metrics.pass_rate must be between 0 and 100", index)
	}
	if point.EvalMetrics.AvgScore < 0 || point.EvalMetrics.AvgScore > 100 {
		return fmt.Errorf("time_series[%d].eval_metrics.avg_score must be between 0 and 100", index)
	}
	if point.EvalMetrics.TotalCount < 0 {
		return fmt.Errorf("time_series[%d].eval_metrics.total_count must be non-negative", index)
	}
	if point.EvalMetrics.PassCount < 0 {
		return fmt.Errorf("time_series[%d].eval_metrics.pass_count must be non-negative", index)
	}
	if point.EvalMetrics.FailCount < 0 {
		return fmt.Errorf("time_series[%d].eval_metrics.fail_count must be non-negative", index)
	}

	// Validate meta metrics
	if point.MetaMetrics.AvgConsistency < 0 || point.MetaMetrics.AvgConsistency > 100 {
		return fmt.Errorf("time_series[%d].meta_metrics.avg_consistency must be between 0 and 100", index)
	}
	if point.MetaMetrics.TotalAgents < 0 {
		return fmt.Errorf("time_series[%d].meta_metrics.total_agents must be non-negative", index)
	}

	return nil
}

// validateAgentMetrics validates a single agent metrics entry
func validateAgentMetrics(agent AgentMetrics, index int) error {
	if agent.AgentName == "" {
		return fmt.Errorf("agents[%d].agent_name is required", index)
	}

	if agent.Metrics.ConsistencyPercentage < 0 || agent.Metrics.ConsistencyPercentage > 100 {
		return fmt.Errorf("agents[%d].metrics.consistency_percentage must be between 0 and 100", index)
	}
	if agent.Metrics.RunCount < 0 {
		return fmt.Errorf("agents[%d].metrics.run_count must be non-negative", index)
	}

	// Validate last_run timestamp
	if agent.Metrics.LastRun == "" {
		return fmt.Errorf("agents[%d].metrics.last_run is required", index)
	}
	if _, err := time.Parse(time.RFC3339, agent.Metrics.LastRun); err != nil {
		return fmt.Errorf("agents[%d].metrics.last_run must be a valid RFC3339 timestamp: %w", index, err)
	}

	return nil
}

// validateFailureCategoryMetrics validates a single failure category metrics entry
func validateFailureCategoryMetrics(fc FailureCategoryMetrics, index int) error {
	// Check if category is valid
	validCategory := false
	for _, validCat := range validFailureCategories {
		if fc.Category == validCat {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return fmt.Errorf("failure_categories[%d]: invalid failure category '%s', must be one of: %s",
			index, fc.Category, strings.Join(validFailureCategories, ", "))
	}

	if fc.Count < 0 {
		return fmt.Errorf("failure_categories[%d].count must be non-negative", index)
	}

	if fc.PassRate < 0 || fc.PassRate > 100 {
		return fmt.Errorf("failure_categories[%d].pass_rate must be between 0 and 100", index)
	}

	return nil
}
