package store

import (
	"database/sql"
	"time"

	"github.com/footprint-tools/cli/internal/log"
)

func InsertEvent(db *sql.DB, e RepoEvent) error {
	_, err := db.Exec(
		`INSERT INTO repo_events
		 (repo_id, repo_path, commit_hash, branch, timestamp, status_id, source_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(repo_id, commit_hash, source_id)
		 DO UPDATE SET timestamp = excluded.timestamp`,
		e.RepoID,
		e.RepoPath,
		e.Commit,
		e.Branch,
		e.Timestamp.Format(time.RFC3339),
		int(e.Status),
		int(e.Source),
	)
	if err != nil {
		log.Error("store: insert event failed: %v (repo=%s, commit=%.7s)", err, e.RepoID, e.Commit)
	}
	return err
}

// MarkOrphanedByRepoID marks all pending events for a repo as orphaned.
// Returns the number of events updated.
func MarkOrphanedByRepoID(db *sql.DB, repoID string) (int64, error) {
	query := `
		UPDATE repo_events
		SET status_id = ?
		WHERE repo_id = ? AND status_id = ?
	`
	result, err := db.Exec(query, int(StatusOrphaned), repoID, int(StatusPending))
	if err != nil {
		log.Error("store: mark orphaned failed: %v (repo=%s)", err, repoID)
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteOrphanedEvents deletes all orphaned events from the database.
// Returns the number of events deleted.
func DeleteOrphanedEvents(db *sql.DB) (int64, error) {
	query := `DELETE FROM repo_events WHERE status_id = ?`
	result, err := db.Exec(query, int(StatusOrphaned))
	if err != nil {
		log.Error("store: delete orphaned failed: %v", err)
		return 0, err
	}
	return result.RowsAffected()
}

// CountOrphanedEvents returns the count of orphaned events.
func CountOrphanedEvents(db *sql.DB) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM repo_events WHERE status_id = ?`
	err := db.QueryRow(query, int(StatusOrphaned)).Scan(&count)
	if err != nil {
		log.Error("store: count orphaned failed: %v", err)
		return 0, err
	}
	return count, nil
}
