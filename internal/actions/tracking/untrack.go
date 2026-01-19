package tracking

import (
	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/repo"
	"github.com/Skryensya/footprint/internal/usage"
)

func Untrack(args []string, flags *dispatchers.ParsedFlags) error {
	return untrack(args, flags, DefaultDeps())
}

func untrack(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	// Check for --id flag first
	idValue := flags.String("--id", "")
	if idValue != "" {
		return untrackByID(idValue, deps)
	}

	// Normal path-based untrack
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

	remoteURL, _ := deps.OriginURL(repoRoot)

	id, err := deps.DeriveID(remoteURL, repoRoot)
	if err != nil {
		return usage.InvalidRepo()
	}

	removed, err := deps.Untrack(id)
	if err != nil {
		return err
	}

	if !removed {
		deps.Printf("repository not tracked: %s\n", id)
		return nil
	}

	deps.Printf("untracked %s\n", id)
	return nil
}

func untrackByID(idStr string, deps Deps) error {
	id := repo.RepoID(idStr)

	removed, err := deps.Untrack(id)
	if err != nil {
		return err
	}

	if !removed {
		deps.Printf("repository not tracked: %s\n", id)
		return nil
	}

	deps.Printf("untracked %s\n", id)
	return nil
}
