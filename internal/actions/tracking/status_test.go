package tracking

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/repo"
)

func TestStatus_GitNotInstalled(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return false },
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Git is not installed")
}

func TestStatus_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.Error(t, err)
}

func TestStatus_RemoteTracked(t *testing.T) {
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
			if remoteURL == "" {
				return "local:/path/to/repo", nil
			}
			return "github.com/user/repo", nil
		},
		IsTracked: func(id repo.RepoID) (bool, error) {
			// Remote is tracked
			return id == "github.com/user/repo", nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "tracked")
}

func TestStatus_LocalTracked(t *testing.T) {
	var capturedPrintfCalls []string

	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			if remoteURL == "" {
				return "local:/path/to/repo", nil
			}
			return "github.com/user/repo", nil
		},
		IsTracked: func(id repo.RepoID) (bool, error) {
			// Only local is tracked
			return id == "local:/path/to/repo", nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintfCalls = append(capturedPrintfCalls, format)
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.NoError(t, err)
	require.NotEmpty(t, capturedPrintfCalls)
	require.Contains(t, capturedPrintfCalls[0], "tracked")
}

func TestStatus_LocalTrackedWithRemoteDetected(t *testing.T) {
	var capturedPrintfCalls []string
	var capturedPrintlnCalls []string

	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			if remoteURL == "" {
				return "local:/path/to/repo", nil
			}
			return "github.com/user/repo", nil
		},
		IsTracked: func(id repo.RepoID) (bool, error) {
			// Only local is tracked, not remote
			return id == "local:/path/to/repo", nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintfCalls = append(capturedPrintfCalls, format)
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				capturedPrintlnCalls = append(capturedPrintlnCalls, a[0].(string))
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.NoError(t, err)
	require.Len(t, capturedPrintfCalls, 2)
	require.Contains(t, capturedPrintfCalls[0], "tracked")
	require.Contains(t, capturedPrintfCalls[1], "remote detected")
	require.Contains(t, capturedPrintlnCalls, "run 'fp adopt' to update identity")
}

func TestStatus_NotTrackedWithRemote(t *testing.T) {
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
			if remoteURL == "" {
				return "local:/path/to/repo", nil
			}
			return "github.com/user/repo", nil
		},
		IsTracked: func(id repo.RepoID) (bool, error) {
			return false, nil // Not tracked
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "not tracked")
}

func TestStatus_NotTrackedLocalOnly(t *testing.T) {
	var capturedPrintf string

	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "local:/path/to/repo", nil
		},
		IsTracked: func(id repo.RepoID) (bool, error) {
			return false, nil // Not tracked
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "not tracked")
}

func TestStatus_WithExplicitPath(t *testing.T) {
	var receivedPath string
	tempDir := t.TempDir()

	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			receivedPath = path
			return path, nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "local:/path", nil
		},
		IsTracked: func(id repo.RepoID) (bool, error) {
			return false, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{tempDir}, flags, deps)

	require.NoError(t, err)
	require.NotEmpty(t, receivedPath)
	require.Contains(t, receivedPath, tempDir)
}

func TestStatus_IsTrackedError(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "local:/path/to/repo", nil
		},
		IsTracked: func(id repo.RepoID) (bool, error) {
			return false, errors.New("database error")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := status([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "database error")
}
