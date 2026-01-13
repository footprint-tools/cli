package actions

import (
	"fmt"
	"os"
	"time"

	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/repo"
	"github.com/Skryensya/footprint/internal/telemetry"
)

func RepoRecord(_ []string, flags []string) error {
	verbose := hasFlag(flags, "--verbose")

	defer func() { _ = recover() }()

	if !git.IsAvailable() {
		if verbose {
			fmt.Println("git not available")
		}
		return nil
	}

	repoRoot, err := git.RepoRoot(".")
	if err != nil {
		if verbose {
			fmt.Println("not in a git repository")
		}
		return nil
	}

	remoteURL, _ := git.OriginURL(repoRoot)

	repoID, err := repo.DeriveID(remoteURL, repoRoot)
	if err != nil {
		if verbose {
			fmt.Println("could not derive repo id")
		}
		return nil
	}

	tracked, err := repo.IsTracked(repoID)
	if err != nil || !tracked {
		if verbose {
			fmt.Println("repository not tracked")
		}
		return nil
	}

	commit, err := git.HeadCommit()
	if err != nil {
		if verbose {
			fmt.Println("could not read HEAD commit")
		}
		return nil
	}

	branch, _ := git.CurrentBranch()

	db, err := telemetry.Open(telemetry.DBPath())
	if err != nil {
		if verbose {
			fmt.Println("could not open telemetry db")
		}
		return nil
	}

	_ = telemetry.Init(db)

	source := resolveSource()

	msg, _ := git.CommitMessage()

	err = telemetry.InsertEvent(db, telemetry.RepoEvent{
		RepoID:        string(repoID),
		RepoPath:      repoRoot,
		Commit:        commit,
		CommitMessage: msg,
		Branch:        branch,
		Timestamp:     time.Now().UTC(),
		Status:        telemetry.StatusPending,
		Source:        source,
	})

	if verbose {
		if err != nil {
			fmt.Printf("failed to record event: %v\n", err)
		} else {
			fmt.Printf(
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

func resolveSource() telemetry.Source {
	switch os.Getenv("FP_HOOK") {
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
