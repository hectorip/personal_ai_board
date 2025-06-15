package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"personal-ai-board/internal/llm/types"
)

// AnthropicProvider implements the LLM provider interface for Anthropic Claude
type AnthropicProvider struct {
	config     types.Config
	httpClient *http.Client
	logger     types.Logger
}

// AnthropicRequest represents a request to the Anthropic API
type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Stream    bool               `json:"stream"`
	Stop      []string           `json:"stop_sequences,omitempty"`
}

// AnthropicMessage represents a message in the Anthropic format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents a response from the Anthropic API
type AnthropicResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []AnthropicContent `json:"content"`
	Model        string             `json:"model"`
	StopReason   string             `json:"stop_reason"`
	StopSequence string             `json:"stop_sequence"`
	Usage        AnthropicUsage     `json:"usage"`
}

// AnthropicContent represents content in the Anthropic response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicUsage represents token usage information
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicError represents an error response from Anthropic
type AnthropicError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config types.Config, logger types.Logger) (*AnthropicProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}

	if config.Model == "" {
		config.Model = "claude-3-sonnet-20240229"
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &AnthropicProvider{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// GenerateResponse implements the Provider interface
func (p *AnthropicProvider) GenerateResponse(ctx context.Context, req types.Request) (*types.Response, error) {
	// Validate request
	if err := types.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Build Anthropic request
	anthropicReq := p.buildAnthropicRequest(req)

	// Make API call
	startTime := time.Now()
	anthropicResp, err := p.callAnthropic(ctx, anthropicReq)
	if err != nil {
		return nil, err
	}
	duration := time.Since(startTime)

	// Convert response
	response := p.convertResponse(anthropicResp, duration)

	return response, nil
}

// buildAnthropicRequest converts our request format to Anthropic format
func (p *AnthropicProvider) buildAnthropicRequest(req types.Request) AnthropicRequest {
	messages := []AnthropicMessage{}

	// Add user message
	messages = append(messages, AnthropicMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	model := req.Model
	if model == "" {
		model = p.config.Model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
	}
	if maxTokens == 0 {
		maxTokens = 4096 // Default for Claude
	}

	return AnthropicRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: maxTokens,
		System:    req.SystemMsg,
		Stream:    false,
	}
}

// callAnthropic makes the actual API call to Anthropic
func (p *AnthropicProvider) callAnthropic(ctx context.Context, req AnthropicRequest) (*AnthropicResponse, error) {
	// Serialize request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/messages", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("User-Agent", "personal-ai-board/1.0")

	p.logger.Debug("Making Anthropic API call",
		"url", url,
		"model", req.Model,
		"messages", len(req.Messages),
		"max_tokens", req.MaxTokens,
	)

	// Make request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var anthropicErr AnthropicError
		if err := json.Unmarshal(body, &anthropicErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("Anthropic API error (%s): %s", anthropicErr.Error.Type, anthropicErr.Error.Message)
	}

	// Parse response
	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate response
	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	return &anthropicResp, nil
}

// convertResponse converts Anthropic response to our response format
func (p *AnthropicProvider) convertResponse(anthropicResp *AnthropicResponse, duration time.Duration) *types.Response {
	// Extract text content
	var content string
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	totalTokens := anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens

	response := &types.Response{
		Content:      content,
		TokensUsed:   totalTokens,
		Model:        anthropicResp.Model,
		Duration:     duration,
		FinishReason: anthropicResp.StopReason,
		Usage: &types.TokenUsage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      totalTokens,
		},
		Metadata: map[string]interface{}{
			"id":            anthropicResp.ID,
			"type":          anthropicResp.Type,
			"role":          anthropicResp.Role,
			"stop_sequence": anthropicResp.StopSequence,
		},
	}

	return response
}

// GetModelInfo implements the Provider interface
func (p *AnthropicProvider) GetModelInfo() types.ModelInfo {
	// Model information based on the configured model
	modelInfo := types.ModelInfo{
		Provider:     "anthropic",
		Name:         p.config.Model,
		Capabilities: []string{"chat", "completion", "system_messages"},
	}

	// Set model-specific information
	switch p.config.Model {
	case "claude-3-opus-20240229":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 200000
		modelInfo.CostPer1K = 0.015 // Input cost, output is higher
	case "claude-3-sonnet-20240229":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 200000
		modelInfo.CostPer1K = 0.003 // Input cost
	case "claude-3-haiku-20240307":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 200000
		modelInfo.CostPer1K = 0.00025 // Input cost
	case "claude-2.1", "claude-2.0":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 200000
		modelInfo.CostPer1K = 0.008 // Input cost
	case "claude-instant-1.2":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 100000
		modelInfo.CostPer1K = 0.0008 // Input cost
	default:
		// Default values for unknown models
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 200000
		modelInfo.CostPer1K = 0.003
	}

	return modelInfo
}

// ValidateConfig implements the Provider interface
func (p *AnthropicProvider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if p.config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if p.config.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Validate model name
	validModels := []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-2.1",
		"claude-2.0",
		"claude-instant-1.2",
	}

	isValid := false
	for _, validModel := range validModels {
		if p.config.Model == validModel {
			isValid = true
			break
		}
	}

	if !isValid {
		p.logger.Warn("Unknown Anthropic model", "model", p.config.Model)
	}

	// Validate max tokens
	if p.config.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	modelInfo := p.GetModelInfo()
	if p.config.MaxTokens > modelInfo.MaxTokens {
		return fmt.Errorf("max_tokens (%d) exceeds model limit (%d)", p.config.MaxTokens, modelInfo.MaxTokens)
	}

	return nil
}

// Name implements the Provider interface
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// TestConnection tests the connection to Anthropic
func (p *AnthropicProvider) TestConnection(ctx context.Context) error {
	testReq := types.Request{
		Prompt:    "Hello, this is a test.",
		MaxTokens: 10,
	}

	_, err := p.GenerateResponse(ctx, testReq)
	return err
}

// EstimateTokens provides a rough estimate of tokens for Anthropic models
func (p *AnthropicProvider) EstimateTokens(text string) int {
	// Anthropic models typically use ~4 characters per token for English
	// This is a rough approximation
	baseEstimate := len(text) / 4

	// Add some overhead for special tokens
	return baseEstimate + 10
}

// CalculateCost estimates the cost of a request
func (p *AnthropicProvider) CalculateCost(usage *types.TokenUsage) float64 {
	modelInfo := p.GetModelInfo()

	// Anthropic has different pricing for input and output tokens
	// For simplicity, we'll use the input token cost as base
	// In production, you'd want separate input/output pricing
	inputCost := (float64(usage.PromptTokens) / 1000.0) * modelInfo.CostPer1K

	// Output tokens typically cost more (e.g., 3x for Claude-3)
	outputCostMultiplier := 3.0
	outputCost := (float64(usage.CompletionTokens) / 1000.0) * modelInfo.CostPer1K * outputCostMultiplier

	return inputCost + outputCost
}
