package main

import (
	"fmt"
	"os"

	"jrn/internal/parser"
	"jrn/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	doc, err := parser.ParseMarkdown("2026-08-25.md")
	if err != nil {
		fmt.Printf("Ошибка парсинга: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		ui.InitialModel(doc),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Ошибка при запуске TUI: %v\n", err)
		os.Exit(1)
	}
}
