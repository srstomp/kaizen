package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDashboardData_JSONSerialization(t *testing.T) {
	// Create a sample dashboard data structure
	now := time.Now().UTC().Truncate(time.Second)
	dayAgo := now.Add(-24 * time.Hour)

	data := DashboardData{
		Metadata: DashboardMetadata{
			GeneratedAt: now.Format(time.RFC3339),
			Version:     "1.0.0",
			DateRange: DateRange{
				Start: dayAgo.Format(time.RFC3339),
				End:   now.Format(time.RFC3339),
			},
		},
		TimeSeries: []TimeSeriesPoint{
			{
				Timestamp: dayAgo.Format(time.RFC3339),
				EvalMetrics: EvalMetrics{
					PassRate:   85.5,
					AvgScore:   87.3,
					TotalCount: 10,
					PassCount:  9,
					FailCount:  1,
				},
				MetaMetrics: MetaMetrics{
					AvgConsistency: 92.0,
					TotalAgents:    3,
				},
			},
		},
		Agents: []AgentMetrics{
			{
				AgentName: "yokay-quality-reviewer",
				Metrics: AgentMetricsData{
					ConsistencyPercentage: 85.0,
					RunCount:              20,
					LastRun:               now.Format(time.RFC3339),
				},
			},
		},
		FailureCategories: []FailureCategoryMetrics{
			{
				Category: "missing-tests",
				Count:    15,
				PassRate: 80.0,
			},
		},
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal DashboardData to JSON: %v", err)
	}

	// Deserialize from JSON
	var deserialized DashboardData
	if err := json.Unmarshal(jsonBytes, &deserialized); err != nil {
		t.Fatalf("Failed to unmarshal DashboardData from JSON: %v", err)
	}

	// Verify key fields
	if deserialized.Metadata.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", deserialized.Metadata.Version)
	}

	if len(deserialized.TimeSeries) != 1 {
		t.Errorf("Expected 1 time series point, got %d", len(deserialized.TimeSeries))
	}

	if deserialized.TimeSeries[0].EvalMetrics.PassRate != 85.5 {
		t.Errorf("Expected pass rate 85.5, got %.1f", deserialized.TimeSeries[0].EvalMetrics.PassRate)
	}

	if len(deserialized.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(deserialized.Agents))
	}

	if deserialized.Agents[0].AgentName != "yokay-quality-reviewer" {
		t.Errorf("Expected agent name 'yokay-quality-reviewer', got '%s'", deserialized.Agents[0].AgentName)
	}

	if len(deserialized.FailureCategories) != 1 {
		t.Errorf("Expected 1 failure category, got %d", len(deserialized.FailureCategories))
	}

	if deserialized.FailureCategories[0].Category != "missing-tests" {
		t.Errorf("Expected category 'missing-tests', got '%s'", deserialized.FailureCategories[0].Category)
	}
}

func TestDashboardData_JSONFieldNames(t *testing.T) {
	// Create minimal data
	data := DashboardData{
		Metadata: DashboardMetadata{
			GeneratedAt: "2026-01-28T12:00:00Z",
			Version:     "1.0.0",
			DateRange: DateRange{
				Start: "2026-01-01T00:00:00Z",
				End:   "2026-01-28T23:59:59Z",
			},
		},
		TimeSeries:        []TimeSeriesPoint{},
		Agents:            []AgentMetrics{},
		FailureCategories: []FailureCategoryMetrics{},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify snake_case field names
	expectedFields := []string{
		"\"generated_at\"",
		"\"version\"",
		"\"date_range\"",
		"\"start\"",
		"\"end\"",
		"\"time_series\"",
		"\"agents\"",
		"\"failure_categories\"",
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain field %s, but it didn't. JSON: %s", field, jsonStr)
		}
	}
}

func TestDashboardData_Validate(t *testing.T) {
	tests := []struct {
		name    string
		data    DashboardData
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid data",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries:        []TimeSeriesPoint{},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: false,
		},
		{
			name: "missing generated_at",
			data: DashboardData{
				Metadata: DashboardMetadata{
					Version: "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries:        []TimeSeriesPoint{},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "generated_at is required",
		},
		{
			name: "missing version",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries:        []TimeSeriesPoint{},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "invalid generated_at timestamp",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "not-a-timestamp",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries:        []TimeSeriesPoint{},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "generated_at must be a valid RFC3339 timestamp",
		},
		{
			name: "invalid date range start",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "invalid",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries:        []TimeSeriesPoint{},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "date_range.start must be a valid RFC3339 timestamp",
		},
		{
			name: "invalid failure category",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries: []TimeSeriesPoint{},
				Agents:     []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{
					{
						Category: "invalid-category",
						Count:    5,
						PassRate: 50.0,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid failure category",
		},
		{
			name: "negative pass rate",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries: []TimeSeriesPoint{
					{
						Timestamp: "2026-01-27T00:00:00Z",
						EvalMetrics: EvalMetrics{
							PassRate:   -5.0,
							AvgScore:   85.0,
							TotalCount: 10,
							PassCount:  5,
							FailCount:  5,
						},
						MetaMetrics: MetaMetrics{
							AvgConsistency: 90.0,
							TotalAgents:    3,
						},
					},
				},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "pass_rate must be between 0 and 100",
		},
		{
			name: "missing date_range.start",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries:        []TimeSeriesPoint{},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "date_range.start is required",
		},
		{
			name: "missing date_range.end",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "",
					},
				},
				TimeSeries:        []TimeSeriesPoint{},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "date_range.end is required",
		},
		{
			name: "missing timestamp in time_series",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries: []TimeSeriesPoint{
					{
						Timestamp: "",
						EvalMetrics: EvalMetrics{
							PassRate:   85.0,
							AvgScore:   85.0,
							TotalCount: 10,
							PassCount:  8,
							FailCount:  2,
						},
						MetaMetrics: MetaMetrics{
							AvgConsistency: 90.0,
							TotalAgents:    3,
						},
					},
				},
				Agents:            []AgentMetrics{},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "timestamp is required",
		},
		{
			name: "missing last_run in agents",
			data: DashboardData{
				Metadata: DashboardMetadata{
					GeneratedAt: "2026-01-28T12:00:00Z",
					Version:     "1.0.0",
					DateRange: DateRange{
						Start: "2026-01-01T00:00:00Z",
						End:   "2026-01-28T23:59:59Z",
					},
				},
				TimeSeries: []TimeSeriesPoint{},
				Agents: []AgentMetrics{
					{
						AgentName: "test-agent",
						Metrics: AgentMetricsData{
							ConsistencyPercentage: 85.0,
							RunCount:              10,
							LastRun:               "",
						},
					},
				},
				FailureCategories: []FailureCategoryMetrics{},
			},
			wantErr: true,
			errMsg:  "last_run is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', but got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}
