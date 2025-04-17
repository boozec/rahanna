package views

import (
	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	chessBoard = `
A B C D E F G H
+---------------+
8 |♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜| 8
7 |♟ ♟ ♟ ♟ ♟ ♟ ♟ ♟| 7
6 |. . . . . . . .| 6
5 |. . . . . . . .| 5
4 |. . . . . . . .| 4
3 |. . . . . . . .| 3
2 |♙ ♙ ♙ ♙ ♙ ♙ ♙ ♙| 2
1 |♖ ♘ ♗ ♕ ♔ ♗ ♘ ♖| 1
+---------------+
A B C D E F G H
`
)

type PlayModel struct {
	// UI dimensions
	width  int
	height int

	// UI state
	err        error
	keys       playKeyMap
	namePrompt textinput.Model
	page       PlayModelPage
	isLoading  bool
	paginator  paginator.Model

	// Game state
	playName      string
	currentGameId int
	game          *database.Game
	network       *multiplayer.GameNetwork
	games         []database.Game
}

// NewPlayModel creates a new play model instance
func NewPlayModel(width, height int) PlayModel {
	namePrompt := createNamePrompt(width)
	p := paginator.New()
	p.PerPage = 10
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

	return PlayModel{
		width:      width,
		height:     height,
		keys:       defaultPlayKeyMap,
		namePrompt: namePrompt,
		page:       LandingPage,
		paginator:  p,
	}
}

func (m PlayModel) Init() tea.Cmd {
	ClearScreen()
	return tea.Batch(textinput.Blink, m.fetchGames())
}

func (m PlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if exit := handleExit(msg); exit != nil {
		return m, exit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case playResponse:
		return m.handlePlayResponse(msg)
	case database.Game:
		return m.handleGameResponse(msg)
	case []database.Game:
		return m.handleGamesResponse(msg)
	case StartGameMsg:
		return m, SwitchModelCmd(NewGameModel(m.width, m.height+1, "peer-2", m.currentGameId, m.network))
	case error:
		return m.handleError(msg)
	}

	return m, nil
}

func (m PlayModel) View() string {
	formWidth := getFormWidth(m.width)
	base := lipgloss.NewStyle().Align(lipgloss.Center).Width(formWidth)

	content := m.renderPageContent(base)

	// Build the main window with error handling
	windowContent := m.buildWindowContent(content, formWidth)

	// Create navigation buttons
	buttons := m.renderNavigationButtons()

	centeredContent := lipgloss.JoinVertical(
		lipgloss.Center,
		getLogo(formWidth),
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
