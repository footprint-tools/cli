package tracking

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/footprint-tools/footprint-cli/internal/git"
	"github.com/footprint-tools/footprint-cli/internal/store"
	"github.com/stretchr/testify/require"
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

	// Create CSV file with header and data
	content := `authored_at,repo,branch,commit,subject,author,author_email,files,additions,deletions,parents,committer,committer_email,committed_at,source,machine
2024-01-15T10:30:00Z,github.com/user/repo,main,abc123,Fix bug,John,john@example.com,3,10,5,parent1,John,john@example.com,2024-01-15T10:30:00Z,post-commit,machine1
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
	require.Equal(t, "2024-01-15T10:30:00Z", record[0]) // authored_at
	require.Equal(t, "github.com/user/repo", record[1]) // repo
	require.Equal(t, "main", record[2])                 // branch
	require.Equal(t, "abc123def456", record[3])         // commit
	require.Equal(t, "Fix bug", record[4])              // subject
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
	require.Equal(t, "Line 1 Line 2Line 3", record[4])
}

func TestWriteCSVSorted_CreatesFileWithHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	records := map[string][]string{
		"repo1:commit1": {"2024-01-15T10:30:00Z", "repo1", "main", "commit1", "msg", "", "", "0", "0", "0", "", "", "", "", "", ""},
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

	require.Equal(t, "authored_at", header[0])
	require.Equal(t, "repo", header[1])
}

func TestWriteCSVSorted_SortsByAuthoredAt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	records := map[string][]string{
		"repo:commit3": {"2024-01-20T10:00:00Z", "repo", "main", "commit3", "third", "", "", "0", "0", "0", "", "", "", "", "", ""},
		"repo:commit1": {"2024-01-10T10:00:00Z", "repo", "main", "commit1", "first", "", "", "0", "0", "0", "", "", "", "", "", ""},
		"repo:commit2": {"2024-01-15T10:00:00Z", "repo", "main", "commit2", "second", "", "", "0", "0", "0", "", "", "", "", "", ""},
	}

	err := writeCSVSorted(path, records)
	require.NoError(t, err)

	// Read and verify order
	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	all, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, all, 4) // header + 3 records
	require.Equal(t, "commit1", all[1][3])
	require.Equal(t, "commit2", all[2][3])
	require.Equal(t, "commit3", all[3][3])
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
		repos[row[1]] = true
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
	require.Equal(t, "first", records[1][3])
	require.Equal(t, "second", records[2][3])
	require.Equal(t, "third", records[3][3])
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
	require.Equal(t, "last_of_2024", records2024[1][3])

	file2025, err := os.Open(filepath.Join(exportDir, "commits.csv"))
	require.NoError(t, err)
	defer func() { _ = file2025.Close() }()

	reader2025 := csv.NewReader(file2025)
	records2025, err := reader2025.ReadAll()
	require.NoError(t, err)
	require.Len(t, records2025, 2)
	require.Equal(t, "first_of_2025", records2025[1][3])
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
		commits[row[3]] = true
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
	require.Equal(t, "feature", records[1][2])        // branch from second export
	require.Equal(t, "BACKFILL", records[1][14])      // source from second export (uppercase)
}

func TestParseCSVIntoMap_LastWriteWins(t *testing.T) {
	records := make(map[string][]string)

	// First version
	content1 := `authored_at,repo,branch,commit,subject,author,author_email,files,additions,deletions,parents,committer,committer_email,committed_at,source,machine
2024-01-15T10:30:00Z,myrepo,main,abc123,First version,,,0,0,0,,,,,post-commit,machine1
`
	parseCSVIntoMap(content1, records)

	// Second version (same key, different data)
	content2 := `authored_at,repo,branch,commit,subject,author,author_email,files,additions,deletions,parents,committer,committer_email,committed_at,source,machine
2024-01-15T10:30:00Z,myrepo,feature,abc123,Second version,,,0,0,0,,,,,backfill,machine2
`
	parseCSVIntoMap(content2, records)

	require.Len(t, records, 1)
	require.Equal(t, "feature", records["myrepo:abc123"][2])      // branch from second
	require.Equal(t, "Second version", records["myrepo:abc123"][4]) // subject from second
}

func TestDoExportWork_OfflineMode_ContinuesWhenPullFails(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	err := ensureExportRepo(exportDir)
	require.NoError(t, err)

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
