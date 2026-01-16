package tracking

import (
	"github.com/Skryensya/footprint/internal/usage"
)

func Untrack(args []string, flags []string) error {
	return untrack(args, flags, DefaultDeps())
}

func untrack(args []string, _ []string, deps Deps) error {
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
