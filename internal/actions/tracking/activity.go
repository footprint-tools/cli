package tracking

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/Skryensya/footprint/internal/telemetry"
)

func Activity(args []string, flags []string) error {
	return activity(args, flags, DefaultDeps())
}

func activity(_ []string, flags []string, deps Deps) error {
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return nil
	}

	var (
		filter  telemetry.EventFilter
		oneline bool
	)

	for _, flag := range flags {
		if flag == "--oneline" {
			oneline = true
			continue
		}

		if strings.HasPrefix(flag, "--status=") {
			if status, ok := parseStatus(strings.TrimPrefix(flag, "--status=")); ok {
				filter.Status = &status
			}
			continue
		}

		if strings.HasPrefix(flag, "--source=") {
			if source, ok := parseSource(strings.TrimPrefix(flag, "--source=")); ok {
				filter.Source = &source
			}
			continue
		}

		if strings.HasPrefix(flag, "--since=") {
			if t, ok := parseDate(strings.TrimPrefix(flag, "--since=")); ok {
				filter.Since = &t
			}
			continue
		}

		if strings.HasPrefix(flag, "--until=") {
			if t, ok := parseDate(strings.TrimPrefix(flag, "--until=")); ok {
				filter.Until = &t
			}
			continue
		}

		if strings.HasPrefix(flag, "--repo=") {
			repoID := strings.TrimPrefix(flag, "--repo=")
			filter.RepoID = &repoID
			continue
		}

		if strings.HasPrefix(flag, "--limit=") {
			if n, err := strconv.Atoi(strings.TrimPrefix(flag, "--limit=")); err == nil && n > 0 {
				filter.Limit = n
			}
			continue
		}
	}

	events, err := deps.ListEvents(db, filter)
	if err != nil || len(events) == 0 {
		return nil
	}

	var output bytes.Buffer

	for _, event := range events {
		if oneline {
			message := truncateMessage(event.CommitMessage, 15)

			output.WriteString(fmt.Sprintf(
				"%s %-9s %-13s %-20s %-8s %.7s %s\n",
				event.Timestamp.Format("2006-01-02 15:04"),
				event.Status.String(),
				event.Source.String(),
				event.RepoID,
				event.Branch,
				event.Commit,
				message,
			))
			continue
		}

		output.WriteString(fmt.Sprintf(
			"%s %-9s %-13s %-20s %-8s %.7s\n    %s\n",
			event.Timestamp.Format("2006-01-02 15:04"),
			event.Status.String(),
			event.Source.String(),
			event.RepoID,
			event.Branch,
			event.Commit,
			event.CommitMessage,
		))
	}

	deps.Pager(output.String())
	return nil
}
