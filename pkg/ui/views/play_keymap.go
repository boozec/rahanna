package views

import (
	"errors"
	"fmt"
	"strconv"

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
	EnterNewGame key.Binding
	StartNewGame key.Binding
	RestoreGame  key.Binding
	GoLogout     key.Binding
	Quit         key.Binding
	NextPage     key.Binding
	PrevPage     key.Binding
}

// Default key bindings for the play model
var defaultPlayKeyMap = playKeyMap{
	EnterNewGame: key.NewBinding(
		key.WithKeys("alt+E", "alt+e"),
		key.WithHelp("Alt+E", "Enter a play using code"),
	),
	StartNewGame: key.NewBinding(
		key.WithKeys("alt+s", "alt+S"),
		key.WithHelp("Alt+S", "Start a new play"),
	),
	RestoreGame: key.NewBinding(
		key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("[0-9]", "Restore a game"),
	),
	GoLogout: key.NewBinding(
		key.WithKeys("alt+Q", "alt+q"),
		key.WithHelp("Alt+Q", "Logout"),
	),
	Quit: key.NewBinding(
		key.WithKeys("Q", "q"),
		key.WithHelp("    Q", "Quit"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→/h", "Next Page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←/l", "Prev Page"),
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

	case key.Matches(msg, m.keys.StartNewGame):
		if m.page == LandingPage {
			m.page = StartGamePage
			if !m.isLoading {
				m.isLoading = true
				return m, m.newGameCallback()
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

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

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

	quitKey := fmt.Sprintf("%s %s",
		altCodeStyle.Render(m.keys.Quit.Help().Key),
		m.keys.Quit.Help().Desc)

	if m.page == LandingPage {
		enterKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.EnterNewGame.Help().Key),
			m.keys.EnterNewGame.Help().Desc)

		restoreKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.RestoreGame.Help().Key),
			m.keys.RestoreGame.Help().Desc)

		startKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.StartNewGame.Help().Key),
			m.keys.StartNewGame.Help().Desc)

		nextPageKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.NextPage.Help().Key),
			m.keys.NextPage.Help().Desc)

		prevPageKey := fmt.Sprintf("%s %s",
			altCodeStyle.Render(m.keys.PrevPage.Help().Key),
			m.keys.PrevPage.Help().Desc)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			enterKey,
			startKey,
			restoreKey,
			lipgloss.JoinHorizontal(lipgloss.Left, prevPageKey, " | ", nextPageKey),
			logoutKey,
			quitKey,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		logoutKey,
		quitKey,
	)
}
