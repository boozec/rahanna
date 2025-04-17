package views

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/network"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

// gameKeyMap defines the key bindings for the game view.
type gameKeyMap struct {
	GoLogout key.Binding
	Quit     key.Binding
}

// defaultGameKeyMap provides the default key bindings for the game view.
var defaultGameKeyMap = gameKeyMap{
	GoLogout: key.NewBinding(
		key.WithKeys("alt+Q", "alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "q"),
		key.WithHelp("   Q", "Quit"),
	),
}

// ChessMoveMsg is a message containing a received chess move.
type ChessMoveMsg string

// UpdateMovesListMsg is a message to update the moves list
type UpdateMovesListMsg struct{}

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

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
	movesList     list.Model
}

// NewGameModel creates a new GameModel.
func NewGameModel(width, height int, peer string, currentGameID int, network *multiplayer.GameNetwork) GameModel {
	listDelegate := list.NewDefaultDelegate()
	listDelegate.ShowDescription = false
	listDelegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(highlightColor).
		Foreground(highlightColor).
		Padding(0, 0, 0, 1)

	moveList := list.New([]list.Item{}, listDelegate, width/4, height/2)
	moveList.Styles.Title = lipgloss.NewStyle().
		Background(highlightColor).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

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
		movesList:     moveList,
	}
}

// Init initializes the GameModel.
func (m GameModel) Init() tea.Cmd {
	ClearScreen()
	return tea.Batch(textinput.Blink, m.getGame(), m.getMoves(), m.updateMovesListCmd())
}

func (m *GameModel) updateMovesListCmd() tea.Cmd {
	return func() tea.Msg {
		return UpdateMovesListMsg{}
	}
}

// Update handles incoming messages and updates the GameModel.
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if exit := handleExit(msg); exit != nil {
		return m, exit
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m, cmd = m.handleWindowSizeMsg(msg)
		cmds = append(cmds, cmd)
	case UpdateMovesListMsg:
		m = m.handleUpdateMovesListMsg()
	case tea.KeyMsg:
		m, cmd = m.handleKeyMsg(msg)
		cmds = append(cmds, cmd)
	case ChessMoveMsg:
		m, cmd = m.handleChessMoveMsg(msg)
		cmds = append(cmds, cmd)
	case database.Game:
		m = m.handleDatabaseGameMsg(msg)
		cmds = append(cmds, m.updateMovesListCmd())
	}

	if m.isMyTurn() {
		m.movesList, cmd = m.movesList.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				selectedItem := m.movesList.SelectedItem()
				if selectedItem != nil {
					moveStr := selectedItem.(item).Title()
					m.network.Server.Send(network.NetworkID(m.peer), []byte(moveStr))
					m.chessGame.MoveStr(moveStr)
					m.turn++
					cmds = append(cmds, m.getMoves(), m.updateMovesListCmd())
				}
			}
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m GameModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (GameModel, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	listWidth := m.width / 4
	m.movesList.SetSize(listWidth, m.height/2)
	return m, m.updateMovesListCmd()
}

func (m GameModel) handleUpdateMovesListMsg() GameModel {
	if m.isMyTurn() && m.game != nil {
		var items []list.Item
		for _, move := range m.chessGame.ValidMoves() {
			items = append(items, item{title: move.String()})
		}
		m.movesList.SetItems(items)
		m.movesList.Title = "Choose a move"
		m.movesList.Select(0)
		m.movesList.SetShowFilter(true)
		m.movesList.SetFilteringEnabled(true)
	}
	return m
}

func (m GameModel) handleKeyMsg(msg tea.KeyMsg) (GameModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.GoLogout):
		return m, logout(m.width, m.height+1)
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}

	return m, nil
}

func (m GameModel) handleChessMoveMsg(msg ChessMoveMsg) (GameModel, tea.Cmd) {
	m.turn++
	err := m.chessGame.MoveStr(string(msg))
	if err != nil {
		fmt.Println("Error applying move:", err)
	}
	return m, tea.Batch(m.getMoves(), m.updateMovesListCmd())
}

func (m GameModel) handleDatabaseGameMsg(msg database.Game) GameModel {
	m.game = &msg
	if m.peer == "peer-2" {
		m.network.Peer = msg.IP2
	} else {
		m.network.Peer = msg.IP1
	}
	return m
}

// View renders the GameModel.
func (m GameModel) View() string {
	formWidth := getFormWidth(m.width)

	if m.game == nil {
		return "Loading game..."
	}

	listWidth := formWidth / 4
	boardWidth := formWidth / 2
	notationWidth := formWidth - listWidth - boardWidth - 2

	listHeight := m.height / 3
	boardHeight := m.height / 3
	notationHeight := m.height - listHeight - boardHeight - 2

	listStyle := lipgloss.NewStyle().Width(listWidth).Height(listHeight).Padding(0, 1)
	boardStyle := lipgloss.NewStyle().Width(boardWidth).Height(boardHeight).Align(lipgloss.Center).Padding(0, 1)
	notationStyle := lipgloss.NewStyle().Width(notationWidth).Height(notationHeight).Padding(0, 1)

	var movesListView string

	if m.isMyTurn() {
		m.movesList.SetSize(listWidth, listHeight-2)
		movesListView = listStyle.Render(m.movesList.View())
	} else {
		movesListView = listStyle.Render(lipgloss.Place(listWidth, listHeight, lipgloss.Center, lipgloss.Center, "Wait your turn"))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#f1c40f")).Render(fmt.Sprintf("%s vs %s", m.game.Player1.Username, m.game.Player2.Username)),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			movesListView,
			boardStyle.Render(
				m.chessGame.Position().Board().Draw(),
			),
			notationStyle.Render(fmt.Sprintf("Moves\n%s", m.chessGame.String())),
		),
	)

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

func (m GameModel) isMyTurn() bool {
	return m.turn%2 == 0 && m.peer == "peer-2" || m.turn%2 == 1 && m.peer == "peer-1"
}
