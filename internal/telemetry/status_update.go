package telemetry

import "database/sql"

func UpdateStatus(
	db *sql.DB,
	repoID string,
	commit string,
	source Source,
	status Status,
) error {
	_, err := db.Exec(
		`UPDATE repo_events
		 SET status_id = ?
		 WHERE repo_id = ? AND commit_hash IS ? AND source_id = ?`,
		int(status),
		repoID,
		commit,
		int(source),
	)
	return err
}
