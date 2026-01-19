package tracking

import (
	"github.com/Skryensya/footprint/internal/dispatchers"
	repodomain "github.com/Skryensya/footprint/internal/repo"
	"github.com/Skryensya/footprint/internal/usage"
)

func Status(args []string, flags *dispatchers.ParsedFlags) error {
	return status(args, flags, DefaultDeps())
}

func status(args []string, _ *dispatchers.ParsedFlags, deps Deps) error {
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

	localID, err := deps.DeriveID("", repoRoot)
	if err != nil {
		return usage.InvalidRepo()
	}

	remoteID := repodomain.RepoID("")
	if remoteURL != "" {
		remoteID, _ = deps.DeriveID(remoteURL, repoRoot)
	}

	isLocalTracked, err := deps.IsTracked(localID)
	if err != nil {
		return err
	}

	isRemoteTracked := false
	if remoteID != "" {
		isRemoteTracked, err = deps.IsTracked(remoteID)
		if err != nil {
			return err
		}
	}

	if isRemoteTracked {
		deps.Printf("tracked %s\n", remoteID)
		return nil
	}

	if isLocalTracked {
		deps.Printf("tracked %s\n", localID)

		if remoteID != "" && localID != remoteID {
			deps.Printf("remote detected %s\n", remoteID)
			deps.Println("run 'fp adopt' to update identity")
		}

		return nil
	}

	if remoteID != "" {
		deps.Printf("not tracked %s\n", remoteID)
	} else {
		deps.Printf("not tracked %s\n", localID)
	}

	return nil
}
