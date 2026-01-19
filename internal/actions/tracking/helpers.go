package tracking

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Skryensya/footprint/internal/store"
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

func truncateMessage(message string, max int) string {
	message = strings.TrimSpace(message)
	if len(message) <= max {
		return message
	}
	return message[:max] + "..."
}

func parseStatus(s string) (store.Status, bool) {
	switch strings.ToLower(s) {
	case "pending":
		return store.StatusPending, true
	case "exported":
		return store.StatusExported, true
	case "orphaned":
		return store.StatusOrphaned, true
	case "skipped":
		return store.StatusSkipped, true
	default:
		return 0, false
	}
}

func parseSource(s string) (store.Source, bool) {
	switch strings.ToLower(s) {
	case "post-commit":
		return store.SourcePostCommit, true
	case "post-rewrite":
		return store.SourcePostRewrite, true
	case "post-checkout":
		return store.SourcePostCheckout, true
	case "post-merge":
		return store.SourcePostMerge, true
	case "pre-push":
		return store.SourcePrePush, true
	case "manual":
		return store.SourceManual, true
	case "backfill":
		return store.SourceBackfill, true
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
