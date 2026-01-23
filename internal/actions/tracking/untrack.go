package tracking

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/log"
	"github.com/footprint-tools/footprint-cli/internal/repo"
	"github.com/footprint-tools/footprint-cli/internal/usage"
)

func Untrack(args []string, flags *dispatchers.ParsedFlags) error {
	return untrack(args, flags, DefaultDeps())
}

func untrack(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	log.Debug("untrack: starting")

	// Check for --id flag first
	idValue := flags.String("--id", "")
	if idValue != "" {
		log.Debug("untrack: using --id flag: %s", idValue)
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

	log.Debug("untrack: repo=%s, path=%s", id, repoRoot)

	removed, err := deps.Untrack(id)
	if err != nil {
		log.Error("untrack: failed to untrack repo: %v", err)
		return err
	}

	if !removed {
		log.Debug("untrack: repo was not tracked")
		_, _ = deps.Printf("repository not tracked: %s\n", id)
		return nil
	}

	log.Info("untrack: stopped tracking %s", id)
	_, _ = deps.Printf("untracked %s\n", id)
	return nil
}

func untrackByID(idStr string, deps Deps) error {
	id := repo.RepoID(idStr)

	removed, err := deps.Untrack(id)
	if err != nil {
		log.Error("untrack: failed to untrack by ID: %v", err)
		return err
	}

	if !removed {
		log.Debug("untrack: repo was not tracked (id=%s)", id)
		_, _ = deps.Printf("repository not tracked: %s\n", id)
		return nil
	}

	log.Info("untrack: stopped tracking %s", id)
	_, _ = deps.Printf("untracked %s\n", id)
	return nil
}
