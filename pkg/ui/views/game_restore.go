package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Catch for `RestoreGameMessage` message from multiplayer
type SendRestoreMsg struct{}

// Catch for `RestoreAckGameMessage` message from multiplayer
type RestoreMoves string

// For `RestoreGameMessage` from multiplayer it fixes the peer with the new
// address and sends back an ACK to the peer' sender
func (m GameModel) handleSendRestoreMsg() tea.Cmd {
	if m.network.Me() == m.playerPeer(1) {
		_ = m.getGame()()
		remote := m.game.IP2
		m.network.AddPeer(m.playerPeer(2), remote)
	} else {
		_ = m.getGame()()
		remote := m.game.IP1
		m.network.AddPeer(m.playerPeer(1), remote)
	}

	// FIXME: add a loading modal
	time.Sleep(2 * time.Second)

	payload := ""

	for _, move := range m.chessGame.Moves() {
		payload += fmt.Sprintf("%s\n", move.String())
	}

	m.err = m.network.Send([]byte("restore-ack"), []byte(payload))

	return nil
}

// Restores the moves for `m.chessGame`
func (m *GameModel) handleRestoreMoves(msg RestoreMoves) tea.Cmd {
	moves := strings.Split(string(msg), "\n")
	for _, move := range moves {
		m.chessGame.MoveStr(move)
	}

	m.turn = len(moves) - 1
	cmds := []tea.Cmd{m.getMoves(), m.updateMovesListCmd()}
	return tea.Batch(cmds...)
}
