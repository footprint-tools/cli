package tracking

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/store"
	"github.com/stretchr/testify/require"
)

func TestTruncateCommit_ShortCommit(t *testing.T) {
	result := truncateCommit("abc123", 10)
	require.Equal(t, "abc123", result)
}

func TestTruncateCommit_LongCommit(t *testing.T) {
	commit := "abc123def456789"
	result := truncateCommit(commit, 10)
	require.Equal(t, "abc123def4", result)
}

func TestTruncateCommit_ExactLength(t *testing.T) {
	commit := "1234567890"
	result := truncateCommit(commit, 10)
	require.Equal(t, "1234567890", result)
}

func TestGroupEventsByRepo_SingleRepo(t *testing.T) {
	events := []store.RepoEvent{
		{RepoID: "repo1", Commit: "abc"},
		{RepoID: "repo1", Commit: "def"},
	}

	result := groupEventsByRepo(events)

	require.Len(t, result, 1)
	require.Len(t, result["repo1"], 2)
}

func TestGroupEventsByRepo_MultipleRepos(t *testing.T) {
	events := []store.RepoEvent{
		{RepoID: "repo1", Commit: "abc"},
		{RepoID: "repo2", Commit: "def"},
		{RepoID: "repo1", Commit: "ghi"},
	}

	result := groupEventsByRepo(events)

	require.Len(t, result, 2)
	require.Len(t, result["repo1"], 2)
	require.Len(t, result["repo2"], 1)
}

func TestGroupEventsByRepo_Empty(t *testing.T) {
	result := groupEventsByRepo(nil)

	require.Empty(t, result)
}

func TestFindNextCSVSuffix_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	result := findNextCSVSuffix(dir)

	require.Equal(t, 1, result)
}

func TestFindNextCSVSuffix_WithExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Create some numbered CSV files
	os.WriteFile(filepath.Join(dir, "commits-0001.csv"), []byte(""), 0600)
	os.WriteFile(filepath.Join(dir, "commits-0003.csv"), []byte(""), 0600)

	result := findNextCSVSuffix(dir)

	require.Equal(t, 4, result, "Should return max+1")
}

func TestFindNextCSVSuffix_IgnoresActiveCSV(t *testing.T) {
	dir := t.TempDir()

	// Create commits.csv (active file)
	os.WriteFile(filepath.Join(dir, "commits.csv"), []byte(""), 0600)

	result := findNextCSVSuffix(dir)

	require.Equal(t, 1, result, "Should ignore commits.csv")
}

func TestFindNextCSVSuffix_IgnoresOtherFiles(t *testing.T) {
	dir := t.TempDir()

	// Create non-matching files
	os.WriteFile(filepath.Join(dir, "other.txt"), []byte(""), 0600)
	os.WriteFile(filepath.Join(dir, "commits.txt"), []byte(""), 0600)

	result := findNextCSVSuffix(dir)

	require.Equal(t, 1, result)
}

func TestCountCSVRows_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	// Write empty file
	os.WriteFile(path, []byte(""), 0600)

	count, err := countCSVRows(path)

	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestCountCSVRows_HeaderOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	// Write header only
	os.WriteFile(path, []byte("col1,col2,col3\n"), 0600)

	count, err := countCSVRows(path)

	require.NoError(t, err)
	require.Equal(t, 0, count, "Header should not count as data row")
}

func TestCountCSVRows_WithData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	content := "col1,col2\ndata1,data2\ndata3,data4\n"
	os.WriteFile(path, []byte(content), 0600)

	count, err := countCSVRows(path)

	require.NoError(t, err)
	require.Equal(t, 2, count, "Should count only data rows")
}

func TestCountCSVRows_FileNotExists(t *testing.T) {
	_, err := countCSVRows("/nonexistent/path/file.csv")

	require.Error(t, err)
}

func TestWriteCSVHeader_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	err := writeCSVHeader(path)

	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestWriteCSVHeader_ContainsHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	err := writeCSVHeader(path)
	require.NoError(t, err)

	// Read file and verify header
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	record, err := reader.Read()
	require.NoError(t, err)

	// Check that header contains expected columns
	require.Contains(t, record, "timestamp")
	require.Contains(t, record, "repo_id")
	require.Contains(t, record, "commit")
	require.Contains(t, record, "branch")
}

func TestWriteCSVHeader_HasRestrictivePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	err := writeCSVHeader(path)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)

	// Check permissions (0600 = rw-------)
	perm := info.Mode().Perm()
	require.Equal(t, os.FileMode(0600), perm)
}

func TestAppendRecord_AppendsData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	// First create file with header
	err := writeCSVHeader(path)
	require.NoError(t, err)

	// Create test event
	event := store.RepoEvent{
		ID:        1,
		RepoID:    "github.com/user/repo",
		Commit:    "abc123def456",
		Branch:    "main",
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Status:    store.StatusPending,
		Source:    store.SourcePostCommit,
	}

	meta := git.CommitMetadata{
		CommitShort:    "abc123def4",
		ParentCommits:  "parent1",
		IsMerge:        false,
		AuthorName:     "John Doe",
		AuthorEmail:    "john@example.com",
		CommitterName:  "John Doe",
		CommitterEmail: "john@example.com",
		FilesChanged:   3,
		Insertions:     10,
		Deletions:      5,
		Subject:        "Fix bug",
	}

	// Append record
	err = appendRecord(path, event, meta)
	require.NoError(t, err)

	// Read and verify
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, records, 2, "Should have header + 1 data row")

	dataRow := records[1]
	require.Contains(t, dataRow[0], "2024-01-15", "Should have timestamp")
	require.Equal(t, "github.com/user/repo", dataRow[1])
	require.Equal(t, "abc123def456", dataRow[2])
	require.Equal(t, "main", dataRow[4])
}

func TestAppendRecord_SanitizesNewlines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	err := writeCSVHeader(path)
	require.NoError(t, err)

	event := store.RepoEvent{
		Timestamp: time.Now().UTC(),
	}

	meta := git.CommitMetadata{
		Subject: "Line 1\nLine 2\rLine 3",
	}

	err = appendRecord(path, event, meta)
	require.NoError(t, err)

	// Read and verify no newlines in message field
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	// The CSV should only have 2 lines: header + data
	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	require.Equal(t, 2, lines, "Should have exactly 2 lines (header + data)")
}

func TestShouldExport_ReturnsNoError(t *testing.T) {
	deps := Deps{
		Now: func() time.Time {
			return time.Now()
		},
	}

	// Test that function doesn't error, regardless of config state
	_, err := shouldExport(deps)

	require.NoError(t, err)
}

func TestRotateCSV_RenamesFile(t *testing.T) {
	dir := t.TempDir()
	activePath := filepath.Join(dir, "commits.csv")

	// Create active CSV
	os.WriteFile(activePath, []byte("test content"), 0600)

	err := rotateCSV(dir)

	require.NoError(t, err)

	// Check that original file is gone
	_, err = os.Stat(activePath)
	require.True(t, os.IsNotExist(err))

	// Check that rotated file exists
	rotatedPath := filepath.Join(dir, "commits-0001.csv")
	_, err = os.Stat(rotatedPath)
	require.NoError(t, err)
}

func TestGetActiveCSV_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()

	path, err := getActiveCSV(dir)

	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "commits.csv"), path)

	// Verify file exists
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestGetActiveCSV_ReturnsExisting(t *testing.T) {
	dir := t.TempDir()

	// Create existing CSV with header
	activePath := filepath.Join(dir, "commits.csv")
	writeCSVHeader(activePath)

	path, err := getActiveCSV(dir)

	require.NoError(t, err)
	require.Equal(t, activePath, path)
}

func TestEnsureExportRepo_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")

	err := ensureExportRepo(exportDir)

	require.NoError(t, err)

	// Check directory exists
	info, err := os.Stat(exportDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())

	// Check git was initialized
	gitDir := filepath.Join(exportDir, ".git")
	_, err = os.Stat(gitDir)
	require.NoError(t, err)
}

func TestCommitExportChanges_EmptyFiles(t *testing.T) {
	dir := t.TempDir()

	// Should not error with empty file list
	err := commitExportChanges(dir, nil)

	require.NoError(t, err)
}
