package views

import (
	"fmt"

	"github.com/boozec/rahanna/internal/network"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// UpdateMovesListMsg is a message to update the moves list
type UpdateMovesListMsg struct{}

// ChessMoveMsg is a message containing a received chess move.
type ChessMoveMsg string

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

func (m *GameModel) getMoves() tea.Cmd {
	m.network.Server.OnReceiveFn = func(msg network.Message) {
		moveStr := string(msg.Payload)
		m.incomingMoves <- moveStr
	}

	return func() tea.Msg {
		move := <-m.incomingMoves
		return ChessMoveMsg(move)
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
			items = append(
				items,
				item{title: fmt.Sprintf("%s → %s", move.S1().String(), move.S2().String())},
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

func (m GameModel) handleChessMoveMsg(msg ChessMoveMsg) (GameModel, tea.Cmd) {
	m.turn++
	err := m.chessGame.MoveStr(string(msg))
	if err != nil {
		m.err = err
	} else {
		m.err = nil
	}
	return m, tea.Batch(m.getMoves(), m.updateMovesListCmd())
}
