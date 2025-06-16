# Personal AI Board

A powerful Go application that simulates a personal advisory board of AI personas to help you make better decisions and analyze complex topics. Create custom boards with unique personalities, run different types of analysis, and get diverse perspectives on your projects and ideas.

## Features

- **AI Personas**: Create and manage AI personas with unique personalities, traits, and individual memories
- **Custom Boards**: Assemble boards with different combinations of personas for specialized analysis
- **Template Boards**: Use pre-designed boards with curated personas for common scenarios
- **Multiple Analysis Modes**:
  - **Discussion**: Interactive discussions between personas
  - **Simulation**: Simulate scenarios and outcomes
  - **Analysis**: Deep analytical breakdowns
  - **Comparison**: Compare different options or ideas
  - **Evaluation**: Systematic evaluation of proposals
  - **Prediction**: Forecast potential outcomes
- **Project Management**: Organize ideas into projects with file attachments and knowledge graphs
- **Concurrent Processing**: High-performance analysis using Go's goroutines
- **Multiple Interfaces**: CLI and Web interfaces available
- **LLM Provider Support**: Easy switching between different LLM providers

## Prerequisites

- Go 1.21 or higher
- SQLite3
- OpenAI API key (or other supported LLM provider)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd personal_ai_board
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o personal-ai-board cmd/cli/main.go
```

## Configuration

The application supports multiple configuration methods with the following priority order:

1. **Default values** - Built-in defaults
2. **YAML configuration file** - From `.personal-ai-board.yaml` or specified file
3. **Environment file (.env)** - From `.env` file in current directory
4. **Environment variables** - System environment variables

### Quick Start with .env File

The easiest way to configure the application:

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and add your API keys:
```bash
# Required: Add at least one LLM provider API key
OPENAI_API_KEY=your_openai_api_key_here
ANTHROPIC_API_KEY=your_anthropic_api_key_here
GOOGLE_API_KEY=your_google_api_key_here

# Optional: Customize other settings
PAB_LLM_DEFAULT_PROVIDER=openai
PAB_LLM_TEMPERATURE=0.7
PAB_LOG_LEVEL=info
```

### Environment Variables

Set your API keys directly:
```bash
export OPENAI_API_KEY="your_api_key_here"
export ANTHROPIC_API_KEY="your_anthropic_key_here"
export GOOGLE_API_KEY="your_google_key_here"
```

### Configuration File

Create a configuration file at `~/.personal-ai-board.yaml`:

```yaml
database:
  path: "personal_ai_board.db"
  max_open_conns: 25
  max_idle_conns: 25
  enable_wal: true
  enable_foreign_keys: true

llm:
  default_provider: "openai"
  default_model: "gpt-4"
  temperature: 0.7
  max_tokens: 1000
  timeout: "30s"
  openai:
    base_url: "https://api.openai.com/v1"

log:
  level: "info"
  format: "text"

analysis:
  max_concurrent: 5
  default_mode: "discussion"

memory:
  retention_days: 90
  short_term_limit: 50
  long_term_limit: 200
```

## Usage

### First Time Setup

1. Run database migrations:
```bash
./personal-ai-board migrate
```

### Interactive Mode

Start the interactive CLI:
```bash
./personal-ai-board
```

### Command Line Usage

#### Create a Persona
```bash
./personal-ai-board create-persona "Tech Visionary" config/traits/visionary.json
```

#### List Personas
```bash
./personal-ai-board list-personas
```

#### Create a Board
```bash
./personal-ai-board create-board "Strategy Board" persona1 persona2 persona3
```

#### List Boards
```bash
./personal-ai-board list-boards
```

#### Run Analysis
```bash
./personal-ai-board analyze board-id "Should we launch this new product?"
```

With specific mode:
```bash
./personal-ai-board analyze board-id "Product launch strategy" --mode simulation
```

#### Other Commands
```bash
# Show version
./personal-ai-board version

# Show help
./personal-ai-board --help

# Use custom config file
./personal-ai-board --config /path/to/config.yaml

# Set custom database path
./personal-ai-board --db-path /path/to/database.db

# Set log level
./personal-ai-board --log-level debug
```

### Configuration Options

You can configure the application using:
- Command line flags
- Environment variables (prefixed with `PAB_`)
- Configuration file
- Default values

Priority order: CLI flags > Environment variables > Config file > Defaults

### Available Persona Traits

The application comes with pre-configured trait templates in `config/traits/`:
- `analytical.json` - Logical, data-driven thinking
- `creative.json` - Innovative, out-of-the-box solutions
- `visionary.json` - Forward-thinking, strategic perspective
- `base.json` - Balanced, general-purpose traits

## Development

### Project Structure

```
personal_ai_board/
├── cmd/cli/           # CLI application entry point
├── internal/          # Private application code
│   ├── db/           # Database layer
│   ├── llm/          # LLM providers and management
│   └── persona/      # Persona logic and memory
├── pkg/              # Public packages
│   ├── config/       # Configuration management
│   └── logger/       # Logging utilities
├── web/              # Web interface
├── config/           # Configuration files and traits
└── docs/             # Documentation
```

### Building from Source

```bash
# Build for current platform
go build -o personal-ai-board cmd/cli/main.go

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o personal-ai-board-linux cmd/cli/main.go
GOOS=windows GOARCH=amd64 go build -o personal-ai-board.exe cmd/cli/main.go
GOOS=darwin GOARCH=amd64 go build -o personal-ai-board-mac cmd/cli/main.go
```

### Running Tests

```bash
go test ./...
```

### Development Mode

```bash
go run cmd/cli/main.go
```

## Architecture

The application follows clean architecture principles with:
- **Core Domain**: Pure Go business logic
- **Infrastructure**: Database, LLM providers, external services
- **Interfaces**: CLI and Web interfaces as separate modules
- **Concurrent Processing**: Goroutines and channels for parallel analysis

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Configuration Details

For detailed configuration options and advanced setup, see [Configuration Guide](docs/CONFIGURATION.md).

## Troubleshooting

### Common Issues

1. **"No LLM provider configured"**
   - Make sure at least one API key is set (OPENAI_API_KEY, ANTHROPIC_API_KEY, or GOOGLE_API_KEY)
   - Check your `.env` file or environment variables
   - Verify `.env` file is in the current directory
   - Use `./personal-ai-board --test-config` to validate configuration

2. **Database connection errors**
   - Ensure SQLite3 is installed
   - Check database file permissions
   - Try running `migrate` command

3. **Command not found**
   - Make sure the binary is built and in your PATH
   - Check that Go is properly installed

4. **Configuration not loading**
   - Verify `.env` file syntax (no spaces around `=`)
   - Check file permissions on configuration files
   - Use `PAB_LOG_LEVEL=debug` to see configuration loading details

### Getting Help

- Use `./personal-ai-board --help` for command help
- Use `./personal-ai-board [command] --help` for specific command help
- Use `./personal-ai-board --list-providers` to see configured providers
- Check the logs with `--log-level debug` for detailed information
- See [Configuration Guide](docs/CONFIGURATION.md) for complete configuration reference
