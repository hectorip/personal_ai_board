package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"personal-ai-board/internal/llm"
	"personal-ai-board/internal/llm/types"
	"personal-ai-board/pkg/logger"
)

// Example demonstrating how to use all three LLM providers:
// OpenAI, Anthropic, and Google Gemini

func main() {
	ctx := context.Background()

	// Initialize logger
	logger := logger.New("info")

	// Create LLM manager
	manager := llm.NewManager(logger)

	// Setup providers
	if err := setupProviders(manager, logger); err != nil {
		log.Fatalf("Failed to setup providers: %v", err)
	}

	// Test prompt
	prompt := "What are the key considerations when building a personal AI advisory board system?"

	// Test each provider
	testOpenAI(ctx, manager, prompt)
	testAnthropic(ctx, manager, prompt)
	testGoogle(ctx, manager, prompt)

	// Demonstrate provider switching
	demonstrateProviderSwitching(ctx, manager)

	// Demonstrate different models
	demonstrateDifferentModels(ctx, manager, logger)
}

func setupProviders(manager *llm.Manager, logger types.Logger) error {
	// Setup OpenAI provider
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config := types.Config{
			Provider:    "openai",
			APIKey:      apiKey,
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   500,
			Timeout:     30 * time.Second,
		}

		provider, err := llm.NewOpenAIProvider(config, logger)
		if err != nil {
			return fmt.Errorf("failed to create OpenAI provider: %w", err)
		}

		if err := manager.RegisterProvider("openai", provider); err != nil {
			return fmt.Errorf("failed to register OpenAI provider: %w", err)
		}

		fmt.Println("✓ OpenAI provider registered")
	} else {
		fmt.Println("⚠ OpenAI API key not found, skipping OpenAI provider")
	}

	// Setup Anthropic provider
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config := types.Config{
			Provider:    "anthropic",
			APIKey:      apiKey,
			BaseURL:     "https://api.anthropic.com",
			Model:       "claude-3-sonnet-20240229",
			Temperature: 0.7,
			MaxTokens:   500,
			Timeout:     60 * time.Second,
		}

		provider, err := llm.NewAnthropicProvider(config, logger)
		if err != nil {
			return fmt.Errorf("failed to create Anthropic provider: %w", err)
		}

		if err := manager.RegisterProvider("anthropic", provider); err != nil {
			return fmt.Errorf("failed to register Anthropic provider: %w", err)
		}

		fmt.Println("✓ Anthropic provider registered")
	} else {
		fmt.Println("⚠ Anthropic API key not found, skipping Anthropic provider")
	}

	// Setup Google provider
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		config := types.Config{
			Provider:    "google",
			APIKey:      apiKey,
			BaseURL:     "https://generativelanguage.googleapis.com",
			Model:       "gemini-1.5-pro",
			Temperature: 0.7,
			MaxTokens:   500,
			Timeout:     30 * time.Second,
		}

		provider, err := llm.NewGoogleProvider(config, logger)
		if err != nil {
			return fmt.Errorf("failed to create Google provider: %w", err)
		}

		if err := manager.RegisterProvider("google", provider); err != nil {
			return fmt.Errorf("failed to register Google provider: %w", err)
		}

		fmt.Println("✓ Google provider registered")
	} else {
		fmt.Println("⚠ Google API key not found, skipping Google provider")
	}

	return nil
}

func testOpenAI(ctx context.Context, manager *llm.Manager, prompt string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TESTING OPENAI PROVIDER")
	fmt.Println(strings.Repeat("=", 60))

	if !isProviderAvailable(manager, "openai") {
		fmt.Println("OpenAI provider not available")
		return
	}

	request := types.Request{
		Prompt:      prompt,
		SystemMsg:   "You are a helpful AI assistant specializing in software architecture and AI systems.",
		Temperature: 0.7,
		MaxTokens:   300,
	}

	response, err := manager.GenerateResponse(ctx, "openai", request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Model: %s\n", response.Model)
	fmt.Printf("Tokens Used: %d\n", response.TokensUsed)
	fmt.Printf("Duration: %v\n", response.Duration)
	fmt.Printf("Response:\n%s\n", response.Content)
}

func testAnthropic(ctx context.Context, manager *llm.Manager, prompt string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TESTING ANTHROPIC PROVIDER")
	fmt.Println(strings.Repeat("=", 60))

	if !isProviderAvailable(manager, "anthropic") {
		fmt.Println("Anthropic provider not available")
		return
	}

	request := types.Request{
		Prompt:      prompt,
		SystemMsg:   "You are Claude, an AI assistant created by Anthropic. You're helpful, harmless, and honest.",
		Temperature: 0.6,
		MaxTokens:   300,
	}

	response, err := manager.GenerateResponse(ctx, "anthropic", request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Model: %s\n", response.Model)
	fmt.Printf("Tokens Used: %d\n", response.TokensUsed)
	fmt.Printf("Duration: %v\n", response.Duration)
	fmt.Printf("Response:\n%s\n", response.Content)
}

func testGoogle(ctx context.Context, manager *llm.Manager, prompt string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TESTING GOOGLE PROVIDER")
	fmt.Println(strings.Repeat("=", 60))

	if !isProviderAvailable(manager, "google") {
		fmt.Println("Google provider not available")
		return
	}

	request := types.Request{
		Prompt:      prompt,
		SystemMsg:   "You are Gemini, Google's advanced AI assistant. Be helpful and comprehensive in your responses.",
		Temperature: 0.8,
		MaxTokens:   300,
	}

	response, err := manager.GenerateResponse(ctx, "google", request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Model: %s\n", response.Model)
	fmt.Printf("Tokens Used: %d\n", response.TokensUsed)
	fmt.Printf("Duration: %v\n", response.Duration)
	fmt.Printf("Response:\n%s\n", response.Content)
}

func demonstrateProviderSwitching(ctx context.Context, manager *llm.Manager) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("DEMONSTRATING PROVIDER SWITCHING")
	fmt.Println(strings.Repeat("=", 60))

	providers := manager.ListProviders()
	if len(providers) == 0 {
		fmt.Println("No providers available for switching demo")
		return
	}

	prompt := "Give me a brief creative idea for an AI-powered productivity tool."

	for _, providerName := range providers {
		fmt.Printf("\n--- Using %s ---\n", providerName)

		request := types.Request{
			Prompt:      prompt,
			Temperature: 0.9, // High creativity
			MaxTokens:   150,
		}

		response, err := manager.GenerateResponse(ctx, providerName, request)
		if err != nil {
			fmt.Printf("Error with %s: %v\n", providerName, err)
			continue
		}

		fmt.Printf("Response: %s\n", response.Content)
		fmt.Printf("Tokens: %d, Duration: %v\n", response.TokensUsed, response.Duration)
	}
}

func demonstrateDifferentModels(ctx context.Context, manager *llm.Manager, logger types.Logger) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("DEMONSTRATING DIFFERENT MODELS")
	fmt.Println(strings.Repeat("=", 60))

	// Test different models if available
	modelConfigs := []struct {
		provider string
		model    string
		apiKey   string
	}{
		{"openai", "gpt-3.5-turbo", os.Getenv("OPENAI_API_KEY")},
		{"anthropic", "claude-3-haiku-20240307", os.Getenv("ANTHROPIC_API_KEY")},
		{"google", "gemini-1.5-flash", os.Getenv("GOOGLE_API_KEY")},
	}

	prompt := "Explain machine learning in one sentence."

	for _, config := range modelConfigs {
		if config.apiKey == "" {
			continue
		}

		fmt.Printf("\n--- Testing %s with %s ---\n", config.provider, config.model)

		// Create a temporary provider with the specific model
		var provider types.Provider
		var err error

		switch config.provider {
		case "openai":
			provider, err = llm.NewOpenAIProvider(types.Config{
				Provider:  "openai",
				APIKey:    config.apiKey,
				BaseURL:   "https://api.openai.com/v1",
				Model:     config.model,
				MaxTokens: 100,
				Timeout:   30 * time.Second,
			}, logger)
		case "anthropic":
			provider, err = llm.NewAnthropicProvider(types.Config{
				Provider:  "anthropic",
				APIKey:    config.apiKey,
				BaseURL:   "https://api.anthropic.com",
				Model:     config.model,
				MaxTokens: 100,
				Timeout:   60 * time.Second,
			}, logger)
		case "google":
			provider, err = llm.NewGoogleProvider(types.Config{
				Provider:  "google",
				APIKey:    config.apiKey,
				BaseURL:   "https://generativelanguage.googleapis.com",
				Model:     config.model,
				MaxTokens: 100,
				Timeout:   30 * time.Second,
			}, logger)
		}

		if err != nil {
			fmt.Printf("Error creating provider: %v\n", err)
			continue
		}

		request := types.Request{
			Prompt:    prompt,
			MaxTokens: 100,
		}

		response, err := provider.GenerateResponse(ctx, request)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Response: %s\n", response.Content)
		fmt.Printf("Model: %s, Tokens: %d\n", response.Model, response.TokensUsed)

		// Show model info
		modelInfo := provider.GetModelInfo()
		fmt.Printf("Model Info - Max Tokens: %d, Context Size: %d, Cost per 1K: $%.4f\n",
			modelInfo.MaxTokens, modelInfo.ContextSize, modelInfo.CostPer1K)
	}
}

func isProviderAvailable(manager *llm.Manager, providerName string) bool {
	providers := manager.ListProviders()
	for _, p := range providers {
		if p == providerName {
			return true
		}
	}
	return false
}

// Example usage:
// 1. Set environment variables:
//    export OPENAI_API_KEY="your_openai_key"
//    export ANTHROPIC_API_KEY="your_anthropic_key"
//    export GOOGLE_API_KEY="your_google_key"
//
// 2. Run the example:
//    go run examples/providers_example.go
//
// This will demonstrate:
// - Setting up multiple LLM providers
// - Making requests to each provider
// - Comparing responses across providers
// - Switching between providers dynamically
// - Using different models within the same provider
