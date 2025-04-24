package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/boozec/rahanna/pkg/p2p"
	tea "github.com/charmbracelet/bubbletea"
)

// Catch for `RestoreGameMessage` message from multiplayer
type SendRestoreMsg p2p.NetworkID

// Catch for `RestoreAckGameMessage` message from multiplayer
type RestoreMoves string

// For `RestoreGameMessage` from multiplayer it fixes the peer with the new
// address and sends back an ACK to the peer' sender
func (m GameModel) handleSendRestoreMsg(source p2p.NetworkID) tea.Cmd {
	_ = m.getGame()()

	peers := map[int]string{
		1: m.game.IP1,
		2: m.game.IP2,
		3: m.game.IP3,
		4: m.game.IP4,
	}

	myPlayerNum := -1
	switch m.network.Me() {
	case m.playerPeer(1):
		myPlayerNum = 1
	case m.playerPeer(2):
		myPlayerNum = 2
	case m.playerPeer(3):
		myPlayerNum = 3
	case m.playerPeer(4):
		myPlayerNum = 4
	}

	// Add all peers to every other peer
	for playerNum, ip := range peers {
		if playerNum != myPlayerNum && ip != "" {
			m.network.AddPeer(m.playerPeer(playerNum), ip)
		}
	}

	// FIXME: add a loading modal
	time.Sleep(2 * time.Second)

	payload := ""

	for _, move := range m.chessGame.Moves() {
		payload += fmt.Sprintf("%s\n", move.String())
	}

	m.err = m.network.Send(source, []byte("restore-ack"), []byte(payload))

	return nil
}

// Restores the moves for `m.chessGame`
func (m *GameModel) handleRestoreMoves(msg RestoreMoves) tea.Cmd {
	moves := strings.Split(string(msg), "\n")
	for _, move := range moves {
		m.chessGame.MoveStr(move)
	}

	cmds := []tea.Cmd{m.getMoves(), m.updateMovesListCmd()}
	return tea.Batch(cmds...)
}
