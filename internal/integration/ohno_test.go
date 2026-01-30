package integration

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// Mock executor for testing
type mockExecExecutor struct {
	lookPathFail  bool
	executeFail   bool
	output        string
	commandCalled []string
	argsCalled    []string
}

func (m *mockExecExecutor) Execute(name string, args ...string) ([]byte, error) {
	m.commandCalled = append(m.commandCalled, name)
	m.argsCalled = args
	if m.executeFail {
		return nil, errors.New("command failed")
	}
	return []byte(m.output), nil
}

func (m *mockExecExecutor) LookPath(name string) (string, error) {
	if m.lookPathFail {
		return "", exec.ErrNotFound
	}
	return "/usr/local/bin/" + name, nil
}

func TestCreateFixTask_Success(t *testing.T) {
	mock := &mockExecExecutor{
		output: "Created task: FIX-123",
	}

	client := &OhnoClient{
		executor: mock,
	}

	params := CreateFixTaskParams{
		Title:       "Fix: Add tests for task-123",
		TaskType:    "bug",
		BlocksTask:  "task-123",
		Description: "Quality review failed for task-123",
		Source:      "kaizen-fix",
	}

	taskID, err := client.CreateFixTask(params)
	if err != nil {
		t.Fatalf("CreateFixTask failed: %v", err)
	}

	if taskID != "FIX-123" {
		t.Errorf("taskID = %s, expected FIX-123", taskID)
	}

	// Verify command was called
	if len(mock.commandCalled) != 1 {
		t.Fatalf("expected 1 command call, got %d", len(mock.commandCalled))
	}

	if mock.commandCalled[0] != "ohno" {
		t.Errorf("command = %s, expected ohno", mock.commandCalled[0])
	}

	// Verify arguments
	expectedArgs := []string{
		"create",
		"Fix: Add tests for task-123",
		"--type", "bug",
		"--blocks", "task-123",
		"--description", "Quality review failed for task-123",
		"--source", "kaizen-fix",
	}

	if len(mock.argsCalled) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.argsCalled))
	}

	for i, expected := range expectedArgs {
		if mock.argsCalled[i] != expected {
			t.Errorf("arg[%d] = %s, expected %s", i, mock.argsCalled[i], expected)
		}
	}
}

func TestCreateFixTask_OhnoNotInstalled(t *testing.T) {
	mock := &mockExecExecutor{
		lookPathFail: true,
	}

	client := &OhnoClient{
		executor: mock,
	}

	params := CreateFixTaskParams{
		Title:       "Fix: Add tests for task-123",
		TaskType:    "bug",
		BlocksTask:  "task-123",
		Description: "Quality review failed",
		Source:      "kaizen-fix",
	}

	_, err := client.CreateFixTask(params)
	if err == nil {
		t.Error("CreateFixTask should fail when ohno is not installed")
	}

	if !strings.Contains(err.Error(), "ohno CLI not found") {
		t.Errorf("error should mention ohno not found, got: %v", err)
	}
}

func TestCreateFixTask_CommandFails(t *testing.T) {
	// LookPath succeeds, but Execute fails
	mock := &mockExecExecutor{
		executeFail: true,
	}

	client := &OhnoClient{
		executor: mock,
	}

	params := CreateFixTaskParams{
		Title:       "Fix: Add tests for task-123",
		TaskType:    "bug",
		BlocksTask:  "task-123",
		Description: "Quality review failed",
		Source:      "kaizen-fix",
	}

	_, err := client.CreateFixTask(params)
	if err == nil {
		t.Error("CreateFixTask should fail when command execution fails")
	}

	if !strings.Contains(err.Error(), "command failed") {
		t.Errorf("error should mention command failed, got: %v", err)
	}
}

func TestCreateFixTask_MinimalParams(t *testing.T) {
	mock := &mockExecExecutor{
		output: "Created task: FIX-456",
	}

	client := &OhnoClient{
		executor: mock,
	}

	params := CreateFixTaskParams{
		Title:      "Fix: Minimal task",
		TaskType:   "bug",
		BlocksTask: "task-456",
	}

	taskID, err := client.CreateFixTask(params)
	if err != nil {
		t.Fatalf("CreateFixTask failed: %v", err)
	}

	if taskID != "FIX-456" {
		t.Errorf("taskID = %s, expected FIX-456", taskID)
	}

	// Verify minimal args (no description or source)
	expectedArgs := []string{
		"create",
		"Fix: Minimal task",
		"--type", "bug",
		"--blocks", "task-456",
	}

	if len(mock.argsCalled) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d", len(expectedArgs), len(mock.argsCalled))
	}
}

func TestCreateFixTask_ParsesTaskID(t *testing.T) {
	testCases := []struct {
		name     string
		output   string
		expected string
		wantErr  bool
	}{
		{
			name:     "standard format",
			output:   "Created task: FIX-123",
			expected: "FIX-123",
			wantErr:  false,
		},
		{
			name:     "with extra text",
			output:   "Successfully created task: TASK-999 and added to backlog",
			expected: "TASK-999",
			wantErr:  false,
		},
		{
			name:     "multiline output",
			output:   "Processing...\nCreated task: BUG-42\nDone!",
			expected: "BUG-42",
			wantErr:  false,
		},
		{
			name:     "no task ID in output",
			output:   "Something went wrong",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockExecExecutor{
				output: tc.output,
			}

			client := &OhnoClient{
				executor: mock,
			}

			params := CreateFixTaskParams{
				Title:      "Test task",
				TaskType:   "bug",
				BlocksTask: "task-1",
			}

			taskID, err := client.CreateFixTask(params)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if taskID != tc.expected {
					t.Errorf("taskID = %s, expected %s", taskID, tc.expected)
				}
			}
		})
	}
}

func TestCreateFixTask_ValidationErrors(t *testing.T) {
	mock := &mockExecExecutor{
		output: "Created task: FIX-123",
	}

	client := &OhnoClient{
		executor: mock,
	}

	testCases := []struct {
		name   string
		params CreateFixTaskParams
	}{
		{
			name: "missing title",
			params: CreateFixTaskParams{
				Title:      "",
				TaskType:   "bug",
				BlocksTask: "task-1",
			},
		},
		{
			name: "missing task type",
			params: CreateFixTaskParams{
				Title:      "Fix something",
				TaskType:   "",
				BlocksTask: "task-1",
			},
		},
		{
			name: "missing blocks task",
			params: CreateFixTaskParams{
				Title:      "Fix something",
				TaskType:   "bug",
				BlocksTask: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.CreateFixTask(tc.params)
			if err == nil {
				t.Error("expected validation error but got none")
			}
		})
	}
}

func TestNewOhnoClient(t *testing.T) {
	client := NewOhnoClient()

	if client == nil {
		t.Fatal("NewOhnoClient returned nil")
	}

	if client.executor == nil {
		t.Error("client executor should not be nil")
	}
}
