package views

import (
	"os"
	"os/exec"

	"golang.org/x/term"
)

var logo = `
▗▄▄▖  ▗▄▖ ▗▖ ▗▖ ▗▄▖ ▗▖  ▗▖▗▖  ▗▖ ▗▄▖ 
▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▛▚▖▐▌▐▛▚▖▐▌▐▌ ▐▌
▐▛▀▚▖▐▛▀▜▌▐▛▀▜▌▐▛▀▜▌▐▌ ▝▜▌▐▌ ▝▜▌▐▛▀▜▌
▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▐▌  ▐▌▐▌  ▐▌▐▌ ▐▌
`

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
	cmd := exec.Command("clear") // Unix (Linux/macOS)
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.Command("cmd", "/c", "cls") // Windows
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func getFormWidth(width int) int {
	formWidth := width * 2 / 3
	if formWidth > 80 {
		formWidth = 80 // Cap at 80 chars for readability
	} else if formWidth < 40 {
		formWidth = width - 4 // For small terminals
	}

	return formWidth
}
