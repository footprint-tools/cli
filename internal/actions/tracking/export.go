package tracking

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Skryensya/footprint/internal/config"
	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/git"
	repodomain "github.com/Skryensya/footprint/internal/repo"
	"github.com/Skryensya/footprint/internal/store"
)

const (
	// CSV rotation thresholds
	maxCSVSizeBytes = 50 * 1024 * 1024 // 50 MB
	maxCSVRowCount  = 200000           // 200k rows

	// CSV filename constants
	activeCSVName = "commits.csv"
)

// CSV header for enriched export
var csvHeader = []string{
	"timestamp",
	"repo_id",
	"commit",
	"commit_short",
	"branch",
	"parent_commits",
	"is_merge",
	"author_name",
	"author_email",
	"committer_name",
	"committer_email",
	"files_changed",
	"insertions",
	"deletions",
	"message",
	"source",
}

// Export handles the manual `fp export` command.
func Export(args []string, flags *dispatchers.ParsedFlags) error {
	return export(args, flags, DefaultDeps())
}

func export(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	force := flags.Has("--force")
	dryRun := flags.Has("--dry-run")
	openDir := flags.Has("--open")
	setRemote := flags.String("--set-remote", "")

	exportRepo := getExportRepo()

	// Handle --open flag
	if openDir {
		return openInFileManager(exportRepo)
	}

	// Handle --set-remote flag
	if setRemote != "" {
		if err := ensureExportRepo(exportRepo); err != nil {
			return fmt.Errorf("could not initialize export repo: %w", err)
		}
		if err := setExportRemote(exportRepo, setRemote); err != nil {
			return fmt.Errorf("could not set remote: %w", err)
		}
		deps.Printf("Remote set to: %s\n", setRemote)
		return nil
	}

	db, err := deps.OpenDB(deps.DBPath())
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	_ = deps.InitDB(db)

	events, err := store.GetPendingEvents(db)
	if err != nil {
		return fmt.Errorf("could not get pending events: %w", err)
	}

	if len(events) == 0 {
		deps.Println("No pending events to export")
		return nil
	}

	if dryRun {
		deps.Printf("Would export %d events:\n", len(events))
		for _, e := range events {
			deps.Printf("  %.7s %s (%s)\n", e.Commit, e.Branch, e.RepoID)
		}
		return nil
	}

	if !force {
		shouldExp, err := shouldExport(deps)
		if err != nil {
			return err
		}
		if !shouldExp {
			deps.Println("Export interval not reached. Use --force to export anyway.")
			return nil
		}
	}

	if err := ensureExportRepo(exportRepo); err != nil {
		return fmt.Errorf("could not initialize export repo: %w", err)
	}

	// Group events by repo_id
	eventsByRepo := groupEventsByRepo(events)

	// Export each repo's events
	var exportedIDs []int64
	var exportedFiles []string

	for repoID, repoEvents := range eventsByRepo {
		ids, files, err := exportRepoEvents(exportRepo, repoID, repoEvents, deps)
		if err != nil {
			// Log error but continue with other repos
			deps.Printf("Warning: failed to export events for %s: %v\n", repoID, err)
			continue
		}
		exportedIDs = append(exportedIDs, ids...)
		exportedFiles = append(exportedFiles, files...)
	}

	if len(exportedFiles) == 0 {
		deps.Println("No events were exported")
		return nil
	}

	// Commit all changes to the export repo
	if err := commitExportChanges(exportRepo, exportedFiles); err != nil {
		return fmt.Errorf("could not commit export: %w", err)
	}

	// Update event statuses
	if err := store.UpdateEventStatuses(db, exportedIDs, store.StatusExported); err != nil {
		return fmt.Errorf("could not update event statuses: %w", err)
	}

	if err := saveExportLast(deps.Now().Unix()); err != nil {
		return fmt.Errorf("could not save export timestamp: %w", err)
	}

	deps.Printf("Exported %d events to %d files\n", len(exportedIDs), len(exportedFiles))

	// Auto-push if remote is configured
	if hasRemote(exportRepo) {
		if err := pushExportRepo(exportRepo); err != nil {
			deps.Printf("Warning: could not push to remote: %v\n", err)
		} else {
			deps.Println("Pushed to remote")
		}
	}

	return nil
}

// MaybeExport checks if it's time to export and does so if needed.
func MaybeExport(db *sql.DB, deps Deps) {
	shouldExp, err := shouldExport(deps)
	if err != nil || !shouldExp {
		return
	}

	events, err := store.GetPendingEvents(db)
	if err != nil || len(events) == 0 {
		return
	}

	exportRepo := getExportRepo()

	if err := ensureExportRepo(exportRepo); err != nil {
		return
	}

	eventsByRepo := groupEventsByRepo(events)

	var exportedIDs []int64
	var exportedFiles []string

	for repoID, repoEvents := range eventsByRepo {
		ids, files, err := exportRepoEvents(exportRepo, repoID, repoEvents, deps)
		if err != nil {
			continue
		}
		exportedIDs = append(exportedIDs, ids...)
		exportedFiles = append(exportedFiles, files...)
	}

	if len(exportedFiles) == 0 {
		return
	}

	if err := commitExportChanges(exportRepo, exportedFiles); err != nil {
		return
	}

	// Auto-push if remote is configured
	if hasRemote(exportRepo) {
		_ = pushExportRepo(exportRepo)
	}

	if err := store.UpdateEventStatuses(db, exportedIDs, store.StatusExported); err != nil {
		return
	}

	_ = saveExportLast(deps.Now().Unix())
}

// groupEventsByRepo groups events by their repo_id.
func groupEventsByRepo(events []store.RepoEvent) map[string][]store.RepoEvent {
	result := make(map[string][]store.RepoEvent)
	for _, e := range events {
		result[e.RepoID] = append(result[e.RepoID], e)
	}
	return result
}

// exportRepoEvents exports events for a single repository.
// Returns the IDs of exported events and the files that were modified.
func exportRepoEvents(exportRepo, repoID string, events []store.RepoEvent, _ Deps) ([]int64, []string, error) {
	// Create per-repo directory
	safeRepoID := repodomain.RepoID(repoID).ToFilesystemSafe()
	repoDir := filepath.Join(exportRepo, "repos", safeRepoID)

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("could not create repo directory: %w", err)
	}

	// Sort events by timestamp (oldest first for append-only)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// Get the repo path from the first event (they all should have the same repo)
	repoPath := ""
	if len(events) > 0 && events[0].RepoPath != "" {
		repoPath = events[0].RepoPath
	}

	var exportedIDs []int64
	var modifiedFiles []string
	modifiedFilesSet := make(map[string]bool)

	for _, e := range events {
		// Enrich event with Git metadata
		var meta git.CommitMetadata
		if repoPath != "" {
			meta = git.GetCommitMetadata(repoPath, e.Commit)
		} else {
			// Fallback: just truncate commit
			meta = git.CommitMetadata{
				CommitShort: truncateCommit(e.Commit, 10),
			}
		}

		// Get or rotate the active CSV file
		csvPath, err := getActiveCSV(repoDir)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get active CSV: %w", err)
		}

		// Append the enriched record
		if err := appendRecord(csvPath, e, meta); err != nil {
			return nil, nil, fmt.Errorf("could not append record: %w", err)
		}

		exportedIDs = append(exportedIDs, e.ID)

		// Track modified files (relative to export repo)
		relPath, _ := filepath.Rel(exportRepo, csvPath)
		if !modifiedFilesSet[relPath] {
			modifiedFilesSet[relPath] = true
			modifiedFiles = append(modifiedFiles, relPath)
		}
	}

	return exportedIDs, modifiedFiles, nil
}

// getActiveCSV returns the path to the active CSV file, performing rotation if needed.
func getActiveCSV(repoDir string) (string, error) {
	activePath := filepath.Join(repoDir, activeCSVName)

	// Check if the active CSV exists
	info, err := os.Stat(activePath)
	if os.IsNotExist(err) {
		// Create new CSV with header
		if err := writeCSVHeader(activePath); err != nil {
			return "", err
		}
		return activePath, nil
	}
	if err != nil {
		return "", err
	}

	// Check if rotation is needed
	needsRotation := false

	// Check size
	if info.Size() >= maxCSVSizeBytes {
		needsRotation = true
	}

	// Check row count
	if !needsRotation {
		rowCount, err := countCSVRows(activePath)
		if err == nil && rowCount >= maxCSVRowCount {
			needsRotation = true
		}
	}

	if needsRotation {
		if err := rotateCSV(repoDir); err != nil {
			return "", err
		}
		// Create new active CSV
		if err := writeCSVHeader(activePath); err != nil {
			return "", err
		}
	}

	return activePath, nil
}

// countCSVRows counts the number of data rows in a CSV file (excluding header).
func countCSVRows(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}

	// Subtract 1 for header row
	if count > 0 {
		count--
	}
	return count, nil
}

// rotateCSV renames the active CSV to the next numbered suffix.
func rotateCSV(repoDir string) error {
	activePath := filepath.Join(repoDir, activeCSVName)

	// Find the next available suffix
	nextSuffix := findNextCSVSuffix(repoDir)
	rotatedName := fmt.Sprintf("commits-%04d.csv", nextSuffix)
	rotatedPath := filepath.Join(repoDir, rotatedName)

	return os.Rename(activePath, rotatedPath)
}

// findNextCSVSuffix finds the next available numeric suffix for rotated CSVs.
func findNextCSVSuffix(repoDir string) int {
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return 1
	}

	maxSuffix := 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "commits-") && strings.HasSuffix(name, ".csv") {
			// Extract the number from commits-NNNN.csv
			numStr := strings.TrimPrefix(name, "commits-")
			numStr = strings.TrimSuffix(numStr, ".csv")
			if num, err := strconv.Atoi(numStr); err == nil && num > maxSuffix {
				maxSuffix = num
			}
		}
	}

	return maxSuffix + 1
}

// writeCSVHeader writes a new CSV file with just the header row.
func writeCSVHeader(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	if err := w.Write(csvHeader); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// appendRecord appends a single enriched record to a CSV file.
func appendRecord(path string, e store.RepoEvent, meta git.CommitMetadata) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)

	// Sanitize commit message: single line, no newlines
	msg := strings.ReplaceAll(e.CommitMessage, "\n", " ")
	msg = strings.ReplaceAll(msg, "\r", "")
	msg = strings.TrimSpace(msg)

	// Build the record
	record := []string{
		e.Timestamp.UTC().Format(time.RFC3339), // timestamp
		e.RepoID,                               // repo_id
		e.Commit,                               // commit (full hash)
		meta.CommitShort,                       // commit_short (10 chars)
		e.Branch,                               // branch
		meta.ParentCommits,                     // parent_commits
		strconv.FormatBool(meta.IsMerge),       // is_merge
		meta.AuthorName,                        // author_name
		meta.AuthorEmail,                       // author_email
		meta.CommitterName,                     // committer_name
		meta.CommitterEmail,                    // committer_email
		strconv.Itoa(meta.FilesChanged),        // files_changed
		strconv.Itoa(meta.Insertions),          // insertions
		strconv.Itoa(meta.Deletions),           // deletions
		msg,                                    // message
		e.Source.String(),                      // source
	}

	if err := w.Write(record); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// truncateCommit returns the first n characters of a commit hash.
func truncateCommit(commit string, n int) string {
	if len(commit) <= n {
		return commit
	}
	return commit[:n]
}

func shouldExport(deps Deps) (bool, error) {
	intervalStr, _ := config.Get("export_interval")
	lastExportStr, _ := config.Get("export_last")

	interval, _ := strconv.Atoi(intervalStr)
	lastExport, _ := strconv.ParseInt(lastExportStr, 10, 64)

	now := deps.Now().Unix()
	return (now - lastExport) >= int64(interval), nil
}

func getExportRepo() string {
	value, _ := config.Get("export_repo")
	return value
}

func ensureExportRepo(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := runGitInDir(path, "init"); err != nil {
			return fmt.Errorf("git init failed: %w", err)
		}
	}

	return nil
}

// commitExportChanges commits all modified files to the export repo.
func commitExportChanges(exportRepo string, files []string) error {
	if len(files) == 0 {
		return nil
	}

	// Add all modified files
	for _, file := range files {
		if err := runGitInDir(exportRepo, "add", file); err != nil {
			return fmt.Errorf("git add failed for %s: %w", file, err)
		}
	}

	// Check if there are changes to commit
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = exportRepo
	if err := cmd.Run(); err == nil {
		// No changes staged, nothing to commit
		return nil
	}

	// Commit with a descriptive message
	msg := fmt.Sprintf("Export %d files", len(files))
	if err := runGitInDir(exportRepo, "commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}

func saveExportLast(timestamp int64) error {
	lines, err := config.ReadLines()
	if err != nil {
		return err
	}

	lines, _ = config.Set(lines, "export_last", strconv.FormatInt(timestamp, 10))
	return config.WriteLines(lines)
}

func runGitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

// openInFileManager opens a directory in the system's file manager.
func openInFileManager(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Run()
}

// setExportRemote sets the remote URL for the export repository.
func setExportRemote(exportRepo, remoteURL string) error {
	// Check if origin already exists
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = exportRepo
	if err := cmd.Run(); err == nil {
		// Origin exists, update it
		return runGitInDir(exportRepo, "remote", "set-url", "origin", remoteURL)
	}
	// Origin doesn't exist, add it
	return runGitInDir(exportRepo, "remote", "add", "origin", remoteURL)
}

// hasRemote checks if the export repository has a remote configured.
func hasRemote(exportRepo string) bool {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = exportRepo
	return cmd.Run() == nil
}

// pushExportRepo pushes the export repository to its remote.
func pushExportRepo(exportRepo string) error {
	// Push to origin, set upstream if needed
	cmd := exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = exportRepo
	return cmd.Run()
}
