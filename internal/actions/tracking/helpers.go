package tracking

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/footprint-tools/cli/internal/store"
)

var statusMap = map[string]store.Status{
	"pending":  store.StatusPending,
	"exported": store.StatusExported,
	"orphaned": store.StatusOrphaned,
	"skipped":  store.StatusSkipped,
}

var sourceMap = map[string]store.Source{
	"post-commit":   store.SourcePostCommit,
	"post-rewrite":  store.SourcePostRewrite,
	"post-checkout": store.SourcePostCheckout,
	"post-merge":    store.SourcePostMerge,
	"pre-push":      store.SourcePrePush,
	"manual":        store.SourceManual,
	"backfill":      store.SourceBackfill,
}

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

func parseStatus(s string) (store.Status, bool) {
	status, ok := statusMap[strings.ToLower(s)]
	return status, ok
}

func parseSource(s string) (store.Source, bool) {
	source, ok := sourceMap[strings.ToLower(s)]
	return source, ok
}

// ValidStatuses returns a list of valid status values for use in error messages.
func ValidStatuses() []string {
	keys := make([]string, 0, len(statusMap))
	for k := range statusMap {
		keys = append(keys, k)
	}
	return keys
}

// ValidSources returns a list of valid source values for use in error messages.
func ValidSources() []string {
	keys := make([]string, 0, len(sourceMap))
	for k := range sourceMap {
		keys = append(keys, k)
	}
	return keys
}
