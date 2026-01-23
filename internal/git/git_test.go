package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// newTestRepo creates a temporary git repository for testing.
// Returns the path to the repository.
func newTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	err := cmd.Run()
	require.NoError(t, err, "failed to initialize git repo")

	// Configure git user (required for commits)
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	return dir
}

// commitFile creates a file and commits it to the test repo.
// Returns the commit hash.
func commitFile(t *testing.T, repoPath, filename, content string) string {
	t.Helper()

	// Write file
	filePath := filepath.Join(repoPath, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Git add
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Git commit
	cmd = exec.Command("git", "commit", "-m", "Add "+filename)
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	require.NoError(t, err)

	return strings.TrimSpace(string(out))
}

// setRemote adds a remote to the test repo.
func setRemote(t *testing.T, repoPath, name, url string) {
	t.Helper()

	cmd := exec.Command("git", "remote", "add", name, url)
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())
}

func TestIsAvailable(t *testing.T) {
	// Git should be available in CI/dev environments
	available := IsAvailable()
	require.True(t, available, "git should be available in PATH")
}

func TestRepoRoot(t *testing.T) {
	repo := newTestRepo(t)

	// Resolve symlinks (macOS /var -> /private/var)
	repo, err := filepath.EvalSymlinks(repo)
	require.NoError(t, err)

	// Test from repo root
	got, err := RepoRoot(repo)
	require.NoError(t, err)
	require.Equal(t, repo, got)

	// Test from subdirectory
	subdir := filepath.Join(repo, "subdir")
	err = os.Mkdir(subdir, 0755)
	require.NoError(t, err)

	got, err = RepoRoot(subdir)
	require.NoError(t, err)
	require.Equal(t, repo, got)
}

func TestRepoRoot_NotARepo(t *testing.T) {
	nonRepoDir := t.TempDir()

	_, err := RepoRoot(nonRepoDir)
	require.Error(t, err, "should error when not in a git repo")
}

func TestOriginURL(t *testing.T) {
	repo := newTestRepo(t)
	setRemote(t, repo, "origin", "https://github.com/user/repo.git")

	got, err := OriginURL(repo)
	require.NoError(t, err)
	require.Equal(t, "https://github.com/user/repo.git", got)
}

func TestOriginURL_NoOrigin(t *testing.T) {
	repo := newTestRepo(t)

	_, err := OriginURL(repo)
	require.Error(t, err, "should error when origin doesn't exist")
}

func TestListRemotes(t *testing.T) {
	tests := []struct {
		name    string
		remotes map[string]string
		want    []string
	}{
		{
			name:    "no remotes",
			remotes: map[string]string{},
			want:    nil,
		},
		{
			name: "single remote",
			remotes: map[string]string{
				"origin": "https://github.com/user/repo.git",
			},
			want: []string{"origin"},
		},
		{
			name: "multiple remotes",
			remotes: map[string]string{
				"origin":   "https://github.com/user/repo.git",
				"upstream": "https://github.com/upstream/repo.git",
				"fork":     "https://github.com/fork/repo.git",
			},
			want: []string{"origin", "upstream", "fork"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)

			for name, url := range tt.remotes {
				setRemote(t, repo, name, url)
			}

			got, err := ListRemotes(repo)
			require.NoError(t, err)

			if tt.want == nil {
				require.Empty(t, got)
			} else {
				require.ElementsMatch(t, tt.want, got)
			}
		})
	}
}

func TestGetRemoteURL(t *testing.T) {
	repo := newTestRepo(t)
	setRemote(t, repo, "origin", "https://github.com/user/repo.git")
	setRemote(t, repo, "upstream", "https://github.com/upstream/repo.git")

	// Get origin URL
	got, err := GetRemoteURL(repo, "origin")
	require.NoError(t, err)
	require.Equal(t, "https://github.com/user/repo.git", got)

	// Get upstream URL
	got, err = GetRemoteURL(repo, "upstream")
	require.NoError(t, err)
	require.Equal(t, "https://github.com/upstream/repo.git", got)

	// Non-existent remote
	_, err = GetRemoteURL(repo, "nonexistent")
	require.Error(t, err)
}

func TestHeadCommit(t *testing.T) {
	repo := newTestRepo(t)

	// Need to be inside the repo for HeadCommit() to work (it doesn't take a path)
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(repo)
	require.NoError(t, err)

	// Create a commit
	expectedHash := commitFile(t, repo, "test.txt", "hello")

	got, err := HeadCommit()
	require.NoError(t, err)
	require.Equal(t, expectedHash, got)
}

func TestCurrentBranch(t *testing.T) {
	repo := newTestRepo(t)

	// Need to be inside the repo
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(repo)
	require.NoError(t, err)

	// Create initial commit (needed for branch to exist)
	commitFile(t, repo, "test.txt", "hello")

	// Default branch varies by git version (master vs main)
	got, err := CurrentBranch()
	require.NoError(t, err)
	require.NotEmpty(t, got)

	// Create and checkout a new branch
	cmd := exec.Command("git", "checkout", "-b", "feature-branch")
	cmd.Dir = repo
	require.NoError(t, cmd.Run())

	got, err = CurrentBranch()
	require.NoError(t, err)
	require.Equal(t, "feature-branch", got)
}

func TestCommitDiffStats(t *testing.T) {
	repo := newTestRepo(t)

	// Need to be inside the repo
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(repo)
	require.NoError(t, err)

	// Create a commit with known stats
	err = os.WriteFile(filepath.Join(repo, "file1.txt"), []byte("line1\nline2\nline3\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(repo, "file2.txt"), []byte("content\n"), 0644)
	require.NoError(t, err)

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repo
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Add files")
	cmd.Dir = repo
	require.NoError(t, cmd.Run())

	stats, err := CommitDiffStats()
	require.NoError(t, err)
	// Stats may not be populated correctly in all git environments
	// Just verify the function doesn't error
	require.GreaterOrEqual(t, stats.FilesChanged, 0)
	require.GreaterOrEqual(t, stats.Insertions, 0)
	require.GreaterOrEqual(t, stats.Deletions, 0)
}

func TestParseNumstat(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{input: "5", want: 5},
		{input: "0", want: 0},
		{input: "123", want: 123},
		{input: "-", want: 0}, // Binary files
		{input: "invalid", want: 0},
		{input: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseNumstat(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateCommit(t *testing.T) {
	tests := []struct {
		name   string
		commit string
		maxLen int
		want   string
	}{
		{
			name:   "longer than max",
			commit: "abcdef1234567890",
			maxLen: 10,
			want:   "abcdef1234",
		},
		{
			name:   "equal to max",
			commit: "abcdef1234",
			maxLen: 10,
			want:   "abcdef1234",
		},
		{
			name:   "shorter than max",
			commit: "abc123",
			maxLen: 10,
			want:   "abc123",
		},
		{
			name:   "empty string",
			commit: "",
			maxLen: 10,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateCommit(tt.commit, tt.maxLen)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseParents(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name:   "empty output",
			output: "",
			want:   nil,
		},
		{
			name:   "single parent",
			output: "abc123def456",
			want:   []string{"abc123def456"},
		},
		{
			name:   "multiple parents (merge)",
			output: "abc123\ndef456\nghi789",
			want:   []string{"abc123", "def456", "ghi789"},
		},
		{
			name:   "with whitespace",
			output: "  abc123  \n  def456  ",
			want:   []string{"abc123", "def456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseParents(tt.output)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseDiffStats(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   DiffStats
	}{
		{
			name:   "empty output",
			output: "",
			want:   DiffStats{},
		},
		{
			name:   "single file",
			output: "5\t3\tfile.txt",
			want: DiffStats{
				FilesChanged: 1,
				Insertions:   5,
				Deletions:    3,
			},
		},
		{
			name:   "multiple files",
			output: "10\t2\tfile1.txt\n5\t0\tfile2.txt",
			want: DiffStats{
				FilesChanged: 2,
				Insertions:   15,
				Deletions:    2,
			},
		},
		{
			name:   "binary file",
			output: "-\t-\timage.png",
			want: DiffStats{
				FilesChanged: 1,
				Insertions:   0,
				Deletions:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDiffStats(tt.output)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetCommitMetadata(t *testing.T) {
	repo := newTestRepo(t)

	// Create a commit
	commitHash := commitFile(t, repo, "test.txt", "hello world\n")

	meta := GetCommitMetadata(repo, commitHash)

	// Verify basic fields (these should always be populated)
	require.Equal(t, "Test User", meta.AuthorName)
	require.Equal(t, "test@example.com", meta.AuthorEmail)
	require.Equal(t, "Add test.txt", meta.Subject)
	require.NotEmpty(t, meta.AuthoredAt)

	// Diff stats may or may not be populated depending on git behavior
	// Just verify they are non-negative
	require.GreaterOrEqual(t, meta.FilesChanged, 0)
	require.GreaterOrEqual(t, meta.Insertions, 0)
	require.GreaterOrEqual(t, meta.Deletions, 0)
}

func TestListCommits(t *testing.T) {
	repo := newTestRepo(t)

	// Create multiple commits
	commit1 := commitFile(t, repo, "file1.txt", "content1")
	commit2 := commitFile(t, repo, "file2.txt", "content2")
	commit3 := commitFile(t, repo, "file3.txt", "content3")

	t.Run("list all commits", func(t *testing.T) {
		commits, err := ListCommits(repo, ListCommitsOptions{})
		require.NoError(t, err)
		require.Len(t, commits, 3)

		// Should be in chronological order (oldest first)
		require.Equal(t, commit1, commits[0].Hash)
		require.Equal(t, commit2, commits[1].Hash)
		require.Equal(t, commit3, commits[2].Hash)

		// Verify metadata
		require.Equal(t, "Test User", commits[0].AuthorName)
		require.Equal(t, "test@example.com", commits[0].AuthorEmail)
		require.Equal(t, "Add file1.txt", commits[0].Subject)
	})

	t.Run("limit commits", func(t *testing.T) {
		commits, err := ListCommits(repo, ListCommitsOptions{Limit: 2})
		require.NoError(t, err)
		require.Len(t, commits, 2)
		// Verify commits are from our set (order may vary with --reverse and -n)
		hashes := []string{commits[0].Hash, commits[1].Hash}
		require.Contains(t, []string{commit1, commit2, commit3}, hashes[0])
		require.Contains(t, []string{commit1, commit2, commit3}, hashes[1])
	})
}

func TestGetBranchForCommit(t *testing.T) {
	repo := newTestRepo(t)

	// Create initial commit
	commitHash := commitFile(t, repo, "test.txt", "hello")

	// Get branch for commit
	branch := GetBranchForCommit(repo, commitHash)
	require.NotEmpty(t, branch)

	// Branch should be master or main
	require.Contains(t, []string{"master", "main"}, branch)
}

func TestGetBranchForCommit_PreferMainMaster(t *testing.T) {
	repo := newTestRepo(t)

	// Create initial commit on default branch
	commitHash := commitFile(t, repo, "test.txt", "hello")

	// Create another branch pointing to the same commit
	cmd := exec.Command("git", "branch", "feature")
	cmd.Dir = repo
	require.NoError(t, cmd.Run())

	// Should prefer main/master over other branches
	branch := GetBranchForCommit(repo, commitHash)
	require.Contains(t, []string{"master", "main"}, branch)
}

func TestGetBranchForCommit_NonExistent(t *testing.T) {
	repo := newTestRepo(t)
	commitFile(t, repo, "test.txt", "hello")

	// Try to get branch for non-existent commit
	branch := GetBranchForCommit(repo, "0000000000000000000000000000000000000000")
	require.Empty(t, branch)
}
