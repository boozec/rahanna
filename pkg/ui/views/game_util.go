package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m GameModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (GameModel, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	listWidth := m.width / 4
	m.availableMovesList.SetSize(listWidth, m.height/2)
	return m, m.updateMovesListCmd()
}

func (m GameModel) buildWindowContent(content string, formWidth int) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
			lipgloss.Center,
			content,
		)),
	)
}

func (m GameModel) isMyTurn() bool {
	return m.turn%2 == 0 && m.peer == "peer-2" || m.turn%2 == 1 && m.peer == "peer-1"
}
