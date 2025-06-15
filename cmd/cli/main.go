package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"personal-ai-board/internal/db"
	"personal-ai-board/internal/llm"
	"personal-ai-board/internal/llm/types"
	"personal-ai-board/pkg/logger"
	"personal-ai-board/web/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "1.0.0-dev"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "personal-ai-board",
	Short: "A personal AI advisory board for decision making and analysis",
	Long: `Personal AI Board helps you make better decisions by simulating
advisory boards of AI personas with unique personalities and expertise.

Create custom boards, analyze ideas, and get diverse perspectives on your
projects and decisions.`,
	Version: version,
	RunE:    runInteractiveMode,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.personal-ai-board.yaml)")
	rootCmd.PersistentFlags().String("db-path", "", "database file path")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")

	// Bind flags to viper
	viper.BindPFlag("database.path", rootCmd.PersistentFlags().Lookup("db-path"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))

	// Add subcommands
	rootCmd.AddCommand(createPersonaCmd())
	rootCmd.AddCommand(listPersonasCmd())
	rootCmd.AddCommand(createBoardCmd())
	rootCmd.AddCommand(listBoardsCmd())
	rootCmd.AddCommand(runAnalysisCmd())
	rootCmd.AddCommand(migrateCmd())
	rootCmd.AddCommand(versionCmd())
}

// initConfig reads in config file and ENV variables
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		// Search for config in home directory
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".personal-ai-board")
	}

	// Environment variables
	viper.SetEnvPrefix("PAB")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func setDefaults() {
	// Database defaults
	viper.SetDefault("database.path", "personal_ai_board.db")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.enable_wal", true)
	viper.SetDefault("database.enable_foreign_keys", true)

	// LLM defaults
	viper.SetDefault("llm.default_provider", "openai")
	viper.SetDefault("llm.default_model", "gpt-4")
	viper.SetDefault("llm.temperature", 0.7)
	viper.SetDefault("llm.max_tokens", 1000)
	viper.SetDefault("llm.timeout", "30s")

	// OpenAI defaults
	viper.SetDefault("llm.openai.base_url", "https://api.openai.com/v1")

	// Logging defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")

	// Analysis defaults
	viper.SetDefault("analysis.max_concurrent", 5)
	viper.SetDefault("analysis.default_mode", "discussion")

	// Memory defaults
	viper.SetDefault("memory.retention_days", 90)
	viper.SetDefault("memory.short_term_limit", 50)
	viper.SetDefault("memory.long_term_limit", 200)
}

// runInteractiveMode runs the interactive CLI mode
func runInteractiveMode(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize components
	logger := logger.New(viper.GetString("log.level"))

	// Initialize database
	dbConfig := &db.Config{
		Path:              viper.GetString("database.path"),
		MaxOpenConns:      viper.GetInt("database.max_open_conns"),
		MaxIdleConns:      viper.GetInt("database.max_idle_conns"),
		EnableWAL:         viper.GetBool("database.enable_wal"),
		EnableForeignKeys: viper.GetBool("database.enable_foreign_keys"),
	}

	database, err := db.Connect(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize LLM manager
	llmManager := llm.NewManager(logger)

	// Register OpenAI provider if API key is available
	openaiKey := viper.GetString("llm.openai.api_key")
	if openaiKey == "" {
		openaiKey = os.Getenv("OPENAI_API_KEY")
	}

	if openaiKey != "" {
		openaiConfig := types.Config{
			Provider:    "openai",
			APIKey:      openaiKey,
			BaseURL:     viper.GetString("llm.openai.base_url"),
			Model:       viper.GetString("llm.default_model"),
			Temperature: viper.GetFloat64("llm.temperature"),
			MaxTokens:   viper.GetInt("llm.max_tokens"),
		}

		openaiProvider, err := llm.NewOpenAIProvider(openaiConfig, logger)
		if err != nil {
			return fmt.Errorf("failed to create OpenAI provider: %w", err)
		}

		if err := llmManager.RegisterProvider("openai", openaiProvider); err != nil {
			return fmt.Errorf("failed to register OpenAI provider: %w", err)
		}

		if err := llmManager.SetDefaultProvider("openai"); err != nil {
			return fmt.Errorf("failed to set default provider: %w", err)
		}
	} else {
		logger.Warn("No OpenAI API key found. Set OPENAI_API_KEY environment variable or configure in config file.")
		return fmt.Errorf("no LLM provider configured")
	}

	// Create app configuration
	appConfig := &cli.Config{
		Database:   database,
		LLMManager: llmManager,
		Logger:     logger,
		ConfigPath: viper.GetString("config.path"),
	}

	// Start interactive CLI
	app := cli.NewApp(appConfig)
	return app.Run(ctx)
}

// Command implementations

func createPersonaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-persona [name] [traits-file]",
		Short: "Create a new persona",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implementation for creating persona
			return fmt.Errorf("not implemented yet")
		},
	}
	return cmd
}

func listPersonasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-personas",
		Short: "List all personas",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implementation for listing personas
			return fmt.Errorf("not implemented yet")
		},
	}
	return cmd
}

func createBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-board [name] [persona-ids...]",
		Short: "Create a new advisory board",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implementation for creating board
			return fmt.Errorf("not implemented yet")
		},
	}
	return cmd
}

func listBoardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-boards",
		Short: "List all boards",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implementation for listing boards
			return fmt.Errorf("not implemented yet")
		},
	}
	return cmd
}

func runAnalysisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [board-id] [prompt]",
		Short: "Run analysis with a board",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implementation for running analysis
			return fmt.Errorf("not implemented yet")
		},
	}

	cmd.Flags().String("mode", "discussion", "analysis mode (discussion, simulation, analysis, comparison, evaluation, prediction)")
	cmd.Flags().String("project", "", "project ID to associate with analysis")

	return cmd
}

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logger.New(viper.GetString("log.level"))

			dbConfig := &db.Config{
				Path:              viper.GetString("database.path"),
				MaxOpenConns:      viper.GetInt("database.max_open_conns"),
				MaxIdleConns:      viper.GetInt("database.max_idle_conns"),
				EnableWAL:         viper.GetBool("database.enable_wal"),
				EnableForeignKeys: viper.GetBool("database.enable_foreign_keys"),
			}

			database, err := db.Connect(dbConfig)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer database.Close()

			logger.Info("Running database migrations...")
			if err := database.Migrate(); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			logger.Info("Migrations completed successfully")
			return nil
		},
	}

	cmd.Flags().Bool("reset", false, "reset database (WARNING: this will delete all data)")

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Personal AI Board v%s\n", version)
		},
	}
}

// Utility functions

func getConfigPath() string {
	if cfgFile != "" {
		return filepath.Dir(cfgFile)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	return home
}
