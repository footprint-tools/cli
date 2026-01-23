package store

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/footprint-cli/internal/store/migrations"
)

// newTestDB creates an in-memory SQLite database with migrations applied.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "failed to open in-memory database")

	t.Cleanup(func() {
		_ = db.Close()
	})

	err = migrations.Run(db)
	require.NoError(t, err, "failed to run migrations")

	return db
}

func TestInsertEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   RepoEvent
		wantErr bool
	}{
		{
			name: "insert new event successfully",
			event: RepoEvent{
				RepoID:    "github.com/user/repo",
				RepoPath:  "/path/to/repo",
				Commit:    "abc123def456",
				Branch:    "main",
				Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Status:    StatusPending,
				Source:    SourcePostCommit,
			},
			wantErr: false,
		},
		{
			name: "insert with different status",
			event: RepoEvent{
				RepoID:    "github.com/user/repo",
				RepoPath:  "/path/to/repo",
				Commit:    "def456abc123",
				Branch:    "develop",
				Timestamp: time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
				Status:    StatusExported,
				Source:    SourceManual,
			},
			wantErr: false,
		},
		{
			name: "insert with different source",
			event: RepoEvent{
				RepoID:    "github.com/user/repo",
				RepoPath:  "/path/to/repo",
				Commit:    "123456789abc",
				Branch:    "feature",
				Timestamp: time.Date(2024, 1, 17, 12, 0, 0, 0, time.UTC),
				Status:    StatusPending,
				Source:    SourceBackfill,
			},
			wantErr: false,
		},
		{
			name: "insert local repo event",
			event: RepoEvent{
				RepoID:    "local:/path/to/local/repo",
				RepoPath:  "/path/to/local/repo",
				Commit:    "fedcba987654",
				Branch:    "main",
				Timestamp: time.Date(2024, 1, 18, 12, 0, 0, 0, time.UTC),
				Status:    StatusPending,
				Source:    SourcePostCommit,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)

			err := InsertEvent(db, tt.event)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify event was inserted
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM repo_events WHERE commit_hash = ?", tt.event.Commit).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 1, count, "event should be inserted")

				// Verify data integrity
				var (
					repoID    string
					repoPath  string
					branch    string
					timestamp string
					statusID  int
					sourceID  int
				)
				err = db.QueryRow(`
					SELECT repo_id, repo_path, branch, timestamp, status_id, source_id
					FROM repo_events
					WHERE commit_hash = ?
				`, tt.event.Commit).Scan(&repoID, &repoPath, &branch, &timestamp, &statusID, &sourceID)
				require.NoError(t, err)
				require.Equal(t, tt.event.RepoID, repoID)
				require.Equal(t, tt.event.RepoPath, repoPath)
				require.Equal(t, tt.event.Branch, branch)
				require.Equal(t, int(tt.event.Status), statusID)
				require.Equal(t, int(tt.event.Source), sourceID)

				// Verify timestamp format (RFC3339)
				parsedTime, err := time.Parse(time.RFC3339, timestamp)
				require.NoError(t, err)
				require.True(t, tt.event.Timestamp.Equal(parsedTime))
			}
		})
	}
}

func TestInsertEvent_OnConflictUpdatesTimestamp(t *testing.T) {
	db := newTestDB(t)

	// Insert initial event
	event1 := RepoEvent{
		RepoID:    "github.com/user/repo",
		RepoPath:  "/path/to/repo",
		Commit:    "abc123",
		Branch:    "main",
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		Status:    StatusPending,
		Source:    SourcePostCommit,
	}
	err := InsertEvent(db, event1)
	require.NoError(t, err)

	// Insert duplicate event with same repo_id, commit_hash, and source_id
	// Should update timestamp (ON CONFLICT behavior)
	event2 := RepoEvent{
		RepoID:    "github.com/user/repo",
		RepoPath:  "/path/to/repo",
		Commit:    "abc123",
		Branch:    "main",
		Timestamp: time.Date(2024, 1, 16, 14, 30, 0, 0, time.UTC), // Different timestamp
		Status:    StatusPending,
		Source:    SourcePostCommit, // Same source
	}
	err = InsertEvent(db, event2)
	require.NoError(t, err)

	// Verify only one event exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM repo_events WHERE commit_hash = ?", event1.Commit).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should still have only one event")

	// Verify timestamp was updated
	var timestamp string
	err = db.QueryRow("SELECT timestamp FROM repo_events WHERE commit_hash = ?", event1.Commit).Scan(&timestamp)
	require.NoError(t, err)

	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	require.NoError(t, err)
	require.True(t, event2.Timestamp.Equal(parsedTime), "timestamp should be updated to event2's timestamp")
}

func TestInsertEvent_DifferentSourceCreatesNewRow(t *testing.T) {
	db := newTestDB(t)

	// Insert event with SourcePostCommit
	event1 := RepoEvent{
		RepoID:    "github.com/user/repo",
		RepoPath:  "/path/to/repo",
		Commit:    "abc123",
		Branch:    "main",
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		Status:    StatusPending,
		Source:    SourcePostCommit,
	}
	err := InsertEvent(db, event1)
	require.NoError(t, err)

	// Insert event with same commit but different source
	event2 := RepoEvent{
		RepoID:    "github.com/user/repo",
		RepoPath:  "/path/to/repo",
		Commit:    "abc123",
		Branch:    "main",
		Timestamp: time.Date(2024, 1, 16, 14, 30, 0, 0, time.UTC),
		Status:    StatusPending,
		Source:    SourceBackfill, // Different source
	}
	err = InsertEvent(db, event2)
	require.NoError(t, err)

	// Verify two events exist (different source_id means different rows)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM repo_events WHERE commit_hash = ?", event1.Commit).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count, "should have two events with different sources")
}

func TestInsertEvent_AllStatuses(t *testing.T) {
	statuses := []Status{
		StatusPending,
		StatusExported,
		StatusOrphaned,
		StatusSkipped,
	}

	for _, status := range statuses {
		t.Run(status.String(), func(t *testing.T) {
			db := newTestDB(t)

			event := RepoEvent{
				RepoID:    "github.com/user/repo",
				RepoPath:  "/path/to/repo",
				Commit:    "abc123",
				Branch:    "main",
				Timestamp: time.Now(),
				Status:    status,
				Source:    SourcePostCommit,
			}

			err := InsertEvent(db, event)
			require.NoError(t, err)

			// Verify status was saved correctly
			var statusID int
			err = db.QueryRow("SELECT status_id FROM repo_events WHERE commit_hash = ?", event.Commit).Scan(&statusID)
			require.NoError(t, err)
			require.Equal(t, int(status), statusID)
		})
	}
}

func TestInsertEvent_AllSources(t *testing.T) {
	sources := []Source{
		SourcePostCommit,
		SourcePostRewrite,
		SourcePostCheckout,
		SourcePostMerge,
		SourcePrePush,
		SourceManual,
		SourceBackfill,
	}

	for _, source := range sources {
		t.Run(source.String(), func(t *testing.T) {
			db := newTestDB(t)

			event := RepoEvent{
				RepoID:    "github.com/user/repo",
				RepoPath:  "/path/to/repo",
				Commit:    "abc123",
				Branch:    "main",
				Timestamp: time.Now(),
				Status:    StatusPending,
				Source:    source,
			}

			err := InsertEvent(db, event)
			require.NoError(t, err)

			// Verify source was saved correctly
			var sourceID int
			err = db.QueryRow("SELECT source_id FROM repo_events WHERE commit_hash = ?", event.Commit).Scan(&sourceID)
			require.NoError(t, err)
			require.Equal(t, int(source), sourceID)
		})
	}
}
