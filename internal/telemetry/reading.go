package telemetry

import (
	"database/sql"
	"strings"
	"time"
)

func ListEvents(
	db *sql.DB,
	status *Status,
	source *Source,
) ([]RepoEvent, error) {

	base := `
		SELECT
			repo_id,
			repo_path,
			commit_hash,
			commit_message,
			branch,
			timestamp,
			status_id,
			source_id
		FROM repo_events
	`

	var (
		clauses []string
		args    []any
	)

	if status != nil {
		clauses = append(clauses, "status_id = ?")
		args = append(args, int(*status))
	}

	if source != nil {
		clauses = append(clauses, "source_id = ?")
		args = append(args, int(*source))
	}

	query := base

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RepoEvent

	for rows.Next() {
		var (
			e        RepoEvent
			ts       string
			statusID int
			sourceID int
		)

		if err := rows.Scan(
			&e.RepoID,
			&e.RepoPath,
			&e.Commit,
			&e.CommitMessage,
			&e.Branch,
			&ts,
			&statusID,
			&sourceID,
		); err != nil {
			return nil, err
		}

		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}

		e.Timestamp = t
		e.Status = Status(statusID)
		e.Source = Source(sourceID)

		out = append(out, e)
	}

	return out, rows.Err()
}
