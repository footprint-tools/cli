package tracking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
)

func Activity(args []string, flags *dispatchers.ParsedFlags) error {
	return activity(args, flags, DefaultDeps())
}

func activity(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	dbPath := deps.DBPath()
	db, err := deps.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w\nHint: Run 'fp setup' to initialize tracking in this repository", dbPath, err)
	}
	defer store.CloseDB(db)

	var filter store.EventFilter

	oneline := flags.Has("--oneline")
	jsonOutput := flags.Has("--json")
	enrich := flags.Has("--enrich")

	if statusStr := flags.String("--status", ""); statusStr != "" {
		status, ok := parseStatus(statusStr)
		if !ok {
			return fmt.Errorf("invalid status '%s': valid values are %s", statusStr, strings.Join(ValidStatuses(), ", "))
		}
		filter.Status = &status
	}

	if sourceStr := flags.String("--source", ""); sourceStr != "" {
		source, ok := parseSource(sourceStr)
		if !ok {
			return fmt.Errorf("invalid source '%s': valid values are %s", sourceStr, strings.Join(ValidSources(), ", "))
		}
		filter.Source = &source
	}

	if sinceStr := flags.String("--since", ""); sinceStr != "" {
		since := flags.Date("--since")
		if since == nil {
			return fmt.Errorf("invalid date '%s' for --since: expected format YYYY-MM-DD", sinceStr)
		}
		filter.Since = since
	}

	if untilStr := flags.String("--until", ""); untilStr != "" {
		until := flags.Date("--until")
		if until == nil {
			return fmt.Errorf("invalid date '%s' for --until: expected format YYYY-MM-DD", untilStr)
		}
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
		if jsonOutput {
			_, _ = deps.Println("[]")
		} else {
			_, _ = deps.Println("no events")
		}
		return nil
	}

	if jsonOutput {
		return outputEventsJSON(events, enrich, deps)
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

func outputEventsJSON(events []store.RepoEvent, enrich bool, deps Deps) error {
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

	out := make([]jsonEvent, 0, len(events))
	for _, e := range events {
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
		out = append(out, je)
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	_, _ = deps.Println(string(data))
	return nil
}
