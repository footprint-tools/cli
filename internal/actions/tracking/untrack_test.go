package tracking

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/repo"
)

func TestUntrack_Success(t *testing.T) {
	var capturedPrintf string
	deps := Deps{
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
		Untrack: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := untrack([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "untracked")
}

func TestUntrack_NotTracked(t *testing.T) {
	var capturedPrintf string
	deps := Deps{
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
		Untrack: func(id repo.RepoID) (bool, error) {
			return false, nil // Not tracked
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := untrack([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "not tracked")
}

func TestUntrack_GitNotInstalled(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return false },
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := untrack([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Git is not installed")
}

func TestUntrack_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := untrack([]string{}, flags, deps)

	require.Error(t, err)
}

func TestUntrack_ByIDFlag(t *testing.T) {
	var capturedPrintf string
	var untrackCalledWith repo.RepoID
	deps := Deps{
		Untrack: func(id repo.RepoID) (bool, error) {
			untrackCalledWith = id
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--id=github.com/user/old-repo"})
	err := untrack([]string{}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, repo.RepoID("github.com/user/old-repo"), untrackCalledWith)
	require.Contains(t, capturedPrintf, "untracked")
}

func TestUntrack_ByIDNotTracked(t *testing.T) {
	var capturedPrintf string
	deps := Deps{
		Untrack: func(id repo.RepoID) (bool, error) {
			return false, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--id=nonexistent/repo"})
	err := untrack([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "not tracked")
}

func TestUntrack_ByIDError(t *testing.T) {
	deps := Deps{
		Untrack: func(id repo.RepoID) (bool, error) {
			return false, errors.New("database error")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--id=github.com/user/repo"})
	err := untrack([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "database error")
}

func TestUntrack_WithPath(t *testing.T) {
	var receivedPath string
	tempDir := t.TempDir()
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			receivedPath = path
			return path, nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/user/repo", nil
		},
		Untrack: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := untrack([]string{tempDir}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, receivedPath, tempDir)
}

func TestUntrack_UntrackError(t *testing.T) {
	deps := Deps{
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
		Untrack: func(id repo.RepoID) (bool, error) {
			return false, errors.New("database error")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := untrack([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "database error")
}
