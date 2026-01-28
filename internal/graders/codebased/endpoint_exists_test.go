package codebased

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEndpointExistsGraderInterface verifies EndpointExistsGrader implements CodeGrader
func TestEndpointExistsGraderInterface(t *testing.T) {
	var _ CodeGrader = (*EndpointExistsGrader)(nil)
}

// TestEndpointExistsGraderName verifies the grader name
func TestEndpointExistsGraderName(t *testing.T) {
	grader := NewEndpointExistsGrader()
	if grader.Name() != "endpoint-exists" {
		t.Errorf("Expected name 'endpoint-exists', got %s", grader.Name())
	}
}

// TestEndpointExistsGraderIsApplicable verifies applicability logic
func TestEndpointExistsGraderIsApplicable(t *testing.T) {
	grader := NewEndpointExistsGrader()

	tests := []struct {
		name     string
		input    GradeInput
		expected bool
	}{
		{
			name: "applicable when JS files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"routes.js"},
				WorkDir:      "/tmp",
			},
			expected: true,
		},
		{
			name: "applicable when TS files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"routes.ts"},
				WorkDir:      "/tmp",
			},
			expected: true,
		},
		{
			name: "applicable when JSX files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"routes.jsx"},
				WorkDir:      "/tmp",
			},
			expected: true,
		},
		{
			name: "applicable when TSX files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"routes.tsx"},
				WorkDir:      "/tmp",
			},
			expected: true,
		},
		{
			name: "not applicable when only Go files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"main.go"},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
		{
			name: "not applicable for chore tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "chore",
				ChangedFiles: []string{"routes.js"},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
		{
			name: "not applicable for spike tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "spike",
				ChangedFiles: []string{"routes.js"},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
		{
			name: "not applicable when no files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{},
				WorkDir:      "/tmp",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grader.IsApplicable(tt.input)
			if result != tt.expected {
				t.Errorf("Expected IsApplicable to be %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEndpointExistsGraderExpressRoutesFound verifies grading when Express routes exist
func TestEndpointExistsGraderExpressRoutesFound(t *testing.T) {
	tmpDir := t.TempDir()
	routesFile := filepath.Join(tmpDir, "routes.js")

	content := `
const express = require('express');
const app = express();

app.get('/users', (req, res) => {
  res.json({ users: [] });
});

app.post('/users/:id', (req, res) => {
  res.json({ created: true });
});

app.put("/items/:id", (req, res) => {
  res.json({ updated: true });
});
`
	if err := os.WriteFile(routesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create routes file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{routesFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.GraderName != "endpoint-exists" {
		t.Errorf("Expected GraderName 'endpoint-exists', got %s", result.GraderName)
	}
	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
	if result.Skipped {
		t.Error("Expected Skipped to be false")
	}
	if !strings.Contains(result.Details, "GET /users") {
		t.Error("Expected Details to contain 'GET /users'")
	}
	if !strings.Contains(result.Details, "POST /users/:id") {
		t.Error("Expected Details to contain 'POST /users/:id'")
	}
	if !strings.Contains(result.Details, "PUT /items/:id") {
		t.Error("Expected Details to contain 'PUT /items/:id'")
	}
}

// TestEndpointExistsGraderRouterPattern verifies detection of router patterns
func TestEndpointExistsGraderRouterPattern(t *testing.T) {
	tmpDir := t.TempDir()
	routesFile := filepath.Join(tmpDir, "userRoutes.js")

	content := `
const express = require('express');
const router = express.Router();

router.get('/profile', (req, res) => {
  res.json({ profile: {} });
});

router.delete('/account', (req, res) => {
  res.json({ deleted: true });
});
`
	if err := os.WriteFile(routesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create routes file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{routesFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
	if !strings.Contains(result.Details, "GET /profile") {
		t.Error("Expected Details to contain 'GET /profile'")
	}
	if !strings.Contains(result.Details, "DELETE /account") {
		t.Error("Expected Details to contain 'DELETE /account'")
	}
}

// TestEndpointExistsGraderAllHTTPMethods verifies all HTTP methods are detected
func TestEndpointExistsGraderAllHTTPMethods(t *testing.T) {
	tmpDir := t.TempDir()
	routesFile := filepath.Join(tmpDir, "api.ts")

	content := `
import express from 'express';
const app = express();

app.get('/get-route', handler);
app.post('/post-route', handler);
app.put('/put-route', handler);
app.patch('/patch-route', handler);
app.delete('/delete-route', handler);
`
	if err := os.WriteFile(routesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create routes file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{routesFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if !strings.Contains(result.Details, "GET /get-route") {
		t.Error("Expected Details to contain 'GET /get-route'")
	}
	if !strings.Contains(result.Details, "POST /post-route") {
		t.Error("Expected Details to contain 'POST /post-route'")
	}
	if !strings.Contains(result.Details, "PUT /put-route") {
		t.Error("Expected Details to contain 'PUT /put-route'")
	}
	if !strings.Contains(result.Details, "PATCH /patch-route") {
		t.Error("Expected Details to contain 'PATCH /patch-route'")
	}
	if !strings.Contains(result.Details, "DELETE /delete-route") {
		t.Error("Expected Details to contain 'DELETE /delete-route'")
	}
}

// TestEndpointExistsGraderNoRoutesFound verifies grading when no routes exist
func TestEndpointExistsGraderNoRoutesFound(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "utils.js")

	content := `
function calculateSum(a, b) {
  return a + b;
}

module.exports = { calculateSum };
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{file},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if result.Passed {
		t.Error("Expected Passed to be false when no routes found")
	}
	if result.Score != 0 {
		t.Errorf("Expected Score 0, got %f", result.Score)
	}
	if !strings.Contains(result.Details, "No endpoints") {
		t.Errorf("Expected Details to indicate no endpoints, got: %s", result.Details)
	}
}

// TestEndpointExistsGraderMixedFiles verifies grading with some files having routes
func TestEndpointExistsGraderMixedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// File with routes
	routesFile := filepath.Join(tmpDir, "routes.js")
	routesContent := `
app.get('/api/data', handler);
`
	if err := os.WriteFile(routesFile, []byte(routesContent), 0644); err != nil {
		t.Fatalf("Failed to write routes file: %v", err)
	}

	// File without routes
	utilsFile := filepath.Join(tmpDir, "utils.js")
	utilsContent := `
function helper() { return 42; }
`
	if err := os.WriteFile(utilsFile, []byte(utilsContent), 0644); err != nil {
		t.Fatalf("Failed to write utils file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{routesFile, utilsFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	// Should pass because at least one file has endpoints
	if !result.Passed {
		t.Errorf("Expected Passed to be true when at least one file has routes, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
	if !strings.Contains(result.Details, "GET /api/data") {
		t.Error("Expected Details to contain 'GET /api/data'")
	}
}

// TestEndpointExistsGraderRelativePaths verifies handling of relative paths
func TestEndpointExistsGraderRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	routesFile := filepath.Join(tmpDir, "routes.js")

	content := `
app.get('/test', handler);
`
	if err := os.WriteFile(routesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create routes file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{"routes.js"}, // Relative path
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	// Should resolve relative path against WorkDir
	if !result.Passed {
		t.Errorf("Expected Passed to be true for relative path, details: %s", result.Details)
	}
	if result.Score != 100 {
		t.Errorf("Expected Score 100, got %f", result.Score)
	}
}

// TestEndpointExistsGraderSkipWhenNotApplicable verifies grader skips appropriately
func TestEndpointExistsGraderSkipWhenNotApplicable(t *testing.T) {
	grader := NewEndpointExistsGrader()

	tests := []struct {
		name       string
		input      GradeInput
		skipReason string
	}{
		{
			name: "skip when no files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{},
				WorkDir:      "/tmp",
			},
			skipReason: "No changed files",
		},
		{
			name: "skip for chore tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "chore",
				ChangedFiles: []string{"routes.js"},
				WorkDir:      "/tmp",
			},
			skipReason: "Not applicable for chore tasks",
		},
		{
			name: "skip for spike tasks",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "spike",
				ChangedFiles: []string{"routes.js"},
				WorkDir:      "/tmp",
			},
			skipReason: "Not applicable for spike tasks",
		},
		{
			name: "skip when only non-JS files changed",
			input: GradeInput{
				TaskID:       "task-123",
				TaskType:     "feature",
				ChangedFiles: []string{"main.go", "utils.py"},
				WorkDir:      "/tmp",
			},
			skipReason: "No JS/TS files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grader.Grade(tt.input)

			if !result.Skipped {
				t.Error("Expected Skipped to be true")
			}
			if result.SkipReason == "" {
				t.Error("Expected SkipReason to be provided")
			}
		})
	}
}

// TestEndpointExistsGraderSingleQuotes verifies single quote handling
func TestEndpointExistsGraderSingleQuotes(t *testing.T) {
	tmpDir := t.TempDir()
	routesFile := filepath.Join(tmpDir, "routes.js")

	content := `
app.get('/single-quotes', handler);
router.post('/another-single', handler);
`
	if err := os.WriteFile(routesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create routes file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{routesFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if !strings.Contains(result.Details, "/single-quotes") {
		t.Error("Expected Details to contain '/single-quotes'")
	}
	if !strings.Contains(result.Details, "/another-single") {
		t.Error("Expected Details to contain '/another-single'")
	}
}

// TestEndpointExistsGraderDoubleQuotes verifies double quote handling
func TestEndpointExistsGraderDoubleQuotes(t *testing.T) {
	tmpDir := t.TempDir()
	routesFile := filepath.Join(tmpDir, "routes.js")

	content := `
app.get("/double-quotes", handler);
router.post("/another-double", handler);
`
	if err := os.WriteFile(routesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create routes file: %v", err)
	}

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{routesFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	if !result.Passed {
		t.Errorf("Expected Passed to be true, details: %s", result.Details)
	}
	if !strings.Contains(result.Details, "/double-quotes") {
		t.Error("Expected Details to contain '/double-quotes'")
	}
	if !strings.Contains(result.Details, "/another-double") {
		t.Error("Expected Details to contain '/another-double'")
	}
}

// TestEndpointExistsGraderFileReadError verifies graceful handling of unreadable files
func TestEndpointExistsGraderFileReadError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid routes file
	validFile := filepath.Join(tmpDir, "valid.js")
	validContent := `
app.get('/valid-endpoint', handler);
`
	if err := os.WriteFile(validFile, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid routes file: %v", err)
	}

	// Reference a file that doesn't exist
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.js")

	grader := NewEndpointExistsGrader()
	input := GradeInput{
		TaskID:       "task-123",
		TaskType:     "feature",
		ChangedFiles: []string{validFile, nonExistentFile},
		WorkDir:      tmpDir,
	}

	result := grader.Grade(input)

	// Should still pass because valid file has endpoints
	// Grader should continue processing other files when one file can't be read
	if !result.Passed {
		t.Errorf("Expected Passed to be true when at least one file is readable, details: %s", result.Details)
	}
	if !strings.Contains(result.Details, "GET /valid-endpoint") {
		t.Error("Expected Details to contain endpoint from readable file")
	}
}
