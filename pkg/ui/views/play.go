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
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var chess string = `
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

type playKeyMap struct {
	EnterNewGame key.Binding
	StartNewGame key.Binding
	GoLogout     key.Binding
	Quit         key.Binding
}

type playResponse struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

var defaultGameKeyMap = playKeyMap{
	EnterNewGame: key.NewBinding(
		key.WithKeys("alt+E", "alt+e"),
		key.WithHelp("Alt+E", "Enter a play using code"),
	),
	StartNewGame: key.NewBinding(
		key.WithKeys("alt+s", "alt+s"),
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

type PlayModelPage int

const (
	LandingPage PlayModelPage = iota
	InsertCodePage
	StartGamePage
)

type PlayModel struct {
	width      int
	height     int
	err        error
	keys       playKeyMap
	namePrompt textinput.Model
	page       PlayModelPage
	isLoading  bool
	playName   string
	play       *database.Game
}

func NewPlayModel(width, height int) PlayModel {
	namePrompt := textinput.New()
	namePrompt.Prompt = " "
	namePrompt.TextStyle = inputStyle
	namePrompt.Placeholder = "rectangular-lake"
	namePrompt.Focus()
	namePrompt.CharLimit = 23
	namePrompt.Width = 23

	return PlayModel{
		width:      width,
		height:     height,
		err:        nil,
		keys:       defaultGameKeyMap,
		namePrompt: namePrompt,
		page:       LandingPage,
		isLoading:  false,
		playName:   "",
		play:       nil,
	}
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
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
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
			if m.page == InsertCodePage {
				if !m.isLoading {
					m.isLoading = true
					return m, m.enterGame()
				}
			}
		}
	case playResponse:
		m.isLoading = false
		m.err = nil
		if msg.Error != "" {
			m.err = fmt.Errorf(msg.Error)
			if msg.Error == "unauthorized" {
				return m, m.logout()
			}
		} else {
			m.playName = msg.Name
		}
		return m, nil
	case database.Game:
		m.isLoading = false
		m.play = &msg
		m.err = nil
		return m, nil
	case error:
		m.isLoading = false
		m.err = msg
	}

	var cmd tea.Cmd = nil

	if m.page == InsertCodePage {
		m.namePrompt, cmd = m.namePrompt.Update(msg)
	}

	return m, tea.Batch(cmd)
}

func (m PlayModel) View() string {
	formWidth := getFormWidth(m.width)

	var content string
	base := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width)

	switch m.page {
	case LandingPage:
		content = chess
		m.namePrompt.Blur()
	case InsertCodePage:
		m.namePrompt.Focus()
		var statusMsg string
		if m.isLoading {
			statusMsg = "Loading..."
			content = base.
				Render(
					lipgloss.NewStyle().
						Align(lipgloss.Center).
						Bold(true).
						Render(statusMsg),
				)
		} else if m.play != nil {
			statusMsg = fmt.Sprintf("You are playing versus %s", lipgloss.NewStyle().Foreground(lipgloss.Color("#e67e22")).Render(m.play.Player1.Username))
			content = base.
				Render(
					lipgloss.NewStyle().
						Align(lipgloss.Center).
						Width(m.width).
						Bold(true).
						Render(statusMsg),
				)
		} else {
			statusMsg = fmt.Sprintf("Press %s to join", lipgloss.NewStyle().Italic(true).Render("Enter"))
			content = base.
				Render(
					lipgloss.JoinVertical(lipgloss.Left,
						lipgloss.NewStyle().Width(23).Render("Insert play code:"),
						m.namePrompt.View(),
						lipgloss.NewStyle().
							Align(lipgloss.Center).
							PaddingTop(2).
							Width(23).
							Bold(true).
							Render(statusMsg),
					),
				)
		}

	case StartGamePage:
		var statusMsg string
		if m.isLoading {
			statusMsg = "Loading..."
		} else if m.playName != "" {
			statusMsg = fmt.Sprintf("Share `%s` to your friend", lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#F39C12")).Render(m.playName))
		}

		content = base.
			Render(statusMsg)
	}

	var windowContent string
	if m.err != nil {
		formError := fmt.Sprintf("Error: %v", m.err.Error())
		windowContent = lipgloss.JoinVertical(
			lipgloss.Center,
			windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
				lipgloss.Center,
				errorStyle.Align(lipgloss.Center).Width(formWidth-4).Render(formError),
				content,
			)),
		)
	} else {
		windowContent = lipgloss.JoinVertical(
			lipgloss.Center,
			windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
				lipgloss.Center,
				content,
			)),
		)
	}

	enterKey := fmt.Sprintf("%s %s", altCodeStyle.Render(m.keys.EnterNewGame.Help().Key), m.keys.EnterNewGame.Help().Desc)
	startKey := fmt.Sprintf("%s %s", altCodeStyle.Render(m.keys.StartNewGame.Help().Key), m.keys.StartNewGame.Help().Desc)
	logoutKey := fmt.Sprintf("%s %s", altCodeStyle.Render(m.keys.GoLogout.Help().Key), m.keys.GoLogout.Help().Desc)
	quitKey := fmt.Sprintf("%s %s", altCodeStyle.Render(m.keys.Quit.Help().Key), m.keys.Quit.Help().Desc)

	// Vertically align the buttons
	buttons := lipgloss.JoinVertical(
		lipgloss.Left,
		enterKey,
		startKey,
		logoutKey,
		quitKey,
	)

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

func (m PlayModel) newGameCallback() tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(".rahannarc")
		if err != nil {
			return playResponse{Error: err.Error()}
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		var authorization string
		for scanner.Scan() {
			authorization = scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error during scanning:", err)
		}

		url := os.Getenv("API_BASE") + "/play"

		port, err := network.GetRandomAvailablePort()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		payload, err := json.Marshal(map[string]string{
			"ip": fmt.Sprintf("%s:%d", network.GetOutboundIP().String(), port),
		})

		if err != nil {
			return playResponse{Error: err.Error()}
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authorization))

		client := &http.Client{}

		resp, err := client.Do(req)

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var response playResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return playResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return playResponse{Error: response.Error}
		}

		var response playResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return playResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return response
	}
}

func (m PlayModel) enterGame() tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(".rahannarc")
		if err != nil {
			return playResponse{Error: err.Error()}
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		var authorization string
		for scanner.Scan() {
			authorization = scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error during scanning:", err)
		}

		url := os.Getenv("API_BASE") + "/enter-game"

		port, err := network.GetRandomAvailablePort()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		payload, err := json.Marshal(map[string]string{
			"ip":   fmt.Sprintf("%s:%d", network.GetOutboundIP().String(), port),
			"name": m.namePrompt.Value(),
		})

		if err != nil {
			return playResponse{Error: err.Error()}
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authorization))

		client := &http.Client{}

		resp, err := client.Do(req)

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var response playResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return playResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return playResponse{Error: response.Error}
		}

		var response database.Game
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
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
