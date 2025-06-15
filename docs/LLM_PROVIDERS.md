# LLM Providers Documentation

This document provides comprehensive information about the LLM providers supported by Personal AI Board.

## Overview

Personal AI Board supports multiple LLM providers to give you flexibility in choosing the best AI model for your specific needs. Each provider has unique strengths and characteristics that make them suitable for different types of analysis and interaction.

## Supported Providers

### 1. OpenAI (GPT Models)

**Best for:** General-purpose conversations, creative tasks, and balanced analysis

#### Configuration
```yaml
llm:
  openai:
    api_key: "${OPENAI_API_KEY}"
    base_url: "https://api.openai.com/v1"
    model: "gpt-4"
    temperature: 0.7
    max_tokens: 1000
    timeout: "30s"
```

#### Supported Models
- `gpt-4` - Most capable model, best reasoning
- `gpt-4-turbo` - Faster version of GPT-4
- `gpt-4o` - Optimized for efficiency
- `gpt-3.5-turbo` - Fast and cost-effective

#### Characteristics
- **Strengths:** Creative writing, code generation, conversational AI
- **Context Size:** 8K-128K tokens depending on model
- **Cost:** $0.0005-$0.06 per 1K tokens
- **Response Time:** Fast (1-3 seconds)

#### Environment Setup
```bash
export OPENAI_API_KEY="your_openai_api_key_here"
```

### 2. Anthropic (Claude Models)

**Best for:** Analytical thinking, research, and detailed reasoning

#### Configuration
```yaml
llm:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    base_url: "https://api.anthropic.com"
    model: "claude-3-sonnet-20240229"
    temperature: 0.7
    max_tokens: 1000
    timeout: "60s"
```

#### Supported Models
- `claude-3-opus-20240229` - Most capable, best for complex reasoning
- `claude-3-sonnet-20240229` - Balanced performance and cost
- `claude-3-haiku-20240307` - Fast and efficient
- `claude-2.1` - Previous generation, still very capable
- `claude-instant-1.2` - Fastest, most cost-effective

#### Characteristics
- **Strengths:** Long-form analysis, research, ethical reasoning
- **Context Size:** 100K-200K tokens
- **Cost:** $0.00025-$0.015 per 1K tokens
- **Response Time:** Moderate (2-5 seconds)

#### Environment Setup
```bash
export ANTHROPIC_API_KEY="your_anthropic_api_key_here"
```

### 3. Google (Gemini Models)

**Best for:** Multimodal tasks, factual information, and structured analysis

#### Configuration
```yaml
llm:
  google:
    api_key: "${GOOGLE_API_KEY}"
    base_url: "https://generativelanguage.googleapis.com"
    model: "gemini-1.5-pro"
    temperature: 0.7
    max_tokens: 1000
    timeout: "30s"
```

#### Supported Models
- `gemini-1.5-pro` - Most capable, large context window
- `gemini-1.5-flash` - Fast and efficient
- `gemini-1.0-pro` - General purpose model
- `gemini-1.0-pro-vision` - Supports image inputs

#### Characteristics
- **Strengths:** Large context windows, multimodal capabilities, factual accuracy
- **Context Size:** 32K-2M tokens depending on model
- **Cost:** $0.00035-$0.0035 per 1K tokens
- **Response Time:** Fast (1-3 seconds)

#### Environment Setup
```bash
export GOOGLE_API_KEY="your_google_api_key_here"
```

## Usage Examples

### Basic Provider Usage

```go
package main

import (
    "context"
    "personal-ai-board/internal/llm"
    "personal-ai-board/internal/llm/types"
)

func main() {
    ctx := context.Background()
    logger := logger.New("info")
    manager := llm.NewManager(logger)

    // Configure OpenAI
    openaiConfig := types.Config{
        Provider:    "openai",
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Model:       "gpt-4",
        Temperature: 0.7,
        MaxTokens:   500,
    }
    
    provider, _ := llm.NewOpenAIProvider(openaiConfig, logger)
    manager.RegisterProvider("openai", provider)

    // Make a request
    request := types.Request{
        Prompt:    "Explain quantum computing in simple terms",
        MaxTokens: 200,
    }
    
    response, _ := manager.GenerateResponse(ctx, "openai", request)
    fmt.Println(response.Content)
}
```

### Switching Between Providers

```go
// Test the same prompt with different providers
prompt := "What are the pros and cons of remote work?"

providers := []string{"openai", "anthropic", "google"}
for _, providerName := range providers {
    response, err := manager.GenerateResponse(ctx, providerName, types.Request{
        Prompt: prompt,
        MaxTokens: 300,
    })
    
    if err == nil {
        fmt.Printf("%s: %s\n", providerName, response.Content)
    }
}
```

## Provider Selection Guidelines

### When to Use OpenAI
- **Creative writing and brainstorming**
- **Code generation and debugging** 
- **Conversational personas**
- **General-purpose analysis**

### When to Use Anthropic
- **Deep analytical thinking**
- **Research and fact-checking**
- **Ethical reasoning and safety**
- **Long-form content analysis**

### When to Use Google
- **Large context requirements**
- **Factual information retrieval**
- **Structured data analysis**
- **Cost-sensitive applications**

## Cost Optimization

### Model Selection by Use Case

| Use Case | Recommended Models | Reasoning |
|----------|-------------------|-----------|
| Quick Q&A | `gpt-3.5-turbo`, `claude-3-haiku`, `gemini-1.5-flash` | Fast and cost-effective |
| Deep Analysis | `gpt-4`, `claude-3-opus`, `gemini-1.5-pro` | Best reasoning capabilities |
| Creative Tasks | `gpt-4`, `claude-3-sonnet` | Strong creative abilities |
| Large Documents | `claude-3-sonnet`, `gemini-1.5-pro` | Large context windows |

### Cost Management Tips

1. **Use appropriate models**: Don't use expensive models for simple tasks
2. **Optimize token usage**: Keep prompts concise but clear
3. **Set budget limits**: Configure daily/monthly spending limits
4. **Monitor usage**: Track token consumption and costs
5. **Implement fallbacks**: Use cheaper models as backups

## Error Handling

### Common Errors and Solutions

#### Rate Limiting
```go
// Implement retry logic with exponential backoff
retryConfig := types.RetryConfig{
    MaxRetries:    3,
    BaseDelay:     time.Second,
    MaxDelay:      30 * time.Second,
    BackoffFactor: 2.0,
}

retryableProvider := llm.NewRetryableProvider(provider, retryConfig, logger)
```

#### API Key Issues
- Ensure environment variables are set correctly
- Check API key permissions and quotas
- Verify API key format for each provider

#### Model Availability
```go
// Check if a model is available before using
modelInfo := provider.GetModelInfo()
if modelInfo.MaxTokens == 0 {
    // Model might not be available
    // Fall back to a different model
}
```

## Performance Optimization

### Connection Pooling
```go
config.Timeout = 30 * time.Second  // Reasonable timeout
```

### Concurrent Requests
```go
// Use goroutines for concurrent requests to different providers
var wg sync.WaitGroup
results := make(map[string]*types.Response)

for _, providerName := range providers {
    wg.Add(1)
    go func(name string) {
        defer wg.Done()
        resp, err := manager.GenerateResponse(ctx, name, request)
        if err == nil {
            results[name] = resp
        }
    }(providerName)
}

wg.Wait()
```

## Security Considerations

### API Key Management
- Store API keys in environment variables, not in code
- Use secrets management systems in production
- Rotate API keys regularly
- Monitor API key usage for suspicious activity

### Request Filtering
- Validate user inputs before sending to LLM providers
- Implement content filtering for sensitive information
- Log requests for audit purposes (without sensitive data)

## Monitoring and Observability

### Metrics to Track
- **Response times** per provider
- **Token usage** and costs
- **Error rates** and types
- **Model performance** comparisons

### Logging
```go
logger.Debug("LLM request", 
    "provider", providerName,
    "model", request.Model,
    "tokens_requested", request.MaxTokens,
    "temperature", request.Temperature,
)
```

## Troubleshooting

### Common Issues

1. **Slow responses**: Check network connectivity and provider status
2. **High costs**: Review model selection and token usage patterns
3. **Poor quality**: Adjust temperature and prompt engineering
4. **Rate limits**: Implement proper retry logic and request throttling

### Debug Mode
```bash
# Enable debug logging to see detailed request/response information
export LOG_LEVEL=debug
./personal-ai-board
```

## Future Providers

We plan to add support for additional providers:
- **Ollama** (local models)
- **Azure OpenAI** (enterprise deployments)
- **Cohere** (specialized language tasks)
- **Together AI** (open-source models)

## Contributing

To add a new LLM provider:

1. Create a new provider file in `internal/llm/providers/`
2. Implement the `types.Provider` interface
3. Add the provider to the factory in `llm.go`
4. Add configuration examples and documentation
5. Write tests for the new provider

See `internal/llm/providers/openai.go` as a reference implementation.

## Support

For issues with specific providers:
- **OpenAI**: Check [OpenAI Status](https://status.openai.com/)
- **Anthropic**: Check [Anthropic Status](https://status.anthropic.com/)
- **Google**: Check [Google Cloud Status](https://status.cloud.google.com/)

For Personal AI Board issues, please create an issue in the GitHub repository.