package views

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// TabType represents the available tabs

type TabType int

var (
	highlightColor   = lipgloss.Color("#7ee2a8")
	tabStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(highlightColor).Padding(0, 2)
	inactiveTabStyle = tabStyle
	activeTabStyle   = tabStyle
	altCodeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Bold(true)
	windowStyle      = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.RoundedBorder())
)

func getTabsRow(tabsText []string, activeTab TabType) string {
	tabs := make([]string, len(tabsText))

	for i, tab := range tabsText {
		if TabType(i) == activeTab {
			tabs[i] = fmt.Sprintf("%s %s", altCodeStyle.Render(fmt.Sprintf("Alt+%d", i+1)), lipgloss.NewStyle().Bold(true).Foreground(highlightColor).Render(tab))
			tabs[i] = activeTabStyle.Foreground(highlightColor).Render(tabs[i])
		} else {
			tabs[i] = fmt.Sprintf("%s %s", altCodeStyle.Render(fmt.Sprintf("Alt+%d", i+1)), lipgloss.NewStyle().Render(tab))
			tabs[i] = inactiveTabStyle.Foreground(highlightColor).Render(tabs[i])
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

}
