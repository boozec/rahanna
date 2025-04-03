package main

import (
	"log"

	"github.com/boozec/rahanna/ui/views"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	views.ClearScreen()
	p := tea.NewProgram(views.LoginModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	views.ClearScreen()
}
