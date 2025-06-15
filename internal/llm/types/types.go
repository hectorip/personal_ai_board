package types

import (
	"context"
	"time"
)

// Provider defines the interface for LLM providers
type Provider interface {
	GenerateResponse(ctx context.Context, req Request) (*Response, error)
	GetModelInfo() ModelInfo
	ValidateConfig() error
	Name() string
}

// Request represents a request to an LLM provider
type Request struct {
	Prompt      string                 `json:"prompt"`
	SystemMsg   string                 `json:"system_message"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Context     map[string]interface{} `json:"context"`
	Model       string                 `json:"model,omitempty"`
}

// Response represents the response from an LLM provider
type Response struct {
	Content      string                 `json:"content"`
	TokensUsed   int                    `json:"tokens_used"`
	Model        string                 `json:"model"`
	Duration     time.Duration          `json:"duration"`
	FinishReason string                 `json:"finish_reason"`
	Usage        *TokenUsage            `json:"usage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TokenUsage provides detailed token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo contains information about the LLM model
type ModelInfo struct {
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	MaxTokens    int      `json:"max_tokens"`
	ContextSize  int      `json:"context_size"`
	CostPer1K    float64  `json:"cost_per_1k"`
	Capabilities []string `json:"capabilities"`
}

// Config represents provider configuration
type Config struct {
	Provider    string                 `json:"provider"`
	APIKey      string                 `json:"api_key"`
	BaseURL     string                 `json:"base_url,omitempty"`
	Model       string                 `json:"model"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Timeout     time.Duration          `json:"timeout"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// Logger interface for LLM operations
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// ProviderStatus represents the status of a provider
type ProviderStatus struct {
	Name      string    `json:"name"`
	Available bool      `json:"available"`
	Error     string    `json:"error,omitempty"`
	LastCheck time.Time `json:"last_check"`
	Model     ModelInfo `json:"model"`
}

// RetryConfig defines retry behavior for LLM requests
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	BaseDelay     time.Duration `json:"base_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}
