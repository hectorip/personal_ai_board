package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const version = "1.0.0-dev"

// ViewType represents different views in the application
type ViewType string

const (
	ViewMenu     ViewType = "menu"
	ViewPersonas ViewType = "personas"
	ViewBoards   ViewType = "boards"
	ViewProjects ViewType = "projects"
	ViewAnalysis ViewType = "analysis"
	ViewSettings ViewType = "settings"
	ViewHelp     ViewType = "help"
)

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Action      string
	Icon        string
}

// Model represents the application state
type Model struct {
	currentView ViewType
	width       int
	height      int
	ready       bool
	cursor      int
	items       []MenuItem
	statusMsg   string
	errorMsg    string
}

// StatusMsg represents a status message
type StatusMsg string

func main() {
	// Handle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("Personal AI Advisory Board v%s\n", version)
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	model := NewModel()
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running CLI: %v\n", err)
		os.Exit(1)
	}
}

// NewModel creates a new application model
func NewModel() *Model {
	menuItems := []MenuItem{
		{"Manage Personas", "Create, edit, and manage AI personas for your advisory board", "personas", "👥"},
		{"Manage Boards", "Create and configure advisory boards with different personas", "boards", "🏛️"},
		{"Manage Projects", "Create and manage projects with ideas and documents", "projects", "📁"},
		{"Run Analysis", "Analyze ideas and projects with your advisory boards", "analysis", "🔍"},
		{"Settings", "Configure application settings and preferences", "settings", "⚙️"},
		{"Help", "View help, documentation, and usage guides", "help", "❓"},
		{"Quit", "Exit the Personal AI Advisory Board application", "quit", "🚪"},
	}

	return &Model{
		currentView: ViewMenu,
		cursor:      0,
		items:       menuItems,
	}
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case StatusMsg:
		m.statusMsg = string(msg)
		m.errorMsg = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

// handleKeyMsg handles keyboard input
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "q":
		if m.currentView == ViewMenu {
			return m, tea.Quit
		}
		m.currentView = ViewMenu
		m.cursor = 0
		m.statusMsg = ""
		m.errorMsg = ""
		return m, nil

	case "esc":
		if m.currentView != ViewMenu {
			m.currentView = ViewMenu
			m.cursor = 0
			m.statusMsg = ""
			m.errorMsg = ""
		}
		return m, nil

	case "up", "k":
		return m.moveCursor(-1)

	case "down", "j":
		return m.moveCursor(1)

	case "enter", " ":
		return m.handleSelection()

	// Global shortcuts (only from main menu)
	case "1":
		if m.currentView == ViewMenu {
			return m.navigateToView(ViewPersonas)
		}
	case "2":
		if m.currentView == ViewMenu {
			return m.navigateToView(ViewBoards)
		}
	case "3":
		if m.currentView == ViewMenu {
			return m.navigateToView(ViewProjects)
		}
	case "4":
		if m.currentView == ViewMenu {
			return m.navigateToView(ViewAnalysis)
		}
	case "5":
		if m.currentView == ViewMenu {
			return m.navigateToView(ViewSettings)
		}
	case "h":
		if m.currentView == ViewMenu {
			return m.navigateToView(ViewHelp)
		}
	}

	return m, nil
}

// moveCursor moves the cursor up or down
func (m *Model) moveCursor(direction int) (tea.Model, tea.Cmd) {
	if m.currentView == ViewMenu {
		m.cursor += direction
		if m.cursor < 0 {
			m.cursor = len(m.items) - 1
		} else if m.cursor >= len(m.items) {
			m.cursor = 0
		}
	}
	return m, nil
}

// navigateToView navigates to a specific view
func (m *Model) navigateToView(view ViewType) (tea.Model, tea.Cmd) {
	m.currentView = view
	m.cursor = 0
	m.statusMsg = ""
	m.errorMsg = ""
	return m, nil
}

// handleSelection handles item selection
func (m *Model) handleSelection() (tea.Model, tea.Cmd) {
	if m.currentView == ViewMenu {
		if m.cursor >= 0 && m.cursor < len(m.items) {
			action := m.items[m.cursor].Action
			switch action {
			case "personas":
				return m.navigateToView(ViewPersonas)
			case "boards":
				return m.navigateToView(ViewBoards)
			case "projects":
				return m.navigateToView(ViewProjects)
			case "analysis":
				return m.navigateToView(ViewAnalysis)
			case "settings":
				return m.navigateToView(ViewSettings)
			case "help":
				return m.navigateToView(ViewHelp)
			case "quit":
				return m, tea.Quit
			}
		}
	} else {
		// Handle selections in other views
		return m, func() tea.Msg {
			return StatusMsg("✨ Feature implementation in progress! This is a working demo of the TUI interface.")
		}
	}
	return m, nil
}

// View implements tea.Model
func (m *Model) View() string {
	if !m.ready {
		return "🚀 Initializing Personal AI Advisory Board..."
	}

	var content string

	switch m.currentView {
	case ViewMenu:
		content = m.renderMainMenu()
	case ViewPersonas:
		content = m.renderPersonasView()
	case ViewBoards:
		content = m.renderBoardsView()
	case ViewProjects:
		content = m.renderProjectsView()
	case ViewAnalysis:
		content = m.renderAnalysisView()
	case ViewSettings:
		content = m.renderSettingsView()
	case ViewHelp:
		content = m.renderHelpView()
	default:
		content = "Unknown view"
	}

	// Add common elements
	var s strings.Builder
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	s.WriteString(content)
	s.WriteString("\n")
	s.WriteString(m.renderFooter())

	return s.String()
}

// renderHeader renders the application header
func (m *Model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(4).
		Width(m.width)

	title := "🤖 Personal AI Advisory Board"
	if m.currentView != ViewMenu {
		title += " - " + m.getViewTitle()
	}

	return titleStyle.Render(title)
}

// renderMainMenu renders the main menu
func (m *Model) renderMainMenu() string {
	var s strings.Builder

	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		PaddingLeft(2).
		MarginBottom(1)
	s.WriteString(breadcrumbStyle.Render("🏠 Home"))
	s.WriteString("\n\n")

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		PaddingLeft(2).
		MarginBottom(2)
	s.WriteString(descStyle.Render("Welcome to your Personal AI Advisory Board! 🎯 Choose an option to get started:"))
	s.WriteString("\n\n")

	// Render menu items
	for i, item := range m.items {
		cursor := "  "
		if m.cursor == i {
			cursor = "→ "
		}

		itemStyle := lipgloss.NewStyle().PaddingLeft(2)
		numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
		titleStyle := lipgloss.NewStyle()
		descStyle := lipgloss.NewStyle().
			PaddingLeft(6).
			Foreground(lipgloss.Color("#888888")).
			MarginBottom(1)

		if m.cursor == i {
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true).
				Background(lipgloss.Color("#1a1a1a"))
			titleStyle = titleStyle.
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true)
			descStyle = descStyle.
				Foreground(lipgloss.Color("#AAAAAA")).
				Background(lipgloss.Color("#1a1a1a"))
			numberStyle = numberStyle.
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true)
		}

		itemNumber := fmt.Sprintf("%d.", i+1)
		itemText := fmt.Sprintf("%s %s %s", cursor, item.Icon, titleStyle.Render(item.Title))

		s.WriteString(itemStyle.Render(fmt.Sprintf("%s %s", numberStyle.Render(itemNumber), itemText)))
		s.WriteString("\n")

		if m.cursor == i {
			s.WriteString(descStyle.Render(fmt.Sprintf("   %s", item.Description)))
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	return s.String()
}

// renderPersonasView renders the personas management view
func (m *Model) renderPersonasView() string {
	return m.renderSimpleView("👥 Personas", "🏠 Home > 👥 Personas",
		"Manage your AI personas. Each persona has unique traits, expertise, and communication styles.",
		[]string{
			"✨ Create New Persona - Design a custom AI persona with unique traits",
			"📋 List All Personas - View your complete persona collection",
			"✏️ Edit Persona - Modify existing persona traits and characteristics",
			"🗑️ Delete Persona - Remove a persona from your collection",
			"📤 Export Personas - Backup your personas to a file",
			"📥 Import Personas - Restore personas from a backup file",
		})
}

// renderBoardsView renders the boards management view
func (m *Model) renderBoardsView() string {
	return m.renderSimpleView("🏛️ Boards", "🏠 Home > 🏛️ Boards",
		"Manage your advisory boards. Combine personas to create diverse expert panels.",
		[]string{
			"🆕 Create New Board - Assemble a new advisory board from your personas",
			"📑 List All Boards - View your complete board collection",
			"📋 Board Templates - Use pre-designed board templates (Executive, Technical, Creative)",
			"⚙️ Edit Board - Modify board composition and settings",
			"🧪 Test Board - Simulate board discussions with sample topics",
			"📊 Board Analytics - View board performance and insights",
		})
}

// renderProjectsView renders the projects management view
func (m *Model) renderProjectsView() string {
	return m.renderSimpleView("📁 Projects", "🏠 Home > 📁 Projects",
		"Organize your ideas into projects and collaborate with your AI advisory board.",
		[]string{
			"🆕 Create Project - Start a new project with goals and objectives",
			"📂 List Projects - View all your active and completed projects",
			"💡 Add Ideas - Brainstorm and capture new ideas within projects",
			"📄 Add Documents - Upload supporting documents, images, and files",
			"📈 Project Analytics - Track progress, milestones, and insights",
			"🗂️ Archive Project - Move completed projects to archive",
		})
}

// renderAnalysisView renders the analysis view
func (m *Model) renderAnalysisView() string {
	return m.renderSimpleView("🔍 Analysis", "🏠 Home > 🔍 Analysis",
		"Run different types of analysis with your advisory boards to gain deep insights.",
		[]string{
			"💬 Discussion Mode - Interactive debate and collaborative exploration",
			"🎭 Simulation Mode - Model scenarios and predict potential outcomes",
			"📊 Analysis Mode - Deep analytical breakdown with data-driven insights",
			"⚖️ Comparison Mode - Compare multiple options side-by-side",
			"🏆 Evaluation Mode - Score and rank alternatives systematically",
			"🔮 Prediction Mode - Forecast future trends and possibilities",
		})
}

// renderSettingsView renders the settings view
func (m *Model) renderSettingsView() string {
	return m.renderSimpleView("⚙️ Settings", "🏠 Home > ⚙️ Settings",
		"Configure application settings and preferences to customize your experience.",
		[]string{
			"🤖 LLM Providers - Configure AI language model providers (OpenAI, Anthropic, Google)",
			"🗄️ Database Settings - Manage data storage and backup preferences",
			"🧠 Memory Settings - Configure persona memory and context retention",
			"🎨 Interface Themes - Customize colors and visual appearance",
			"📤 Export Settings - Backup entire configuration for portability",
			"📥 Import Settings - Restore settings from backup files",
		})
}

// renderHelpView renders the help view
func (m *Model) renderHelpView() string {
	var s strings.Builder

	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		PaddingLeft(2).
		MarginBottom(1)
	s.WriteString(breadcrumbStyle.Render("🏠 Home > ❓ Help"))
	s.WriteString("\n\n")

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E74C3C")).
		Bold(true).
		PaddingLeft(2)
	s.WriteString(titleStyle.Render("📚 Personal AI Advisory Board - Help & Documentation"))
	s.WriteString("\n\n")

	quickStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2ECC71")).
		Bold(true).
		PaddingLeft(2)
	s.WriteString(quickStyle.Render("🚀 Quick Start Guide:"))
	s.WriteString("\n")

	steps := []string{
		"1. 👥 Create personas with unique traits, expertise, and personalities",
		"2. 🏛️ Assemble advisory boards by combining complementary personas",
		"3. 📁 Create projects to organize your ideas, goals, and documents",
		"4. 🔍 Run analysis sessions to get insights from your advisory board",
		"5. 📊 Review results and iterate on your ideas with expert feedback",
	}

	stepStyle := lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
	for _, step := range steps {
		s.WriteString(stepStyle.Render(step))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	shortcutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3498DB")).
		Bold(true).
		PaddingLeft(2)
	s.WriteString(shortcutStyle.Render("⌨️ Keyboard Shortcuts:"))
	s.WriteString("\n")

	shortcuts := []string{
		"↑/↓ or j/k    - Navigate menu items up and down",
		"Enter/Space   - Select the currently highlighted item",
		"Esc or q      - Return to main menu from any view",
		"1-5           - Quick access to main sections from menu",
		"h             - Show this help screen from menu",
		"Ctrl+C        - Force quit the application",
	}

	for _, shortcut := range shortcuts {
		s.WriteString(stepStyle.Render(shortcut))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	conceptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B59B6")).
		Bold(true).
		PaddingLeft(2)
	s.WriteString(conceptStyle.Render("💡 Key Concepts:"))
	s.WriteString("\n")

	concepts := []string{
		"👤 Persona - An AI character with specific traits, expertise, and communication style",
		"🏛️ Board - A collection of personas that form your advisory committee",
		"📁 Project - A container for related ideas, documents, and analysis sessions",
		"🔍 Analysis - Different modes of getting insights from your board on topics",
		"💭 Memory - Each persona maintains context and learns from interactions",
	}

	for _, concept := range concepts {
		s.WriteString(stepStyle.Render(concept))
		s.WriteString("\n")
	}

	return s.String()
}

// renderSimpleView renders a simple view with title, breadcrumb, description and items
func (m *Model) renderSimpleView(title, breadcrumb, description string, items []string) string {
	var s strings.Builder

	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		PaddingLeft(2).
		MarginBottom(1)
	s.WriteString(breadcrumbStyle.Render(breadcrumb))
	s.WriteString("\n\n")

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		PaddingLeft(2).
		MarginBottom(2)
	s.WriteString(descStyle.Render(description))
	s.WriteString("\n\n")

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(lipgloss.Color("#AAAAAA"))

	for _, item := range items {
		s.WriteString(itemStyle.Render(item))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	comingSoonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F39C12")).
		PaddingLeft(2).
		Bold(true)
	s.WriteString(comingSoonStyle.Render("🚧 Full implementation in progress! Press Enter to see status message."))

	return s.String()
}

// renderFooter renders the application footer
func (m *Model) renderFooter() string {
	var footerParts []string

	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Background(lipgloss.Color("#1a1a1a")).
			PaddingLeft(2)
		footerParts = append(footerParts, statusStyle.Render(m.statusMsg))
	}

	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#1a1a1a")).
			PaddingLeft(2)
		footerParts = append(footerParts, errorStyle.Render("❌ "+m.errorMsg))
	}

	navHelp := m.getNavigationHelp()
	if navHelp != "" {
		navStyle := lipgloss.NewStyle().Faint(true).PaddingLeft(2)
		footerParts = append(footerParts, navStyle.Render(navHelp))
	}

	return strings.Join(footerParts, "\n")
}

// getViewTitle returns the title for the current view
func (m *Model) getViewTitle() string {
	switch m.currentView {
	case ViewPersonas:
		return "👥 Personas"
	case ViewBoards:
		return "🏛️ Boards"
	case ViewProjects:
		return "📁 Projects"
	case ViewAnalysis:
		return "🔍 Analysis"
	case ViewSettings:
		return "⚙️ Settings"
	case ViewHelp:
		return "❓ Help"
	default:
		return "🏠 Menu"
	}
}

// getNavigationHelp returns navigation help text
func (m *Model) getNavigationHelp() string {
	base := "Navigation: ↑/↓ or j/k to move, Enter/Space to select"

	if m.currentView == ViewMenu {
		return base + ", 1-5 for quick access, q to quit"
	}
	return base + ", Esc or q to return to menu"
}

// printHelp prints usage information
func printHelp() {
	fmt.Printf("Personal AI Advisory Board v%s\n\n", version)
	fmt.Println("DESCRIPTION:")
	fmt.Println("  A terminal user interface for managing AI personas, advisory boards,")
	fmt.Println("  projects, and running analysis sessions with your virtual advisors.")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  personal-ai-board [options]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --version, -v     Show version information")
	fmt.Println("  --help, -h        Show this help message")
	fmt.Println()
	fmt.Println("KEYBOARD SHORTCUTS:")
	fmt.Println("  ↑/↓ or j/k       Navigate menu items")
	fmt.Println("  Enter/Space      Select current item")
	fmt.Println("  Esc or q         Return to main menu")
	fmt.Println("  1-5              Quick access to sections")
	fmt.Println("  h                Show help")
	fmt.Println("  Ctrl+C           Force quit")
	fmt.Println()
	fmt.Println("FEATURES:")
	fmt.Println("  • Persona Management - Create and manage AI personalities")
	fmt.Println("  • Board Assembly - Combine personas into advisory boards")
	fmt.Println("  • Project Organization - Manage ideas and documents")
	fmt.Println("  • Analysis Modes - Get insights from your AI board")
	fmt.Println("  • Configurable Settings - Customize your experience")
	fmt.Println()
	fmt.Println("For more information, visit the project documentation.")
}
