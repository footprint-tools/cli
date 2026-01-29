package tracking

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/footprint-tools/cli/internal/dispatchers"
	repodomain "github.com/footprint-tools/cli/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestBackfill_DryRun_GitNotAvailable(t *testing.T) {
	var printed string
	deps := Deps{
		GitIsAvailable: func() bool { return false },
		Printf: func(format string, a ...any) (int, error) {
			printed = fmt.Sprintf(format, a...)
			return len(printed), nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--dry-run"})

	err := backfill(nil, flags, deps)

	require.Error(t, err, "Should error when git is not available")
}

func TestBackfill_DryRun_InvalidPath(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
	}

	flags := dispatchers.NewParsedFlags([]string{"--dry-run"})

	err := backfill([]string{"/nonexistent/path"}, flags, deps)

	require.Error(t, err)
}

func TestBackfill_DryRun_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags([]string{"--dry-run"})

	err := backfill([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestBackfill_Background_GitNotAvailable(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return false },
	}

	flags := dispatchers.NewParsedFlags([]string{""})

	err := backfill(nil, flags, deps)

	require.Error(t, err)
}

func TestBackfill_Background_InvalidPath(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
	}

	flags := dispatchers.NewParsedFlags([]string{""})

	err := backfill([]string{"/nonexistent/path"}, flags, deps)

	require.Error(t, err)
}

func TestBackfill_Background_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags([]string{""})

	err := backfill([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestBackfill_Background_InvalidRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return path, nil
		},
		OriginURL: func(path string) (string, error) {
			return "", nil
		},
		DeriveID: func(remoteURL, path string) (repodomain.RepoID, error) {
			return "", errors.New("invalid repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags([]string{""})

	err := backfill([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestBackfill_Background_DBOpenError(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return path, nil
		},
		OriginURL: func(path string) (string, error) {
			return "git@github.com:user/repo.git", nil
		},
		DeriveID: func(remoteURL, path string) (repodomain.RepoID, error) {
			return "github.com/user/repo", nil
		},
		DBPath: func() string {
			return "/nonexistent/db.sqlite"
		},
		OpenDB: func(path string) (*sql.DB, error) {
			return nil, errors.New("failed to open db")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags([]string{""})

	err := backfill([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestDoBackfillDryRun_GitNotAvailable(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return false },
	}

	flags := dispatchers.NewParsedFlags([]string{"--dry-run"})

	err := doBackfillDryRun(nil, flags, deps)

	require.Error(t, err)
}

func TestDoBackfillDryRun_InvalidPath(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
	}

	flags := dispatchers.NewParsedFlags([]string{"--dry-run"})

	err := doBackfillDryRun([]string{"/nonexistent/path"}, flags, deps)

	require.Error(t, err)
}

func TestDoBackfillDryRun_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags([]string{"--dry-run"})

	err := doBackfillDryRun([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestDoBackfillWork_GitNotAvailable(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return false },
	}

	flags := dispatchers.NewParsedFlags([]string{""})

	err := doBackfillText(nil, flags, deps)

	require.Error(t, err)
}

func TestDoBackfillWork_InvalidPath(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
	}

	flags := dispatchers.NewParsedFlags([]string{""})

	err := doBackfillText([]string{"/nonexistent/path"}, flags, deps)

	require.Error(t, err)
}

func TestDoBackfillWork_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags([]string{""})

	err := doBackfillText([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestDoBackfillWork_InvalidRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return path, nil
		},
		OriginURL: func(path string) (string, error) {
			return "", nil
		},
		DeriveID: func(remoteURL, path string) (repodomain.RepoID, error) {
			return "", errors.New("invalid repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags([]string{""})

	err := doBackfillText([]string{dir}, flags, deps)

	require.Error(t, err)
}
