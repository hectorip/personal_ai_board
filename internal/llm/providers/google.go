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

// GoogleProvider implements the LLM provider interface for Google Gemini
type GoogleProvider struct {
	config     types.Config
	httpClient *http.Client
	logger     types.Logger
}

// GoogleRequest represents a request to the Google Gemini API
type GoogleRequest struct {
	Contents          []GoogleContent        `json:"contents"`
	SystemInstruction *GoogleContent         `json:"systemInstruction,omitempty"`
	GenerationConfig  GoogleGenerationConfig `json:"generationConfig"`
	SafetySettings    []GoogleSafetySetting  `json:"safetySettings,omitempty"`
}

// GoogleContent represents content in the Google Gemini format
type GoogleContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GooglePart `json:"parts"`
}

// GooglePart represents a part of content
type GooglePart struct {
	Text string `json:"text"`
}

// GoogleGenerationConfig represents generation configuration
type GoogleGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// GoogleSafetySetting represents safety settings
type GoogleSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GoogleResponse represents a response from the Google Gemini API
type GoogleResponse struct {
	Candidates     []GoogleCandidate `json:"candidates"`
	UsageMetadata  GoogleUsage       `json:"usageMetadata"`
	PromptFeedback GoogleFeedback    `json:"promptFeedback,omitempty"`
}

// GoogleCandidate represents a candidate response
type GoogleCandidate struct {
	Content       GoogleContent        `json:"content"`
	FinishReason  string               `json:"finishReason"`
	Index         int                  `json:"index"`
	SafetyRatings []GoogleSafetyRating `json:"safetyRatings,omitempty"`
}

// GoogleSafetyRating represents safety rating
type GoogleSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GoogleUsage represents token usage information
type GoogleUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GoogleFeedback represents prompt feedback
type GoogleFeedback struct {
	BlockReason   string               `json:"blockReason,omitempty"`
	SafetyRatings []GoogleSafetyRating `json:"safetyRatings,omitempty"`
}

// GoogleError represents an error response from Google
type GoogleError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// NewGoogleProvider creates a new Google Gemini provider
func NewGoogleProvider(config types.Config, logger types.Logger) (*GoogleProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Google API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com"
	}

	if config.Model == "" {
		config.Model = "gemini-1.5-pro"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &GoogleProvider{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// GenerateResponse implements the Provider interface
func (p *GoogleProvider) GenerateResponse(ctx context.Context, req types.Request) (*types.Response, error) {
	// Validate request
	if err := types.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Build Google request
	googleReq := p.buildGoogleRequest(req)

	// Make API call
	startTime := time.Now()
	googleResp, err := p.callGoogle(ctx, googleReq)
	if err != nil {
		return nil, err
	}
	duration := time.Since(startTime)

	// Convert response
	response := p.convertResponse(googleResp, duration)

	return response, nil
}

// buildGoogleRequest converts our request format to Google format
func (p *GoogleProvider) buildGoogleRequest(req types.Request) GoogleRequest {
	contents := []GoogleContent{
		{
			Role: "user",
			Parts: []GooglePart{
				{Text: req.Prompt},
			},
		},
	}

	// Temperature from request or config
	temperature := req.Temperature
	if temperature == 0 {
		temperature = p.config.Temperature
	}

	// Max tokens from request or config
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
	}
	if maxTokens == 0 {
		maxTokens = 2048 // Default for Gemini
	}

	generationConfig := GoogleGenerationConfig{
		Temperature:     temperature,
		MaxOutputTokens: maxTokens,
		TopP:            0.95,
		TopK:            40,
	}

	googleReq := GoogleRequest{
		Contents:         contents,
		GenerationConfig: generationConfig,
		SafetySettings:   p.getDefaultSafetySettings(),
	}

	// Add system instruction if provided
	if req.SystemMsg != "" {
		googleReq.SystemInstruction = &GoogleContent{
			Parts: []GooglePart{
				{Text: req.SystemMsg},
			},
		}
	}

	return googleReq
}

// getDefaultSafetySettings returns default safety settings
func (p *GoogleProvider) getDefaultSafetySettings() []GoogleSafetySetting {
	return []GoogleSafetySetting{
		{
			Category:  "HARM_CATEGORY_HARASSMENT",
			Threshold: "BLOCK_MEDIUM_AND_ABOVE",
		},
		{
			Category:  "HARM_CATEGORY_HATE_SPEECH",
			Threshold: "BLOCK_MEDIUM_AND_ABOVE",
		},
		{
			Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
			Threshold: "BLOCK_MEDIUM_AND_ABOVE",
		},
		{
			Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
			Threshold: "BLOCK_MEDIUM_AND_ABOVE",
		},
	}
}

// callGoogle makes the actual API call to Google Gemini
func (p *GoogleProvider) callGoogle(ctx context.Context, req GoogleRequest) (*GoogleResponse, error) {
	// Serialize request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	model := p.config.Model
	if model == "" {
		model = "gemini-1.5-pro"
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.config.BaseURL, model, p.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "personal-ai-board/1.0")

	p.logger.Debug("Making Google Gemini API call",
		"url", url,
		"model", model,
		"contents", len(req.Contents),
		"max_tokens", req.GenerationConfig.MaxOutputTokens,
		"temperature", req.GenerationConfig.Temperature,
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
		var googleErr GoogleError
		if err := json.Unmarshal(body, &googleErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("Google API error (%d): %s", googleErr.Error.Code, googleErr.Error.Message)
	}

	// Parse response
	var googleResp GoogleResponse
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate response
	if len(googleResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	// Check for prompt feedback issues
	if googleResp.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("prompt blocked: %s", googleResp.PromptFeedback.BlockReason)
	}

	return &googleResp, nil
}

// convertResponse converts Google response to our response format
func (p *GoogleProvider) convertResponse(googleResp *GoogleResponse, duration time.Duration) *types.Response {
	candidate := googleResp.Candidates[0]

	// Extract text content
	var content strings.Builder
	for _, part := range candidate.Content.Parts {
		content.WriteString(part.Text)
	}

	response := &types.Response{
		Content:      content.String(),
		TokensUsed:   googleResp.UsageMetadata.TotalTokenCount,
		Model:        p.config.Model,
		Duration:     duration,
		FinishReason: candidate.FinishReason,
		Usage: &types.TokenUsage{
			PromptTokens:     googleResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: googleResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      googleResp.UsageMetadata.TotalTokenCount,
		},
		Metadata: map[string]interface{}{
			"index":          candidate.Index,
			"safety_ratings": candidate.SafetyRatings,
		},
	}

	return response
}

// GetModelInfo implements the Provider interface
func (p *GoogleProvider) GetModelInfo() types.ModelInfo {
	// Model information based on the configured model
	modelInfo := types.ModelInfo{
		Provider:     "google",
		Name:         p.config.Model,
		Capabilities: []string{"chat", "completion", "system_messages", "multimodal"},
	}

	// Set model-specific information
	switch p.config.Model {
	case "gemini-1.5-pro", "gemini-1.5-pro-latest":
		modelInfo.MaxTokens = 8192
		modelInfo.ContextSize = 2097152 // 2M tokens
		modelInfo.CostPer1K = 0.0035    // Input cost per 1K tokens
	case "gemini-1.5-flash", "gemini-1.5-flash-latest":
		modelInfo.MaxTokens = 8192
		modelInfo.ContextSize = 1048576 // 1M tokens
		modelInfo.CostPer1K = 0.00035   // Input cost per 1K tokens
	case "gemini-1.0-pro", "gemini-1.0-pro-latest":
		modelInfo.MaxTokens = 2048
		modelInfo.ContextSize = 32768 // 32K tokens
		modelInfo.CostPer1K = 0.0005  // Input cost per 1K tokens
	case "gemini-1.0-pro-vision":
		modelInfo.MaxTokens = 2048
		modelInfo.ContextSize = 16384 // 16K tokens
		modelInfo.CostPer1K = 0.0025  // Input cost per 1K tokens
	default:
		// Default values for unknown models
		modelInfo.MaxTokens = 2048
		modelInfo.ContextSize = 32768
		modelInfo.CostPer1K = 0.0035
	}

	return modelInfo
}

// ValidateConfig implements the Provider interface
func (p *GoogleProvider) ValidateConfig() error {
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
		"gemini-1.5-pro",
		"gemini-1.5-pro-latest",
		"gemini-1.5-flash",
		"gemini-1.5-flash-latest",
		"gemini-1.0-pro",
		"gemini-1.0-pro-latest",
		"gemini-1.0-pro-vision",
	}

	isValid := false
	for _, validModel := range validModels {
		if p.config.Model == validModel {
			isValid = true
			break
		}
	}

	if !isValid {
		p.logger.Warn("Unknown Google model", "model", p.config.Model)
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
func (p *GoogleProvider) Name() string {
	return "google"
}

// TestConnection tests the connection to Google Gemini
func (p *GoogleProvider) TestConnection(ctx context.Context) error {
	testReq := types.Request{
		Prompt:      "Hello, this is a test.",
		Temperature: 0.1,
		MaxTokens:   10,
	}

	_, err := p.GenerateResponse(ctx, testReq)
	return err
}

// EstimateTokens provides a rough estimate of tokens for Google models
func (p *GoogleProvider) EstimateTokens(text string) int {
	// Google models typically use ~4 characters per token for English
	// This is a rough approximation
	baseEstimate := len(text) / 4

	// Add some overhead for special tokens
	return baseEstimate + 10
}

// CalculateCost estimates the cost of a request
func (p *GoogleProvider) CalculateCost(usage *types.TokenUsage) float64 {
	modelInfo := p.GetModelInfo()

	// Google has different pricing for input and output tokens
	// Input tokens
	inputCost := (float64(usage.PromptTokens) / 1000.0) * modelInfo.CostPer1K

	// Output tokens typically cost more (e.g., 2x for Gemini)
	outputCostMultiplier := 2.0
	outputCost := (float64(usage.CompletionTokens) / 1000.0) * modelInfo.CostPer1K * outputCostMultiplier

	return inputCost + outputCost
}
