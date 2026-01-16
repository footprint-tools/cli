package tracking

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Skryensya/footprint/internal/telemetry"
)

func resolvePath(args []string) (string, error) {
	p := "."
	if len(args) > 0 {
		p = args[0]
	}

	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", os.ErrInvalid
	}

	return abs, nil
}

func hasFlag(flags []string, name string) bool {
	for _, f := range flags {
		if f == name {
			return true
		}
	}
	return false
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

func parseDate(s string) (time.Time, bool) {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}
