package git

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/footprint-tools/footprint-cli/internal/log"
)

// dateArgPattern validates git date arguments to prevent injection
// Accepts: ISO dates, relative dates (e.g., "2 weeks ago"), and common formats
var dateArgPattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-:+/.]+$`)

// commitHashPattern validates git commit hashes (SHA-1 or SHA-256)
// SHA-1: 40 hex characters, SHA-256: 64 hex characters
// Also accepts short hashes (7+ characters) and HEAD/branch references
var commitHashPattern = regexp.MustCompile(`^[a-fA-F0-9]{7,64}$|^HEAD$|^[a-zA-Z0-9_\-./]+$`)

// isValidCommitRef checks if a string looks like a valid git commit reference
func isValidCommitRef(ref string) bool {
	if ref == "" {
		return false
	}
	// Reject obvious injection attempts
	if strings.ContainsAny(ref, ";&|`$(){}[]<>\\\"'") {
		return false
	}
	return commitHashPattern.MatchString(ref)
}

// DiffStats contains statistics from a git diff.
type DiffStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

func parseNumstat(v string) int {
	if v == "-" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func IsAvailable() bool {
	path, err := exec.LookPath("git")
	if err != nil {
		return false
	}
	// Verify git is functional by running a simple command
	cmd := exec.Command(path, "--version")
	return cmd.Run() == nil
}

func RepoRoot(path string) (string, error) {
	return runGit("-C", path, "rev-parse", "--show-toplevel")
}

func OriginURL(repoRoot string) (string, error) {
	return runGit("-C", repoRoot, "remote", "get-url", "origin")
}

// ListRemotes returns all remote names for a repository.
func ListRemotes(repoRoot string) ([]string, error) {
	out, err := runGit("-C", repoRoot, "remote")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return []string{}, nil
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	remotes := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			remotes = append(remotes, line)
		}
	}
	return remotes, nil
}

// GetRemoteURL returns the URL for a specific remote.
func GetRemoteURL(repoRoot, remoteName string) (string, error) {
	return runGit("-C", repoRoot, "remote", "get-url", remoteName)
}

func HeadCommit() (string, error) {
	return runGit("rev-parse", "HEAD")
}

func CommitMessage() (string, error) {
	out, err := exec.Command(
		"git", "show", "-s", "--format=%s",
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func CurrentBranch() (string, error) {
	return runGit("rev-parse", "--abbrev-ref", "HEAD")
}

func CommitAuthor() (string, error) {
	return runGit("show", "-s", "--format=%an <%ae>", "HEAD")
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Debug("git: command failed: git %s: %v", strings.Join(args, " "), err)
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// CommitMetadata contains enriched data from Git for a specific commit.
type CommitMetadata struct {
	AuthoredAt     string // Author date in RFC3339 format
	ParentCommits  string // Space-separated list of parent commit hashes
	AuthorName     string
	AuthorEmail    string
	CommitterName  string
	CommitterEmail string
	Subject        string // Commit message (first line)
	Body           string // Commit message body (after first line)
	FilesChanged   int
	Insertions     int
	Deletions      int
}

// GetCommitMetadata retrieves enriched metadata for a specific commit from a repository.
// repoPath is the path to the repository, commit is the full commit hash.
// Returns empty values (not errors) if the data cannot be retrieved.
func GetCommitMetadata(repoPath, commit string) CommitMetadata {
	meta := CommitMetadata{}

	// Validate commit reference format
	if !isValidCommitRef(commit) {
		log.Warn("git: invalid commit reference format: %s", commit)
		return meta
	}

	// Get parent commits using git rev-list
	if parents, err := runGitInRepo(repoPath, "rev-parse", commit+"^@"); err == nil {
		parentList := parseParents(parents)
		meta.ParentCommits = strings.Join(parentList, " ")
	}

	// Get author, committer info, author date, subject, and body using git show with format
	// Format: author_name%x00author_email%x00author_date_iso%x00committer_name%x00committer_email%x00subject%x00body
	format := "%an%x00%ae%x00%aI%x00%cn%x00%ce%x00%s%x00%b"
	if info, err := runGitInRepo(repoPath, "show", "-s", "--format="+format, commit); err == nil {
		parts := strings.Split(info, "\x00")
		if len(parts) >= 6 {
			meta.AuthorName = parts[0]
			meta.AuthorEmail = parts[1]
			meta.AuthoredAt = parts[2]
			meta.CommitterName = parts[3]
			meta.CommitterEmail = parts[4]
			meta.Subject = parts[5]
			if len(parts) >= 7 {
				meta.Body = strings.TrimSpace(parts[6])
			}
		}
	}

	// Get diff stats using git diff-tree
	if stats, err := runGitInRepo(repoPath, "diff-tree", "--no-commit-id", "--numstat", "-r", commit); err == nil {
		diffStats := parseDiffStats(stats)
		meta.FilesChanged = diffStats.FilesChanged
		meta.Insertions = diffStats.Insertions
		meta.Deletions = diffStats.Deletions
	}

	return meta
}

// runGitInRepo runs a git command in the specified repository directory.
func runGitInRepo(repoPath string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	return runGit(fullArgs...)
}

// truncateCommit returns the first maxLen characters of a commit hash.
func truncateCommit(commit string, maxLen int) string {
	if len(commit) <= maxLen {
		return commit
	}
	return commit[:maxLen]
}

// parseParents parses the output of git rev-parse commit^@ into a slice of parent hashes.
func parseParents(output string) []string {
	if strings.TrimSpace(output) == "" {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	parents := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			parents = append(parents, line)
		}
	}
	return parents
}

// parseDiffStats parses git diff-tree --numstat output into DiffStats.
func parseDiffStats(output string) DiffStats {
	stats := DiffStats{}
	if strings.TrimSpace(output) == "" {
		return stats
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}

		ins := parseNumstat(parts[0])
		del := parseNumstat(parts[1])

		stats.FilesChanged++
		stats.Insertions += ins
		stats.Deletions += del
	}

	return stats
}

// HistoryCommit represents a commit from git log.
type HistoryCommit struct {
	Hash       string
	AuthorName string
	AuthorEmail string
	AuthorDate string // ISO 8601 format
	Subject    string
}

// ListCommitsOptions configures the ListCommits query.
type ListCommitsOptions struct {
	Since string // Date filter: commits after this date
	Until string // Date filter: commits before this date
	Limit int    // Max number of commits (0 = unlimited)
}

// ListCommits returns commits from git log in chronological order (oldest first).
// repoPath is the path to the repository.
func ListCommits(repoPath string, opts ListCommitsOptions) ([]HistoryCommit, error) {
	// Format: hash%x00author_name%x00author_email%x00author_date_iso%x00subject
	format := "%H%x00%an%x00%ae%x00%aI%x00%s"

	args := []string{"-C", repoPath, "log", "--format=" + format, "--reverse"}

	if opts.Since != "" && dateArgPattern.MatchString(opts.Since) {
		args = append(args, "--since="+opts.Since)
	}
	if opts.Until != "" && dateArgPattern.MatchString(opts.Until) {
		args = append(args, "--until="+opts.Until)
	}
	if opts.Limit > 0 {
		args = append(args, "-n", strconv.Itoa(opts.Limit))
	}

	out, err := runGit(args...)
	if err != nil {
		log.Error("git: list commits failed: %v", err)
		return nil, err
	}

	if strings.TrimSpace(out) == "" {
		return []HistoryCommit{}, nil
	}

	lines := strings.Split(out, "\n")
	var commits []HistoryCommit

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\x00", 5)
		if len(parts) < 5 {
			log.Warn("git: skipping malformed commit line (expected 5 fields, got %d)", len(parts))
			continue
		}

		commits = append(commits, HistoryCommit{
			Hash:        parts[0],
			AuthorName:  parts[1],
			AuthorEmail: parts[2],
			AuthorDate:  parts[3],
			Subject:     parts[4],
		})
	}

	log.Debug("git: listed %d commits", len(commits))
	return commits, nil
}

// GetBranchForCommit tries to infer which branch a commit belongs to.
// Returns the branch name or empty string if unable to determine.
func GetBranchForCommit(repoPath, commit string) string {
	// Get all branches that contain this commit
	out, err := runGitInRepo(repoPath, "branch", "--contains", commit, "--format=%(refname:short)")
	if err != nil || strings.TrimSpace(out) == "" {
		return ""
	}

	branches := strings.Split(strings.TrimSpace(out), "\n")
	if len(branches) == 0 {
		return ""
	}

	// If only one branch, use it
	if len(branches) == 1 {
		return strings.TrimSpace(branches[0])
	}

	// Prefer main/master if present
	for _, b := range branches {
		b = strings.TrimSpace(b)
		if b == "main" || b == "master" {
			return b
		}
	}

	// Return the first branch
	return strings.TrimSpace(branches[0])
}
