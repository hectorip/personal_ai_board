package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"personal-ai-board/internal/llm/types"
)

// Manager manages multiple LLM providers
type Manager struct {
	providers       map[string]types.Provider
	defaultProvider string
	logger          types.Logger
}

// NewManager creates a new LLM manager
func NewManager(logger types.Logger) *Manager {
	return &Manager{
		providers: make(map[string]types.Provider),
		logger:    logger,
	}
}

// RegisterProvider registers a new LLM provider
func (m *Manager) RegisterProvider(name string, provider types.Provider) error {
	if err := provider.ValidateConfig(); err != nil {
		return fmt.Errorf("provider validation failed: %w", err)
	}

	m.providers[name] = provider

	// Set as default if it's the first provider
	if m.defaultProvider == "" {
		m.defaultProvider = name
	}

	m.logger.Info("LLM provider registered", "provider", name)
	return nil
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (types.Provider, error) {
	if name == "" {
		name = m.defaultProvider
	}

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// SetDefaultProvider sets the default provider
func (m *Manager) SetDefaultProvider(name string) error {
	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	m.defaultProvider = name
	m.logger.Info("Default LLM provider set", "provider", name)
	return nil
}

// ListProviders returns all registered provider names
func (m *Manager) ListProviders() []string {
	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// GenerateResponse generates a response using the specified provider
func (m *Manager) GenerateResponse(ctx context.Context, providerName string, req types.Request) (*types.Response, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	m.logger.Debug("Generating LLM response",
		"provider", providerName,
		"model", req.Model,
		"temperature", req.Temperature,
		"max_tokens", req.MaxTokens,
		"prompt_length", len(req.Prompt),
	)

	resp, err := provider.GenerateResponse(ctx, req)
	if err != nil {
		m.logger.Error("LLM generation failed",
			"provider", providerName,
			"error", err,
			"duration", time.Since(startTime),
		)
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	resp.Duration = time.Since(startTime)

	m.logger.Debug("LLM response generated",
		"provider", providerName,
		"model", resp.Model,
		"tokens_used", resp.TokensUsed,
		"duration", resp.Duration,
		"finish_reason", resp.FinishReason,
		"response_length", len(resp.Content),
	)

	return resp, nil
}

// ProviderFactory creates providers based on configuration
type ProviderFactory struct {
	logger types.Logger
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(logger types.Logger) *ProviderFactory {
	return &ProviderFactory{logger: logger}
}

// CreateProvider creates a provider based on the configuration
func (f *ProviderFactory) CreateProvider(config types.Config) (types.Provider, error) {
	switch strings.ToLower(config.Provider) {
	case "openai":
		return NewOpenAIProvider(config, f.logger)
	case "anthropic":
		return NewAnthropicProvider(config, f.logger)
	case "google", "gemini":
		return NewGoogleProvider(config, f.logger)
	case "ollama":
		return NewOllamaProvider(config, f.logger)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// TokenCounter provides token counting utilities
type TokenCounter struct{}

// EstimateTokens provides a rough estimate of token count
// This is a simplified implementation - in production you'd use tiktoken or similar
func (tc *TokenCounter) EstimateTokens(text string) int {
	// Very rough approximation: ~4 characters per token for English
	return len(text) / 4
}

// EstimateRequestTokens estimates the total tokens for a request
func (tc *TokenCounter) EstimateRequestTokens(req types.Request) int {
	total := tc.EstimateTokens(req.Prompt)

	if req.SystemMsg != "" {
		total += tc.EstimateTokens(req.SystemMsg)
	}

	// Add some overhead for request formatting
	total += 50

	return total
}

// HealthChecker checks the health of LLM providers
type HealthChecker struct {
	manager *Manager
	logger  types.Logger
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(manager *Manager, logger types.Logger) *HealthChecker {
	return &HealthChecker{
		manager: manager,
		logger:  logger,
	}
}

// CheckProvider checks the health of a specific provider
func (hc *HealthChecker) CheckProvider(ctx context.Context, providerName string) types.ProviderStatus {
	status := types.ProviderStatus{
		Name:      providerName,
		LastCheck: time.Now(),
	}

	provider, err := hc.manager.GetProvider(providerName)
	if err != nil {
		status.Error = err.Error()
		return status
	}

	status.Model = provider.GetModelInfo()

	// Test with a simple request
	testReq := types.Request{
		Prompt:      "Hello",
		Temperature: 0.1,
		MaxTokens:   10,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = provider.GenerateResponse(ctx, testReq)
	if err != nil {
		status.Error = err.Error()
		hc.logger.Warn("Provider health check failed", "provider", providerName, "error", err)
	} else {
		status.Available = true
		hc.logger.Debug("Provider health check passed", "provider", providerName)
	}

	return status
}

// CheckAllProviders checks the health of all registered providers
func (hc *HealthChecker) CheckAllProviders(ctx context.Context) []types.ProviderStatus {
	providers := hc.manager.ListProviders()
	statuses := make([]types.ProviderStatus, len(providers))

	for i, providerName := range providers {
		statuses[i] = hc.CheckProvider(ctx, providerName)
	}

	return statuses
}

// RetryableProvider wraps a provider with retry logic
type RetryableProvider struct {
	provider types.Provider
	config   types.RetryConfig
	logger   types.Logger
}

// NewRetryableProvider creates a provider with retry capabilities
func NewRetryableProvider(provider types.Provider, config types.RetryConfig, logger types.Logger) *RetryableProvider {
	return &RetryableProvider{
		provider: provider,
		config:   config,
		logger:   logger,
	}
}

// GenerateResponse implements Provider interface with retry logic
func (rp *RetryableProvider) GenerateResponse(ctx context.Context, req types.Request) (*types.Response, error) {
	var lastErr error
	delay := rp.config.BaseDelay

	for attempt := 0; attempt <= rp.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}

			rp.logger.Debug("Retrying LLM request",
				"attempt", attempt+1,
				"max_attempts", rp.config.MaxRetries+1,
				"delay", delay,
			)

			// Exponential backoff
			delay = time.Duration(float64(delay) * rp.config.BackoffFactor)
			if delay > rp.config.MaxDelay {
				delay = rp.config.MaxDelay
			}
		}

		resp, err := rp.provider.GenerateResponse(ctx, req)
		if err == nil {
			if attempt > 0 {
				rp.logger.Info("LLM request succeeded after retries", "attempts", attempt+1)
			}
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if !types.IsRetryableError(err) {
			rp.logger.Debug("Non-retryable error encountered", "error", err)
			break
		}

		rp.logger.Warn("LLM request failed, will retry",
			"attempt", attempt+1,
			"error", err,
		)
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", rp.config.MaxRetries+1, lastErr)
}

// GetModelInfo implements Provider interface
func (rp *RetryableProvider) GetModelInfo() types.ModelInfo {
	return rp.provider.GetModelInfo()
}

// ValidateConfig implements Provider interface
func (rp *RetryableProvider) ValidateConfig() error {
	return rp.provider.ValidateConfig()
}

// Name implements Provider interface
func (rp *RetryableProvider) Name() string {
	return rp.provider.Name()
}
