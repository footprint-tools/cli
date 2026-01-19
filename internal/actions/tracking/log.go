package tracking

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/store"
)

const pollInterval = 300 * time.Millisecond

func Log(args []string, flags *dispatchers.ParsedFlags) error {
	return logCmd(args, flags, DefaultDeps())
}

func logCmd(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Ensure database is initialized
	if err := deps.InitDB(db); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Parse flags
	oneline := flags.Has("--oneline")

	// Get current max ID as starting point (we only want new events)
	lastID, err := store.GetMaxEventID(db)
	if err != nil {
		// If we can't get max ID, start from 0
		lastID = 0
	}

	fmt.Fprintln(os.Stderr, "Watching for new events... (Ctrl+C to stop)")

	// Setup signal handling for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	// Polling loop
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			events, err := store.ListEventsSince(db, lastID)
			if err != nil {
				continue
			}

			for _, event := range events {
				fmt.Fprintln(os.Stdout, FormatEvent(event, oneline))
				lastID = event.ID
			}
		}
	}
}
