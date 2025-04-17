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
			items = append(items, item{title: move.String()})
		}
		m.movesList.SetItems(items)
		m.movesList.Title = "Choose a move"
		m.movesList.Select(0)
		m.movesList.SetShowFilter(true)
		m.movesList.SetFilteringEnabled(true)
		m.movesList.ResetFilter()
	}
	return m
}

func (m GameModel) handleChessMoveMsg(msg ChessMoveMsg) (GameModel, tea.Cmd) {
	m.turn++
	err := m.chessGame.MoveStr(string(msg))
	if err != nil {
		fmt.Println("Error applying move:", err)
	}
	return m, tea.Batch(m.getMoves(), m.updateMovesListCmd())
}
