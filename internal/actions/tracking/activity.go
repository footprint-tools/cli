package tracking

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/git"
	"github.com/footprint-tools/footprint-cli/internal/store"
)

func Activity(args []string, flags *dispatchers.ParsedFlags) error {
	return activity(args, flags, DefaultDeps())
}

func activity(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.CloseDB(db)

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

	// Validate and parse limit flag
	if limitStr := flags.String("--limit", ""); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return fmt.Errorf("invalid limit value '%s': must be a positive integer", limitStr)
		}
		if limit <= 0 {
			return fmt.Errorf("invalid limit value %d: must be greater than 0", limit)
		}
		filter.Limit = limit
	}

	events, err := deps.ListEvents(db, filter)
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}
	if len(events) == 0 {
		_, _ = deps.Println("no events")
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
