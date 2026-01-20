package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestListEvents_NoFilters(t *testing.T) {
	db := newTestDB(t)

	// Seed events
	events := []RepoEvent{
		{
			RepoID:    "github.com/user/repo1",
			RepoPath:  "/path/to/repo1",
			Commit:    "abc123",
			Branch:    "main",
			Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Status:    StatusPending,
			Source:    SourcePostCommit,
		},
		{
			RepoID:    "github.com/user/repo2",
			RepoPath:  "/path/to/repo2",
			Commit:    "def456",
			Branch:    "develop",
			Timestamp: time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
			Status:    StatusExported,
			Source:    SourceManual,
		},
		{
			RepoID:    "github.com/user/repo3",
			RepoPath:  "/path/to/repo3",
			Commit:    "ghi789",
			Branch:    "feature",
			Timestamp: time.Date(2024, 1, 17, 12, 0, 0, 0, time.UTC),
			Status:    StatusPending,
			Source:    SourceBackfill,
		},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// List all events without filters
	got, err := ListEvents(db, EventFilter{})
	require.NoError(t, err)
	require.Len(t, got, 3)

	// Verify ORDER BY timestamp DESC (newest first)
	require.Equal(t, "ghi789", got[0].Commit)
	require.Equal(t, "def456", got[1].Commit)
	require.Equal(t, "abc123", got[2].Commit)
}

func TestListEvents_FilterByStatus(t *testing.T) {
	db := newTestDB(t)

	// Seed events with different statuses
	events := []RepoEvent{
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo2", RepoPath: "/path2", Commit: "abc2", Branch: "main", Timestamp: time.Now(), Status: StatusExported, Source: SourcePostCommit},
		{RepoID: "repo3", RepoPath: "/path3", Commit: "abc3", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo4", RepoPath: "/path4", Commit: "abc4", Branch: "main", Timestamp: time.Now(), Status: StatusOrphaned, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Filter by StatusPending
	pending := StatusPending
	got, err := ListEvents(db, EventFilter{Status: &pending})
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, StatusPending, e.Status)
	}

	// Filter by StatusExported
	exported := StatusExported
	got, err = ListEvents(db, EventFilter{Status: &exported})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, StatusExported, got[0].Status)
}

func TestListEvents_FilterBySource(t *testing.T) {
	db := newTestDB(t)

	// Seed events with different sources
	events := []RepoEvent{
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo2", RepoPath: "/path2", Commit: "abc2", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourceManual},
		{RepoID: "repo3", RepoPath: "/path3", Commit: "abc3", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourceBackfill},
		{RepoID: "repo4", RepoPath: "/path4", Commit: "abc4", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Filter by SourcePostCommit
	postCommit := SourcePostCommit
	got, err := ListEvents(db, EventFilter{Source: &postCommit})
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, SourcePostCommit, e.Source)
	}

	// Filter by SourceBackfill
	backfill := SourceBackfill
	got, err = ListEvents(db, EventFilter{Source: &backfill})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, SourceBackfill, got[0].Source)
}

func TestListEvents_FilterBySince(t *testing.T) {
	db := newTestDB(t)

	// Seed events with different timestamps
	events := []RepoEvent{
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo2", RepoPath: "/path2", Commit: "abc2", Branch: "main", Timestamp: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo3", RepoPath: "/path3", Commit: "abc3", Branch: "main", Timestamp: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Filter by Since (>= 2024-01-15)
	since := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got, err := ListEvents(db, EventFilter{Since: &since})
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		require.True(t, e.Timestamp.Equal(since) || e.Timestamp.After(since))
	}
}

func TestListEvents_FilterByUntil(t *testing.T) {
	db := newTestDB(t)

	// Seed events with different timestamps
	events := []RepoEvent{
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo2", RepoPath: "/path2", Commit: "abc2", Branch: "main", Timestamp: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo3", RepoPath: "/path3", Commit: "abc3", Branch: "main", Timestamp: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Filter by Until (<= 2024-01-15)
	until := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got, err := ListEvents(db, EventFilter{Until: &until})
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		require.True(t, e.Timestamp.Equal(until) || e.Timestamp.Before(until))
	}
}

func TestListEvents_FilterByRepoID(t *testing.T) {
	db := newTestDB(t)

	// Seed events with different repo IDs
	events := []RepoEvent{
		{RepoID: "github.com/user/repo1", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "github.com/user/repo2", RepoPath: "/path2", Commit: "abc2", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "github.com/user/repo1", RepoPath: "/path1", Commit: "abc3", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Filter by repo1
	repoID := "github.com/user/repo1"
	got, err := ListEvents(db, EventFilter{RepoID: &repoID})
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, repoID, e.RepoID)
	}
}

func TestListEvents_CombinedFilters(t *testing.T) {
	db := newTestDB(t)

	// Seed events
	events := []RepoEvent{
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc2", Branch: "main", Timestamp: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc3", Branch: "main", Timestamp: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), Status: StatusExported, Source: SourcePostCommit},
		{RepoID: "repo2", RepoPath: "/path2", Commit: "abc4", Branch: "main", Timestamp: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Status: StatusPending, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Filter by repo1 + StatusPending + Since 2024-01-12
	repoID := "repo1"
	status := StatusPending
	since := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	got, err := ListEvents(db, EventFilter{
		RepoID: &repoID,
		Status: &status,
		Since:  &since,
	})
	require.NoError(t, err)
	require.Len(t, got, 1) // Only abc2 matches all filters
	require.Equal(t, "abc2", got[0].Commit)
}

func TestListEvents_Limit(t *testing.T) {
	db := newTestDB(t)

	// Seed 10 events
	for i := 0; i < 10; i++ {
		event := RepoEvent{
			RepoID:    "repo1",
			RepoPath:  "/path1",
			Commit:    string(rune('a' + i)),
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    StatusPending,
			Source:    SourcePostCommit,
		}
		err := InsertEvent(db, event)
		require.NoError(t, err)
	}

	// Limit to 5
	got, err := ListEvents(db, EventFilter{Limit: 5})
	require.NoError(t, err)
	require.Len(t, got, 5)

	// Limit to 1
	got, err = ListEvents(db, EventFilter{Limit: 1})
	require.NoError(t, err)
	require.Len(t, got, 1)

	// No limit
	got, err = ListEvents(db, EventFilter{})
	require.NoError(t, err)
	require.Len(t, got, 10)
}

func TestGetMaxEventID(t *testing.T) {
	db := newTestDB(t)

	// Empty database returns 0
	maxID, err := GetMaxEventID(db)
	require.NoError(t, err)
	require.Equal(t, int64(0), maxID)

	// Insert event
	event := RepoEvent{
		RepoID:    "repo1",
		RepoPath:  "/path1",
		Commit:    "abc123",
		Branch:    "main",
		Timestamp: time.Now(),
		Status:    StatusPending,
		Source:    SourcePostCommit,
	}
	err = InsertEvent(db, event)
	require.NoError(t, err)

	// Get max ID (should be 1 now)
	maxID, err = GetMaxEventID(db)
	require.NoError(t, err)
	require.Equal(t, int64(1), maxID)

	// Insert another event
	event2 := RepoEvent{
		RepoID:    "repo2",
		RepoPath:  "/path2",
		Commit:    "def456",
		Branch:    "main",
		Timestamp: time.Now(),
		Status:    StatusPending,
		Source:    SourcePostCommit,
	}
	err = InsertEvent(db, event2)
	require.NoError(t, err)

	// Get max ID (should be 2 now)
	maxID, err = GetMaxEventID(db)
	require.NoError(t, err)
	require.Equal(t, int64(2), maxID)
}

func TestListEventsSince(t *testing.T) {
	db := newTestDB(t)

	// Insert 5 events
	for i := 0; i < 5; i++ {
		event := RepoEvent{
			RepoID:    "repo1",
			RepoPath:  "/path1",
			Commit:    string(rune('a' + i)),
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    StatusPending,
			Source:    SourcePostCommit,
		}
		err := InsertEvent(db, event)
		require.NoError(t, err)
	}

	// Get events after ID 3 (should get IDs 4 and 5)
	got, err := ListEventsSince(db, 3)
	require.NoError(t, err)
	require.Len(t, got, 2)

	// Verify ORDER BY id ASC
	require.Equal(t, int64(4), got[0].ID)
	require.Equal(t, int64(5), got[1].ID)

	// Get events after ID 5 (should be empty)
	got, err = ListEventsSince(db, 5)
	require.NoError(t, err)
	require.Len(t, got, 0)

	// Get events after ID 0 (should get all)
	got, err = ListEventsSince(db, 0)
	require.NoError(t, err)
	require.Len(t, got, 5)
}

func TestUpdateEventStatuses(t *testing.T) {
	db := newTestDB(t)

	// Insert 3 pending events
	for i := 0; i < 3; i++ {
		event := RepoEvent{
			RepoID:    "repo1",
			RepoPath:  "/path1",
			Commit:    string(rune('a' + i)),
			Branch:    "main",
			Timestamp: time.Now(),
			Status:    StatusPending,
			Source:    SourcePostCommit,
		}
		err := InsertEvent(db, event)
		require.NoError(t, err)
	}

	// Update status of first 2 events to Exported
	err := UpdateEventStatuses(db, []int64{1, 2}, StatusExported)
	require.NoError(t, err)

	// Verify statuses were updated
	pending := StatusPending
	got, err := ListEvents(db, EventFilter{Status: &pending})
	require.NoError(t, err)
	require.Len(t, got, 1)

	exported := StatusExported
	got, err = ListEvents(db, EventFilter{Status: &exported})
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestUpdateEventStatuses_EmptyList(t *testing.T) {
	db := newTestDB(t)

	// Update with empty list should not error
	err := UpdateEventStatuses(db, []int64{}, StatusExported)
	require.NoError(t, err)
}

func TestMigratePendingRepoID(t *testing.T) {
	db := newTestDB(t)

	// Insert events with different statuses
	events := []RepoEvent{
		{RepoID: "old-id", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "old-id", RepoPath: "/path1", Commit: "abc2", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "old-id", RepoPath: "/path1", Commit: "abc3", Branch: "main", Timestamp: time.Now(), Status: StatusExported, Source: SourcePostCommit},
		{RepoID: "other-id", RepoPath: "/path2", Commit: "abc4", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Migrate pending events from old-id to new-id
	count, err := MigratePendingRepoID(db, "old-id", "new-id")
	require.NoError(t, err)
	require.Equal(t, int64(2), count) // Only 2 pending events with old-id

	// Verify migration
	newID := "new-id"
	got, err := ListEvents(db, EventFilter{RepoID: &newID})
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, StatusPending, e.Status)
	}

	// Verify exported event was not migrated
	oldID := "old-id"
	got, err = ListEvents(db, EventFilter{RepoID: &oldID})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, StatusExported, got[0].Status)
}

func TestGetPendingEvents(t *testing.T) {
	db := newTestDB(t)

	// Insert events with different statuses
	events := []RepoEvent{
		{RepoID: "repo1", RepoPath: "/path1", Commit: "abc1", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
		{RepoID: "repo2", RepoPath: "/path2", Commit: "abc2", Branch: "main", Timestamp: time.Now(), Status: StatusExported, Source: SourcePostCommit},
		{RepoID: "repo3", RepoPath: "/path3", Commit: "abc3", Branch: "main", Timestamp: time.Now(), Status: StatusPending, Source: SourcePostCommit},
	}

	for _, e := range events {
		err := InsertEvent(db, e)
		require.NoError(t, err)
	}

	// Get pending events
	got, err := GetPendingEvents(db)
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, StatusPending, e.Status)
	}
}
