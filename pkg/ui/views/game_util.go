package views

import (
	"fmt"

	"github.com/boozec/rahanna/pkg/p2p"
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
	return m.turn%2 == 0 && m.network.Me() == m.playerPeer(1) || m.turn%2 == 1 && m.network.Me() == m.playerPeer(2)
}

func (m GameModel) playerPeer(n int) p2p.NetworkID {
	if m.game == nil {
		return p2p.EmptyNetworkID
	}
	return p2p.NetworkID(fmt.Sprintf("%s-%d", m.game.Name, n))
}
