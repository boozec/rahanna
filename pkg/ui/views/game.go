package views

import (
	"fmt"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Keyboard controls
type gameKeyMap struct {
	EnterNewGame key.Binding
	StartNewGame key.Binding
	GoLogout     key.Binding
	Quit         key.Binding
}

// Default key bindings for the game model
var defaultGameKeyMap = gameKeyMap{
	GoLogout: key.NewBinding(
		key.WithKeys("alt+Q", "alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "q"),
		key.WithHelp("    Q", "Quit"),
	),
}

type GameModel struct {
	// UI dimensions
	width  int
	height int

	// UI state
	keys playKeyMap

	// Game state
	game    *database.Game
	network *multiplayer.GameNetwork
}

func NewGameModel(width, height int, game *database.Game, network *multiplayer.GameNetwork) GameModel {
	return GameModel{
		width:   width,
		height:  height,
		game:    game,
		network: network,
	}
}

// Init function for GameModel
func (m GameModel) Init() tea.Cmd {
	ClearScreen()
	return textinput.Blink
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if exit := handleExit(msg); exit != nil {
		return m, exit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// View function for GameModel
func (m GameModel) View() string {
	formWidth := getFormWidth(m.width)
	// base := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width)

	content := "abc"

	// Build the main window with error handling
	windowContent := m.buildWindowContent(content, formWidth)

	// Create navigation buttons
	buttons := m.renderNavigationButtons()

	centeredContent := lipgloss.JoinVertical(
		lipgloss.Center,
		getLogo(m.width),
		windowContent,
		lipgloss.NewStyle().MarginTop(2).Render(buttons),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		centeredContent,
	)
}

func (m GameModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	return m, nil
}

func (m GameModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.GoLogout):
		return m, logout(m.width, m.height+1)

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}

	return m, nil
}

func (m GameModel) buildWindowContent(content string, formWidth int) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
			lipgloss.Center,
			content,
		)),
	)
}

func (m GameModel) renderNavigationButtons() string {
	logoutKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.GoLogout.Help().Key),
		m.keys.GoLogout.Help().Desc)

	quitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Quit.Help().Key),
		m.keys.Quit.Help().Desc)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		logoutKey,
		quitKey,
	)
}
