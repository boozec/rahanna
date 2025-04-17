package views

import (
	"fmt"
	"strings"

	"github.com/boozec/rahanna/internal/api/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m PlayModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	return m, nil
}

func (m *PlayModel) handleError(msg error) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.err = msg
	return m, nil
}

func formatGamesForPage(games []database.Game, altCodeStyle lipgloss.Style) []string {
	var gamesStrings []string
	gamesStrings = append(gamesStrings, "Games list")

	longestName := 0
	for _, game := range games {
		if len(game.Name) > longestName {
			longestName = len(game.Name)
		}
	}

	for i, game := range games {
		indexStr := altCodeStyle.Render(fmt.Sprintf("[%d] ", i))
		nameStr := game.Name
		dateStr := game.UpdatedAt.Format("2006-01-02 15:04")

		padding := longestName - len(nameStr)
		paddingStr := strings.Repeat(" ", padding+4)

		line := lipgloss.JoinHorizontal(lipgloss.Left,
			indexStr,
			nameStr,
			paddingStr,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#d35400")).Render(dateStr),
		)
		gamesStrings = append(gamesStrings, line)
	}
	return gamesStrings
}

func (m PlayModel) buildWindowContent(content string, formWidth int) string {
	if m.err != nil {
		formError := fmt.Sprintf("Error: %v", m.err.Error())
		return lipgloss.JoinVertical(
			lipgloss.Center,
			windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
				lipgloss.Center,
				errorStyle.Align(lipgloss.Center).Width(formWidth-4).Render(formError),
				content,
			)),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		windowStyle.Width(formWidth).Render(lipgloss.JoinVertical(
			lipgloss.Center,
			content,
		)),
	)
}
