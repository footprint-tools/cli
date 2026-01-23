package tracking

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	repodomain "github.com/footprint-tools/footprint-cli/internal/repo"
	"github.com/footprint-tools/footprint-cli/internal/usage"
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
		_, _ = deps.Printf("tracked %s\n", remoteID)
		return nil
	}

	if isLocalTracked {
		_, _ = deps.Printf("tracked %s\n", localID)

		if remoteID != "" && localID != remoteID {
			_, _ = deps.Printf("remote detected %s\n", remoteID)
			_, _ = deps.Println("run 'fp adopt' to update identity")
		}

		return nil
	}

	if remoteID != "" {
		_, _ = deps.Printf("not tracked %s\n", remoteID)
	} else {
		_, _ = deps.Printf("not tracked %s\n", localID)
	}

	return nil
}
