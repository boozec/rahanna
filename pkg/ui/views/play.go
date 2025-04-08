package views

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/network"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/key"
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

	// Game state
	playName string
	play     *database.Game
	network  *multiplayer.PlayNetwork
}

// NewPlayModel creates a new play model instance
func NewPlayModel(width, height int) PlayModel {
	namePrompt := createNamePrompt()

	return PlayModel{
		width:      width,
		height:     height,
		err:        nil,
		keys:       defaultPlayKeyMap,
		namePrompt: namePrompt,
		page:       LandingPage,
		isLoading:  false,
		playName:   "",
		play:       nil,
	}
}

// Create and configure the name input prompt
func createNamePrompt() textinput.Model {
	namePrompt := textinput.New()
	namePrompt.Prompt = " "
	namePrompt.TextStyle = inputStyle
	namePrompt.Placeholder = "rectangular-lake"
	namePrompt.Focus()
	namePrompt.CharLimit = 23
	namePrompt.Width = 23

	return namePrompt
}

func (m PlayModel) Init() tea.Cmd {
	ClearScreen()
	return textinput.Blink
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
	base := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width)

	content := m.renderPageContent(base)

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
		return m, m.logout()

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyEnter:
		if m.page == InsertCodePage && !m.isLoading {
			m.isLoading = true
			return m, m.enterGame()
		}
	}

	return m, nil
}

func (m *PlayModel) handlePlayResponse(msg playResponse) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.err = nil

	if msg.Error != "" {
		m.err = fmt.Errorf(msg.Error)
		if msg.Error == "unauthorized" {
			return m, m.logout()
		}
	} else {
		m.playName = msg.Ok.Name
		m.network = multiplayer.NewPlayNetwork("peer-1", msg.Ok.IP, msg.Ok.Port)
	}

	return m, nil
}

func (m *PlayModel) handleGameResponse(msg database.Game) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.play = &msg
	m.err = nil
	return m, nil
}

func (m *PlayModel) handleError(msg error) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.err = msg
	return m, nil
}

func (m *PlayModel) renderPageContent(base lipgloss.Style) string {
	switch m.page {
	case LandingPage:
		m.namePrompt.Blur()
		return chessBoard

	case InsertCodePage:
		m.namePrompt.Focus()
		return m.renderInsertCodeContent(base)

	case StartGamePage:
		return m.renderStartGameContent(base)
	}

	return ""
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
	if m.play != nil {
		playerName := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e67e22")).
			Render(m.play.Player1.Username)

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
			lipgloss.NewStyle().Width(23).Render("Insert play code:"),
			m.namePrompt.View(),
			lipgloss.NewStyle().
				Align(lipgloss.Center).
				PaddingTop(2).
				Width(23).
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

		return lipgloss.JoinVertical(
			lipgloss.Left,
			enterKey,
			startKey,
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

func (m PlayModel) logout() tea.Cmd {
	if err := os.Remove(".rahannarc"); err != nil {
		return nil
	}
	return SwitchModelCmd(NewAuthModel(m.width, m.height+1))
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
