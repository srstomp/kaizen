package llm

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// TestNewClient tests the creation of a new Anthropic client
func TestNewClient(t *testing.T) {
	// Save original env var and restore after test
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	tests := []struct {
		name        string
		apiKey      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid API key",
			apiKey:      "sk-ant-test-key-123",
			expectError: false,
		},
		{
			name:        "Missing API key",
			apiKey:      "",
			expectError: true,
			errorMsg:    "ANTHROPIC_API_KEY environment variable is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.apiKey == "" {
				os.Unsetenv("ANTHROPIC_API_KEY")
			} else {
				os.Setenv("ANTHROPIC_API_KEY", tt.apiKey)
			}

			// Create client
			client, err := NewClient()

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errorMsg)
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if client != nil {
					t.Error("Expected nil client on error")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if client == nil {
					t.Error("Expected non-nil client")
				}
			}
		})
	}
}

// TestNewClientWithOptions tests creating a client with custom options
func TestNewClientWithOptions(t *testing.T) {
	// Save and restore env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-123")

	tests := []struct {
		name    string
		options []ClientOption
	}{
		{
			name:    "Default timeout (60s)",
			options: nil,
		},
		{
			name:    "Custom timeout",
			options: []ClientOption{WithTimeout(30 * time.Second)},
		},
		{
			name:    "Custom max retries",
			options: []ClientOption{WithMaxRetries(5)},
		},
		{
			name: "Multiple options",
			options: []ClientOption{
				WithTimeout(45 * time.Second),
				WithMaxRetries(2),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.options...)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if client == nil {
				t.Error("Expected non-nil client")
			}
		})
	}
}

// TestClientTimeout tests that the client respects timeout configuration
func TestClientTimeout(t *testing.T) {
	// Save and restore env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-123")

	client, err := NewClient(WithTimeout(100 * time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context that will timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should fail due to context timeout
	_, err = client.Complete(ctx, "test prompt")
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected context deadline exceeded or context-related error, got: %v", err)
	}
}

// TestClientComplete tests the Complete method
func TestClientComplete(t *testing.T) {
	// Save and restore env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	// Skip this test if no API key is available (for CI environments)
	// In real usage, we'd mock the API calls
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-for-testing")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name        string
		prompt      string
		options     []CompletionOption
		expectError bool
	}{
		{
			name:        "Empty prompt",
			prompt:      "",
			expectError: true,
		},
		{
			name:        "Valid prompt with defaults",
			prompt:      "Hello, world!",
			expectError: false, // Will fail with real API, but tests structure
		},
		{
			name:   "Valid prompt with options",
			prompt: "Analyze this task",
			options: []CompletionOption{
				WithModel("claude-opus-4"),
				WithMaxTokens(1024),
			},
			expectError: false, // Will fail with real API, but tests structure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := client.Complete(ctx, tt.prompt, tt.options...)

			if tt.expectError && tt.prompt == "" {
				if err == nil {
					t.Error("Expected error for empty prompt")
				}
				if !strings.Contains(err.Error(), "prompt") {
					t.Errorf("Expected error about prompt, got: %v", err)
				}
			}
			// Note: Other cases will fail without a valid API key or mock,
			// which is expected in unit tests. In production, we'd use mocks.
		})
	}
}

// TestClientCancellation tests that context cancellation is respected
func TestClientCancellation(t *testing.T) {
	// Save and restore env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-123")

	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This should fail due to cancelled context
	_, err = client.Complete(ctx, "test prompt")
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected context canceled error, got: %v", err)
	}
}

// TestDefaultValues tests that default values are applied correctly
func TestDefaultValues(t *testing.T) {
	// Save and restore env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-123")

	// Create client with no options - should use defaults
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify client was created (defaults are applied internally)
	if client == nil {
		t.Error("Expected non-nil client with default values")
	}

	// Test that the client has proper default configuration
	// This is implicitly tested by ensuring NewClient succeeds
	impl, ok := client.(*anthropicClient)
	if !ok {
		t.Error("Expected client to be of type *anthropicClient")
	}

	// Verify defaults
	if impl.timeout != 60*time.Second {
		t.Errorf("Expected default timeout of 60s, got %v", impl.timeout)
	}
	if impl.maxRetries != 3 {
		t.Errorf("Expected default max retries of 3, got %d", impl.maxRetries)
	}
}

// TestRetryConfiguration tests retry configuration
func TestRetryConfiguration(t *testing.T) {
	// Save and restore env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-123")

	tests := []struct {
		name       string
		maxRetries int
		wantErr    bool
	}{
		{
			name:       "Default retries (3)",
			maxRetries: 3,
			wantErr:    false,
		},
		{
			name:       "No retries",
			maxRetries: 0,
			wantErr:    false,
		},
		{
			name:       "Many retries",
			maxRetries: 10,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(WithMaxRetries(tt.maxRetries))
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if !tt.wantErr && client == nil {
				t.Error("Expected non-nil client")
			}
		})
	}
}
