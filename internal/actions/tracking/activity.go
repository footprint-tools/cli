package tracking

import (
	"bytes"
	"fmt"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/store"
)

func Activity(args []string, flags *dispatchers.ParsedFlags) error {
	return activity(args, flags, DefaultDeps())
}

func activity(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var filter store.EventFilter

	oneline := flags.Has("--oneline")
	enrich := flags.Has("--enrich")

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

	if since := flags.Date("--since"); since != nil {
		filter.Since = since
	}

	if until := flags.Date("--until"); until != nil {
		filter.Until = until
	}

	if repoID := flags.String("--repo", ""); repoID != "" {
		filter.RepoID = &repoID
	}

	if limit := flags.Int("--limit", 0); limit > 0 {
		filter.Limit = limit
	}

	events, err := deps.ListEvents(db, filter)
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}
	if len(events) == 0 {
		return nil
	}

	var output bytes.Buffer

	for _, event := range events {
		if enrich {
			meta := git.GetCommitMetadata(event.RepoPath, event.Commit)
			output.WriteString(FormatEventEnriched(event, meta, oneline))
		} else {
			output.WriteString(FormatEvent(event, oneline))
		}
		output.WriteString("\n")
	}

	deps.Pager(output.String())
	return nil
}
