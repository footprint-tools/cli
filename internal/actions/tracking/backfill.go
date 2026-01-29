package tracking

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/usage"
)

// Backfill imports historical commits from a git repository into the database.
func Backfill(args []string, flags *dispatchers.ParsedFlags) error {
	return backfill(args, flags, DefaultDeps())
}

func backfill(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	jsonOutput := flags.Has("--json")

	if flags.Has("--dry-run") {
		if jsonOutput {
			return doBackfillDryRunJSON(args, flags, deps)
		}
		return doBackfillDryRun(args, flags, deps)
	}

	if jsonOutput {
		return doBackfillJSON(args, flags, deps)
	}

	return doBackfillText(args, flags, deps)
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

// doBackfillText performs the backfill and prints text output.
func doBackfillText(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
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

	if len(commits) == 0 {
		_, _ = deps.Println("No commits found to import")
		return nil
	}

	_, _ = deps.Printf("Found %d commits to import...\n", len(commits))

	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	_ = deps.InitDB(db)

	branchOverride := flags.String("--branch", "")

	imported := 0
	skipped := 0
	for _, c := range commits {
		branch := branchOverride
		if branch == "" {
			branch = git.GetBranchForCommit(repoRoot, c.Hash)
			if branch == "" {
				branch = "unknown"
			}
		}

		timestamp, err := time.Parse(time.RFC3339, c.AuthorDate)
		if err != nil {
			timestamp = time.Now().UTC()
		}

		event := store.RepoEvent{
			RepoID:    repoID,
			RepoPath:  repoRoot,
			Commit:    c.Hash,
			Branch:    branch,
			Timestamp: timestamp.UTC(),
			Status:    store.StatusPending,
			Source:    store.SourceBackfill,
		}

		if err := deps.InsertEvent(db, event); err == nil {
			imported++
		} else {
			skipped++
		}
	}

	_, _ = deps.Printf("Imported %d commits (%d skipped)\n", imported, skipped)
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

	_, _ = deps.Printf("Repository: %s\n", repoID)
	_, _ = deps.Printf("Found %d commits to import:\n\n", len(commits))

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

		_, _ = deps.Printf("  %.7s %s %s \"%s\"\n", c.Hash, c.AuthorDate[:10], branch, subject)
	}

	return nil
}

// doBackfillDryRunJSON shows what would be imported as JSON.
func doBackfillDryRunJSON(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
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

	branchOverride := flags.String("--branch", "")

	type commitEntry struct {
		Hash       string `json:"hash"`
		Branch     string `json:"branch"`
		AuthorDate string `json:"author_date"`
		Subject    string `json:"subject"`
	}

	type dryRunResult struct {
		RepoID  string        `json:"repo_id"`
		Path    string        `json:"path"`
		Count   int           `json:"count"`
		Commits []commitEntry `json:"commits"`
	}

	result := dryRunResult{
		RepoID:  repoID,
		Path:    repoRoot,
		Count:   len(commits),
		Commits: make([]commitEntry, 0, len(commits)),
	}

	for _, c := range commits {
		branch := branchOverride
		if branch == "" {
			branch = git.GetBranchForCommit(repoRoot, c.Hash)
			if branch == "" {
				branch = "unknown"
			}
		}

		result.Commits = append(result.Commits, commitEntry{
			Hash:       c.Hash,
			Branch:     branch,
			AuthorDate: c.AuthorDate,
			Subject:    c.Subject,
		})
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	_, _ = deps.Println(string(data))
	return nil
}

// doBackfillJSON runs backfill synchronously and outputs JSON result.
func doBackfillJSON(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
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

	type backfillResult struct {
		RepoID   string `json:"repo_id"`
		Path     string `json:"path"`
		Found    int    `json:"found"`
		Imported int    `json:"imported"`
		Skipped  int    `json:"skipped"`
	}

	result := backfillResult{
		RepoID: repoID,
		Path:   repoRoot,
		Found:  len(commits),
	}

	if len(commits) == 0 {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		_, _ = deps.Println(string(data))
		return nil
	}

	// Open database
	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	_ = deps.InitDB(db)

	branchOverride := flags.String("--branch", "")

	// Insert each commit as an event
	for _, c := range commits {
		branch := branchOverride
		if branch == "" {
			branch = git.GetBranchForCommit(repoRoot, c.Hash)
			if branch == "" {
				branch = "unknown"
			}
		}

		timestamp, err := time.Parse(time.RFC3339, c.AuthorDate)
		if err != nil {
			timestamp = time.Now().UTC()
		}

		event := store.RepoEvent{
			RepoID:    repoID,
			RepoPath:  repoRoot,
			Commit:    c.Hash,
			Branch:    branch,
			Timestamp: timestamp.UTC(),
			Status:    store.StatusPending,
			Source:    store.SourceBackfill,
		}

		if err := deps.InsertEvent(db, event); err == nil {
			result.Imported++
		} else {
			result.Skipped++
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	_, _ = deps.Println(string(data))
	return nil
}
