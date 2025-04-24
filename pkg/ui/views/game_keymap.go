package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

// gameKeyMap defines the key bindings for the game view.
type gameKeyMap struct {
	Abandon key.Binding
	Quit    key.Binding
	Exit    key.Binding
}

// defaultGameKeyMap provides the default key bindings for the game view.
var defaultGameKeyMap = gameKeyMap{
	Abandon: key.NewBinding(
		key.WithKeys("A", "a"),
		key.WithHelp("     A", "Abandon"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "q"),
		key.WithHelp("     Q", "Quit"),
	),
	Exit: key.NewBinding(
		key.WithKeys("ctrl+c", "ctrl+C"),
		key.WithHelp("Ctrl+C", "Exit"),
	),
}

func (m GameModel) handleKeyMsg(msg tea.KeyMsg) (GameModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Abandon):
		// Abandon game only if it is not finished
		if m.game.Outcome == "*" {
			var outcome string
			if m.network.Me() == m.playerPeer(1) || m.network.Me() == m.playerPeer(3) {
				outcome = string(chess.BlackWon)
			} else {
				outcome = string(chess.WhiteWon)
			}

			m.network.SendAll([]byte("abandon"), []byte("🏳️"))
			return m, m.endGame(outcome)
		}
	case key.Matches(msg, m.keys.Quit):
		return m, SwitchModelCmd(NewPlayModel(m.width, m.height))
	}

	return m, nil
}

func (m GameModel) renderNavigationButtons() string {
	var abandonKey string
	if m.game.Outcome == "*" {
		abandonKey = fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.Abandon.Help().Key),
			m.keys.Abandon.Help().Desc)
	}

	quitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Quit.Help().Key),
		m.keys.Quit.Help().Desc)

	exitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Exit.Help().Key),
		m.keys.Exit.Help().Desc)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		abandonKey,
		quitKey,
		exitKey,
	)
}
