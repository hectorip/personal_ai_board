package config

import (
	"time"
)

// Config represents the application configuration
type Config struct {
	Database DatabaseConfig `json:"database" yaml:"database"`
	LLM      LLMConfig      `json:"llm" yaml:"llm"`
	Log      LogConfig      `json:"log" yaml:"log"`
	Analysis AnalysisConfig `json:"analysis" yaml:"analysis"`
	Memory   MemoryConfig   `json:"memory" yaml:"memory"`
	Web      WebConfig      `json:"web" yaml:"web"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Path              string        `json:"path" yaml:"path"`
	MaxOpenConns      int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns      int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime   time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime   time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	EnableWAL         bool          `json:"enable_wal" yaml:"enable_wal"`
	EnableForeignKeys bool          `json:"enable_foreign_keys" yaml:"enable_foreign_keys"`
}

// LLMConfig contains LLM provider configuration
type LLMConfig struct {
	DefaultProvider string                    `json:"default_provider" yaml:"default_provider"`
	DefaultModel    string                    `json:"default_model" yaml:"default_model"`
	Temperature     float64                   `json:"temperature" yaml:"temperature"`
	MaxTokens       int                       `json:"max_tokens" yaml:"max_tokens"`
	Timeout         time.Duration             `json:"timeout" yaml:"timeout"`
	OpenAI          OpenAIConfig              `json:"openai" yaml:"openai"`
	Anthropic       AnthropicConfig           `json:"anthropic" yaml:"anthropic"`
	Ollama          OllamaConfig              `json:"ollama" yaml:"ollama"`
	Retry           RetryConfig               `json:"retry" yaml:"retry"`
}

// OpenAIConfig contains OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey  string `json:"api_key" yaml:"api_key"`
	BaseURL string `json:"base_url" yaml:"base_url"`
}

// AnthropicConfig contains Anthropic-specific configuration
type AnthropicConfig struct {
	APIKey  string `json:"api_key" yaml:"api_key"`
	BaseURL string `json:"base_url" yaml:"base_url"`
}

// OllamaConfig contains Ollama-specific configuration
type OllamaConfig struct {
	BaseURL string `json:"base_url" yaml:"base_url"`
}

// RetryConfig contains retry configuration for LLM requests
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	BaseDelay     time.Duration `json:"base_delay" yaml:"base_delay"`
	MaxDelay      time.Duration `json:"max_delay" yaml:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor" yaml:"backoff_factor"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

// AnalysisConfig contains analysis configuration
type AnalysisConfig struct {
	MaxConcurrent int    `json:"max_concurrent" yaml:"max_concurrent"`
	DefaultMode   string `json:"default_mode" yaml:"default_mode"`
	Timeout       time.Duration `json:"timeout" yaml:"timeout"`
}

// MemoryConfig contains memory management configuration
type MemoryConfig struct {
	RetentionDays    int     `json:"retention_days" yaml:"retention_days"`
	ShortTermLimit   int     `json:"short_term_limit" yaml:"short_term_limit"`
	LongTermLimit    int     `json:"long_term_limit" yaml:"long_term_limit"`
	DecayRate        float64 `json:"decay_rate" yaml:"decay_rate"`
}

// WebConfig contains web server configuration
type WebConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	StaticDir    string        `json:"static_dir" yaml:"static_dir"`
	TemplateDir  string        `json:"template_dir" yaml:"template_dir"`
}

// Default returns a configuration with default values
func Default() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path:              "personal_ai_board.db",
			MaxOpenConns:      25,
			MaxIdleConns:      25,
			ConnMaxLifetime:   time.Hour,
			ConnMaxIdleTime:   time.Minute * 5,
			EnableWAL:         true,
			EnableForeignKeys: true,
		},
		LLM: LLMConfig{
			DefaultProvider: "openai",
			DefaultModel:    "gpt-4",
			Temperature:     0.7,
			MaxTokens:       1000,
			Timeout:         30 * time.Second,
			OpenAI: OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
			},
			Anthropic: AnthropicConfig{
				BaseURL: "https://api.anthropic.com/v1",
			},
			Ollama: OllamaConfig{
				BaseURL: "http://localhost:11434",
			},
			Retry: RetryConfig{
				MaxRetries:    3,
				BaseDelay:     time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Analysis: AnalysisConfig{
			MaxConcurrent: 5,
			DefaultMode:   "discussion",
			Timeout:       5 * time.Minute,
		},
		Memory: MemoryConfig{
			RetentionDays:  90,
			ShortTermLimit: 50,
			LongTermLimit:  200,
			DecayRate:      0.95,
		},
		Web: WebConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			StaticDir:    "web/static",
			TemplateDir:  "web/templates",
		},
	}
}