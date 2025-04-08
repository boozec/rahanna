package main

import (
	"log"

	"github.com/boozec/rahanna/pkg/ui/views"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	views.ClearScreen()

	p := tea.NewProgram(views.NewRahannaModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	views.ClearScreen()
}
