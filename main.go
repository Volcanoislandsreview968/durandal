package main

import (
	"fmt"
	"os"

	"github.com/blumenwagen/durandal/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := app.NewModel()

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Durandal: %v\n", err)
		os.Exit(1)
	}
}
