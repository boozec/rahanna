package views

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model holds the state for login page
type loginModel struct {
	username  textinput.Model
	password  textinput.Model
	focus     int
	err       error
	isLoading bool
	token     string
	width     int
	height    int
}

// Response from API
type loginResponse struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

// Initialize loginModel
func LoginModel() loginModel {
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7EE2A8"))

	username := textinput.New()
	username.Prompt = " "
	username.TextStyle = inputStyle
	username.Placeholder = "mario.rossi"
	username.Focus()
	username.CharLimit = 156
	username.Width = 30

	password := textinput.New()
	password.Prompt = " "
	password.TextStyle = inputStyle
	password.Placeholder = "*****"
	password.EchoMode = textinput.EchoPassword
	password.CharLimit = 156
	password.Width = 30

	width, height := GetTerminalSize()

	return loginModel{
		username:  username,
		password:  password,
		focus:     0,
		err:       nil,
		isLoading: false,
		token:     "",
		width:     width,
		height:    height,
	}
}

// Init function
func (m loginModel) Init() tea.Cmd {
	ClearScreen()
	return textinput.Blink
}

// Update function
func (m loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.focus = 0
			m.username.Focus()
			m.password.Blur()
		case tea.KeyDown:
			m.focus = 1
			m.username.Blur()
			m.password.Focus()
		case tea.KeyEnter:
			if !m.isLoading {
				m.isLoading = true
				return m, m.loginCallback()
			}
		case tea.KeyTab:
			m.focus = (m.focus + 1) % 2
			if m.focus == 0 {
				m.username.Focus()
				m.password.Blur()
			} else {
				m.username.Blur()
				m.password.Focus()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case loginResponse:
		m.isLoading = false
		if msg.Error != "" {
			m.err = fmt.Errorf(msg.Error)
			m.focus = 0
			m.username.Focus()
			m.password.Blur()
		} else {
			m.token = msg.Token
			ClearScreen()
			f, err := os.OpenFile(".rahannarc", os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				m.err = err
				break
			}
			defer f.Close()

			f.Write([]byte(m.token))

			return m, tea.Quit
		}
	case error:
		m.isLoading = false
		m.err = msg
		m.focus = 0
		m.username.Focus()
		m.password.Blur()
	}

	var cmd tea.Cmd
	m.username, cmd = m.username.Update(msg)
	cmdPassword := tea.Batch(cmd)
	m.password, cmd = m.password.Update(msg)
	return m, tea.Batch(cmd, cmdPassword)
}

// Login API callback
func (m loginModel) loginCallback() tea.Cmd {
	return func() tea.Msg {
		url := os.Getenv("API_BASE") + "/auth/login"

		payload, err := json.Marshal(map[string]string{
			"username": m.username.Value(),
			"password": m.password.Value(),
		})

		if err != nil {
			return loginResponse{Error: err.Error()}
		}

		resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil {
			return loginResponse{Error: err.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var response loginResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return loginResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return loginResponse{Error: response.Error}
		}

		var response loginResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return loginResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return response
	}
}

// View function (UI rendering)
func (m loginModel) View() string {
	width, height := m.width, m.height
	formWidth := getFormWidth(width)

	// Styles
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7ee2a8")).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00ffcc")).
		Padding(1, 2).
		Align(lipgloss.Center).
		Width(formWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7ee2a8")).
		Align(lipgloss.Center).
		Width(formWidth - 4) // Account for padding

	labelStyle := lipgloss.NewStyle().
		Width(10).
		Align(lipgloss.Right)

	inputWrapStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(formWidth - 4) // Account for padding

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff0000")).
		Align(lipgloss.Center).
		Width(formWidth - 4) // Account for padding

	statusStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Bold(true).
		Width(formWidth - 4) // Account for padding

	// Error message
	formError := ""
	if m.err != nil {
		formError = fmt.Sprintf("Error: %v", m.err.Error())
	}

	// Status message
	statusMsg := fmt.Sprintf("Press %s to login", lipgloss.NewStyle().Italic(true).Render("Enter"))
	if m.isLoading {
		statusMsg = "Logging in..."
	}

	form := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("Sign in to your account"),
		"\n",
		errorStyle.Render(formError),
		inputWrapStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				labelStyle.Render("Username:"),
				m.username.View(),
			),
		),
		inputWrapStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				labelStyle.Render("Password:"),
				m.password.View(),
			),
		),
		"\n",
		statusStyle.Render(statusMsg),
	)

	// Wrap content inside a border
	borderedForm := borderStyle.Render(form)

	// Center logo and form in available space
	contentHeight := lipgloss.Height(logo) + lipgloss.Height(borderedForm) + 2
	paddingTop := (height - contentHeight) / 2
	if paddingTop < 0 {
		paddingTop = 0
	}

	// Combine logo and form with vertical centering
	output := lipgloss.NewStyle().
		MarginTop(paddingTop).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				logoStyle.Render(logo),
				lipgloss.PlaceHorizontal(width, lipgloss.Center, borderedForm),
			),
		)

	return output
}
