package llm

import (
	"fmt"
	"personal-ai-board/internal/llm/providers"
	"personal-ai-board/internal/llm/types"
)

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config types.Config, logger types.Logger) (types.Provider, error) {
	return providers.NewOpenAIProvider(config, logger)
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config types.Config, logger types.Logger) (types.Provider, error) {
	return providers.NewAnthropicProvider(config, logger)
}

// NewGoogleProvider creates a new Google Gemini provider
func NewGoogleProvider(config types.Config, logger types.Logger) (types.Provider, error) {
	return providers.NewGoogleProvider(config, logger)
}

// NewOllamaProvider creates a new Ollama provider (placeholder)
func NewOllamaProvider(config types.Config, logger types.Logger) (types.Provider, error) {
	// TODO: Implement Ollama provider
	return nil, fmt.Errorf("Ollama provider not yet implemented")
}
