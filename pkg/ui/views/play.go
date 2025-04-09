package views

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/network"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/key"
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

type PlayModelPage int

var start = make(chan int)

const (
	LandingPage PlayModelPage = iota
	InsertCodePage
	StartGamePage
)

type responseOk struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port int    `json:"int"`
}

// API response types
type playResponse struct {
	Ok    responseOk
	Error string `json:"error"`
}

// Keyboard controls
type playKeyMap struct {
	EnterNewGame key.Binding
	StartNewGame key.Binding
	GoLogout     key.Binding
	Quit         key.Binding
	NextPage     key.Binding
	PrevPage     key.Binding
}

// Default key bindings for the play model
var defaultPlayKeyMap = playKeyMap{
	EnterNewGame: key.NewBinding(
		key.WithKeys("alt+E", "alt+e"),
		key.WithHelp("Alt+E", "Enter a play using code"),
	),
	StartNewGame: key.NewBinding(
		key.WithKeys("alt+s", "alt+S"),
		key.WithHelp("Alt+S", "Start a new play"),
	),
	GoLogout: key.NewBinding(
		key.WithKeys("alt+Q", "alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "q"),
		key.WithHelp("    Q", "Quit"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→/h", "Next Page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←/l", "Prev Page"),
	),
}

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
	playName string
	game     *database.Game
	network  *multiplayer.GameNetwork
	games    []database.Game // Store the list of games
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

// Create and configure the name input prompt
func createNamePrompt(width int) textinput.Model {
	namePrompt := textinput.New()
	namePrompt.Prompt = " "
	namePrompt.TextStyle = inputStyle
	namePrompt.Placeholder = "rectangular-lake"
	namePrompt.Focus()
	namePrompt.CharLimit = 23
	namePrompt.Width = getFormWidth(width)

	return namePrompt
}

func (m PlayModel) Init() tea.Cmd {
	ClearScreen()
	return tea.Batch(textinput.Blink, m.fetchGames())
}

func (m PlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if exit := handleExit(msg); exit != nil {
		return m, exit
	}

	select {
	case <-start:
		return m, SwitchModelCmd(NewGameModel(m.width, m.height+1, m.game, m.network))
	default:
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
	case error:
		return m.handleError(msg)
	}

	// Handle input updates when on the InsertCodePage
	if m.page == InsertCodePage {
		var cmd tea.Cmd
		m.namePrompt, cmd = m.namePrompt.Update(msg)
		return m, cmd
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

func (m PlayModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	return m, nil
}

func (m PlayModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.EnterNewGame):
		m.page = InsertCodePage
		return m, nil

	case key.Matches(msg, m.keys.StartNewGame):
		m.page = StartGamePage
		if !m.isLoading {
			m.isLoading = true
			return m, m.newGameCallback()
		}

	case key.Matches(msg, m.keys.GoLogout):
		return m, logout(m.width, m.height+1)

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyEnter:
		if m.page == InsertCodePage && !m.isLoading {
			m.isLoading = true
			return m, m.enterGame()
		}
	}

	m.paginator, _ = m.paginator.Update(msg)

	return m, nil
}

func (m *PlayModel) handlePlayResponse(msg playResponse) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.err = nil

	if msg.Error != "" {
		m.err = fmt.Errorf("%s", msg.Error)
		if msg.Error == "unauthorized" {
			return m, logout(m.width, m.height+1)
		}
	} else {
		m.playName = msg.Ok.Name

		m.network = multiplayer.NewGameNetwork("peer-1", msg.Ok.IP, msg.Ok.Port, func() {
			close(start)
		})
	}

	return m, nil
}

func (m *PlayModel) handleGameResponse(msg database.Game) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.game = &msg
	m.err = nil
	return m, nil
}

func (m *PlayModel) handleGamesResponse(msg []database.Game) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.games = msg
	m.err = nil
	m.paginator.SetTotalPages(len(m.games))
	return m, nil
}

func (m *PlayModel) handleError(msg error) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.err = msg
	return m, nil
}

func (m PlayModel) renderPageContent(base lipgloss.Style) string {
	switch m.page {
	case LandingPage:
		m.namePrompt.Blur()
		if len(m.games) == 0 {
			return base.Render(chessBoard)
		} else {
			start, end := m.paginator.GetSliceBounds(len(m.games))
			gamesStrings := formatGamesForPage(m.games[start:end], altCodeStyle)
			pageInfo := m.paginator.View()
			return base.Render(lipgloss.JoinVertical(lipgloss.Center, strings.Join(gamesStrings, "\n"), pageInfo))
		}
	case InsertCodePage:
		m.namePrompt.Focus()
		return m.renderInsertCodeContent(base)

	case StartGamePage:
		return m.renderStartGameContent(base)
	}

	return ""
}

func formatGamesForPage(games []database.Game, altCodeStyle lipgloss.Style) []string {
	var gamesStrings []string
	gamesStrings = append(gamesStrings, "Games list")

	longestName := 0
	for _, game := range games {
		if len(game.Name) > longestName {
			longestName = len(game.Name)
		}
	}

	for i, game := range games {
		indexStr := altCodeStyle.Render(fmt.Sprintf("[%d] ", i))
		nameStr := game.Name
		dateStr := game.UpdatedAt.Format("2006-01-02 15:04")

		padding := longestName - len(nameStr)
		paddingStr := strings.Repeat(" ", padding+4)

		line := lipgloss.JoinHorizontal(lipgloss.Left,
			indexStr,
			nameStr,
			paddingStr,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#d35400")).Render(dateStr),
		)
		gamesStrings = append(gamesStrings, line)
	}
	return gamesStrings
}

func (m PlayModel) renderInsertCodeContent(base lipgloss.Style) string {
	// When loading, show loading status
	if m.isLoading {
		return base.Render(
			lipgloss.NewStyle().
				Align(lipgloss.Center).
				Bold(true).
				Render("Loading..."),
		)
	}

	// When we have a play, show who we're playing against
	if m.game != nil {
		playerName := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e67e22")).
			Render(m.game.Player1.Username)

		statusMsg := fmt.Sprintf("You are playing versus %s", playerName)

		return base.Render(
			lipgloss.NewStyle().
				Align(lipgloss.Center).
				Width(m.width).
				Bold(true).
				Render(statusMsg),
		)
	}

	// Default: show input prompt
	return base.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Render("Insert play code:"),
			m.namePrompt.View(),
			lipgloss.NewStyle().
				Align(lipgloss.Center).
				PaddingTop(2).
				Bold(true).
				Render(fmt.Sprintf("Press %s to join",
					lipgloss.NewStyle().Italic(true).Render("Enter"))),
		),
	)
}

func (m PlayModel) renderStartGameContent(base lipgloss.Style) string {
	var statusMsg string

	if m.isLoading {
		statusMsg = "Loading..."
	} else if m.playName != "" {
		gameCode := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#F39C12")).
			Render(m.playName)

		statusMsg = fmt.Sprintf("Share `%s` to your friend", gameCode)
	}

	return base.Render(statusMsg)
}

func (m PlayModel) buildWindowContent(content string, formWidth int) string {
	if m.err != nil {
		formError := fmt.Sprintf("Error: %v", m.err.Error())
		return lipgloss.JoinVertical(
			lipgloss.Center,
			windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
				lipgloss.Center,
				errorStyle.Align(lipgloss.Center).Width(formWidth-4).Render(formError),
				content,
			)),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
			lipgloss.Center,
			content,
		)),
	)
}

func (m PlayModel) renderNavigationButtons() string {
	logoutKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.GoLogout.Help().Key),
		m.keys.GoLogout.Help().Desc)

	quitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Quit.Help().Key),
		m.keys.Quit.Help().Desc)

	if m.page == LandingPage {
		enterKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.EnterNewGame.Help().Key),
			m.keys.EnterNewGame.Help().Desc)

		startKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.StartNewGame.Help().Key),
			m.keys.StartNewGame.Help().Desc)

		nextPageKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.NextPage.Help().Key),
			m.keys.NextPage.Help().Desc)

		prevPageKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.PrevPage.Help().Key),
			m.keys.PrevPage.Help().Desc)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			enterKey,
			startKey,
			lipgloss.JoinHorizontal(lipgloss.Left, prevPageKey, " | ", nextPageKey),
			logoutKey,
			quitKey,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		logoutKey,
		quitKey,
	)
}

func (m *PlayModel) newGameCallback() tea.Cmd {
	return func() tea.Msg {
		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Set up network connection
		port, err := network.GetRandomAvailablePort()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		ip := network.GetOutboundIP().String()
		// FIXME: ip
		ip = "0.0.0.0"

		// Prepare request payload
		payload, err := json.Marshal(map[string]string{
			"ip": fmt.Sprintf("%s:%d", ip, port),
		})
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Send API request
		url := os.Getenv("API_BASE") + "/play"
		resp, err := sendAPIRequest("POST", url, payload, authorization)
		if err != nil {
			return playResponse{Error: err.Error()}
		}
		defer resp.Body.Close()

		// Handle response
		if resp.StatusCode != http.StatusOK {
			var response playResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return playResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return playResponse{Error: response.Error}
		}

		// Decode successful response
		var response struct {
			Name  string `json:"name"`
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return playResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return playResponse{Ok: responseOk{Name: response.Name, IP: ip, Port: port}}
	}
}

func (m PlayModel) enterGame() tea.Cmd {
	return func() tea.Msg {
		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Set up network connection
		port, err := network.GetRandomAvailablePort()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		ip := network.GetOutboundIP().String()
		// FIXME: ip
		ip = "0.0.0.0"

		// Prepare request payload
		payload, err := json.Marshal(map[string]string{
			"ip":   fmt.Sprintf("%s:%d", ip, port),
			"name": m.namePrompt.Value(),
		})
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Send API request
		url := os.Getenv("API_BASE") + "/enter-game"
		resp, err := sendAPIRequest("POST", url, payload, authorization)
		if err != nil {
			return playResponse{Error: err.Error()}
		}
		defer resp.Body.Close()

		// Handle response
		if resp.StatusCode != http.StatusOK {
			var response playResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return playResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return playResponse{Error: response.Error}
		}

		// Decode successful response
		var response database.Game
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return playResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return response
	}
}

// getAuthorizationToken reads the authentication token from the .rahannarc file
func getAuthorizationToken() (string, error) {
	f, err := os.Open(".rahannarc")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var authorization string
	for scanner.Scan() {
		authorization = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading auth token: %v", err)
	}

	return authorization, nil
}

// sendAPIRequest sends an HTTP request to the API with the given parameters
func sendAPIRequest(method, url string, payload []byte, authorization string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authorization))

	client := &http.Client{}
	return client.Do(req)
}

func (m *PlayModel) fetchGames() tea.Cmd {
	return func() tea.Msg {
		var games []database.Game
		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return games
		}

		// Send API request
		url := os.Getenv("API_BASE") + "/play"
		resp, err := sendAPIRequest("GET", url, nil, authorization)
		if err != nil {
			return games
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
			return []database.Game{}
		}
		return games
	}
}
