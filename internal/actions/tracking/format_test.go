package tracking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
)

func TestFormatEvent_Oneline(t *testing.T) {
	event := store.RepoEvent{
		ID:        1,
		RepoID:    "github.com/test/repo",
		Commit:    "abc1234567890",
		Branch:    "main",
		Source:    store.SourcePostCommit,
		Timestamp: time.Now(),
	}

	output := formatEvent(event, true)

	require.Contains(t, output, "abc1234")
	require.Contains(t, output, "main")
	require.Contains(t, output, "github.com/test/repo")
}

func TestFormatEvent_Multiline(t *testing.T) {
	event := store.RepoEvent{
		ID:        1,
		RepoID:    "github.com/test/repo",
		Commit:    "abc1234567890",
		Branch:    "main",
		Source:    store.SourcePostCommit,
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	output := formatEvent(event, false)

	require.Contains(t, output, "abc1234")
	require.Contains(t, output, "main")
	require.Contains(t, output, "github.com/test/repo")
	// Date format is configurable (default: "Jan 02"), check for date components
	require.Contains(t, output, "Jan")
	require.Contains(t, output, "15")
}

func TestFormatSource_AllSources(t *testing.T) {
	tests := []struct {
		source store.Source
		expect string
	}{
		{store.SourcePostCommit, "POST-COMMIT"},
		{store.SourcePostRewrite, "POST-REWRITE"},
		{store.SourcePostCheckout, "POST-CHECKOUT"},
		{store.SourcePostMerge, "POST-MERGE"},
		{store.SourcePrePush, "PRE-PUSH"},
		{store.SourceBackfill, "BACKFILL"},
		{store.SourceManual, "MANUAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			output := formatSource(tt.source)
			require.Contains(t, output, tt.expect)
		})
	}
}

func TestFormatSource_Unknown(t *testing.T) {
	output := formatSource(store.Source(99))
	require.NotEmpty(t, output)
}

func TestFormatEventEnriched_Oneline(t *testing.T) {
	event := store.RepoEvent{
		ID:        1,
		RepoID:    "github.com/test/repo",
		Commit:    "abc1234567890",
		Branch:    "main",
		Source:    store.SourcePostCommit,
		Timestamp: time.Now(),
	}

	meta := git.CommitMetadata{
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Subject:     "Fix a bug in the code",
	}

	output := formatEventEnriched(event, meta, true)

	require.Contains(t, output, "abc1234")
	require.Contains(t, output, "github.com/test/repo")
	require.Contains(t, output, "Fix a bug")
}

func TestFormatEventEnriched_Oneline_LongSubject(t *testing.T) {
	event := store.RepoEvent{
		ID:        1,
		RepoID:    "github.com/test/repo",
		Commit:    "abc1234567890",
		Branch:    "main",
		Source:    store.SourcePostCommit,
		Timestamp: time.Now(),
	}

	meta := git.CommitMetadata{
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Subject:     "This is a very long commit message that should be truncated for display purposes",
	}

	output := formatEventEnriched(event, meta, true)

	// Should be truncated with ...
	require.Contains(t, output, "...")
}

func TestFormatEventEnriched_Multiline(t *testing.T) {
	event := store.RepoEvent{
		ID:        1,
		RepoID:    "github.com/test/repo",
		Commit:    "abc1234567890",
		Branch:    "main",
		Source:    store.SourcePostCommit,
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	meta := git.CommitMetadata{
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Subject:     "Fix a bug",
	}

	output := formatEventEnriched(event, meta, false)

	require.Contains(t, output, "abc1234")
	require.Contains(t, output, "github.com/test/repo")
	require.Contains(t, output, "Test User")
	require.Contains(t, output, "test@example.com")
	require.Contains(t, output, "Fix a bug")
	// Date format is configurable (default: "Jan 02"), check for date components
	require.Contains(t, output, "Jan")
	require.Contains(t, output, "15")
}
