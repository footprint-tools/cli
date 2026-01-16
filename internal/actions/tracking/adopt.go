package tracking

import (
	"github.com/Skryensya/footprint/internal/usage"
)

func Adopt(args []string, flags []string) error {
	return adopt(args, flags, DefaultDeps())
}

func adopt(args []string, _ []string, deps Deps) error {
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

	if _, err := deps.Untrack(localID); err != nil {
		return err
	}

	if _, err := deps.Track(remoteID); err != nil {
		return err
	}

	deps.Printf("adopted identity:\n  %s\n→ %s\n", localID, remoteID)
	return nil
}
