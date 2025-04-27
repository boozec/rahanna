package views

import (
	"fmt"
	"math/rand"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/pkg/p2p"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/notnil/chess"
)

// UpdateMovesListMsg is a message to update the moves list
type UpdateMovesListMsg struct{}

// ChessMoveMsg is a message containing a received chess move.
type ChessMoveMsg string

type SendNewTurnMsg struct{}
type SaveTurnMsg string

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

func (m *GameModel) getMoves() tea.Cmd {
	m.network.AddReceiveFunction(func(msg p2p.Message) {
		gm := multiplayer.GameMove{
			Source:  msg.Source,
			Type:    msg.Type,
			Payload: msg.Payload,
		}
		m.incomingMoves <- gm
	})

	return func() tea.Msg {
		move := <-m.incomingMoves

		switch multiplayer.MoveType(string(move.Type)) {
		case multiplayer.AbandonGameMessage:
			return EndGameMsg{abandoned: true}
		case multiplayer.DefineTurnMessage:
			return SaveTurnMsg(string(move.Payload))
		case multiplayer.RestoreGameMessage:
			return SendRestoreMsg(move.Source)
		case multiplayer.RestoreAckGameMessage:
			return RestoreMoves(string(move.Payload))
		default:
			return ChessMoveMsg(string(move.Payload))
		}
	}
}

func (m *GameModel) updateMovesListCmd() tea.Cmd {
	return func() tea.Msg {
		return UpdateMovesListMsg{}
	}
}

func (m GameModel) handleUpdateMovesListMsg() GameModel {
	if m.isMyTurn() && m.game != nil {
		var items []list.Item
		for _, move := range m.chessGame.ValidMoves() {
			var promo string
			if move.Promo().String() != "" {
				promo = " " + move.Promo().String()
			}
			items = append(
				items,
				item{title: fmt.Sprintf("%s → %s%s", move.S1().String(), move.S2().String(), promo)},
			)
		}
		m.availableMovesList.SetItems(items)
		m.availableMovesList.Title = "Choose a move"
		m.availableMovesList.Select(0)
		m.availableMovesList.SetShowFilter(true)
		m.availableMovesList.SetFilteringEnabled(true)
		m.availableMovesList.ResetFilter()
	}
	return m
}

func (m GameModel) handleDefineTurnMsg() (GameModel, tea.Cmd) {
	cmds := []tea.Cmd{m.getMoves(), m.updateMovesListCmd()}

	switch m.game.Type {
	case database.SingleGameType:
		if m.network.Me() == m.playerPeer(1) {
			m.turn = m.playerPeer(2)
		} else {
			m.turn = m.playerPeer(1)
		}
	case database.PairGameType:
		switch m.game.MoveChoose {
		case database.SequentialChooseType:
			switch m.network.Me() {
			case m.playerPeer(1):
				m.turn = m.playerPeer(2)
			case m.playerPeer(2):
				m.turn = m.playerPeer(3)
			case m.playerPeer(3):
				m.turn = m.playerPeer(4)
			case m.playerPeer(4):
				m.turn = m.playerPeer(1)
			}
		case database.RandomChooseType:
			var players []int
			switch m.network.Me() {
			case m.playerPeer(1):
				players = []int{2, 4}
			case m.playerPeer(3):
				players = []int{2, 4}
			case m.playerPeer(2):
				players = []int{1, 3}
			case m.playerPeer(4):
				players = []int{1, 3}
			}
			m.turn = m.playerPeer(players[rand.Intn(len(players))])
		default:
			panic("should not be here")
		}
	}

	m.network.SendAll([]byte(string(multiplayer.DefineTurnMessage)), []byte(string(m.turn)))

	return m, tea.Batch(cmds...)
}

func (m GameModel) handleSaveTurnMsg(msg SaveTurnMsg) (GameModel, tea.Cmd) {
	cmds := []tea.Cmd{m.getMoves(), m.updateMovesListCmd()}

	m.turn = p2p.NetworkID(msg)

	return m, tea.Batch(cmds...)
}

func (m GameModel) handleChessMoveMsg(msg ChessMoveMsg) (GameModel, tea.Cmd) {
	m.err = m.chessGame.MoveStr(string(msg))
	cmds := []tea.Cmd{m.getMoves(), m.updateMovesListCmd()}

	if m.chessGame.Outcome() != chess.NoOutcome {
		cmds = append(cmds, m.endGame(m.chessGame.Outcome().String(), false))
	}

	return m, tea.Batch(cmds...)
}

func (m GameModel) sendNewTurnCmd() tea.Cmd {
	return func() tea.Msg {
		return SendNewTurnMsg{}
	}
}
