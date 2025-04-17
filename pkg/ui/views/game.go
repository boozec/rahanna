package views

import (
	"fmt"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/network"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

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
