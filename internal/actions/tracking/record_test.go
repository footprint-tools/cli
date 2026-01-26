package tracking

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/repo"
	"github.com/footprint-tools/footprint-cli/internal/store"
)

func TestRecord_SuccessFromHook(t *testing.T) {
	var insertedEvent store.RepoEvent
	fixedNow := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	deps := Deps{
		Getenv: func(key string) string {
			if key == "FP_SOURCE" {
				return "post-commit"
			}
			return ""
		},
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/user/repo", nil
		},
		HeadCommit: func() (string, error) {
			return "abc123def456", nil
		},
		CurrentBranch: func() (string, error) {
			return "main", nil
		},
		DBPath: func() string {
			return ":memory:"
		},
		OpenDB: func(path string) (*sql.DB, error) {
			db, _ := sql.Open("sqlite3", ":memory:")
			return db, nil
		},
		InitDB: func(db *sql.DB) error {
			return nil
		},
		InsertEvent: func(db *sql.DB, event store.RepoEvent) error {
			insertedEvent = event
			return nil
		},
		Now: func() time.Time {
			return fixedNow
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := record([]string{}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, "github.com/user/repo", insertedEvent.RepoID)
	require.Equal(t, "abc123def456", insertedEvent.Commit)
	require.Equal(t, "main", insertedEvent.Branch)
	require.Equal(t, store.StatusPending, insertedEvent.Status)
	require.Equal(t, store.SourcePostCommit, insertedEvent.Source)
	require.Equal(t, fixedNow.UTC(), insertedEvent.Timestamp)
}

func TestRecord_SuccessWithManualFlag(t *testing.T) {
	var capturedPrintf string
	fixedNow := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	deps := Deps{
		Getenv: func(key string) string {
			return "" // No FP_SOURCE
		},
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/user/repo", nil
		},
		HeadCommit: func() (string, error) {
			return "abc123", nil
		},
		CurrentBranch: func() (string, error) {
			return "feature", nil
		},
		DBPath: func() string {
			return ":memory:"
		},
		OpenDB: func(path string) (*sql.DB, error) {
			db, _ := sql.Open("sqlite3", ":memory:")
			return db, nil
		},
		InitDB: func(db *sql.DB) error {
			return nil
		},
		InsertEvent: func(db *sql.DB, event store.RepoEvent) error {
			return nil
		},
		Now: func() time.Time {
			return fixedNow
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--manual"})
	err := record([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "recorded")
}

func TestRecord_SuccessWithVerboseFlag(t *testing.T) {
	var capturedPrintf string

	deps := Deps{
		Getenv: func(key string) string {
			return "post-merge"
		},
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/user/repo", nil
		},
		HeadCommit: func() (string, error) {
			return "def456", nil
		},
		CurrentBranch: func() (string, error) {
			return "main", nil
		},
		DBPath: func() string {
			return ":memory:"
		},
		OpenDB: func(path string) (*sql.DB, error) {
			db, _ := sql.Open("sqlite3", ":memory:")
			return db, nil
		},
		InitDB: func(db *sql.DB) error {
			return nil
		},
		InsertEvent: func(db *sql.DB, event store.RepoEvent) error {
			return nil
		},
		Now: func() time.Time {
			return time.Now()
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--verbose"})
	err := record([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "recorded")
}

func TestRecord_GitNotAvailable(t *testing.T) {
	var errorShown bool

	deps := Deps{
		Getenv: func(key string) string {
			return "post-commit" // From hook to avoid the note message
		},
		GitIsAvailable: func() bool { return false },
		Println: func(a ...any) (int, error) {
			// Check if it's the error message about git not being available
			if len(a) > 0 {
				if str, ok := a[0].(string); ok && str == "git not available" {
					errorShown = true
				}
			}
			return 0, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	// Without --verbose or --manual, should not show error (hook mode)
	flags := dispatchers.NewParsedFlags([]string{})
	err := record([]string{}, flags, deps)

	require.NoError(t, err)
	require.False(t, errorShown)

	// With --verbose, should show error
	errorShown = false
	flags = dispatchers.NewParsedFlags([]string{"--verbose"})
	err = record([]string{}, flags, deps)

	require.NoError(t, err)
	require.True(t, errorShown)
}

func TestRecord_NotInGitRepo(t *testing.T) {
	var errorShown bool

	deps := Deps{
		Getenv: func(key string) string {
			return "post-commit" // From hook to avoid note message
		},
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				if str, ok := a[0].(string); ok && str == "not in a git repository" {
					errorShown = true
				}
			}
			return 0, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	// Without --verbose or --manual, should not show error (hook mode)
	flags := dispatchers.NewParsedFlags([]string{})
	err := record([]string{}, flags, deps)

	require.NoError(t, err)
	require.False(t, errorShown)

	// With --manual, should show error
	errorShown = false
	flags = dispatchers.NewParsedFlags([]string{"--manual"})
	err = record([]string{}, flags, deps)

	require.NoError(t, err)
	require.True(t, errorShown)
}

func TestRecord_DifferentSourceTypes(t *testing.T) {
	tests := []struct {
		envValue       string
		expectedSource store.Source
	}{
		{"post-commit", store.SourcePostCommit},
		{"post-rewrite", store.SourcePostRewrite},
		{"post-checkout", store.SourcePostCheckout},
		{"post-merge", store.SourcePostMerge},
		{"pre-push", store.SourcePrePush},
		{"", store.SourceManual},
		{"unknown", store.SourceManual},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			var insertedEvent store.RepoEvent

			deps := Deps{
				Getenv: func(key string) string {
					if key == "FP_SOURCE" {
						return tt.envValue
					}
					return ""
				},
				GitIsAvailable: func() bool { return true },
				RepoRoot: func(path string) (string, error) {
					return "/path/to/repo", nil
				},
				OriginURL: func(repoRoot string) (string, error) {
					return "https://github.com/user/repo.git", nil
				},
				DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
					return "github.com/user/repo", nil
				},
				HeadCommit: func() (string, error) {
					return "abc123", nil
				},
				CurrentBranch: func() (string, error) {
					return "main", nil
				},
				DBPath: func() string {
					return ":memory:"
				},
				OpenDB: func(path string) (*sql.DB, error) {
					db, _ := sql.Open("sqlite3", ":memory:")
					return db, nil
				},
				InitDB: func(db *sql.DB) error {
					return nil
				},
				InsertEvent: func(db *sql.DB, event store.RepoEvent) error {
					insertedEvent = event
					return nil
				},
				Now: func() time.Time {
					return time.Now()
				},
				Println: func(a ...any) (int, error) {
					return 0, nil
				},
				Printf: func(format string, a ...any) (int, error) {
					return 0, nil
				},
			}

			flags := dispatchers.NewParsedFlags([]string{})
			err := record([]string{}, flags, deps)

			require.NoError(t, err)
			require.Equal(t, tt.expectedSource, insertedEvent.Source)
		})
	}
}

func TestRecord_DatabaseOpenError(t *testing.T) {
	var capturedPrintln string

	deps := Deps{
		Getenv: func(key string) string {
			return ""
		},
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/user/repo", nil
		},
		HeadCommit: func() (string, error) {
			return "abc123", nil
		},
		CurrentBranch: func() (string, error) {
			return "main", nil
		},
		DBPath: func() string {
			return "/invalid/path/db.sqlite"
		},
		OpenDB: func(path string) (*sql.DB, error) {
			return nil, errors.New("failed to open db")
		},
		Println: func(a ...any) (int, error) {
			capturedPrintln = "called"
			return 0, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	// Should exit gracefully even with DB error
	flags := dispatchers.NewParsedFlags([]string{})
	err := record([]string{}, flags, deps)

	require.NoError(t, err)

	// With --verbose, should show error
	capturedPrintln = ""
	flags = dispatchers.NewParsedFlags([]string{"--verbose"})
	err = record([]string{}, flags, deps)

	require.NoError(t, err)
	require.NotEmpty(t, capturedPrintln)
}

func TestRecord_InsertEventError(t *testing.T) {
	var capturedPrintf string

	deps := Deps{
		Getenv: func(key string) string {
			return ""
		},
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/user/repo", nil
		},
		HeadCommit: func() (string, error) {
			return "abc123", nil
		},
		CurrentBranch: func() (string, error) {
			return "main", nil
		},
		DBPath: func() string {
			return ":memory:"
		},
		OpenDB: func(path string) (*sql.DB, error) {
			db, _ := sql.Open("sqlite3", ":memory:")
			return db, nil
		},
		InitDB: func(db *sql.DB) error {
			return nil
		},
		InsertEvent: func(db *sql.DB, event store.RepoEvent) error {
			return errors.New("insert failed")
		},
		Now: func() time.Time {
			return time.Now()
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	// Should not return error even when insert fails
	flags := dispatchers.NewParsedFlags([]string{"--verbose"})
	err := record([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "failed")
}

func TestResolveSource(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     store.Source
	}{
		{"post-commit", "post-commit", store.SourcePostCommit},
		{"post-rewrite", "post-rewrite", store.SourcePostRewrite},
		{"post-checkout", "post-checkout", store.SourcePostCheckout},
		{"post-merge", "post-merge", store.SourcePostMerge},
		{"pre-push", "pre-push", store.SourcePrePush},
		{"empty defaults to manual", "", store.SourceManual},
		{"unknown defaults to manual", "unknown-hook", store.SourceManual},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := Deps{
				Getenv: func(key string) string {
					if key == "FP_SOURCE" {
						return tt.envValue
					}
					return ""
				},
			}

			got := resolveSource(deps)
			require.Equal(t, tt.want, got)
		})
	}
}
