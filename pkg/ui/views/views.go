package views

import (
	"errors"
	"os"

	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var logo = `
▗▄▄▖  ▗▄▖ ▗▖ ▗▖ ▗▄▖ ▗▖  ▗▖▗▖  ▗▖ ▗▄▖ 
▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▛▚▖▐▌▐▛▚▖▐▌▐▌ ▐▌
▐▛▀▚▖▐▛▀▜▌▐▛▀▜▌▐▛▀▜▌▐▌ ▝▜▌▐▌ ▝▜▌▐▛▀▜▌
▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌  ▐▌▐▌  ▐▌▐▌ ▐▌
`

var (
	highlightColor = lipgloss.Color("#7ee2a8")
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	altCodeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Bold(true)
	windowStyle    = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.RoundedBorder())
	inputStyle     = lipgloss.NewStyle().Foreground(highlightColor)
)

// Get terminal size dynamically
func GetTerminalSize() (width, height int) {
	fd := int(os.Stdin.Fd())
	if w, h, err := term.GetSize(fd); err == nil {
		return w, h
	}
	return 80, 24 // Default size if detection fails
}

// Clear terminal screen
func ClearScreen() {
	if len(os.Getenv("DEBUG")) == 0 {
		cmd := exec.Command("clear")
		if os.Getenv("OS") == "Windows_NT" {
			cmd = exec.Command("cmd", "/c", "cls")
		}
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func getFormWidth(width int) int {
	formWidth := width * 2 / 2
	if formWidth > 160 {
		formWidth = 160
	} else if formWidth < 40 {
		formWidth = width - 4
	}

	return formWidth
}

type RahannaModel struct {
	width        int
	height       int
	currentModel tea.Model
	auth         AuthModel
	play         PlayModel
}

func NewRahannaModel() RahannaModel {
	width, height := GetTerminalSize()

	auth := NewAuthModel(width, height)
	play := NewPlayModel(width, height)

	var currentModel tea.Model = auth

	if _, err := os.Stat(".rahannarc"); !errors.Is(err, os.ErrNotExist) {
		currentModel = play
	}

	return RahannaModel{
		width:        width,
		height:       height,
		currentModel: currentModel,
		auth:         auth,
		play:         play,
	}
}

func (m RahannaModel) Init() tea.Cmd {
	return m.currentModel.Init()
}

type switchModel struct {
	model tea.Model
}

func SwitchModelCmd(model tea.Model) tea.Cmd {
	s := switchModel{
		model: model,
	}

	return func() tea.Msg {
		return s
	}
}

func (m RahannaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchModel:
		m.currentModel = msg.model
		return m, nil
	}
	var cmd tea.Cmd
	m.currentModel, cmd = m.currentModel.Update(msg)
	return m, cmd
}

func (m RahannaModel) View() string {
	return m.currentModel.View()
}

func handleExit(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return tea.Quit
		}
	}

	return nil
}

func getLogo(width int) string {
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7ee2a8")).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)

	return logoStyle.Render(logo)

}

func logout(width, height int) tea.Cmd {
	if err := os.Remove(".rahannarc"); err != nil {
		return nil
	}
	return SwitchModelCmd(NewAuthModel(width, height+1))
}
