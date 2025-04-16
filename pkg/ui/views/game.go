package views

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// gameKeyMap defines the key bindings for the game view.
type gameKeyMap struct {
	EnterNewGame key.Binding
	StartNewGame key.Binding
	GoLogout     key.Binding
	Quit         key.Binding
}

// defaultGameKeyMap provides the default key bindings for the game view.
var defaultGameKeyMap = gameKeyMap{
	GoLogout: key.NewBinding(
		key.WithKeys("alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("Q", "Quit"),
	),
}

// GameModel represents the state of the game view.
type GameModel struct {
	// UI dimensions
	width  int
	height int

	// UI state
	keys playKeyMap

	// Game state
	peer          string
	currentGameID int
	game          *database.Game
	network       *multiplayer.GameNetwork
}

// NewGameModel creates a new GameModel.
func NewGameModel(width, height int, peer string, currentGameID int, network *multiplayer.GameNetwork) GameModel {
	return GameModel{
		width:         width,
		height:        height,
		peer:          peer,
		currentGameID: currentGameID,
		network:       network,
	}
}

// Init initializes the GameModel.
func (m GameModel) Init() tea.Cmd {
	ClearScreen()
	return tea.Batch(textinput.Blink, m.getGame())
}

// Update handles incoming messages and updates the GameModel.
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if exit := handleExit(msg); exit != nil {
		return m, exit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case database.Game:
		return m.handleGetGameResponse(msg)
	}

	return m, nil
}

// View renders the GameModel.
func (m GameModel) View() string {
	formWidth := getFormWidth(m.width)

	var content string
	if m.game != nil {
		otherPlayer := ""
		if m.peer == "peer-1" {
			otherPlayer = m.game.Player2.Username
		} else {
			otherPlayer = m.game.Player1.Username
		}
		content = fmt.Sprintf("You're playing versus %s", otherPlayer)
	}

	windowContent := m.buildWindowContent(content, formWidth)
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

func (m *GameModel) handleGetGameResponse(msg database.Game) (tea.Model, tea.Cmd) {
	m.game = &msg
	if m.peer == "peer-1" {
		m.network.Peer = msg.IP2
	} else {
		m.network.Peer = msg.IP1
	}
	return m, nil
}

func (m *GameModel) getGame() tea.Cmd {
	return func() tea.Msg {
		var game database.Game

		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return nil
		}

		// Send API request
		url := fmt.Sprintf("%s/play/%d", os.Getenv("API_BASE"), m.currentGameID)
		resp, err := sendAPIRequest("GET", url, nil, authorization)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&game); err != nil {
			return nil
		}

		// Establish peer connection
		if m.peer == "peer-1" {
			if game.IP2 != "" {
				remote := game.IP2
				go m.network.Server.AddPeer("peer-2", remote)
			}
		} else {
			if game.IP1 != "" {
				remote := game.IP1
				go m.network.Server.AddPeer("peer-1", remote)
			}
		}

		return game
	}
}
