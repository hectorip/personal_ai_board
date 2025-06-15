package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"personal-ai-board/internal/llm/types"
)

// OpenAIProvider implements the LLM provider interface for OpenAI
type OpenAIProvider struct {
	config     types.Config
	httpClient *http.Client
	logger     types.Logger
}

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream"`
	Stop        []string        `json:"stop,omitempty"`
	User        string          `json:"user,omitempty"`
}

// OpenAIMessage represents a message in the OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents a response from the OpenAI API
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

// OpenAIChoice represents a choice in the OpenAI response
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage represents token usage information
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError represents an error response from OpenAI
type OpenAIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config types.Config, logger types.Logger) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Model == "" {
		config.Model = "gpt-4"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &OpenAIProvider{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// GenerateResponse implements the Provider interface
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, req types.Request) (*types.Response, error) {
	// Validate request
	if err := types.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Build OpenAI request
	openaiReq := p.buildOpenAIRequest(req)

	// Make API call
	startTime := time.Now()
	openaiResp, err := p.callOpenAI(ctx, openaiReq)
	if err != nil {
		return nil, err
	}
	duration := time.Since(startTime)

	// Convert response
	response := p.convertResponse(openaiResp, duration)

	return response, nil
}

// buildOpenAIRequest converts our request format to OpenAI format
func (p *OpenAIProvider) buildOpenAIRequest(req types.Request) OpenAIRequest {
	messages := []OpenAIMessage{}

	// Add system message if present
	if req.SystemMsg != "" {
		messages = append(messages, OpenAIMessage{
			Role:    "system",
			Content: req.SystemMsg,
		})
	}

	// Add user message
	messages = append(messages, OpenAIMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	model := req.Model
	if model == "" {
		model = p.config.Model
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = p.config.Temperature
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
	}

	return OpenAIRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}
}

// callOpenAI makes the actual API call to OpenAI
func (p *OpenAIProvider) callOpenAI(ctx context.Context, req OpenAIRequest) (*OpenAIResponse, error) {
	// Serialize request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))
	httpReq.Header.Set("User-Agent", "personal-ai-board/1.0")

	p.logger.Debug("Making OpenAI API call",
		"url", url,
		"model", req.Model,
		"messages", len(req.Messages),
		"temperature", req.Temperature,
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
		var openaiErr OpenAIError
		if err := json.Unmarshal(body, &openaiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("OpenAI API error (%s): %s", openaiErr.Error.Type, openaiErr.Error.Message)
	}

	// Parse response
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate response
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &openaiResp, nil
}

// convertResponse converts OpenAI response to our response format
func (p *OpenAIProvider) convertResponse(openaiResp *OpenAIResponse, duration time.Duration) *types.Response {
	choice := openaiResp.Choices[0]

	response := &types.Response{
		Content:      choice.Message.Content,
		TokensUsed:   openaiResp.Usage.TotalTokens,
		Model:        openaiResp.Model,
		Duration:     duration,
		FinishReason: choice.FinishReason,
		Usage: &types.TokenUsage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
		Metadata: map[string]interface{}{
			"id":      openaiResp.ID,
			"object":  openaiResp.Object,
			"created": openaiResp.Created,
		},
	}

	return response
}

// GetModelInfo implements the Provider interface
func (p *OpenAIProvider) GetModelInfo() types.ModelInfo {
	// Model information based on the configured model
	modelInfo := types.ModelInfo{
		Provider:     "openai",
		Name:         p.config.Model,
		Capabilities: []string{"chat", "completion", "system_messages"},
	}

	// Set model-specific information
	switch p.config.Model {
	case "gpt-4", "gpt-4-0613", "gpt-4-32k", "gpt-4-32k-0613":
		modelInfo.MaxTokens = 8192
		modelInfo.ContextSize = 8192
		modelInfo.CostPer1K = 0.03
		if strings.Contains(p.config.Model, "32k") {
			modelInfo.MaxTokens = 32768
			modelInfo.ContextSize = 32768
			modelInfo.CostPer1K = 0.06
		}
	case "gpt-4-turbo", "gpt-4-turbo-preview", "gpt-4-0125-preview", "gpt-4-1106-preview":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 128000
		modelInfo.CostPer1K = 0.01
	case "gpt-4o", "gpt-4o-2024-05-13":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 128000
		modelInfo.CostPer1K = 0.005
	case "gpt-3.5-turbo", "gpt-3.5-turbo-0125", "gpt-3.5-turbo-1106":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 16385
		modelInfo.CostPer1K = 0.0005
	case "gpt-3.5-turbo-16k", "gpt-3.5-turbo-16k-0613":
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 16385
		modelInfo.CostPer1K = 0.001
	default:
		// Default values for unknown models
		modelInfo.MaxTokens = 4096
		modelInfo.ContextSize = 8192
		modelInfo.CostPer1K = 0.02
	}

	return modelInfo
}

// ValidateConfig implements the Provider interface
func (p *OpenAIProvider) ValidateConfig() error {
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
		"gpt-4", "gpt-4-0613", "gpt-4-32k", "gpt-4-32k-0613",
		"gpt-4-turbo", "gpt-4-turbo-preview", "gpt-4-0125-preview", "gpt-4-1106-preview",
		"gpt-4o", "gpt-4o-2024-05-13",
		"gpt-3.5-turbo", "gpt-3.5-turbo-0125", "gpt-3.5-turbo-1106",
		"gpt-3.5-turbo-16k", "gpt-3.5-turbo-16k-0613",
	}

	isValid := false
	for _, validModel := range validModels {
		if p.config.Model == validModel {
			isValid = true
			break
		}
	}

	if !isValid {
		p.logger.Warn("Unknown OpenAI model", "model", p.config.Model)
	}

	// Validate temperature
	if p.config.Temperature < 0 || p.config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
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
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// TestConnection tests the connection to OpenAI
func (p *OpenAIProvider) TestConnection(ctx context.Context) error {
	testReq := types.Request{
		Prompt:      "Hello, this is a test.",
		Temperature: 0.1,
		MaxTokens:   10,
	}

	_, err := p.GenerateResponse(ctx, testReq)
	return err
}

// GetAvailableModels retrieves the list of available models from OpenAI
func (p *OpenAIProvider) GetAvailableModels(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/models", p.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))
	req.Header.Set("User-Agent", "personal-ai-board/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]string, len(modelsResp.Data))
	for i, model := range modelsResp.Data {
		models[i] = model.ID
	}

	return models, nil
}

// EstimateTokens provides a rough estimate of tokens for OpenAI models
func (p *OpenAIProvider) EstimateTokens(text string) int {
	// Very rough approximation for GPT models
	// In production, you'd use tiktoken library

	// GPT models typically use ~4 characters per token for English
	baseEstimate := len(text) / 4

	// Add some overhead for special tokens
	return baseEstimate + 10
}

// CalculateCost estimates the cost of a request
func (p *OpenAIProvider) CalculateCost(usage *types.TokenUsage) float64 {
	modelInfo := p.GetModelInfo()

	// Calculate cost based on total tokens and model pricing
	totalTokens := float64(usage.TotalTokens)
	costPer1K := modelInfo.CostPer1K

	return (totalTokens / 1000.0) * costPer1K
}
