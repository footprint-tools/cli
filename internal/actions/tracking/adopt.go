package tracking

import (
	"os"
	"path/filepath"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/paths"
	repodomain "github.com/Skryensya/footprint/internal/repo"
	"github.com/Skryensya/footprint/internal/store"
	"github.com/Skryensya/footprint/internal/usage"
)

func Adopt(args []string, flags *dispatchers.ParsedFlags) error {
	return adopt(args, flags, DefaultDeps())
}

func adopt(args []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	if !deps.GitIsAvailable() {
		return usage.GitNotInstalled()
	}

	path, err := resolvePath(args)
	if err != nil {
		return usage.InvalidPath()
	}

	repoRoot, err := deps.RepoRoot(path)
	if err != nil {
		return usage.NotInGitRepo()
	}

	remoteURL, err := deps.OriginURL(repoRoot)
	if err != nil || remoteURL == "" {
		return usage.MissingRemote()
	}

	localID, err := deps.DeriveID("", repoRoot)
	if err != nil {
		return usage.InvalidRepo()
	}

	remoteID, err := deps.DeriveID(remoteURL, repoRoot)
	if err != nil {
		return usage.InvalidRepo()
	}

	isLocalTracked, err := deps.IsTracked(localID)
	if err != nil {
		return err
	}

	if !isLocalTracked {
		return usage.InvalidRepo()
	}

	// Update tracking
	if _, err := deps.Untrack(localID); err != nil {
		return err
	}

	if _, err := deps.Track(remoteID); err != nil {
		return err
	}

	deps.Printf("adopted identity:\n  %s\n→ %s\n", localID, remoteID)

	// Migrate pending events in database
	db, err := deps.OpenDB(deps.DBPath())
	if err == nil {
		defer db.Close()
		_ = deps.InitDB(db)

		migrated, err := store.MigratePendingRepoID(db, string(localID), string(remoteID))
		if err == nil && migrated > 0 {
			deps.Printf("migrated %d pending events\n", migrated)
		}
	}

	// Rename export directory if it exists
	if renamed := renameExportDir(localID, remoteID); renamed {
		deps.Println("renamed export directory")
	}

	return nil
}

// renameExportDir renames the export directory from old repo ID to new repo ID.
// Returns true if the directory was renamed.
func renameExportDir(oldID, newID repodomain.RepoID) bool {
	exportRepo := paths.ExportRepoDir()
	reposDir := filepath.Join(exportRepo, "repos")

	oldDir := filepath.Join(reposDir, oldID.ToFilesystemSafe())
	newDir := filepath.Join(reposDir, newID.ToFilesystemSafe())

	// Check if old directory exists
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return false
	}

	// Check if new directory already exists (don't overwrite)
	if _, err := os.Stat(newDir); err == nil {
		return false
	}

	// Rename the directory
	if err := os.Rename(oldDir, newDir); err != nil {
		return false
	}

	return true
}
