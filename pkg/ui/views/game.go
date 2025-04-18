package views

import (
	"fmt"
	"strings"

	"github.com/boozec/rahanna/internal/api/database"
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
	err  error
	keys gameKeyMap

	// Game state
	currentGameID      int
	game               *database.Game
	network            *multiplayer.GameNetwork
	chessGame          *chess.Game
	incomingMoves      chan multiplayer.GameMove
	turn               int
	availableMovesList list.Model
}

// NewGameModel creates a new GameModel.
func NewGameModel(width, height int, currentGameID int, network *multiplayer.GameNetwork) GameModel {
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
		width:              width,
		height:             height,
		keys:               defaultGameKeyMap,
		currentGameID:      currentGameID,
		network:            network,
		chessGame:          chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		incomingMoves:      make(chan multiplayer.GameMove),
		turn:               0,
		availableMovesList: moveList,
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
		m, cmd = m.handleDatabaseGameMsg(msg)
		cmds = append(cmds, cmd, m.updateMovesListCmd())
	case EndGameMsg:
		if msg.abandoned {
			if m.network.Me() == m.playerPeer(1) {
				m.game.Outcome = string(chess.WhiteWon)
			} else {
				m.game.Outcome = string(chess.BlackWon)
			}
			m, cmd = m.handleDatabaseGameMsg(*m.game)
			cmds = append(cmds, cmd)
		}

		m.err = m.network.Close()
	case error:
		m.err = msg
	}

	if m.isMyTurn() {
		m.availableMovesList, cmd = m.availableMovesList.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				selectedItem := m.availableMovesList.SelectedItem()
				if selectedItem != nil {
					moveStr := strings.Replace(selectedItem.(item).Title(), " → ", "", 1)
					moveStr = strings.Replace(moveStr, " ", "", 1)
					err := m.chessGame.MoveStr(moveStr)
					if err != nil {
						m.err = err
					} else {
						m.turn++
						m.network.Send([]byte("new-move"), []byte(moveStr))
						m.err = nil
					}
					cmds = append(cmds, m.getMoves(), m.updateMovesListCmd())

					if m.chessGame.Outcome() != chess.NoOutcome {
						cmds = append(cmds, m.endGame(m.chessGame.Outcome().String()))
					}
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

	var availableMovesListView string

	if m.game.Outcome == chess.NoOutcome.String() {
		if m.isMyTurn() {
			m.availableMovesList.SetSize(listWidth, listHeight-2)
			availableMovesListView = listStyle.Render(m.availableMovesList.View())
		} else {
			availableMovesListView = listStyle.Render(lipgloss.Place(listWidth, listHeight, lipgloss.Center, lipgloss.Center, "Wait your turn"))
		}
	} else {
		var outcome string
		switch m.game.Outcome {
		case string(chess.WhiteWon):
			outcome = "White won"
			if m.network.Me() == m.playerPeer(1) {
				outcome += " (YOU)"
			}
		case string(chess.BlackWon):
			outcome = "Black won"
			if m.network.Me() == m.playerPeer(2) {
				outcome += " (YOU)"
			}
		case string(chess.Draw):
			outcome = "Draw"
		default:
			outcome = "NoOutcome"
		}

		availableMovesListView = listStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Background(highlightColor).Foreground(lipgloss.Color("230")).Padding(0, 1).MarginBottom(1).Render("Result"),
				outcome,
				m.game.Outcome,
			),
		)
	}

	var movesListStr string

	for i, move := range m.chessGame.Moves() {
		s1 := move.S1().String()
		s2 := move.S2().String()
		var promo string

		if move.Promo().String() != "" {
			promo = " " + move.Promo().String()
		}

		if i%2 == 0 {
			movesListStr += altCodeStyle.Render(fmt.Sprintf("[%d]", i/2)) + fmt.Sprintf(" %s → %s%s", s1, s2, promo)
		} else {
			movesListStr += fmt.Sprintf(", %s → %s%s\n", s1, s2, promo)
		}
	}

	// TODO: a faster solution withoout strings.Split and strings.Join
	moves := strings.Split(movesListStr, "\n")
	start := 0
	if len(moves) > 10 {
		start = len(moves) - 10 - 1
	}

	movesListStr = strings.Join(moves[start:], "\n")

	var errorStr string
	if m.err != nil {
		errorStr = m.err.Error()
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#f1c40f")).Render(fmt.Sprintf("♔ %s vs ♚ %s", m.game.Player1.Username, m.game.Player2.Username)),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			availableMovesListView,
			boardStyle.Render(
				m.chessGame.Position().Board().Draw(),
			),
			notationStyle.Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					lipgloss.NewStyle().Background(highlightColor).Foreground(lipgloss.Color("230")).Padding(0, 1).MarginBottom(1).Render("Moves"),
					movesListStr,
				),
			),
		),
	)

	windowContent := m.buildWindowContent(content, formWidth)
	buttons := m.renderNavigationButtons()

	centeredContent := lipgloss.JoinVertical(
		lipgloss.Center,
		getLogo(m.width),
		windowContent,
		errorStyle.Width(formWidth/2).Render(errorStr),
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
