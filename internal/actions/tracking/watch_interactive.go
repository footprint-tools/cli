package tracking

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/store"
	"golang.org/x/term"
)

// WatchInteractive runs the interactive watch TUI
func WatchInteractive(args []string, flags *dispatchers.ParsedFlags) error {
	return watchInteractive(args, flags, DefaultDeps())
}

func watchInteractive(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	// Check for terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("interactive watch requires an interactive terminal")
	}

	// Open database
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Ensure database is initialized
	if err := deps.InitDB(db); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Get current max ID as starting point (we only want new events)
	lastID, err := store.GetMaxEventID(db)
	if err != nil {
		lastID = 0
	}

	// Create model
	m := newWatchModel(db, lastID)

	// Run program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = p.Run()
	return err
}
