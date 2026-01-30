package integration

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// CreateFixTaskParams contains parameters for creating a fix task in ohno
type CreateFixTaskParams struct {
	Title       string // Task title
	TaskType    string // "bug", "feature", etc.
	BlocksTask  string // Task ID this blocks
	Description string // Optional detailed description
	Source      string // Optional source label (e.g., "kaizen-fix")
}

// execExecutor defines the interface for executing commands
// This allows us to mock exec.Command for testing
type execExecutor interface {
	Execute(name string, args ...string) ([]byte, error)
	LookPath(name string) (string, error)
}

// realExecExecutor implements execExecutor using real os/exec
type realExecExecutor struct{}

func (r *realExecExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

func (r *realExecExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// OhnoClient wraps the ohno CLI for creating tasks
type OhnoClient struct {
	executor execExecutor
}

// NewOhnoClient creates a new OhnoClient
func NewOhnoClient() *OhnoClient {
	return &OhnoClient{
		executor: &realExecExecutor{},
	}
}

// CreateFixTask creates a fix task in ohno and returns the task ID
func (c *OhnoClient) CreateFixTask(params CreateFixTaskParams) (string, error) {
	// Validate required parameters
	if err := c.validateParams(params); err != nil {
		return "", err
	}

	// Check if ohno is installed
	if _, err := c.executor.LookPath("ohno"); err != nil {
		return "", fmt.Errorf("ohno CLI not found: %w", err)
	}

	// Build command arguments
	args := c.buildArgs(params)

	// Execute ohno command
	output, err := c.executor.Execute("ohno", args...)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Parse task ID from output
	taskID, err := c.parseTaskID(string(output))
	if err != nil {
		return "", fmt.Errorf("failed to parse task ID from output: %w", err)
	}

	return taskID, nil
}

// validateParams validates the required parameters
func (c *OhnoClient) validateParams(params CreateFixTaskParams) error {
	if params.Title == "" {
		return fmt.Errorf("title is required")
	}
	if params.TaskType == "" {
		return fmt.Errorf("task type is required")
	}
	if params.BlocksTask == "" {
		return fmt.Errorf("blocks task is required")
	}
	return nil
}

// buildArgs builds the command arguments for ohno create
func (c *OhnoClient) buildArgs(params CreateFixTaskParams) []string {
	args := []string{
		"create",
		params.Title,
		"--type", params.TaskType,
		"--blocks", params.BlocksTask,
	}

	// Add optional parameters if provided
	if params.Description != "" {
		args = append(args, "--description", params.Description)
	}

	if params.Source != "" {
		args = append(args, "--source", params.Source)
	}

	return args
}

// parseTaskID extracts the task ID from ohno command output
// Expected format: "Created task: TASK-123" or similar
func (c *OhnoClient) parseTaskID(output string) (string, error) {
	// Look for pattern like "TASK-123", "FIX-456", "BUG-789", etc.
	// Task IDs are typically: uppercase letters, dash, numbers
	re := regexp.MustCompile(`\b([A-Z]+-\d+)\b`)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 2 {
		return "", fmt.Errorf("no task ID found in output: %s", strings.TrimSpace(output))
	}

	return matches[1], nil
}
