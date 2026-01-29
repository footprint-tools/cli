package store

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/log"
	"github.com/footprint-tools/cli/internal/store/migrations"
)

// Store wraps a SQLite database connection for event storage.
// It implements the domain.EventStore interface.
type Store struct {
	db   *sql.DB
	path string
}

// New creates a new Store with the given database path.
// Runs migrations automatically.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Configure SQLite for better concurrency and reliability
	if err = configureSQLite(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure database: %w", err)
	}

	setDBPermissions(path)

	if err = migrations.Run(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &Store{db: db, path: path}, nil
}

// configureSQLite sets pragmas for better concurrency and reliability.
// WAL mode allows concurrent reads during writes and is more resilient to corruption.
// busy_timeout prevents "database is locked" errors during concurrent access.
func configureSQLite(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",   // Write-Ahead Logging for better concurrency
		"PRAGMA busy_timeout=5000",  // Wait up to 5 seconds if locked
		"PRAGMA synchronous=NORMAL", // Safe with WAL, better performance
		"PRAGMA foreign_keys=ON",    // Enforce foreign key constraints
		"PRAGMA cache_size=-64000",  // 64MB cache (negative = KB)
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("%s: %w", pragma, err)
		}
	}

	// SQLite works best with a single connection for writes
	db.SetMaxOpenConns(1)

	return nil
}

// NewWithDB creates a Store from an existing database connection.
// Useful for testing with pre-configured databases.
func NewWithDB(db *sql.DB) *Store {
	return &Store{db: db, path: ""}
}

// DB returns the underlying database connection.
// Use sparingly - prefer using Store methods.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// CloseDB closes a database connection and logs any errors.
// Intended for use in defer statements where errors would otherwise be ignored.
func CloseDB(db *sql.DB) {
	if db == nil {
		return
	}
	if err := db.Close(); err != nil {
		// Use fmt to stderr since log package may not be initialized
		fmt.Fprintf(os.Stderr, "store: close database: %v\n", err)
	}
}

// closeRows closes a rows iterator and logs any errors.
// Intended for use in defer statements where errors would otherwise be ignored.
func closeRows(rows *sql.Rows) {
	if rows == nil {
		return
	}
	if err := rows.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "store: close rows: %v\n", err)
	}
}

// setDBPermissions sets restrictive file permissions on the database and its WAL/SHM files.
func setDBPermissions(path string) {
	if path == ":memory:" {
		return
	}
	_ = os.Chmod(path, 0600)
	_ = os.Chmod(path+"-wal", 0600)
	_ = os.Chmod(path+"-shm", 0600)
}

// Insert adds a new event to the store.
func (s *Store) Insert(event domain.RepoEvent) error {
	_, err := s.db.Exec(
		`INSERT INTO repo_events
		 (repo_id, repo_path, commit_hash, branch, timestamp, status_id, source_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(repo_id, commit_hash, source_id)
		 DO UPDATE SET timestamp = excluded.timestamp`,
		event.RepoID.String(),
		event.RepoPath,
		event.Commit,
		event.Branch,
		event.Timestamp.Format(time.RFC3339),
		int(event.Status),
		int(event.Source),
	)
	return err
}

// List returns events matching the given filter.
func (s *Store) List(filter domain.EventFilter) ([]domain.RepoEvent, error) {
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

	if !filter.RepoID.IsEmpty() {
		filterClauses = append(filterClauses, "repo_id = ?")
		filterArgs = append(filterArgs, filter.RepoID.String())
	}

	if filter.SinceID > 0 {
		filterClauses = append(filterClauses, "id > ?")
		filterArgs = append(filterArgs, filter.SinceID)
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

	rows, err := s.db.Query(queryBuilder.String(), filterArgs...)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	var out []domain.RepoEvent

	for rows.Next() {
		e, err := s.scanRepoEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

// GetPending returns all events with pending status.
func (s *Store) GetPending() ([]domain.RepoEvent, error) {
	pending := domain.StatusPending
	return s.List(domain.EventFilter{Status: &pending})
}

// UpdateStatus updates the status of multiple events.
func (s *Store) UpdateStatus(ids []int64, status domain.EventStatus) error {
	if len(ids) == 0 {
		return nil
	}

	args := make([]any, len(ids)+1)
	args[0] = int(status)
	for i, id := range ids {
		args[i+1] = id
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	query := fmt.Sprintf("UPDATE repo_events SET status_id = ? WHERE id IN (%s)", placeholders)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		log.Error("store: update status failed: %v (count=%d)", err, len(ids))
	}
	return err
}

// GetMaxID returns the highest event ID.
func (s *Store) GetMaxID() (int64, error) {
	var maxID sql.NullInt64
	err := s.db.QueryRow("SELECT MAX(id) FROM repo_events").Scan(&maxID)
	if err != nil {
		return 0, err
	}
	if !maxID.Valid {
		return 0, nil
	}
	return maxID.Int64, nil
}

// ListSince returns events with ID greater than the given ID.
func (s *Store) ListSince(id int64) ([]domain.RepoEvent, error) {
	query := `
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
		WHERE id > ?
		ORDER BY id ASC
	`

	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	var out []domain.RepoEvent

	for rows.Next() {
		e, err := s.scanRepoEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

// MigrateRepoID changes the repo ID for all pending events.
func (s *Store) MigrateRepoID(oldID, newID domain.RepoID) (int64, error) {
	query := `
		UPDATE repo_events
		SET repo_id = ?
		WHERE repo_id = ? AND status_id = ?
	`
	result, err := s.db.Exec(query, newID.String(), oldID.String(), int(domain.StatusPending))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// scanRepoEvent scans a single row into a domain.RepoEvent.
func (s *Store) scanRepoEvent(rows *sql.Rows) (domain.RepoEvent, error) {
	var (
		e        domain.RepoEvent
		repoID   string
		ts       string
		statusID int
		sourceID int
	)

	if err := rows.Scan(
		&e.ID,
		&repoID,
		&e.RepoPath,
		&e.Commit,
		&e.Branch,
		&ts,
		&statusID,
		&sourceID,
	); err != nil {
		return domain.RepoEvent{}, err
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return domain.RepoEvent{}, err
	}

	e.RepoID = domain.RepoID(repoID)
	e.Timestamp = t
	e.Status = domain.EventStatus(statusID)
	e.Source = domain.EventSource(sourceID)

	return e, nil
}

// MarkOrphaned marks all events for a repo as orphaned.
// Returns the number of events updated.
func (s *Store) MarkOrphaned(repoID domain.RepoID) (int64, error) {
	query := `
		UPDATE repo_events
		SET status_id = ?
		WHERE repo_id = ? AND status_id = ?
	`
	result, err := s.db.Exec(query, int(domain.StatusOrphaned), repoID.String(), int(domain.StatusPending))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteOrphaned deletes all orphaned events from the database.
// Returns the number of events deleted.
func (s *Store) DeleteOrphaned() (int64, error) {
	query := `DELETE FROM repo_events WHERE status_id = ?`
	result, err := s.db.Exec(query, int(domain.StatusOrphaned))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CountOrphaned returns the count of orphaned events.
func (s *Store) CountOrphaned() (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM repo_events WHERE status_id = ?`
	err := s.db.QueryRow(query, int(domain.StatusOrphaned)).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ListDistinctRepos returns all unique repository IDs that have recorded events.
func (s *Store) ListDistinctRepos() ([]domain.RepoID, error) {
	query := `SELECT DISTINCT repo_id FROM repo_events ORDER BY repo_id`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	var repos []domain.RepoID
	for rows.Next() {
		var repoID string
		if err := rows.Scan(&repoID); err != nil {
			return nil, err
		}
		repos = append(repos, domain.RepoID(repoID))
	}

	return repos, rows.Err()
}

// Verify Store implements domain.EventStore
var _ domain.EventStore = (*Store)(nil)
