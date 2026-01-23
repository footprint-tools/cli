package tracking

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/log"
	"github.com/footprint-tools/footprint-cli/internal/store"
)

func Record(args []string, flags *dispatchers.ParsedFlags) error {
	return record(args, flags, DefaultDeps())
}

func record(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	verbose := flags.Has("--verbose")
	manual := flags.Has("--manual")

	// Show note when running manually (no FP_SOURCE env var)
	isFromHook := deps.Getenv("FP_SOURCE") != ""
	log.Debug("record: starting (source=%s, fromHook=%v)", deps.Getenv("FP_SOURCE"), isFromHook)

	if !isFromHook && !manual {
		_, _ = deps.Println("Note: fp record is usually executed automatically by git hooks.")
	}

	// Show errors when running manually or with --verbose
	showErrors := verbose || manual

	if !deps.GitIsAvailable() {
		if showErrors {
			_, _ = deps.Println("git not available")
		}
		return nil
	}

	repoRoot, err := deps.RepoRoot(".")
	if err != nil {
		if showErrors {
			_, _ = deps.Println("not in a git repository")
		}
		return nil
	}

	remoteURL, _ := deps.OriginURL(repoRoot)

	repoID, err := deps.DeriveID(remoteURL, repoRoot)
	if err != nil {
		if showErrors {
			_, _ = deps.Println("could not derive repo id")
		}
		return nil
	}

	tracked, err := deps.IsTracked(repoID)
	if err != nil || !tracked {
		// Warning: hooks are running in an untracked repository
		// This might indicate hooks weren't cleaned up properly
		if isFromHook {
			log.Warn("fp record: repository not tracked but hooks are active (repo=%s, path=%s)", repoID, repoRoot)
		}
		if showErrors {
			_, _ = deps.Println("repository not tracked")
		}
		return nil
	}

	commit, err := deps.HeadCommit()
	if err != nil {
		if showErrors {
			_, _ = deps.Println("could not read HEAD commit")
		}
		return nil
	}

	branch, _ := deps.CurrentBranch()
	log.Debug("record: repo=%s, commit=%.7s, branch=%s, path=%s", repoID, commit, branch, repoRoot)

	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		// Critical error: log it always
		log.Error("fp record: failed to open database: %v (repo=%s, commit=%.7s)", err, repoID, commit)
		if showErrors {
			_, _ = deps.Println("could not open store db")
		}
		return nil
	}
	defer store.CloseDB(db)

	if err := deps.InitDB(db); err != nil {
		// Critical error: DB initialization failed
		log.Error("fp record: failed to initialize database: %v (repo=%s, commit=%.7s)", err, repoID, commit)
		if showErrors {
			_, _ = deps.Printf("failed to initialize database: %v\n", err)
		}
		return nil
	}

	source := resolveSource(deps)

	err = deps.InsertEvent(db, store.RepoEvent{
		RepoID:    string(repoID),
		RepoPath:  repoRoot,
		Commit:    commit,
		Branch:    branch,
		Timestamp: deps.Now().UTC(),
		Status:    store.StatusPending,
		Source:    source,
	})

	if err != nil {
		// Critical error: failed to record event
		log.Error("fp record: failed to insert event: %v (repo=%s, commit=%.7s, source=%s)", err, repoID, commit, source.String())
	} else {
		log.Info("record: event saved (repo=%s, commit=%.7s, source=%s)", repoID, commit, source.String())
	}

	if showErrors {
		if err != nil {
			_, _ = deps.Printf("failed to record event: %v\n", err)
		} else {
			_, _ = deps.Printf(
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
