package tracking

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
)

const pollInterval = 300 * time.Millisecond

func Log(args []string, flags *dispatchers.ParsedFlags) error {
	// Route to interactive mode if --interactive or -i flag is present
	if flags.Has("--interactive") || flags.Has("-i") {
		return WatchInteractive(args, flags)
	}
	return logCmd(args, flags, DefaultDeps())
}

func logCmd(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.CloseDB(db)

	// Ensure database is initialized
	if err := deps.InitDB(db); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Parse display flags
	oneline := flags.Has("--oneline")
	jsonOutput := flags.Has("--json")
	enrich := flags.Has("--enrich")

	// Parse filter flags
	var filter store.EventFilter

	if statusStr := flags.String("--status", ""); statusStr != "" {
		if status, ok := parseStatus(statusStr); ok {
			filter.Status = &status
		}
	}

	if sourceStr := flags.String("--source", ""); sourceStr != "" {
		if source, ok := parseSource(sourceStr); ok {
			filter.Source = &source
		}
	}

	if repoID := flags.String("--repo", ""); repoID != "" {
		filter.RepoID = &repoID
	}

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
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
			// Context cancelled, goroutine can exit
		}
	}()

	// Polling loop
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			events, err := store.ListEventsSinceFiltered(db, lastID, filter)
			if err != nil {
				continue
			}

			for _, event := range events {
				if jsonOutput {
					outputEventJSON(event, enrich)
				} else if enrich {
					meta := git.GetCommitMetadata(event.RepoPath, event.Commit)
					_, _ = fmt.Fprintln(os.Stdout, FormatEventEnriched(event, meta, oneline))
				} else {
					_, _ = fmt.Fprintln(os.Stdout, FormatEvent(event, oneline))
				}
				// Note: int64 overflow is not a practical concern (max ~9 quintillion).
				// A negative ID would indicate database corruption.
				if event.ID > 0 {
					lastID = event.ID
				}
			}
		}
	}
}

func outputEventJSON(e store.RepoEvent, enrich bool) {
	type jsonEvent struct {
		ID        int64  `json:"id"`
		RepoID    string `json:"repo_id"`
		RepoPath  string `json:"repo_path"`
		Commit    string `json:"commit"`
		Branch    string `json:"branch"`
		Timestamp string `json:"timestamp"`
		Status    string `json:"status"`
		Source    string `json:"source"`
		Author    string `json:"author,omitempty"`
		Message   string `json:"message,omitempty"`
	}

	je := jsonEvent{
		ID:        e.ID,
		RepoID:    e.RepoID,
		RepoPath:  e.RepoPath,
		Commit:    e.Commit,
		Branch:    e.Branch,
		Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Status:    e.Status.String(),
		Source:    e.Source.String(),
	}
	if enrich {
		meta := git.GetCommitMetadata(e.RepoPath, e.Commit)
		je.Author = meta.AuthorName
		je.Message = meta.Subject
	}

	data, _ := json.Marshal(je)
	_, _ = fmt.Fprintln(os.Stdout, string(data))
}
