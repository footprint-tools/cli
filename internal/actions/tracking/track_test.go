package tracking

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/repo"
)

func TestTrack_Success(t *testing.T) {
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
		Track: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "tracking")
}

func TestTrack_AlreadyTracked(t *testing.T) {
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
		Track: func(id repo.RepoID) (bool, error) {
			return false, nil // Not added (already tracked)
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "already tracking")
}

func TestTrack_GitNotInstalled(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return false },
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Git is not installed")
}

func TestTrack_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.Error(t, err)
}

func TestTrack_WithRemoteFlag(t *testing.T) {
	var usedRemote string
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		GetRemoteURL: func(repoRoot, remoteName string) (string, error) {
			usedRemote = remoteName
			return "https://github.com/user/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/user/repo", nil
		},
		Track: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--remote=upstream"})
	err := track([]string{}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, "upstream", usedRemote)
}

func TestTrack_InvalidRemoteFlag(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		GetRemoteURL: func(repoRoot, remoteName string) (string, error) {
			return "", errors.New("remote not found")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--remote=nonexistent"})
	err := track([]string{}, flags, deps)

	require.Error(t, err)
}

func TestTrack_NoRemoteUsesLocal(t *testing.T) {
	var derivedRemoteURL string
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		ListRemotes: func(repoRoot string) ([]string, error) {
			return []string{}, nil // No remotes
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			derivedRemoteURL = remoteURL
			return "local:/path/to/repo", nil
		},
		Track: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.NoError(t, err)
	require.Empty(t, derivedRemoteURL, "should use empty remote URL for local repos")
}

func TestTrack_SingleRemoteWithoutOrigin(t *testing.T) {
	var usedRemote string
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		ListRemotes: func(repoRoot string) ([]string, error) {
			return []string{"upstream"}, nil
		},
		GetRemoteURL: func(repoRoot, remoteName string) (string, error) {
			usedRemote = remoteName
			return "https://github.com/upstream/repo.git", nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			return "github.com/upstream/repo", nil
		},
		Track: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, "upstream", usedRemote)
}

func TestTrack_MultipleRemotesWithoutOrigin(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		ListRemotes: func(repoRoot string) ([]string, error) {
			return []string{"upstream", "fork"}, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "multiple remotes")
}

func TestTrack_PrefersOriginOverOtherRemotes(t *testing.T) {
	var derivedRemoteURL string
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "/path/to/repo", nil
		},
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/origin/repo.git", nil
		},
		// ListRemotes should NOT be called if origin exists
		ListRemotes: func(repoRoot string) ([]string, error) {
			t.Fatal("should not list remotes when origin exists")
			return nil, nil
		},
		DeriveID: func(remoteURL, repoRoot string) (repo.RepoID, error) {
			derivedRemoteURL = remoteURL
			return "github.com/origin/repo", nil
		},
		Track: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := track([]string{}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, "https://github.com/origin/repo.git", derivedRemoteURL)
}

func TestTrack_WithExplicitPath(t *testing.T) {
	var receivedPath string
	tempDir := t.TempDir() // Use a real directory that exists
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
		Track: func(id repo.RepoID) (bool, error) {
			return true, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	// Use a real temp directory so resolvePath() doesn't fail
	err := track([]string{tempDir}, flags, deps)

	require.NoError(t, err)
	require.NotEmpty(t, receivedPath)
	require.Contains(t, receivedPath, tempDir)
}

func TestResolveRemoteURL_WithRemoteFlag(t *testing.T) {
	deps := Deps{
		GetRemoteURL: func(repoRoot, remoteName string) (string, error) {
			if remoteName == "upstream" {
				return "https://github.com/upstream/repo.git", nil
			}
			return "", errors.New("not found")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--remote=upstream"})
	url, err := resolveRemoteURL("/repo", flags, deps)

	require.NoError(t, err)
	require.Equal(t, "https://github.com/upstream/repo.git", url)
}

func TestResolveRemoteURL_PreferOrigin(t *testing.T) {
	deps := Deps{
		OriginURL: func(repoRoot string) (string, error) {
			return "https://github.com/origin/repo.git", nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	url, err := resolveRemoteURL("/repo", flags, deps)

	require.NoError(t, err)
	require.Equal(t, "https://github.com/origin/repo.git", url)
}

func TestResolveRemoteURL_SingleRemote(t *testing.T) {
	deps := Deps{
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		ListRemotes: func(repoRoot string) ([]string, error) {
			return []string{"upstream"}, nil
		},
		GetRemoteURL: func(repoRoot, remoteName string) (string, error) {
			return "https://github.com/upstream/repo.git", nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	url, err := resolveRemoteURL("/repo", flags, deps)

	require.NoError(t, err)
	require.Equal(t, "https://github.com/upstream/repo.git", url)
}

func TestResolveRemoteURL_NoRemotes(t *testing.T) {
	deps := Deps{
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		ListRemotes: func(repoRoot string) ([]string, error) {
			return []string{}, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	url, err := resolveRemoteURL("/repo", flags, deps)

	require.NoError(t, err)
	require.Empty(t, url, "should return empty string for local repos")
}

func TestResolveRemoteURL_MultipleRemotesAmbiguous(t *testing.T) {
	deps := Deps{
		OriginURL: func(repoRoot string) (string, error) {
			return "", errors.New("no origin")
		},
		ListRemotes: func(repoRoot string) ([]string, error) {
			return []string{"upstream", "fork", "other"}, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	_, err := resolveRemoteURL("/repo", flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "multiple remotes")
}
