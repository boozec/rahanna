package views

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PlayModel struct {
	width  int
	height int
}

func NewPlayModel(width, height int) PlayModel {

	return PlayModel{
		width:  width,
		height: height,
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
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, SwitchModelCmd(NewAuthModel(m.width, m.height))

		}
	}
	return m, nil
}

func (m PlayModel) View() string {
	width, height := m.width, m.height

	// Create the window with tab content
	ui := lipgloss.JoinVertical(lipgloss.Center,
		windowStyle.Width(getFormWidth(width)).Render("New Play"),
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
