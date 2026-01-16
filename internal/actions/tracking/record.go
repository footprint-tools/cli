package tracking

import (
	"github.com/Skryensya/footprint/internal/telemetry"
)

func Record(args []string, flags []string) error {
	return record(args, flags, DefaultDeps())
}

func record(_ []string, flags []string, deps Deps) error {
	verbose := hasFlag(flags, "--verbose")
	manual := hasFlag(flags, "--manual")

	defer func() { _ = recover() }()

	// Show note when running manually (no FP_SOURCE env var)
	if deps.Getenv("FP_SOURCE") == "" && !manual {
		deps.Println("Note: fp record is usually executed automatically by git hooks.")
	}

	if !deps.GitIsAvailable() {
		if verbose {
			deps.Println("git not available")
		}
		return nil
	}

	repoRoot, err := deps.RepoRoot(".")
	if err != nil {
		if verbose {
			deps.Println("not in a git repository")
		}
		return nil
	}

	remoteURL, _ := deps.OriginURL(repoRoot)

	repoID, err := deps.DeriveID(remoteURL, repoRoot)
	if err != nil {
		if verbose {
			deps.Println("could not derive repo id")
		}
		return nil
	}

	tracked, err := deps.IsTracked(repoID)
	if err != nil || !tracked {
		if verbose {
			deps.Println("repository not tracked")
		}
		return nil
	}

	commit, err := deps.HeadCommit()
	if err != nil {
		if verbose {
			deps.Println("could not read HEAD commit")
		}
		return nil
	}

	branch, _ := deps.CurrentBranch()

	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		if verbose {
			deps.Println("could not open telemetry db")
		}
		return nil
	}

	_ = deps.InitDB(db)

	source := resolveSource(deps)

	msg, _ := deps.CommitMessage()

	err = deps.InsertEvent(db, telemetry.RepoEvent{
		RepoID:        string(repoID),
		RepoPath:      repoRoot,
		Commit:        commit,
		CommitMessage: msg,
		Branch:        branch,
		Timestamp:     deps.Now().UTC(),
		Status:        telemetry.StatusPending,
		Source:        source,
	})

	if verbose {
		if err != nil {
			deps.Printf("failed to record event: %v\n", err)
		} else {
			deps.Printf(
				"recorded %.7s on %s (%s) [%s]\n",
				commit,
				branch,
				repoID,
				source.String(),
			)
		}
	}

	return nil
}

func resolveSource(deps Deps) telemetry.Source {
	switch deps.Getenv("FP_SOURCE") {
	case "post-commit":
		return telemetry.SourcePostCommit
	case "post-rewrite":
		return telemetry.SourcePostRewrite
	case "post-checkout":
		return telemetry.SourcePostCheckout
	case "post-merge":
		return telemetry.SourcePostMerge
	case "pre-push":
		return telemetry.SourcePrePush
	default:
		return telemetry.SourceManual
	}
}
