package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	LLM      LLMConfig      `yaml:"llm"`
	Log      LogConfig      `yaml:"log"`
	Analysis AnalysisConfig `yaml:"analysis"`
	Memory   MemoryConfig   `yaml:"memory"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path              string `yaml:"path"`
	MaxOpenConns      int    `yaml:"max_open_conns"`
	MaxIdleConns      int    `yaml:"max_idle_conns"`
	EnableWAL         bool   `yaml:"enable_wal"`
	EnableForeignKeys bool   `yaml:"enable_foreign_keys"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	DefaultProvider string                 `yaml:"default_provider"`
	DefaultModel    string                 `yaml:"default_model"`
	Temperature     float64                `yaml:"temperature"`
	MaxTokens       int                    `yaml:"max_tokens"`
	Timeout         string                 `yaml:"timeout"`
	OpenAI          ProviderConfig         `yaml:"openai"`
	Anthropic       ProviderConfig         `yaml:"anthropic"`
	Google          ProviderConfig         `yaml:"google"`
	Providers       map[string]interface{} `yaml:"providers"`
}

// ProviderConfig represents a specific provider configuration
type ProviderConfig struct {
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// AnalysisConfig represents analysis configuration
type AnalysisConfig struct {
	MaxConcurrent int    `yaml:"max_concurrent"`
	DefaultMode   string `yaml:"default_mode"`
}

// MemoryConfig represents memory configuration
type MemoryConfig struct {
	RetentionDays  int `yaml:"retention_days"`
	ShortTermLimit int `yaml:"short_term_limit"`
	LongTermLimit  int `yaml:"long_term_limit"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path:              "personal_ai_board.db",
			MaxOpenConns:      25,
			MaxIdleConns:      25,
			EnableWAL:         true,
			EnableForeignKeys: true,
		},
		LLM: LLMConfig{
			DefaultProvider: "openai",
			DefaultModel:    "gpt-4",
			Temperature:     0.7,
			MaxTokens:       1000,
			Timeout:         "30s",
			OpenAI: ProviderConfig{
				BaseURL: "https://api.openai.com/v1",
			},
			Anthropic: ProviderConfig{
				BaseURL: "https://api.anthropic.com",
			},
			Google: ProviderConfig{
				BaseURL: "https://generativelanguage.googleapis.com",
			},
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Analysis: AnalysisConfig{
			MaxConcurrent: 5,
			DefaultMode:   "discussion",
		},
		Memory: MemoryConfig{
			RetentionDays:  90,
			ShortTermLimit: 50,
			LongTermLimit:  200,
		},
	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Load .env file first (if it exists)
	loadDotEnv()

	// Load from file if provided
	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	return config, nil
}

// LoadDefault loads configuration from default locations
func LoadDefault() (*Config, error) {
	config := DefaultConfig()

	// Load .env file first (if it exists)
	loadDotEnv()

	// Try to load from default locations
	configPaths := []string{
		".personal-ai-board.yaml",
		".personal-ai-board.yml",
	}

	// Add home directory paths
	if home, err := os.UserHomeDir(); err == nil {
		configPaths = append(configPaths,
			filepath.Join(home, ".personal-ai-board.yaml"),
			filepath.Join(home, ".personal-ai-board.yml"),
		)
	}

	// Try each path
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			if err := loadFromFile(config, path); err != nil {
				return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
			}
			break
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	return config, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(config *Config, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(config)
}

// loadDotEnv loads environment variables from .env file
func loadDotEnv() {
	// Try to load .env file from current directory
	envPaths := []string{
		".env",
	}

	// Add home directory .env path
	if home, err := os.UserHomeDir(); err == nil {
		envPaths = append(envPaths, filepath.Join(home, ".personal-ai-board.env"))
	}

	// Try each .env file path
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				// Successfully loaded .env file
				return
			}
		}
	}

	// If no .env file found, that's okay - we'll use system environment variables
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Database configuration
	if path := os.Getenv("PAB_DATABASE_PATH"); path != "" {
		config.Database.Path = path
	}
	if val := os.Getenv("PAB_DATABASE_MAX_OPEN_CONNS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			config.Database.MaxOpenConns = i
		}
	}
	if val := os.Getenv("PAB_DATABASE_MAX_IDLE_CONNS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			config.Database.MaxIdleConns = i
		}
	}
	if val := os.Getenv("PAB_DATABASE_ENABLE_WAL"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.Database.EnableWAL = b
		}
	}

	// LLM configuration
	if provider := os.Getenv("PAB_LLM_DEFAULT_PROVIDER"); provider != "" {
		config.LLM.DefaultProvider = provider
	}
	if model := os.Getenv("PAB_LLM_DEFAULT_MODEL"); model != "" {
		config.LLM.DefaultModel = model
	}
	if val := os.Getenv("PAB_LLM_TEMPERATURE"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			config.LLM.Temperature = f
		}
	}
	if val := os.Getenv("PAB_LLM_MAX_TOKENS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			config.LLM.MaxTokens = i
		}
	}

	// Provider API keys from standard environment variables
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		config.LLM.OpenAI.APIKey = key
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		config.LLM.Anthropic.APIKey = key
	}
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		config.LLM.Google.APIKey = key
	}

	// Provider API keys from PAB prefixed environment variables
	if key := os.Getenv("PAB_LLM_OPENAI_API_KEY"); key != "" {
		config.LLM.OpenAI.APIKey = key
	}
	if key := os.Getenv("PAB_LLM_ANTHROPIC_API_KEY"); key != "" {
		config.LLM.Anthropic.APIKey = key
	}
	if key := os.Getenv("PAB_LLM_GOOGLE_API_KEY"); key != "" {
		config.LLM.Google.APIKey = key
	}

	// Log configuration
	if level := os.Getenv("PAB_LOG_LEVEL"); level != "" {
		config.Log.Level = level
	}
	if format := os.Getenv("PAB_LOG_FORMAT"); format != "" {
		config.Log.Format = format
	}

	// Analysis configuration
	if val := os.Getenv("PAB_ANALYSIS_MAX_CONCURRENT"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			config.Analysis.MaxConcurrent = i
		}
	}
	if mode := os.Getenv("PAB_ANALYSIS_DEFAULT_MODE"); mode != "" {
		config.Analysis.DefaultMode = mode
	}

	// Memory configuration
	if val := os.Getenv("PAB_MEMORY_RETENTION_DAYS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			config.Memory.RetentionDays = i
		}
	}
	if val := os.Getenv("PAB_MEMORY_SHORT_TERM_LIMIT"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			config.Memory.ShortTermLimit = i
		}
	}
	if val := os.Getenv("PAB_MEMORY_LONG_TERM_LIMIT"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			config.Memory.LongTermLimit = i
		}
	}
}

// GetString returns a string value, handling environment variable expansion
func (c *Config) GetString(value string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envVar := value[2 : len(value)-1]
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue
		}
	}
	return value
}

// GetTimeout parses the timeout string and returns a time.Duration
func (c *Config) GetTimeout() time.Duration {
	if duration, err := time.ParseDuration(c.LLM.Timeout); err == nil {
		return duration
	}
	return 30 * time.Second // Default timeout
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate database path
	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if c.Log.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)",
			c.Log.Level, strings.Join(validLogLevels, ", "))
	}

	// Validate LLM configuration
	if c.LLM.Temperature < 0 || c.LLM.Temperature > 2 {
		return fmt.Errorf("LLM temperature must be between 0 and 2")
	}

	if c.LLM.MaxTokens <= 0 {
		return fmt.Errorf("LLM max tokens must be positive")
	}

	// Validate analysis mode
	validModes := []string{"discussion", "simulation", "analysis", "comparison", "evaluation", "prediction"}
	validMode := false
	for _, mode := range validModes {
		if c.Analysis.DefaultMode == mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return fmt.Errorf("invalid analysis mode: %s (must be one of: %s)",
			c.Analysis.DefaultMode, strings.Join(validModes, ", "))
	}

	return nil
}

// HasProvider checks if a provider is configured with an API key
func (c *Config) HasProvider(provider string) bool {
	switch strings.ToLower(provider) {
	case "openai":
		return c.GetString(c.LLM.OpenAI.APIKey) != ""
	case "anthropic":
		return c.GetString(c.LLM.Anthropic.APIKey) != ""
	case "google", "gemini":
		return c.GetString(c.LLM.Google.APIKey) != ""
	default:
		return false
	}
}

// GetProviderConfig returns the configuration for a specific provider
func (c *Config) GetProviderConfig(provider string) (ProviderConfig, bool) {
	switch strings.ToLower(provider) {
	case "openai":
		return c.LLM.OpenAI, true
	case "anthropic":
		return c.LLM.Anthropic, true
	case "google", "gemini":
		return c.LLM.Google, true
	default:
		return ProviderConfig{}, false
	}
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	return encoder.Encode(c)
}
