package tracking

import (
	"github.com/Skryensya/footprint/internal/usage"
)

func Track(args []string, flags []string) error {
	return track(args, flags, DefaultDeps())
}

func track(args []string, _ []string, deps Deps) error {
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

	added, err := deps.Track(id)
	if err != nil {
		return err
	}

	if !added {
		deps.Printf("already tracking %s\n", id)
		return nil
	}

	deps.Printf("tracking %s\n", id)
	return nil
}
