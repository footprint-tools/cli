package tracking

import (
	"fmt"

	"github.com/Skryensya/footprint/internal/store"
	"github.com/Skryensya/footprint/internal/ui/style"
)

// FormatEvent formats a single event for display.
// If oneline is true, uses compact single-line format.
func FormatEvent(e store.RepoEvent, oneline bool) string {
	if oneline {
		// source(colored) commit(bold) repo(muted) branch message(plain)
		return fmt.Sprintf(
			"%s %s %s %s %s",
			formatSource(e.Source),
			style.Header(fmt.Sprintf("%.7s", e.Commit)),
			style.Muted(e.RepoID),
			e.Branch,
			truncateMessage(e.CommitMessage, 40),
		)
	}

	// Multi-line format
	header := fmt.Sprintf("%s %s %s %s",
		formatSource(e.Source),
		style.Header(fmt.Sprintf("%.7s", e.Commit)),
		e.Branch,
		style.Muted(e.RepoID),
	)

	author := ""
	if e.Author != "" {
		author = fmt.Sprintf("\n%s", e.Author)
	}

	return fmt.Sprintf(
		"%s%s\n%s\n\n    %s\n",
		header,
		author,
		style.Muted(e.Timestamp.Format("Mon Jan 2 15:04:05 2006")),
		e.CommitMessage,
	)
}

// formatSource applies distinct colors to each hook source.
func formatSource(s store.Source) string {
	switch s {
	case store.SourcePostCommit:
		return style.Color1(s.String()) // cyan
	case store.SourcePostMerge:
		return style.Color2(s.String()) // magenta
	case store.SourcePostCheckout:
		return style.Color3(s.String()) // blue
	case store.SourcePostRewrite:
		return style.Color4(s.String()) // yellow
	case store.SourcePrePush:
		return style.Color5(s.String()) // green
	case store.SourceManual:
		return style.Color6(s.String()) // red
	default:
		return s.String()
	}
}
