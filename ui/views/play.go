package views

import (
	"errors"
	"fmt"
	"os"

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
	EnterNewPlay key.Binding
	StartNewPlay key.Binding
	GoLogout     key.Binding
	Quit         key.Binding
}

var defaultPlayKeyMap = playKeyMap{
	EnterNewPlay: key.NewBinding(
		key.WithKeys("alt+E", "alt+e"),
		key.WithHelp("Alt+E", "Enter a play using code"),
	),
	StartNewPlay: key.NewBinding(
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
	StartPlayPage
)

type PlayModel struct {
	width      int
	height     int
	err        error
	keys       playKeyMap
	namePrompt textinput.Model
	page       PlayModelPage
	isLoading  bool
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
		keys:       defaultPlayKeyMap,
		namePrompt: namePrompt,
		page:       LandingPage,
		isLoading:  false,
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
		case key.Matches(msg, m.keys.EnterNewPlay):
			m.page = InsertCodePage
			return m, nil
		case key.Matches(msg, m.keys.StartNewPlay):
			// TODO: handle new play
			return m, nil
		case key.Matches(msg, m.keys.GoLogout):
			if err := os.Remove(".rahannarc"); err != nil {
				m.err = err
				return m, nil
			}
			return m, SwitchModelCmd(NewAuthModel(m.width, m.height+1))
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case msg.Type == tea.KeyEnter:
			if m.page == InsertCodePage {
				m.err = errors.New("Can't join for now...")
			}
		}
	}

	var cmd tea.Cmd = nil

	if m.page == InsertCodePage {
		m.namePrompt, cmd = m.namePrompt.Update(msg)
	}

	return m, tea.Batch(cmd)
}

func (m PlayModel) View() string {
	formWidth := getFormWidth(m.width)

	// Error message
	formError := ""
	if m.err != nil {
		formError = fmt.Sprintf("Error: %v", m.err.Error())
	}

	// Status message
	statusMsg := fmt.Sprintf("Press %s to join", lipgloss.NewStyle().Italic(true).Render("Enter"))
	if m.isLoading {
		statusMsg = "Creating account..."
	}

	var content string

	switch m.page {
	case LandingPage:
		content = chess
		m.namePrompt.Blur()
	case InsertCodePage:
		m.namePrompt.Focus()
		content = m.namePrompt.View()

		content = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(m.width).
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

	windowContent := lipgloss.JoinVertical(
		lipgloss.Center,
		windowStyle.
			Width(formWidth).
			Render(lipgloss.JoinVertical(
				lipgloss.Center,
				errorStyle.Align(lipgloss.Center).Width(formWidth-4).Render(formError),
				content,
			)),
	)

	enterKey := fmt.Sprintf("%s %s", altCodeStyle.Render(m.keys.EnterNewPlay.Help().Key), m.keys.EnterNewPlay.Help().Desc)
	startKey := fmt.Sprintf("%s %s", altCodeStyle.Render(m.keys.StartNewPlay.Help().Key), m.keys.StartNewPlay.Help().Desc)
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
