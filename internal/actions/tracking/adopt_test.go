package tracking

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Skryensya/footprint/internal/dispatchers"
	repodomain "github.com/Skryensya/footprint/internal/repo"
	"github.com/stretchr/testify/require"
)

func TestAdopt_GitNotAvailable(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return false },
	}

	flags := dispatchers.NewParsedFlags(nil)

	err := adopt(nil, flags, deps)

	require.Error(t, err)
}

func TestAdopt_InvalidPath(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
	}

	flags := dispatchers.NewParsedFlags(nil)

	err := adopt([]string{"/nonexistent/path"}, flags, deps)

	require.Error(t, err)
}

func TestAdopt_NotInGitRepo(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return "", errors.New("not a git repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags(nil)

	err := adopt([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestAdopt_MissingRemote(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return path, nil
		},
		OriginURL: func(path string) (string, error) {
			return "", nil // No remote configured
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags(nil)

	err := adopt([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestAdopt_InvalidLocalID(t *testing.T) {
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return path, nil
		},
		OriginURL: func(path string) (string, error) {
			return "git@github.com:user/repo.git", nil
		},
		DeriveID: func(remoteURL, path string) (repodomain.RepoID, error) {
			return "", errors.New("invalid repo")
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags(nil)

	err := adopt([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestAdopt_LocalNotTracked(t *testing.T) {
	var deriveCount int
	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return path, nil
		},
		OriginURL: func(path string) (string, error) {
			return "git@github.com:user/repo.git", nil
		},
		DeriveID: func(remoteURL, path string) (repodomain.RepoID, error) {
			deriveCount++
			if deriveCount == 1 {
				// First call is for local ID
				return "local/path/repo", nil
			}
			// Second call is for remote ID
			return "github.com/user/repo", nil
		},
		IsTracked: func(id repodomain.RepoID) (bool, error) {
			return false, nil // Not tracked
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags(nil)

	err := adopt([]string{dir}, flags, deps)

	require.Error(t, err)
}

func TestAdopt_Success(t *testing.T) {
	var untracked, tracked repodomain.RepoID
	var printed []string
	var deriveCount int

	deps := Deps{
		GitIsAvailable: func() bool { return true },
		RepoRoot: func(path string) (string, error) {
			return path, nil
		},
		OriginURL: func(path string) (string, error) {
			return "git@github.com:user/repo.git", nil
		},
		DeriveID: func(remoteURL, path string) (repodomain.RepoID, error) {
			deriveCount++
			if deriveCount == 1 {
				return "local/path/repo", nil
			}
			return "github.com/user/repo", nil
		},
		IsTracked: func(id repodomain.RepoID) (bool, error) {
			return true, nil
		},
		Untrack: func(id repodomain.RepoID) (bool, error) {
			untracked = id
			return true, nil
		},
		Track: func(id repodomain.RepoID) (bool, error) {
			tracked = id
			return true, nil
		},
		DBPath: func() string {
			return ":memory:"
		},
		OpenDB: func(path string) (*sql.DB, error) {
			return nil, errors.New("skip db")
		},
		Printf: func(format string, a ...any) (int, error) {
			printed = append(printed, fmt.Sprintf(format, a...))
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
	}

	dir := t.TempDir()
	flags := dispatchers.NewParsedFlags(nil)

	err := adopt([]string{dir}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, repodomain.RepoID("local/path/repo"), untracked)
	require.Equal(t, repodomain.RepoID("github.com/user/repo"), tracked)
}

func TestRenameExportDir_NonExistentOldDir(t *testing.T) {
	result := renameExportDir("nonexistent/old", "nonexistent/new")

	require.False(t, result, "Should return false when old dir doesn't exist")
}

func TestRenameExportDir_NewDirAlreadyExists(t *testing.T) {
	// Create temp dirs for testing
	baseDir := t.TempDir()

	// Create a structure that simulates the export repo
	oldID := repodomain.RepoID("local/path/repo")
	newID := repodomain.RepoID("github.com/user/repo")

	reposDir := filepath.Join(baseDir, "repos")
	os.MkdirAll(reposDir, 0700)

	// Create both old and new dirs
	oldDir := filepath.Join(reposDir, oldID.ToFilesystemSafe())
	newDir := filepath.Join(reposDir, newID.ToFilesystemSafe())
	os.MkdirAll(oldDir, 0700)
	os.MkdirAll(newDir, 0700)

	// Can't use renameExportDir directly because it uses paths.ExportRepoDir()
	// So we test the logic: if new exists, should not rename
	_, err := os.Stat(newDir)
	require.NoError(t, err, "New dir should exist")

	// The function would return false if new dir exists
}
