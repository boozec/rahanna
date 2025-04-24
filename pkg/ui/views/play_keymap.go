package views

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

type PlayModelPage int

const (
	LandingPage PlayModelPage = iota
	InsertCodePage
	StartGamePage
)

// Keyboard controls
type playKeyMap struct {
	EnterNewGame       key.Binding
	StartNewSingleGame key.Binding
	StartNewPairGame   key.Binding
	RestoreGame        key.Binding
	GoLogout           key.Binding
	NextPage           key.Binding
	PrevPage           key.Binding
	Exit               key.Binding
}

// Default key bindings for the play model
var defaultPlayKeyMap = playKeyMap{
	EnterNewGame: key.NewBinding(
		key.WithKeys("alt+E", "alt+e"),
		key.WithHelp("Alt+E", "Enter a play using code"),
	),
	StartNewSingleGame: key.NewBinding(
		key.WithKeys("alt+s", "alt+S"),
		key.WithHelp("Alt+S", "Start a new single play"),
	),
	StartNewPairGame: key.NewBinding(
		key.WithKeys("alt+p", "alt+P"),
		key.WithHelp("Alt+P", "Start a new pair play"),
	),
	RestoreGame: key.NewBinding(
		key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("[0-9]", "Restore a game"),
	),
	GoLogout: key.NewBinding(
		key.WithKeys("alt+Q", "alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→/h", "Next Page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←/l", "Prev Page"),
	),
	Exit: key.NewBinding(
		key.WithKeys("ctrl+c", "ctrl+C"),
		key.WithHelp("Ctrl+C", "Exit"),
	),
}

func (m PlayModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keys.EnterNewGame):
		if m.page == LandingPage {
			m.page = InsertCodePage
			return m, cmd
		}

	case key.Matches(msg, m.keys.StartNewSingleGame):
		if m.page == LandingPage {
			m.page = StartGamePage
			if !m.isLoading {
				m.isLoading = true
				return m, m.newGameCallback(database.SingleGameType)
			}

			return m, cmd
		}
	case key.Matches(msg, m.keys.StartNewPairGame):
		if m.page == LandingPage {
			m.page = StartGamePage
			if !m.isLoading {
				m.isLoading = true
				return m, m.newGameCallback(database.PairGameType)
			}

			return m, cmd
		}

	case key.Matches(msg, m.keys.RestoreGame):
		idx, err := strconv.Atoi(msg.String())
		m.err = err
		if err == nil {
			gameIndex := m.paginator.Page*m.paginator.PerPage + idx
			if gameIndex < len(m.games) {
				m.gameToRestore = &m.games[gameIndex]
				if m.gameToRestore.Outcome != chess.NoOutcome.String() {
					m.err = errors.New("this game is closed")
				} else {
					m.err = nil
					m.namePrompt.SetValue(m.gameToRestore.Name)
					return m, m.enterGame()
				}
			}
		}

	case key.Matches(msg, m.keys.GoLogout):
		return m, logout(m.width, m.height+1)

	case msg.Type == tea.KeyEnter:
		if m.page == InsertCodePage && !m.isLoading {
			m.isLoading = true
			return m, m.enterGame()
		}
	}

	m.paginator, _ = m.paginator.Update(msg)

	if m.page == InsertCodePage {
		m.namePrompt, cmd = m.namePrompt.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m PlayModel) renderNavigationButtons() string {
	logoutKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.GoLogout.Help().Key),
		m.keys.GoLogout.Help().Desc)

	exitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Exit.Help().Key),
		m.keys.Exit.Help().Desc)

	if m.page == LandingPage {
		enterKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.EnterNewGame.Help().Key),
			m.keys.EnterNewGame.Help().Desc)

		restoreKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.RestoreGame.Help().Key),
			m.keys.RestoreGame.Help().Desc)

		startSingleKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.StartNewSingleGame.Help().Key),
			m.keys.StartNewSingleGame.Help().Desc)

		startPairKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.StartNewPairGame.Help().Key),
			m.keys.StartNewPairGame.Help().Desc)

		nextPageKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.NextPage.Help().Key),
			m.keys.NextPage.Help().Desc)

		prevPageKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.PrevPage.Help().Key),
			m.keys.PrevPage.Help().Desc)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			enterKey,
			startSingleKey,
			startPairKey,
			restoreKey,
			lipgloss.JoinHorizontal(lipgloss.Left, prevPageKey, " | ", nextPageKey),
			logoutKey,
			exitKey,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		logoutKey,
		exitKey,
	)
}
