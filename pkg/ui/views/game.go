package views

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/network"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

// gameKeyMap defines the key bindings for the game view.
type gameKeyMap struct {
	GoLogout   key.Binding
	RandomMove key.Binding
	Quit       key.Binding
}

// defaultGameKeyMap provides the default key bindings for the game view.
var defaultGameKeyMap = gameKeyMap{
	RandomMove: key.NewBinding(
		key.WithKeys("R", "r"),
		key.WithHelp("R", "Random Move"),
	),
	GoLogout: key.NewBinding(
		key.WithKeys("alt+Q", "alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "q"),
		key.WithHelp("    Q", "Quit"),
	),
}

// ChessMoveMsg is a message containing a received chess move.
type ChessMoveMsg string

// GameModel represents the state of the game view.
type GameModel struct {
	// UI dimensions
	width  int
	height int

	// UI state
	keys gameKeyMap

	// Game state
	peer          string
	currentGameID int
	game          *database.Game
	network       *multiplayer.GameNetwork
	chessGame     *chess.Game
	incomingMoves chan string
	turn          int
}

// NewGameModel creates a new GameModel.
func NewGameModel(width, height int, peer string, currentGameID int, network *multiplayer.GameNetwork) GameModel {
	return GameModel{
		width:         width,
		height:        height,
		keys:          defaultGameKeyMap,
		peer:          peer,
		currentGameID: currentGameID,
		network:       network,
		chessGame:     chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		incomingMoves: make(chan string),
		turn:          0,
	}
}

// Init initializes the GameModel.
func (m GameModel) Init() tea.Cmd {
	ClearScreen()
	return tea.Batch(textinput.Blink, m.getGame(), m.getMoves())
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
	case ChessMoveMsg:
		m.turn++
		err := m.chessGame.MoveStr(string(msg))
		if err != nil {
			fmt.Println("Error applying move:", err)
		}
		return m, m.getMoves()
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
		yourTurn := ""
		if m.turn%2 == 0 && m.peer == "peer-2" || m.turn%2 == 1 && m.peer == "peer-1" {
			yourTurn = "[YOUR TURN]"
		}

		content = fmt.Sprintf("%s vs %s\n%s\n\n%s\n%s",
			m.game.Player1.Username,
			m.game.Player2.Username,
			lipgloss.NewStyle().Foreground(highlightColor).Render(yourTurn),
			m.chessGame.Position().Board().Draw(),
			m.chessGame.String(),
		)
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
	case key.Matches(msg, m.keys.RandomMove):
		if m.turn%2 == 0 && m.peer == "peer-2" || m.turn%2 == 1 && m.peer == "peer-1" {
			moves := m.chessGame.ValidMoves()
			move := moves[rand.Intn(len(moves))]
			m.network.Server.Send(network.NetworkID(m.peer), []byte(move.String()))
			m.chessGame.MoveStr(move.String())
			m.turn++
		}
		return m, nil
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
	randomMoveKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.RandomMove.Help().Key),
		m.keys.RandomMove.Help().Desc)

	logoutKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.GoLogout.Help().Key),
		m.keys.GoLogout.Help().Desc)

	quitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Quit.Help().Key),
		m.keys.Quit.Help().Desc)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		randomMoveKey,
		logoutKey,
		quitKey,
	)
}

func (m *GameModel) handleGetGameResponse(msg database.Game) (tea.Model, tea.Cmd) {
	m.game = &msg
	if m.peer == "peer-2" {
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
		if m.peer == "peer-2" {
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

func (m *GameModel) getMoves() tea.Cmd {
	m.network.Server.OnReceiveFn = func(msg network.Message) {
		moveStr := string(msg.Payload)
		m.incomingMoves <- moveStr
	}

	return func() tea.Msg {
		move := <-m.incomingMoves
		return ChessMoveMsg(move)
	}
}
