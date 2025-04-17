package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Create and configure the name input prompt
func createNamePrompt(width int) textinput.Model {
	namePrompt := textinput.New()
	namePrompt.Prompt = " "
	namePrompt.TextStyle = inputStyle
	namePrompt.Placeholder = "rectangular-lake"
	namePrompt.Focus()
	namePrompt.CharLimit = 23
	namePrompt.Width = getFormWidth(width)

	return namePrompt
}

func (m PlayModel) renderPageContent(base lipgloss.Style) string {
	switch m.page {
	case LandingPage:
		m.namePrompt.Blur()
		if len(m.games) == 0 {
			return base.Render(chessBoard)
		} else {
			start, end := m.paginator.GetSliceBounds(len(m.games))
			gamesStrings := formatGamesForPage(m.games[start:end], altCodeStyle)
			pageInfo := m.paginator.View()
			return base.Render(lipgloss.JoinVertical(lipgloss.Center, strings.Join(gamesStrings, "\n"), pageInfo))
		}
	case InsertCodePage:
		return m.renderInsertCodeContent(base)

	case StartGamePage:
		return m.renderStartGameContent(base)
	}

	return ""
}

func (m PlayModel) renderInsertCodeContent(base lipgloss.Style) string {
	// When loading, show loading status
	if m.isLoading {
		return base.Render(
			lipgloss.NewStyle().
				Align(lipgloss.Center).
				Bold(true).
				Render("Loading..."),
		)
	}

	// Default: show input prompt
	return base.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Render("Insert play code:"),
			m.namePrompt.View(),
			lipgloss.NewStyle().
				Align(lipgloss.Center).
				PaddingTop(2).
				Bold(true).
				Render(fmt.Sprintf("Press %s to join",
					lipgloss.NewStyle().Italic(true).Render("Enter"))),
		),
	)
}

func (m PlayModel) renderStartGameContent(base lipgloss.Style) string {
	var statusMsg string

	if m.isLoading {
		statusMsg = "Loading..."
	} else if m.playName != "" {
		gameCode := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#F39C12")).
			Render(m.playName)

		statusMsg = fmt.Sprintf("Share `%s` to your friend", gameCode)
	}

	return base.Render(statusMsg)
}
