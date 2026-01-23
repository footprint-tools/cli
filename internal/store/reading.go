package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/footprint-tools/footprint-cli/internal/log"
)

type EventFilter struct {
	Status *Status
	Source *Source
	Since  *time.Time
	Until  *time.Time
	RepoID *string
	Limit  int
}

// scanRepoEvent scans a single row into a RepoEvent.
func scanRepoEvent(rows *sql.Rows) (RepoEvent, error) {
	var (
		e        RepoEvent
		ts       string
		statusID int
		sourceID int
	)

	if err := rows.Scan(
		&e.ID,
		&e.RepoID,
		&e.RepoPath,
		&e.Commit,
		&e.Branch,
		&ts,
		&statusID,
		&sourceID,
	); err != nil {
		return RepoEvent{}, err
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return RepoEvent{}, err
	}

	e.Timestamp = t
	e.Status = Status(statusID)
	e.Source = Source(sourceID)

	return e, nil
}

func ListEvents(db *sql.DB, filter EventFilter) ([]RepoEvent, error) {

	base := `
		SELECT
			id,
			repo_id,
			repo_path,
			commit_hash,
			branch,
			timestamp,
			status_id,
			source_id
		FROM repo_events
	`

	var (
		filterClauses []string
		filterArgs    []any
	)

	if filter.Status != nil {
		filterClauses = append(filterClauses, "status_id = ?")
		filterArgs = append(filterArgs, int(*filter.Status))
	}

	if filter.Source != nil {
		filterClauses = append(filterClauses, "source_id = ?")
		filterArgs = append(filterArgs, int(*filter.Source))
	}

	if filter.Since != nil {
		filterClauses = append(filterClauses, "timestamp >= ?")
		filterArgs = append(filterArgs, filter.Since.Format(time.RFC3339))
	}

	if filter.Until != nil {
		filterClauses = append(filterClauses, "timestamp <= ?")
		filterArgs = append(filterArgs, filter.Until.Format(time.RFC3339))
	}

	if filter.RepoID != nil {
		filterClauses = append(filterClauses, "repo_id = ?")
		filterArgs = append(filterArgs, *filter.RepoID)
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(base)

	if len(filterClauses) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(filterClauses, " AND "))
	}

	queryBuilder.WriteString(" ORDER BY timestamp DESC")

	if filter.Limit > 0 {
		queryBuilder.WriteString(" LIMIT ?")
		filterArgs = append(filterArgs, filter.Limit)
	}

	rows, err := db.Query(queryBuilder.String(), filterArgs...)
	if err != nil {
		log.Error("store: list events query failed: %v", err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []RepoEvent

	for rows.Next() {
		e, err := scanRepoEvent(rows)
		if err != nil {
			log.Error("store: scan event row failed: %v", err)
			return nil, err
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

// GetMaxEventID returns the highest event ID in the database.
// Returns 0 if no events exist.
func GetMaxEventID(db *sql.DB) (int64, error) {
	var maxID sql.NullInt64
	err := db.QueryRow("SELECT MAX(id) FROM repo_events").Scan(&maxID)
	if err != nil {
		return 0, err
	}
	if !maxID.Valid {
		return 0, nil
	}
	return maxID.Int64, nil
}

// ListEventsSince returns events with ID greater than afterID, ordered by ID ascending.
// Used for polling new events in real-time.
func ListEventsSince(db *sql.DB, afterID int64) ([]RepoEvent, error) {
	return ListEventsSinceFiltered(db, afterID, EventFilter{})
}

// ListEventsSinceFiltered returns events with ID greater than afterID that match the filter.
// Used for polling new events in real-time with optional filtering.
func ListEventsSinceFiltered(db *sql.DB, afterID int64, filter EventFilter) ([]RepoEvent, error) {
	var (
		filterClauses = []string{"id > ?"}
		filterArgs    = []any{afterID}
	)

	if filter.Status != nil {
		filterClauses = append(filterClauses, "status_id = ?")
		filterArgs = append(filterArgs, int(*filter.Status))
	}

	if filter.Source != nil {
		filterClauses = append(filterClauses, "source_id = ?")
		filterArgs = append(filterArgs, int(*filter.Source))
	}

	if filter.RepoID != nil {
		filterClauses = append(filterClauses, "repo_id = ?")
		filterArgs = append(filterArgs, *filter.RepoID)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			repo_id,
			repo_path,
			commit_hash,
			branch,
			timestamp,
			status_id,
			source_id
		FROM repo_events
		WHERE %s
		ORDER BY id ASC
	`, strings.Join(filterClauses, " AND "))

	rows, err := db.Query(query, filterArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []RepoEvent

	for rows.Next() {
		e, err := scanRepoEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

// GetPendingEvents returns all events with status=pending for export.
func GetPendingEvents(db *sql.DB) ([]RepoEvent, error) {
	pending := StatusPending
	return ListEvents(db, EventFilter{Status: &pending})
}

// UpdateEventStatuses updates the status for a list of event IDs.
func UpdateEventStatuses(db *sql.DB, ids []int64, status Status) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids)+1)
	args[0] = int(status)

	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf(
		"UPDATE repo_events SET status_id = ? WHERE id IN (%s)",
		strings.Join(placeholders, ","),
	)

	_, err := db.Exec(query, args...)
	if err != nil {
		log.Error("store: update event statuses failed: %v (count=%d)", err, len(ids))
	}
	return err
}

// MigratePendingRepoID updates repo_id for all pending events from oldID to newID.
// Returns the number of events updated.
func MigratePendingRepoID(db *sql.DB, oldID, newID string) (int64, error) {
	query := `
		UPDATE repo_events
		SET repo_id = ?
		WHERE repo_id = ? AND status_id = ?
	`
	result, err := db.Exec(query, newID, oldID, int(StatusPending))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
