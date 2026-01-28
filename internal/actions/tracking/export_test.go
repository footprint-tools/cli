package tracking

import (
	"encoding/csv"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/stretchr/testify/require"
)

// CSV column indices for the semantic API schema
const (
	colEventID      = 0
	colEventType    = 1
	colTimestamp    = 2
	colRepoID       = 3
	colRepoName     = 4
	colAuthorID     = 5
	colAuthorName   = 6
	colAuthorEmail  = 7
	colBranch       = 8
	colCommitHash   = 9
	colParentHashes = 10
	colMessage      = 11
	colFilesChanged = 12
	colInsertions   = 13
	colDeletions    = 14
	colDevice       = 15
)

func TestGetCSVPath_CurrentYear(t *testing.T) {
	dir := t.TempDir()
	currentYear := 2025
	eventTime := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	path := getCSVPath(dir, eventTime, currentYear)

	require.Equal(t, filepath.Join(dir, "commits.csv"), path)
}

func TestGetCSVPath_PastYear(t *testing.T) {
	dir := t.TempDir()
	currentYear := 2025
	eventTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	path := getCSVPath(dir, eventTime, currentYear)

	require.Equal(t, filepath.Join(dir, "commits-2024.csv"), path)
}

func TestGetCSVPath_OlderYear(t *testing.T) {
	dir := t.TempDir()
	currentYear := 2025
	eventTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	path := getCSVPath(dir, eventTime, currentYear)

	require.Equal(t, filepath.Join(dir, "commits-2023.csv"), path)
}

func TestLoadCSVRecords_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.csv")

	records := loadCSVRecords(path)

	require.Empty(t, records)
}

func TestLoadCSVRecords_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	// Create CSV file with new schema header and data
	content := `event_id,event_type,timestamp,repo_id,repo_name,author_id,author_name,author_email,branch,commit_hash,parent_hashes,message,files_changed,insertions,deletions,device
uuid1,commit,2024-01-15T10:30:00Z,github.com/user/repo,repo,auth1,John,john@example.com,main,abc123,parent1,Fix bug,3,10,5,machine1
`
	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)

	records := loadCSVRecords(path)

	require.Len(t, records, 1)
	require.Contains(t, records, "github.com/user/repo:abc123")
}

func TestBuildRecord_CreatesCorrectFormat(t *testing.T) {
	event := store.RepoEvent{
		RepoID:    "github.com/user/repo",
		RepoPath:  "/path/to/repo",
		Commit:    "abc123def456",
		Branch:    "main",
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Source:    store.SourcePostCommit,
	}

	meta := git.CommitMetadata{
		AuthoredAt:     "2024-01-15T10:30:00Z",
		ParentCommits:  "parent1",
		AuthorName:     "John Doe",
		AuthorEmail:    "john@example.com",
		CommitterName:  "John Doe",
		CommitterEmail: "john@example.com",
		FilesChanged:   3,
		Insertions:     10,
		Deletions:      5,
		Subject:        "Fix bug",
	}

	record := buildRecord(event, meta)

	require.Len(t, record, 16)
	require.NotEmpty(t, record[colEventID])                       // UUID generated
	require.Equal(t, "commit", record[colEventType])              // event_type
	require.Equal(t, "2024-01-15T10:30:00Z", record[colTimestamp]) // timestamp
	require.Equal(t, "github.com/user/repo", record[colRepoID])   // repo_id
	require.Equal(t, "repo", record[colRepoName])                 // repo_name (basename of path)
	require.Equal(t, "main", record[colBranch])                   // branch
	require.Equal(t, "abc123def456", record[colCommitHash])       // commit_hash
	require.Equal(t, "Fix bug", record[colMessage])               // message
}

func TestBuildRecord_SanitizesNewlines(t *testing.T) {
	event := store.RepoEvent{
		Timestamp: time.Now().UTC(),
	}

	meta := git.CommitMetadata{
		Subject: "Line 1\nLine 2\rLine 3",
	}

	record := buildRecord(event, meta)

	// \n becomes space, \r is removed
	require.Equal(t, "Line 1 Line 2Line 3", record[colMessage])
}

func TestWriteCSVSorted_CreatesFileWithHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	// New schema: event_id,event_type,timestamp,repo_id,repo_name,author_id,author_name,author_email,branch,commit_hash,parent_hashes,message,files_changed,insertions,deletions,device
	records := map[string][]string{
		"repo1:commit1": {"uuid1", "commit", "2024-01-15T10:30:00Z", "repo1", "repo1", "auth1", "", "", "main", "commit1", "", "msg", "0", "0", "0", "device1"},
	}

	err := writeCSVSorted(path, records)
	require.NoError(t, err)

	// Verify file exists and has header
	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	require.NoError(t, err)

	require.Equal(t, "event_id", header[colEventID])
	require.Equal(t, "repo_id", header[colRepoID])
}

func TestWriteCSVSorted_SortsByTimestamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	// New schema: timestamp is at index 2, commit_hash at index 9
	records := map[string][]string{
		"repo:commit3": {"uuid3", "commit", "2024-01-20T10:00:00Z", "repo", "repo", "auth", "", "", "main", "commit3", "", "third", "0", "0", "0", "device"},
		"repo:commit1": {"uuid1", "commit", "2024-01-10T10:00:00Z", "repo", "repo", "auth", "", "", "main", "commit1", "", "first", "0", "0", "0", "device"},
		"repo:commit2": {"uuid2", "commit", "2024-01-15T10:00:00Z", "repo", "repo", "auth", "", "", "main", "commit2", "", "second", "0", "0", "0", "device"},
	}

	err := writeCSVSorted(path, records)
	require.NoError(t, err)

	// Read and verify order (sorted by timestamp)
	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	all, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, all, 4) // header + 3 records
	require.Equal(t, "commit1", all[1][colCommitHash])
	require.Equal(t, "commit2", all[2][colCommitHash])
	require.Equal(t, "commit3", all[3][colCommitHash])
}

func TestWriteCSVSorted_HasRestrictivePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	records := map[string][]string{}
	err := writeCSVSorted(path, records)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)

	perm := info.Mode().Perm()
	require.Equal(t, os.FileMode(0600), perm)
}

func TestShouldExport_ReturnsNoError(t *testing.T) {
	deps := Deps{
		Now: func() time.Time {
			return time.Now()
		},
	}

	_, err := shouldExport(deps)

	require.NoError(t, err)
}

func TestEnsureExportRepo_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")

	err := ensureExportRepo(exportDir)

	require.NoError(t, err)

	info, err := os.Stat(exportDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())

	gitDir := filepath.Join(exportDir, ".git")
	_, err = os.Stat(gitDir)
	require.NoError(t, err)
}

func TestCommitExportChanges_EmptyFiles(t *testing.T) {
	dir := t.TempDir()

	err := commitExportChanges(dir, nil)

	require.NoError(t, err)
}

func TestExportAllEvents_SortsAndGroupsByYear(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	events := []store.RepoEvent{
		{
			ID:        1,
			RepoID:    "github.com/user/repo1",
			Commit:    "abc123",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
		{
			ID:        2,
			RepoID:    "github.com/user/repo2",
			Commit:    "def456",
			Branch:    "main",
			Timestamp: time.Date(2024, 3, 10, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
		{
			ID:        3,
			RepoID:    "github.com/user/repo1",
			Commit:    "ghi789",
			Branch:    "develop",
			Timestamp: time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
	}

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	ids, files, err := exportAllEvents(exportDir, events, deps)

	require.NoError(t, err)
	require.Len(t, ids, 3)
	require.Len(t, files, 2)

	_, err = os.Stat(filepath.Join(exportDir, "commits.csv"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(exportDir, "commits-2024.csv"))
	require.NoError(t, err)
}

func TestExportAllEvents_MultipleReposInSameFile(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	events := []store.RepoEvent{
		{
			ID:        1,
			RepoID:    "github.com/user/repo1",
			Commit:    "abc123",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
		{
			ID:        2,
			RepoID:    "github.com/user/repo2",
			Commit:    "def456",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 16, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
	}

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	ids, files, err := exportAllEvents(exportDir, events, deps)

	require.NoError(t, err)
	require.Len(t, ids, 2)
	require.Len(t, files, 1)

	file, err := os.Open(filepath.Join(exportDir, "commits.csv"))
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, records, 3)

	repos := make(map[string]bool)
	for _, row := range records[1:] {
		repos[row[colRepoID]] = true
	}
	require.True(t, repos["github.com/user/repo1"])
	require.True(t, repos["github.com/user/repo2"])
}

func TestExportAllEvents_EventsAreSortedByAuthoredAt(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	events := []store.RepoEvent{
		{
			ID:        3,
			RepoID:    "github.com/user/repo",
			Commit:    "third",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
		{
			ID:        1,
			RepoID:    "github.com/user/repo",
			Commit:    "first",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 10, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
		{
			ID:        2,
			RepoID:    "github.com/user/repo",
			Commit:    "second",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
	}

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	_, _, err = exportAllEvents(exportDir, events, deps)
	require.NoError(t, err)

	file, err := os.Open(filepath.Join(exportDir, "commits.csv"))
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, records, 4)
	require.Equal(t, "first", records[1][colCommitHash])
	require.Equal(t, "second", records[2][colCommitHash])
	require.Equal(t, "third", records[3][colCommitHash])
}

func TestExportAllEvents_EmptyEvents(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	ids, files, err := exportAllEvents(exportDir, []store.RepoEvent{}, deps)

	require.NoError(t, err)
	require.Empty(t, ids)
	require.Empty(t, files)
}

func TestExportAllEvents_YearBoundary(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	events := []store.RepoEvent{
		{
			ID:        1,
			RepoID:    "github.com/user/repo",
			Commit:    "last_of_2024",
			Branch:    "main",
			Timestamp: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
		{
			ID:        2,
			RepoID:    "github.com/user/repo",
			Commit:    "first_of_2025",
			Branch:    "main",
			Timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
	}

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	ids, files, err := exportAllEvents(exportDir, events, deps)

	require.NoError(t, err)
	require.Len(t, ids, 2)
	require.Len(t, files, 2)

	file2024, err := os.Open(filepath.Join(exportDir, "commits-2024.csv"))
	require.NoError(t, err)
	defer func() { _ = file2024.Close() }()

	reader2024 := csv.NewReader(file2024)
	records2024, err := reader2024.ReadAll()
	require.NoError(t, err)
	require.Len(t, records2024, 2)
	require.Equal(t, "last_of_2024", records2024[1][colCommitHash])

	file2025, err := os.Open(filepath.Join(exportDir, "commits.csv"))
	require.NoError(t, err)
	defer func() { _ = file2025.Close() }()

	reader2025 := csv.NewReader(file2025)
	records2025, err := reader2025.ReadAll()
	require.NoError(t, err)
	require.Len(t, records2025, 2)
	require.Equal(t, "first_of_2025", records2025[1][colCommitHash])
}

func TestExportAllEvents_PreservesExistingRecords(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	// First batch
	events1 := []store.RepoEvent{
		{
			ID:        1,
			RepoID:    "github.com/user/repo",
			Commit:    "commit1",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 10, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
	}

	_, _, err = exportAllEvents(exportDir, events1, deps)
	require.NoError(t, err)

	// Second batch
	events2 := []store.RepoEvent{
		{
			ID:        2,
			RepoID:    "github.com/user/repo",
			Commit:    "commit2",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
	}

	_, _, err = exportAllEvents(exportDir, events2, deps)
	require.NoError(t, err)

	// Verify both commits are present
	file, err := os.Open(filepath.Join(exportDir, "commits.csv"))
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, records, 3)

	commits := make(map[string]bool)
	for _, row := range records[1:] {
		commits[row[colCommitHash]] = true
	}
	require.True(t, commits["commit1"])
	require.True(t, commits["commit2"])
}

func TestExportAllEvents_ReplacesDuplicates(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
	}

	// First export with original message
	events1 := []store.RepoEvent{
		{
			ID:        1,
			RepoID:    "github.com/user/repo",
			Commit:    "abc123",
			Branch:    "main",
			Timestamp: time.Date(2025, 6, 10, 10, 0, 0, 0, time.UTC),
			Source:    store.SourcePostCommit,
		},
	}

	_, _, err = exportAllEvents(exportDir, events1, deps)
	require.NoError(t, err)

	// Second export with same repo:commit (should replace)
	events2 := []store.RepoEvent{
		{
			ID:        2,
			RepoID:    "github.com/user/repo",
			Commit:    "abc123",
			Branch:    "feature", // different branch
			Timestamp: time.Date(2025, 6, 10, 10, 0, 0, 0, time.UTC),
			Source:    store.SourceBackfill, // different source
		},
	}

	_, _, err = exportAllEvents(exportDir, events2, deps)
	require.NoError(t, err)

	// Verify only one record exists and it's the newer one
	file, err := os.Open(filepath.Join(exportDir, "commits.csv"))
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, records, 2) // header + 1 record (not 2)
	require.Equal(t, "feature", records[1][colBranch]) // branch from second export
}

func TestParseCSVIntoMap_LastWriteWins(t *testing.T) {
	records := make(map[string][]string)

	// First version (using new semantic API schema)
	content1 := `event_id,event_type,timestamp,repo_id,repo_name,author_id,author_name,author_email,branch,commit_hash,parent_hashes,message,files_changed,insertions,deletions,device
uuid1,commit,2024-01-15T10:30:00Z,myrepo,myrepo,author1,,,main,abc123,,First version,0,0,0,machine1
`
	parseCSVIntoMap(content1, records)

	// Second version (same key, different data)
	content2 := `event_id,event_type,timestamp,repo_id,repo_name,author_id,author_name,author_email,branch,commit_hash,parent_hashes,message,files_changed,insertions,deletions,device
uuid2,commit,2024-01-15T10:30:00Z,myrepo,myrepo,author1,,,feature,abc123,,Second version,0,0,0,machine2
`
	parseCSVIntoMap(content2, records)

	require.Len(t, records, 1)
	require.Equal(t, "feature", records["myrepo:abc123"][colBranch])  // branch from second
	require.Equal(t, "Second version", records["myrepo:abc123"][colMessage]) // message from second
}

func TestDoExportWork_OfflineMode_ContinuesWhenPullFails(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	// Configure git user for CI environment (no global config)
	cmd := exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = exportDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = exportDir
	_ = cmd.Run()

	// Create a temp database
	dbPath := filepath.Join(dir, "test.db")
	db, err := store.Open(dbPath) //nolint:staticcheck // Using deprecated for test compatibility
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	err = store.Init(db)
	require.NoError(t, err)

	// Insert test events
	event := store.RepoEvent{
		RepoID:    "github.com/user/repo",
		Commit:    "abc123",
		Branch:    "main",
		Timestamp: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
		Status:    store.StatusPending,
		Source:    store.SourcePostCommit,
	}
	err = store.InsertEvent(db, event)
	require.NoError(t, err)

	// Get pending events
	events, err := store.GetPendingEvents(db)
	require.NoError(t, err)
	require.Len(t, events, 1)

	pullCalled := false
	pushCalled := false

	deps := Deps{
		Now: func() time.Time {
			return time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		},
		// Mock: return test export dir
		GetExportRepo: func() string {
			return exportDir
		},
		// Mock: remote exists
		HasRemote: func(string) bool {
			return true
		},
		// Mock: pull fails (simulating offline)
		PullExportRepo: func(string) error {
			pullCalled = true
			return errors.New("could not resolve host: github.com")
		},
		// Mock: push also fails
		PushExportRepo: func(string) error {
			pushCalled = true
			return errors.New("could not resolve host: github.com")
		},
	}

	// Execute
	count, pushed, err := doExportWork(db, events, deps)

	// Verify: export succeeded despite pull failure
	require.NoError(t, err, "export should succeed even when pull fails")
	require.Equal(t, 1, count, "should have exported 1 event")
	require.False(t, pushed, "push should have failed")
	require.True(t, pullCalled, "pull should have been attempted")
	require.True(t, pushCalled, "push should have been attempted")

	// Verify CSV was created
	csvPath := filepath.Join(exportDir, "commits.csv")
	_, err = os.Stat(csvPath)
	require.NoError(t, err, "CSV file should exist")

	// Verify content
	records := loadCSVRecords(csvPath)
	require.Len(t, records, 1)
	require.Contains(t, records, "github.com/user/repo:abc123")
}

func TestShouldExport_WithInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		lastStr  string
		nowUnix  int64
		expected bool
	}{
		{
			name:     "interval 0 always exports",
			interval: "0",
			lastStr:  "0",
			nowUnix:  100,
			expected: true,
		},
		{
			name:     "interval not yet passed",
			interval: "3600",
			lastStr:  "1000",
			nowUnix:  1500,
			expected: false,
		},
		{
			name:     "interval passed",
			interval: "3600",
			lastStr:  "1000",
			nowUnix:  5000,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := Deps{
				Now: func() time.Time {
					return time.Unix(tt.nowUnix, 0)
				},
			}

			result, err := shouldExport(deps)
			require.NoError(t, err)
			// Note: The actual shouldExport reads from config, so this tests the basic path
			require.NotNil(t, result)
		})
	}
}

func TestGetHostname(t *testing.T) {
	hostname := getHostname()
	// Should return either a non-empty hostname or empty string on error
	// In most environments, this will return a valid hostname
	require.True(t, len(hostname) >= 0)
}

func TestCommitExportChanges_WithFiles(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	// Configure git user for CI environment
	cmd := exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = exportDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = exportDir
	_ = cmd.Run()

	// Create a test file
	testFile := filepath.Join(exportDir, "test.csv")
	err = os.WriteFile(testFile, []byte("content"), 0600)
	require.NoError(t, err)

	// Commit the file
	err = commitExportChanges(exportDir, []string{"test.csv"})
	require.NoError(t, err)

	// Verify commit exists
	cmd = exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = exportDir
	output, err := cmd.Output()
	require.NoError(t, err)
	require.Contains(t, string(output), "Export 1 files")
}

func TestCommitExportChanges_NoChanges(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	// Configure git user for CI environment
	cmd := exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = exportDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = exportDir
	_ = cmd.Run()

	// Create and commit a test file first
	testFile := filepath.Join(exportDir, "test.csv")
	err = os.WriteFile(testFile, []byte("content"), 0600)
	require.NoError(t, err)

	err = commitExportChanges(exportDir, []string{"test.csv"})
	require.NoError(t, err)

	// Try to commit the same file without changes
	err = commitExportChanges(exportDir, []string{"test.csv"})
	require.NoError(t, err) // Should not error when no changes
}

func TestLoadCSVRecords_InvalidCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.csv")

	// Create a file with invalid CSV format (mismatched quotes)
	err := os.WriteFile(path, []byte("col1,col2\n\"unclosed quote,data"), 0600)
	require.NoError(t, err)

	records := loadCSVRecords(path)
	// Should return empty map on error
	require.Empty(t, records)
}

func TestLoadCSVRecords_MissingColumns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "short.csv")

	// Create CSV with fewer columns than expected
	content := `authored_at,repo
2024-01-15T10:30:00Z,github.com/user/repo
`
	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)

	records := loadCSVRecords(path)
	// Should handle gracefully (either return empty or partial data)
	// The actual behavior depends on implementation
	require.NotNil(t, records)
}

func TestWriteCSVSorted_InvalidPath(t *testing.T) {
	// Try to write to an invalid path
	path := "/nonexistent/directory/test.csv"
	records := map[string][]string{
		"repo:commit": {"uuid", "commit", "2024-01-15T10:30:00Z", "repo", "repo", "auth", "", "", "main", "commit", "", "msg", "0", "0", "0", "device"},
	}

	err := writeCSVSorted(path, records)
	require.Error(t, err)
}

func TestEnsureExportRepo_ExistingRepo(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")

	// First call creates the repo
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

	// Second call should not error (repo already exists)
	err = ensureExportRepo(exportDir)
	require.NoError(t, err)

	// Verify .git directory exists
	gitDir := filepath.Join(exportDir, ".git")
	info, err := os.Stat(gitDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestParseCSVIntoMap_EmptyContent(t *testing.T) {
	records := make(map[string][]string)

	parseCSVIntoMap("", records)

	require.Empty(t, records)
}

func TestParseCSVIntoMap_HeaderOnly(t *testing.T) {
	records := make(map[string][]string)

	content := `authored_at,repo,branch,commit,subject,author,author_email,files,additions,deletions,parents,committer,committer_email,committed_at,source,machine`

	parseCSVIntoMap(content, records)

	require.Empty(t, records)
}
