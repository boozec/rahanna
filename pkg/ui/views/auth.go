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

const (
	SignInTab TabType = iota
	SignUpTab
)

// AuthModel is the main container model for both login and signup tabsuth
type AuthModel struct {
	loginModel  loginModel
	signupModel signupModel
	activeTab   TabType
	width       int
	height      int
}

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

// Model holds the state for signup page
type signupModel struct {
	loginModel
	confirmPassword textinput.Model
}

// Response from API
type authResponse struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

// Initialize AuthModel which contains both tabs
func NewAuthModel(width, height int) AuthModel {
	return AuthModel{
		loginModel:  initLoginModel(width, height),
		signupModel: initSignupModel(width, height),
		activeTab:   SignInTab,
		width:       width,
		height:      height,
	}
}

// Initialize loginModel
func initLoginModel(width, height int) loginModel {
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

// Initialize signupModel
func initSignupModel(width, height int) signupModel {
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

	confirmPassword := textinput.New()
	confirmPassword.Prompt = " "
	confirmPassword.TextStyle = inputStyle
	confirmPassword.Placeholder = "*****"
	confirmPassword.EchoMode = textinput.EchoPassword
	confirmPassword.CharLimit = 156
	confirmPassword.Width = 30

	return signupModel{
		loginModel: loginModel{
			username:  username,
			password:  password,
			focus:     0,
			err:       nil,
			isLoading: false,
			token:     "",
			width:     width,
			height:    height,
		},
		confirmPassword: confirmPassword,
	}
}

// Init function for AuthModel
func (m AuthModel) Init() tea.Cmd {
	ClearScreen()
	return textinput.Blink
}

func (m AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if exit := handleExit(msg); exit != nil {
		return m, exit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "alt+1":
			// Switch to sign-in tab
			if m.activeTab != SignInTab {
				m.activeTab = SignInTab
				m.loginModel.focus = 0
				m.loginModel.username.Focus()
				m.loginModel.password.Blur()
				m.signupModel.username.Blur()
				m.signupModel.password.Blur()
				m.signupModel.confirmPassword.Blur()
			}
			return m, nil

		case "alt+2":
			// Switch to sign-up tab
			if m.activeTab != SignUpTab {
				m.activeTab = SignUpTab
				m.signupModel.focus = 0
				m.signupModel.username.Focus()
				m.signupModel.password.Blur()
				m.signupModel.confirmPassword.Blur()
				m.loginModel.username.Blur()
				m.loginModel.password.Blur()
			}
			return m, nil

		}
	}

	if m.activeTab == SignInTab {
		var cmd tea.Cmd
		m.loginModel, cmd = m.loginModel.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.signupModel, cmd = m.signupModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View function for AuthModel
func (m AuthModel) View() string {
	width, height := m.width, m.height

	// Get the content of the active tab
	var tabContent string
	if m.activeTab == SignInTab {
		tabContent = m.loginModel.renderContent()
	} else {
		tabContent = m.signupModel.renderContent()
	}

	// Create the window with tab content
	ui := lipgloss.JoinVertical(lipgloss.Center,
		getTabsRow([]string{"Sign In", "Sign Up"}, m.activeTab),
		windowStyle.Width(getFormWidth(width)).Render(tabContent),
	)

	// Center logo and form in available space
	contentHeight := lipgloss.Height(logo) + lipgloss.Height(ui) + 2
	paddingTop := (height - contentHeight) / 2
	if paddingTop < 0 {
		paddingTop = 0
	}

	// Combine logo and tabs with vertical centering
	output := lipgloss.NewStyle().
		MarginTop(paddingTop).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				getLogo(m.width),
				lipgloss.PlaceHorizontal(width, lipgloss.Center, ui),
			),
		)

	return output
}

// Update function for loginModel
func (m loginModel) Update(msg tea.Msg) (loginModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.focus = (m.focus - 1) % 2
			if m.focus < 0 {
				m.focus = 1
			}
			m.updateFocus()
		case tea.KeyDown:
			m.focus = (m.focus + 1) % 2
			m.updateFocus()
		case tea.KeyEnter:
			if !m.isLoading {
				m.isLoading = true
				return m, m.loginCallback()
			}
		case tea.KeyTab:
			m.focus = (m.focus + 1) % 2
			m.updateFocus()
		}
	case authResponse:
		m.isLoading = false
		if msg.Error != "" {
			m.err = fmt.Errorf("%s", msg.Error)
			m.focus = 0
			m.updateFocus()
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
			return m, SwitchModelCmd(NewPlayModel(m.width, m.height))
		}
	case error:
		m.isLoading = false
		m.err = msg
		m.focus = 0
		m.updateFocus()
	}

	var cmd tea.Cmd
	m.username, cmd = m.username.Update(msg)
	cmdPassword := tea.Batch(cmd)
	m.password, cmd = m.password.Update(msg)
	return m, tea.Batch(cmd, cmdPassword)
}

// Update function for signupModel
func (m signupModel) Update(msg tea.Msg) (signupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.focus = (m.focus - 1) % 3
			if m.focus < 0 {
				m.focus = 2
			}
			m.updateFocus()
		case tea.KeyDown:
			m.focus = (m.focus + 1) % 3
			m.updateFocus()
		case tea.KeyEnter:
			if !m.isLoading {
				m.isLoading = true
				return m, m.signupCallback()
			}
		case tea.KeyTab:
			m.focus = (m.focus + 1) % 3
			m.updateFocus()
		}
	case authResponse:
		m.isLoading = false
		if msg.Error != "" {
			m.err = fmt.Errorf("%s", msg.Error)
			m.focus = 0
			m.updateFocus()
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
			return m, SwitchModelCmd(NewPlayModel(m.width, m.height))
		}
	case error:
		m.isLoading = false
		m.err = msg
		m.focus = 0
		m.updateFocus()
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.username, cmd = m.username.Update(msg)
	cmds = append(cmds, cmd)

	m.password, cmd = m.password.Update(msg)
	cmds = append(cmds, cmd)

	m.confirmPassword, cmd = m.confirmPassword.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Helper function to update input focus for signup
func (m *signupModel) updateFocus() {
	m.username.Blur()
	m.password.Blur()
	m.confirmPassword.Blur()

	switch m.focus {
	case 0:
		m.username.Focus()
	case 1:
		m.password.Focus()
	case 2:
		m.confirmPassword.Focus()
	}
}

// Helper function to update input focus for signin
func (m *loginModel) updateFocus() {
	m.username.Blur()
	m.password.Blur()

	switch m.focus {
	case 0:
		m.username.Focus()
	case 1:
		m.password.Focus()
	}
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
			return authResponse{Error: err.Error()}
		}

		resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil {
			return authResponse{Error: err.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var response authResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return authResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return authResponse{Error: response.Error}
		}

		var response authResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return authResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return response
	}
}

// Signup API callback
func (m signupModel) signupCallback() tea.Cmd {
	return func() tea.Msg {
		// Validate that passwords match
		if m.password.Value() != m.confirmPassword.Value() {
			return authResponse{Error: "Passwords do not match"}
		}

		url := os.Getenv("API_BASE") + "/auth/register"

		payload, err := json.Marshal(map[string]string{
			"username": m.username.Value(),
			"password": m.password.Value(),
		})

		if err != nil {
			return authResponse{Error: err.Error()}
		}

		resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil {
			return authResponse{Error: err.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			var response authResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return authResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return authResponse{Error: response.Error}
		}

		var response authResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return authResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return response
	}
}

// Render content of the login tab
func (m loginModel) renderContent() string {
	formWidth := getFormWidth(m.width)

	// Styles
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
		errorStyle.Align(lipgloss.Center).Width(formWidth-4).Render(formError),
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

	return form
}

// Render content of the signup tab
func (m signupModel) renderContent() string {
	formWidth := getFormWidth(m.width)

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7ee2a8")).
		Align(lipgloss.Center).
		Width(formWidth - 4) // Account for padding

	labelStyle := lipgloss.NewStyle().
		Width(16).
		Align(lipgloss.Right)

	inputWrapStyle := lipgloss.NewStyle().
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
	statusMsg := fmt.Sprintf("Press %s to register", lipgloss.NewStyle().Italic(true).Render("Enter"))
	if m.isLoading {
		statusMsg = "Creating account..."
	}

	form := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("Create a new account"),
		"\n",
		errorStyle.Align(lipgloss.Center).Width(formWidth-4).Render(formError),
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
		inputWrapStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				labelStyle.Render("Confirm:"),
				m.confirmPassword.View(),
			),
		),
		"\n",
		statusStyle.Render(statusMsg),
	)

	return form
}
