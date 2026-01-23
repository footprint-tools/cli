package tracking

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/log"
	"github.com/footprint-tools/footprint-cli/internal/usage"
)

func Track(args []string, flags *dispatchers.ParsedFlags) error {
	return track(args, flags, DefaultDeps())
}

func track(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	log.Debug("track: starting")

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

	// Determine the remote URL to use
	remoteURL, err := resolveRemoteURL(repoRoot, flags, deps)
	if err != nil {
		log.Debug("track: failed to resolve remote URL: %v", err)
		return err
	}

	id, err := deps.DeriveID(remoteURL, repoRoot)
	if err != nil {
		return usage.InvalidRepo()
	}

	log.Debug("track: repo=%s, path=%s", id, repoRoot)

	added, err := deps.Track(id)
	if err != nil {
		log.Error("track: failed to track repo: %v", err)
		return err
	}

	if !added {
		log.Debug("track: repo already tracked")
		_, _ = deps.Printf("already tracking %s\n", id)
		return nil
	}

	log.Info("track: now tracking %s", id)
	_, _ = deps.Printf("tracking %s\n", id)
	return nil
}

// resolveRemoteURL returns the remote URL to use.
// Priority: --remote flag > origin > single remote > error if ambiguous.
func resolveRemoteURL(repoRoot string, flags *dispatchers.ParsedFlags, deps Deps) (string, error) {
	// Check for --remote flag
	specifiedRemote := flags.String("--remote", "")
	if specifiedRemote != "" {
		url, err := deps.GetRemoteURL(repoRoot, specifiedRemote)
		if err != nil {
			return "", usage.MissingRemote()
		}
		return url, nil
	}

	// Try origin first
	if url, err := deps.OriginURL(repoRoot); err == nil && url != "" {
		return url, nil
	}

	// No origin, check available remotes
	remotes, err := deps.ListRemotes(repoRoot)
	if err != nil || len(remotes) == 0 {
		// No remotes at all, will use local path
		return "", nil
	}

	if len(remotes) == 1 {
		// Exactly one remote, use it
		url, err := deps.GetRemoteURL(repoRoot, remotes[0])
		if err != nil {
			return "", nil
		}
		return url, nil
	}

	// Multiple remotes without origin - ambiguous
	return "", usage.AmbiguousRemote(remotes)
}

