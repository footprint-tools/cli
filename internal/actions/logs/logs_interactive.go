package logs

import (
	"errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/paths"
	"golang.org/x/term"
)

// Interactive runs the interactive logs TUI
func Interactive(args []string, flags *dispatchers.ParsedFlags) error {
	return interactive(args, flags, DefaultDeps())
}

func interactive(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	// Check for terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("interactive logs requires an interactive terminal")
	}

	logPath := deps.LogFilePath()
	if logPath == "" {
		logPath = paths.LogFilePath()
	}

	// Create model
	m := newLogsModel(logPath)

	// Run program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
