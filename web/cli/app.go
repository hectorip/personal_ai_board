package cli

import (
	"context"
	"fmt"

	"personal-ai-board/internal/db"
	"personal-ai-board/internal/llm"
	"personal-ai-board/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Config contains configuration for the CLI app
type Config struct {
	Database   *db.Database
	LLMManager *llm.Manager
	Logger     logger.Logger
	ConfigPath string
}

// App represents the CLI application
type App struct {
	config *Config
	model  tea.Model
}

// NewApp creates a new CLI application
func NewApp(config *Config) *App {
	model := NewMainModel(config)
	
	return &App{
		config: config,
		model:  model,
	}
}

// Run starts the interactive CLI application
func (a *App) Run(ctx context.Context) error {
	a.config.Logger.Info("Starting Personal AI Board CLI")
	
	program := tea.NewProgram(
		a.model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	_, err := program.Run()
	return err
}

// MainModel represents the main application model
type MainModel struct {
	config     *Config
	state      appState
	width      int
	height     int
	err        error
}

type appState int

const (
	stateMain appState = iota
	statePersonas
	stateBoards
	stateAnalysis
	stateSettings
	stateLoading
	stateError
)

// NewMainModel creates a new main model
func NewMainModel(config *Config) *MainModel {
	return &MainModel{
		config: config,
		state:  stateMain,
	}
}

// Init implements tea.Model
func (m *MainModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.state = statePersonas
			return m, nil
		case "2":
			m.state = stateBoards
			return m, nil
		case "3":
			m.state = stateAnalysis
			return m, nil
		case "4":
			m.state = stateSettings
			return m, nil
		case "esc":
			m.state = stateMain
			return m, nil
		}
	}
	
	return m, nil
}

// View implements tea.Model
func (m *MainModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	
	switch m.state {
	case stateMain:
		return m.renderMainMenu()
	case statePersonas:
		return m.renderPersonasView()
	case stateBoards:
		return m.renderBoardsView()
	case stateAnalysis:
		return m.renderAnalysisView()
	case stateSettings:
		return m.renderSettingsView()
	case stateLoading:
		return m.renderLoadingView()
	case stateError:
		return m.renderErrorView()
	default:
		return m.renderMainMenu()
	}
}

// Rendering methods

func (m *MainModel) renderMainMenu() string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(2).
		PaddingLeft(4).
		Width(m.width)

	title := style.Render("Personal AI Advisory Board")
	
	menu := lipgloss.NewStyle().
		MarginTop(2).
		MarginLeft(4).
		Render(`
Welcome to your Personal AI Advisory Board!

Main Menu:
  1. Manage Personas
  2. Manage Boards  
  3. Run Analysis
  4. Settings

Navigation:
  • Use number keys to select options
  • Press 'q' or Ctrl+C to quit
  • Press 'Esc' to return to main menu
`)

	status := m.renderStatusBar()
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		menu,
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		status,
	)
}

func (m *MainModel) renderPersonasView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#FF6B6B")).
		PaddingTop(1).
		PaddingLeft(4).
		Width(m.width).
		Render("Persona Management")
	
	content := lipgloss.NewStyle().
		MarginTop(2).
		MarginLeft(4).
		Render(`
Persona Management:
  c. Create new persona
  l. List all personas
  d. Delete persona
  e. Edit persona traits
  i. Import persona from config

Available Personas:
  • Loading personas...

Press 'Esc' to return to main menu
`)

	status := m.renderStatusBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		status,
	)
}

func (m *MainModel) renderBoardsView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#4ECDC4")).
		PaddingTop(1).
		PaddingLeft(4).
		Width(m.width).
		Render("Board Management")
	
	content := lipgloss.NewStyle().
		MarginTop(2).
		MarginLeft(4).
		Render(`
Board Management:
  c. Create new board
  l. List all boards
  d. Delete board
  e. Edit board composition
  t. Browse board templates

Available Boards:
  • Loading boards...

Press 'Esc' to return to main menu
`)

	status := m.renderStatusBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		status,
	)
}

func (m *MainModel) renderAnalysisView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#45B7D1")).
		PaddingTop(1).
		PaddingLeft(4).
		Width(m.width).
		Render("Analysis Center")
	
	content := lipgloss.NewStyle().
		MarginTop(2).
		MarginLeft(4).
		Render(`
Analysis Options:
  n. New analysis session
  r. Recent analysis results
  p. Manage projects
  h. Analysis history

Analysis Modes:
  • Discussion - Interactive board discussion
  • Simulation - Scenario modeling
  • Analysis - Structured evaluation
  • Comparison - Side-by-side comparison
  • Evaluation - Scoring and ranking
  • Prediction - Outcome forecasting

Press 'Esc' to return to main menu
`)

	status := m.renderStatusBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		status,
	)
}

func (m *MainModel) renderSettingsView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#96CEB4")).
		PaddingTop(1).
		PaddingLeft(4).
		Width(m.width).
		Render("Settings")
	
	// Get provider status
	providers := m.config.LLMManager.ListProviders()
	providerStatus := "No providers configured"
	if len(providers) > 0 {
		providerStatus = fmt.Sprintf("Available: %v", providers)
	}

	content := lipgloss.NewStyle().
		MarginTop(2).
		MarginLeft(4).
		Render(fmt.Sprintf(`
System Settings:
  l. Configure LLM providers
  d. Database management
  m. Memory settings
  e. Export/Import data
  h. Health check

Current Configuration:
  • Database: %s
  • LLM Providers: %s
  • Log Level: info

Press 'Esc' to return to main menu
`, m.config.Database.GetConfig().Path, providerStatus))

	status := m.renderStatusBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		status,
	)
}

func (m *MainModel) renderLoadingView() string {
	return lipgloss.NewStyle().
		MarginTop(m.height/2).
		MarginLeft(m.width/2-10).
		Render("Loading...")
}

func (m *MainModel) renderErrorView() string {
	errorMsg := "An error occurred"
	if m.err != nil {
		errorMsg = m.err.Error()
	}
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		MarginTop(2).
		MarginLeft(4).
		Render(fmt.Sprintf("Error: %s\n\nPress 'Esc' to return to main menu", errorMsg))
}

func (m *MainModel) renderStatusBar() string {
	if m.height <= 10 {
		return ""
	}
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Background(lipgloss.Color("#E6E6E6")).
		PaddingLeft(2).
		PaddingRight(2).
		Width(m.width)
	
	statusText := fmt.Sprintf("Personal AI Board v1.0.0 | %dx%d | Press 'q' to quit", m.width, m.height)
	
	return style.Render(statusText)
}