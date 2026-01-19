package tracking

import (
	"fmt"

	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/store"
	"github.com/Skryensya/footprint/internal/ui/style"
)

// FormatEvent formats a single event for display.
// If oneline is true, uses compact single-line format.
func FormatEvent(e store.RepoEvent, oneline bool) string {
	if oneline {
		// source(colored) commit(bold) repo(muted) branch
		return fmt.Sprintf(
			"%s %s %s %s",
			formatSource(e.Source),
			style.Header(fmt.Sprintf("%.7s", e.Commit)),
			style.Muted(e.RepoID),
			e.Branch,
		)
	}

	// Multi-line format
	return fmt.Sprintf(
		"%s %s %s %s\n%s\n",
		formatSource(e.Source),
		style.Header(fmt.Sprintf("%.7s", e.Commit)),
		e.Branch,
		style.Muted(e.RepoID),
		style.Muted(e.Timestamp.Format("Mon Jan 2 15:04:05 2006")),
	)
}

// formatSource applies distinct colors to each hook source.
func formatSource(source store.Source) string {
	switch source {
	case store.SourcePostCommit:
		return style.Color1(source.String()) // cyan
	case store.SourcePostMerge:
		return style.Color2(source.String()) // magenta
	case store.SourcePostCheckout:
		return style.Color3(source.String()) // blue
	case store.SourcePostRewrite:
		return style.Color4(source.String()) // yellow
	case store.SourcePrePush:
		return style.Color5(source.String()) // green
	case store.SourceManual:
		return style.Color6(source.String()) // red
	default:
		return source.String()
	}
}

// FormatEventEnriched formats a single event with git metadata (author, commit message).
// If oneline is true, uses compact single-line format with truncated subject.
func FormatEventEnriched(e store.RepoEvent, meta git.CommitMetadata, oneline bool) string {
	if oneline {
		// source commit repo branch "message"
		subject := meta.Subject
		if len(subject) > 40 {
			subject = subject[:37] + "..."
		}
		return fmt.Sprintf("%s %s %s %s %s",
			formatSource(e.Source),
			style.Header(fmt.Sprintf("%.7s", e.Commit)),
			style.Muted(e.RepoID),
			e.Branch,
			style.Muted(fmt.Sprintf("\"%s\"", subject)),
		)
	}

	// Multiline enriched
	return fmt.Sprintf("%s %s %s %s\n%s\n%s <%s>\n\n    %s\n",
		formatSource(e.Source),
		style.Header(fmt.Sprintf("%.7s", e.Commit)),
		e.Branch,
		style.Muted(e.RepoID),
		style.Muted(e.Timestamp.Format("Mon Jan 2 15:04:05 2006")),
		meta.AuthorName,
		meta.AuthorEmail,
		meta.Subject,
	)
}
