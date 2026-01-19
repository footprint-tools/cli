package tracking

import (
	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/store"
)

func Record(args []string, flags *dispatchers.ParsedFlags) error {
	return record(args, flags, DefaultDeps())
}

func record(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	verbose := flags.Has("--verbose")
	manual := flags.Has("--manual")

	// Show note when running manually (no FP_SOURCE env var)
	isFromHook := deps.Getenv("FP_SOURCE") != ""
	if !isFromHook && !manual {
		deps.Println("Note: fp record is usually executed automatically by git hooks.")
	}

	// Show errors when running manually or with --verbose
	showErrors := verbose || manual

	if !deps.GitIsAvailable() {
		if showErrors {
			deps.Println("git not available")
		}
		return nil
	}

	repoRoot, err := deps.RepoRoot(".")
	if err != nil {
		if showErrors {
			deps.Println("not in a git repository")
		}
		return nil
	}

	remoteURL, _ := deps.OriginURL(repoRoot)

	repoID, err := deps.DeriveID(remoteURL, repoRoot)
	if err != nil {
		if showErrors {
			deps.Println("could not derive repo id")
		}
		return nil
	}

	tracked, err := deps.IsTracked(repoID)
	if err != nil || !tracked {
		if showErrors {
			deps.Println("repository not tracked")
		}
		return nil
	}

	commit, err := deps.HeadCommit()
	if err != nil {
		if showErrors {
			deps.Println("could not read HEAD commit")
		}
		return nil
	}

	branch, _ := deps.CurrentBranch()

	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		if showErrors {
			deps.Println("could not open store db")
		}
		return nil
	}
	defer db.Close()

	_ = deps.InitDB(db)

	source := resolveSource(deps)

	msg, _ := deps.CommitMessage()
	author, _ := deps.CommitAuthor()

	err = deps.InsertEvent(db, store.RepoEvent{
		RepoID:        string(repoID),
		RepoPath:      repoRoot,
		Commit:        commit,
		CommitMessage: msg,
		Branch:        branch,
		Author:        author,
		Timestamp:     deps.Now().UTC(),
		Status:        store.StatusPending,
		Source:        source,
	})

	if showErrors {
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

	// Check if we should auto-export
	if err == nil {
		MaybeExport(db, deps)
	}

	return nil
}

func resolveSource(deps Deps) store.Source {
	switch deps.Getenv("FP_SOURCE") {
	case "post-commit":
		return store.SourcePostCommit
	case "post-rewrite":
		return store.SourcePostRewrite
	case "post-checkout":
		return store.SourcePostCheckout
	case "post-merge":
		return store.SourcePostMerge
	case "pre-push":
		return store.SourcePrePush
	default:
		return store.SourceManual
	}
}
