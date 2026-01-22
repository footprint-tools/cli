package tracking

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/log"
	"github.com/Skryensya/footprint/internal/store"
	"github.com/Skryensya/footprint/internal/usage"
)

// Backfill imports historical commits from a git repository into the database.
func Backfill(args []string, flags *dispatchers.ParsedFlags) error {
	return backfill(args, flags, DefaultDeps())
}

func backfill(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	// Check if running in background mode
	if flags.Has("--background") {
		return doBackfillWork(args, flags, deps)
	}

	// Check for dry-run (runs in foreground)
	if flags.Has("--dry-run") {
		return doBackfillDryRun(args, flags, deps)
	}

	// Foreground mode: launch background process and watch
	return launchBackfillAndWatch(args, flags, deps)
}

// launchBackfillAndWatch starts the backfill in background and runs watch.
func launchBackfillAndWatch(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	// Build the command to run backfill in background
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}

	// Construct background command args
	cmdArgs := []string{"backfill", "--background"}
	cmdArgs = append(cmdArgs, args...)

	// Pass through filter flags
	for _, f := range flags.Raw() {
		if f != "--dry-run" { // Don't pass dry-run to background
			cmdArgs = append(cmdArgs, f)
		}
	}

	// Start background process
	cmd := exec.Command(execPath, cmdArgs...)
	cmd.Stdout = nil // Detach stdout
	cmd.Stderr = nil // Detach stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start background process: %w", err)
	}

	// Reap the background process to prevent zombies
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Debug("backfill: background process exited with error: %v", err)
		}
	}()

	deps.Println("Starting backfill in background...")

	// Now run the watch command in foreground
	return Log([]string{}, dispatchers.NewParsedFlags([]string{"--oneline"}))
}

// setupBackfill validates the environment and resolves the repository.
func setupBackfill(args []string, deps Deps) (repoID string, repoRoot string, err error) {
	if !deps.GitIsAvailable() {
		return "", "", usage.GitNotInstalled()
	}

	path, err := resolvePath(args)
	if err != nil {
		return "", "", usage.InvalidPath()
	}

	repoRoot, err = deps.RepoRoot(path)
	if err != nil {
		return "", "", usage.NotInGitRepo()
	}

	remoteURL, _ := deps.OriginURL(repoRoot)
	id, err := deps.DeriveID(remoteURL, repoRoot)
	if err != nil {
		return "", "", usage.InvalidRepo()
	}

	return string(id), repoRoot, nil
}

// doBackfillWork performs the actual backfill (runs in background).
func doBackfillWork(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	log.Debug("backfill: starting background work")

	repoID, repoRoot, err := setupBackfill(args, deps)
	if err != nil {
		return err
	}

	log.Debug("backfill: repo=%s, path=%s", repoID, repoRoot)

	// Parse filter options
	opts := git.ListCommitsOptions{
		Since: flags.String("--since", ""),
		Until: flags.String("--until", ""),
		Limit: flags.Int("--limit", 0),
	}

	// Get commits from git log
	commits, err := git.ListCommits(repoRoot, opts)
	if err != nil {
		log.Error("backfill: failed to list commits: %v", err)
		return fmt.Errorf("could not list commits: %w", err)
	}

	if len(commits) == 0 {
		log.Debug("backfill: no commits found")
		return nil
	}

	log.Debug("backfill: found %d commits to import", len(commits))

	// Open database
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		log.Error("backfill: failed to open database: %v", err)
		return fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	_ = deps.InitDB(db)

	// Check for --branch override
	branchOverride := flags.String("--branch", "")

	// Insert each commit as an event
	imported := 0
	for _, c := range commits {
		// Determine branch
		branch := branchOverride
		if branch == "" {
			branch = git.GetBranchForCommit(repoRoot, c.Hash)
			if branch == "" {
				branch = "unknown"
			}
		}

		// Parse author date
		timestamp, err := time.Parse(time.RFC3339, c.AuthorDate)
		if err != nil {
			log.Warn("backfill: could not parse author date '%s' for commit %s, using current time", c.AuthorDate, c.Hash[:7])
			timestamp = time.Now().UTC()
		}

		// Create event
		event := store.RepoEvent{
			RepoID:    repoID,
			RepoPath:  repoRoot,
			Commit:    c.Hash,
			Branch:    branch,
			Timestamp: timestamp.UTC(),
			Status:    store.StatusPending,
			Source:    store.SourceBackfill,
		}

		// Insert (UPSERT handles duplicates)
		if err := deps.InsertEvent(db, event); err == nil {
			imported++
		}
	}

	log.Info("backfill: imported %d commits for %s", imported, repoID)
	return nil
}

// doBackfillDryRun shows what would be imported without doing it.
func doBackfillDryRun(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	repoID, repoRoot, err := setupBackfill(args, deps)
	if err != nil {
		return err
	}

	opts := git.ListCommitsOptions{
		Since: flags.String("--since", ""),
		Until: flags.String("--until", ""),
		Limit: flags.Int("--limit", 0),
	}

	commits, err := git.ListCommits(repoRoot, opts)
	if err != nil {
		return fmt.Errorf("could not list commits: %w", err)
	}

	deps.Printf("Repository: %s\n", repoID)
	deps.Printf("Found %d commits to import:\n\n", len(commits))

	branchOverride := flags.String("--branch", "")

	for _, c := range commits {
		branch := branchOverride
		if branch == "" {
			branch = git.GetBranchForCommit(repoRoot, c.Hash)
			if branch == "" {
				branch = "unknown"
			}
		}

		// Truncate subject if too long
		subject := c.Subject
		if len(subject) > 50 {
			subject = subject[:47] + "..."
		}

		deps.Printf("  %.7s %s %s \"%s\"\n", c.Hash, c.AuthorDate[:10], branch, subject)
	}

	return nil
}
