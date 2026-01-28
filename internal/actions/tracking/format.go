package tracking

import (
	"fmt"

	"github.com/footprint-tools/cli/internal/format"
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/ui/style"
)

const (
	maxSubjectLengthOneline = 40
	truncatedSubjectLength  = 37
)

var sourceStylers = map[store.Source]func(string) string{
	store.SourcePostCommit:   style.Color1,
	store.SourcePostRewrite:  style.Color2,
	store.SourcePostCheckout: style.Color3,
	store.SourcePostMerge:    style.Color4,
	store.SourcePrePush:      style.Color5,
	store.SourceBackfill:     style.Color6,
	store.SourceManual:       style.Color7,
}

// FormatEvent formats a single event for display.
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
		style.Muted(format.Full(e.Timestamp)),
	)
}

func formatSource(source store.Source) string {
	if styler, ok := sourceStylers[source]; ok {
		return styler(source.String())
	}
	return source.String()
}

// FormatEventEnriched formats a single event with git metadata (author, commit message).
// If oneline is true, uses compact single-line format with truncated subject.
func FormatEventEnriched(e store.RepoEvent, meta git.CommitMetadata, oneline bool) string {
	if oneline {
		// source commit repo branch "message"
		subject := meta.Subject
		if len(subject) > maxSubjectLengthOneline {
			subject = subject[:truncatedSubjectLength] + "..."
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
		style.Muted(format.Full(e.Timestamp)),
		meta.AuthorName,
		meta.AuthorEmail,
		meta.Subject,
	)
}
