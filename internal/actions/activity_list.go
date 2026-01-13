package actions

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Skryensya/footprint/internal/telemetry"
	"github.com/Skryensya/footprint/internal/ui"
)

func ActivityList(args []string, flags []string) error {
	db, err := telemetry.Open(telemetry.DBPath())
	if err != nil {
		return nil
	}

	var (
		filterStatus *telemetry.Status
		filterSource *telemetry.Source
		oneline      bool
	)

	for _, flag := range flags {
		if flag == "--oneline" {
			oneline = true
			continue
		}

		if strings.HasPrefix(flag, "--status=") {
			if status, ok := parseStatus(strings.TrimPrefix(flag, "--status=")); ok {
				filterStatus = &status
			}
			continue
		}

		if strings.HasPrefix(flag, "--source=") {
			if source, ok := parseSource(strings.TrimPrefix(flag, "--source=")); ok {
				filterSource = &source
			}
			continue
		}
	}

	events, err := telemetry.ListEvents(db, filterStatus, filterSource)
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

	ui.Pager(output.String())
	return nil
}

func truncateMessage(message string, max int) string {
	message = strings.TrimSpace(message)
	if len(message) <= max {
		return message
	}
	return message[:max] + "..."
}

func parseStatus(s string) (telemetry.Status, bool) {
	switch strings.ToLower(s) {
	case "pending":
		return telemetry.StatusPending, true
	case "exported":
		return telemetry.StatusExported, true
	case "orphaned":
		return telemetry.StatusOrphaned, true
	case "skipped":
		return telemetry.StatusSkipped, true
	default:
		return 0, false
	}
}

func parseSource(s string) (telemetry.Source, bool) {
	switch strings.ToLower(s) {
	case "post-commit":
		return telemetry.SourcePostCommit, true
	case "post-rewrite":
		return telemetry.SourcePostRewrite, true
	case "post-checkout":
		return telemetry.SourcePostCheckout, true
	case "post-merge":
		return telemetry.SourcePostMerge, true
	case "pre-push":
		return telemetry.SourcePrePush, true
	case "manual":
		return telemetry.SourceManual, true
	default:
		return 0, false
	}
}
