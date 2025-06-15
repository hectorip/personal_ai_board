package types

import (
	"fmt"
	"strings"
)

// ValidateRequest validates an LLM request
func ValidateRequest(req Request) error {
	if strings.TrimSpace(req.Prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if req.Temperature < 0 || req.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if req.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	if req.MaxTokens > 32768 { // Reasonable upper limit
		return fmt.Errorf("max_tokens too large: %d", req.MaxTokens)
	}

	return nil
}

// ValidateConfig validates a provider configuration
func ValidateConfig(config Config) error {
	if strings.TrimSpace(config.Provider) == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if strings.TrimSpace(config.APIKey) == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if strings.TrimSpace(config.Model) == "" {
		return fmt.Errorf("model cannot be empty")
	}

	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if config.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		BaseDelay:     1000000000,  // 1 second in nanoseconds
		MaxDelay:      30000000000, // 30 seconds in nanoseconds
		BackoffFactor: 2.0,
	}
}

// IsRetryableError determines if an error is worth retrying
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Network errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") {
		return true
	}

	// Rate limit errors
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429") {
		return true
	}

	// Server errors (5xx)
	if strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") {
		return true
	}

	// Temporary service issues
	if strings.Contains(errStr, "temporarily unavailable") ||
		strings.Contains(errStr, "service unavailable") {
		return true
	}

	return false
}
