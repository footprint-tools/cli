package store

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/store/migrations"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	err = migrations.Run(db)
	require.NoError(t, err)

	return NewWithDB(db)
}

func TestStore_Insert(t *testing.T) {
	s := newTestStore(t)

	event := domain.RepoEvent{
		RepoID:    domain.RepoID("github.com/test/repo"),
		RepoPath:  "/path/to/repo",
		Commit:    "abc1234",
		Branch:    "main",
		Timestamp: time.Now().Truncate(time.Second),
		Status:    domain.StatusPending,
		Source:    domain.SourcePostCommit,
	}

	err := s.Insert(event)
	require.NoError(t, err)

	events, err := s.List(domain.EventFilter{})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, event.RepoID, events[0].RepoID)
	require.Equal(t, event.Commit, events[0].Commit)
}

func TestStore_Insert_Upsert(t *testing.T) {
	s := newTestStore(t)

	event := domain.RepoEvent{
		RepoID:    domain.RepoID("github.com/test/repo"),
		RepoPath:  "/path/to/repo",
		Commit:    "abc1234",
		Branch:    "main",
		Timestamp: time.Now().Truncate(time.Second),
		Status:    domain.StatusPending,
		Source:    domain.SourcePostCommit,
	}

	// Insert first time
	err := s.Insert(event)
	require.NoError(t, err)

	// Insert again with same repo_id, commit, source - should update
	event.Timestamp = time.Now().Add(time.Hour).Truncate(time.Second)
	err = s.Insert(event)
	require.NoError(t, err)

	events, err := s.List(domain.EventFilter{})
	require.NoError(t, err)
	require.Len(t, events, 1, "should upsert, not create duplicate")
}

func TestStore_List_WithFilters(t *testing.T) {
	s := newTestStore(t)

	now := time.Now().Truncate(time.Second)

	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: now,
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo2"),
			RepoPath:  "/path/to/repo2",
			Commit:    "def5678",
			Branch:    "develop",
			Timestamp: now.Add(-time.Hour),
			Status:    domain.StatusExported,
			Source:    domain.SourceBackfill,
		},
	}

	for _, e := range events {
		err := s.Insert(e)
		require.NoError(t, err)
	}

	// Filter by status
	pending := domain.StatusPending
	result, err := s.List(domain.EventFilter{Status: &pending})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, domain.RepoID("github.com/test/repo1"), result[0].RepoID)

	// Filter by source
	backfill := domain.SourceBackfill
	result, err = s.List(domain.EventFilter{Source: &backfill})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, domain.RepoID("github.com/test/repo2"), result[0].RepoID)

	// Filter by repo ID
	result, err = s.List(domain.EventFilter{RepoID: domain.RepoID("github.com/test/repo1")})
	require.NoError(t, err)
	require.Len(t, result, 1)

	// Filter with limit
	result, err = s.List(domain.EventFilter{Limit: 1})
	require.NoError(t, err)
	require.Len(t, result, 1)
}

func TestStore_GetPending(t *testing.T) {
	s := newTestStore(t)

	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo2"),
			RepoPath:  "/path/to/repo2",
			Commit:    "def5678",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusExported,
			Source:    domain.SourcePostCommit,
		},
	}

	for _, e := range events {
		require.NoError(t, s.Insert(e))
	}

	pending, err := s.GetPending()
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, domain.StatusPending, pending[0].Status)
}

func TestStore_UpdateStatus(t *testing.T) {
	s := newTestStore(t)

	event := domain.RepoEvent{
		RepoID:    domain.RepoID("github.com/test/repo"),
		RepoPath:  "/path/to/repo",
		Commit:    "abc1234",
		Branch:    "main",
		Timestamp: time.Now(),
		Status:    domain.StatusPending,
		Source:    domain.SourcePostCommit,
	}

	require.NoError(t, s.Insert(event))

	events, err := s.List(domain.EventFilter{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	err = s.UpdateStatus([]int64{events[0].ID}, domain.StatusExported)
	require.NoError(t, err)

	updated, err := s.List(domain.EventFilter{})
	require.NoError(t, err)
	require.Equal(t, domain.StatusExported, updated[0].Status)
}

func TestStore_UpdateStatus_Empty(t *testing.T) {
	s := newTestStore(t)

	err := s.UpdateStatus([]int64{}, domain.StatusExported)
	require.NoError(t, err, "empty IDs should not error")
}

func TestStore_GetMaxID(t *testing.T) {
	s := newTestStore(t)

	// Empty database
	maxID, err := s.GetMaxID()
	require.NoError(t, err)
	require.Equal(t, int64(0), maxID)

	// Insert event
	event := domain.RepoEvent{
		RepoID:    domain.RepoID("github.com/test/repo"),
		RepoPath:  "/path/to/repo",
		Commit:    "abc1234",
		Branch:    "main",
		Timestamp: time.Now(),
		Status:    domain.StatusPending,
		Source:    domain.SourcePostCommit,
	}
	require.NoError(t, s.Insert(event))

	maxID, err = s.GetMaxID()
	require.NoError(t, err)
	require.Equal(t, int64(1), maxID)
}

func TestStore_ListSince(t *testing.T) {
	s := newTestStore(t)

	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo2"),
			RepoPath:  "/path/to/repo2",
			Commit:    "def5678",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourceManual,
		},
	}

	for _, e := range events {
		require.NoError(t, s.Insert(e))
	}

	// Get events since ID 1
	result, err := s.ListSince(1)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, int64(2), result[0].ID)
}

func TestStore_MigrateRepoID(t *testing.T) {
	s := newTestStore(t)

	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("local/old-repo"),
			RepoPath:  "/path/to/repo",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("local/old-repo"),
			RepoPath:  "/path/to/repo",
			Commit:    "def5678",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusExported, // Already exported, should not be migrated
			Source:    domain.SourcePostCommit,
		},
	}

	for _, e := range events {
		require.NoError(t, s.Insert(e))
	}

	count, err := s.MigrateRepoID(
		domain.RepoID("local/old-repo"),
		domain.RepoID("github.com/user/repo"),
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), count, "only pending events should be migrated")

	// Verify migration
	result, err := s.List(domain.EventFilter{RepoID: domain.RepoID("github.com/user/repo")})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, domain.StatusPending, result[0].Status)
}

func TestStore_Close(t *testing.T) {
	s, err := New(":memory:")
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)
}

func TestNew_InvalidPath(t *testing.T) {
	_, err := New("/nonexistent/path/to/db.sqlite")
	// The error happens on ping, not open
	require.Error(t, err)
}

func TestStore_MarkOrphaned(t *testing.T) {
	s := newTestStore(t)

	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "def5678",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusExported, // Already exported, should not be marked
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo2"),
			RepoPath:  "/path/to/repo2",
			Commit:    "ghi9012",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
	}

	for _, e := range events {
		require.NoError(t, s.Insert(e))
	}

	// Mark repo1 events as orphaned
	count, err := s.MarkOrphaned(domain.RepoID("github.com/test/repo1"))
	require.NoError(t, err)
	require.Equal(t, int64(1), count, "only pending events should be marked")

	// Verify the orphaned event
	orphaned := domain.StatusOrphaned
	result, err := s.List(domain.EventFilter{Status: &orphaned})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, domain.RepoID("github.com/test/repo1"), result[0].RepoID)
	require.Equal(t, "abc1234", result[0].Commit)
}

func TestStore_DeleteOrphaned(t *testing.T) {
	s := newTestStore(t)

	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusOrphaned,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo2"),
			RepoPath:  "/path/to/repo2",
			Commit:    "def5678",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
	}

	for _, e := range events {
		require.NoError(t, s.Insert(e))
	}

	// Delete orphaned events
	count, err := s.DeleteOrphaned()
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Verify only non-orphaned event remains
	result, err := s.List(domain.EventFilter{})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, domain.RepoID("github.com/test/repo2"), result[0].RepoID)
}

func TestStore_CountOrphaned(t *testing.T) {
	s := newTestStore(t)

	// Initially no orphaned events
	count, err := s.CountOrphaned()
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Add some events
	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusOrphaned,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo2"),
			RepoPath:  "/path/to/repo2",
			Commit:    "def5678",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusOrphaned,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo3"),
			RepoPath:  "/path/to/repo3",
			Commit:    "ghi9012",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
	}

	for _, e := range events {
		require.NoError(t, s.Insert(e))
	}

	count, err = s.CountOrphaned()
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestStore_ListDistinctRepos(t *testing.T) {
	s := newTestStore(t)

	// Empty database
	repos, err := s.ListDistinctRepos()
	require.NoError(t, err)
	require.Empty(t, repos)

	// Add events with different repo IDs
	events := []domain.RepoEvent{
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "abc1234",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo1"),
			RepoPath:  "/path/to/repo1",
			Commit:    "def5678",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
		{
			RepoID:    domain.RepoID("github.com/test/repo2"),
			RepoPath:  "/path/to/repo2",
			Commit:    "ghi9012",
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    domain.StatusPending,
			Source:    domain.SourcePostCommit,
		},
	}

	for _, e := range events {
		require.NoError(t, s.Insert(e))
	}

	repos, err = s.ListDistinctRepos()
	require.NoError(t, err)
	require.Len(t, repos, 2)
	require.Contains(t, repos, domain.RepoID("github.com/test/repo1"))
	require.Contains(t, repos, domain.RepoID("github.com/test/repo2"))
}
