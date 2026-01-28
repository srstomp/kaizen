package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client is the interface for interacting with the Anthropic API
type Client interface {
	// Complete sends a prompt to the LLM and returns the response
	Complete(ctx context.Context, prompt string, options ...CompletionOption) (string, error)
}

// anthropicClient implements the Client interface using the Anthropic SDK
type anthropicClient struct {
	client     anthropic.Client
	timeout    time.Duration
	maxRetries int
}

// ClientOption is a function that configures the client
type ClientOption func(*clientConfig)

// clientConfig holds configuration for the client
type clientConfig struct {
	timeout    time.Duration
	maxRetries int
}

// CompletionOption is a function that configures a completion request
type CompletionOption func(*completionConfig)

// completionConfig holds configuration for a completion request
type completionConfig struct {
	model     string
	maxTokens int64
}

// Default configuration values
const (
	defaultTimeout    = 60 * time.Second
	defaultMaxRetries = 3
	defaultModel      = "claude-opus-4"
	defaultMaxTokens  = 4096
)

// WithTimeout sets the client timeout duration
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retry attempts
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *clientConfig) {
		c.maxRetries = maxRetries
	}
}

// WithModel sets the model to use for completion
func WithModel(model string) CompletionOption {
	return func(c *completionConfig) {
		c.model = model
	}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(maxTokens int64) CompletionOption {
	return func(c *completionConfig) {
		c.maxTokens = maxTokens
	}
}

// NewClient creates a new Anthropic API client with the specified options
func NewClient(opts ...ClientOption) (Client, error) {
	// Check for API key in environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY environment variable is not set")
	}

	// Apply default configuration
	config := &clientConfig{
		timeout:    defaultTimeout,
		maxRetries: defaultMaxRetries,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(config)
	}

	// Create SDK client options
	sdkOpts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	// Add max retries option
	if config.maxRetries > 0 {
		sdkOpts = append(sdkOpts, option.WithMaxRetries(config.maxRetries))
	}

	// Create the underlying Anthropic SDK client
	sdkClient := anthropic.NewClient(sdkOpts...)

	return &anthropicClient{
		client:     sdkClient,
		timeout:    config.timeout,
		maxRetries: config.maxRetries,
	}, nil
}

// Complete sends a prompt to the Anthropic API and returns the response
func (c *anthropicClient) Complete(ctx context.Context, prompt string, options ...CompletionOption) (string, error) {
	// Validate prompt
	if prompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Apply default completion configuration
	config := &completionConfig{
		model:     defaultModel,
		maxTokens: defaultMaxTokens,
	}

	// Apply custom options
	for _, opt := range options {
		opt(config)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Retry logic with exponential backoff
	var lastErr error
	maxAttempts := c.maxRetries + 1 // retries + initial attempt

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check context before retry
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Exponential backoff: 1s, 2s, 4s
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		// Attempt API call
		response, err := c.makeAPICall(ctx, prompt, config)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on context errors
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// makeAPICall performs the actual API call to Anthropic
func (c *anthropicClient) makeAPICall(ctx context.Context, prompt string, config *completionConfig) (string, error) {
	// Build the message request
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(config.model),
		MaxTokens: config.maxTokens,
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					{
						OfText: &anthropic.TextBlockParam{
							Text: prompt,
						},
					},
				},
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	// Extract text from response
	if len(message.Content) == 0 {
		return "", errors.New("API returned empty response")
	}

	// Get the first text block from the response
	for _, block := range message.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", errors.New("API response contained no text content")
}
