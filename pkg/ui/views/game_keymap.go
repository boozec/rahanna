package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// gameKeyMap defines the key bindings for the game view.
type gameKeyMap struct {
	GoLogout key.Binding
	Quit     key.Binding
}

// defaultGameKeyMap provides the default key bindings for the game view.
var defaultGameKeyMap = gameKeyMap{
	GoLogout: key.NewBinding(
		key.WithKeys("alt+Q", "alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "q"),
		key.WithHelp("   Q", "Quit"),
	),
}

func (m GameModel) handleKeyMsg(msg tea.KeyMsg) (GameModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.GoLogout):
		return m, logout(m.width, m.height+1)
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}

	return m, nil
}

func (m GameModel) renderNavigationButtons() string {
	logoutKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.GoLogout.Help().Key),
		m.keys.GoLogout.Help().Desc)

	quitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Quit.Help().Key),
		m.keys.Quit.Help().Desc)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		logoutKey,
		quitKey,
	)
}
