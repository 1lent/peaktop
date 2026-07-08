package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/brodie/peaktop/internal/ui"
)

func main() {
	interval := flag.Int("i", 500, "update interval in milliseconds")
	flag.Parse()

	if *interval < 100 {
		*interval = 100
	}
	if *interval > 5000 {
		*interval = 5000
	}

	model := ui.NewModel(time.Duration(*interval) * time.Millisecond)
	program := tea.NewProgram(&model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "peaktop: %v\n", err)
		os.Exit(1)
	}
}
