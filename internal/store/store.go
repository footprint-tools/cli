package store

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Skryensya/footprint/internal/domain"
	"github.com/Skryensya/footprint/internal/store/migrations"
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
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	setDBPermissions(path)

	if err = migrations.Run(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &Store{db: db, path: path}, nil
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
		clauses []string
		args    []any
	)

	if filter.Status != nil {
		clauses = append(clauses, "status_id = ?")
		args = append(args, int(*filter.Status))
	}

	if filter.Source != nil {
		clauses = append(clauses, "source_id = ?")
		args = append(args, int(*filter.Source))
	}

	if filter.Since != nil {
		clauses = append(clauses, "timestamp >= ?")
		args = append(args, filter.Since.Format(time.RFC3339))
	}

	if filter.Until != nil {
		clauses = append(clauses, "timestamp <= ?")
		args = append(args, filter.Until.Format(time.RFC3339))
	}

	if !filter.RepoID.IsEmpty() {
		clauses = append(clauses, "repo_id = ?")
		args = append(args, filter.RepoID.String())
	}

	if filter.SinceID > 0 {
		clauses = append(clauses, "id > ?")
		args = append(args, filter.SinceID)
	}

	query := base

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	_, err := s.db.Exec(query, args...)
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
	defer rows.Close()

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

// Verify Store implements domain.EventStore
var _ domain.EventStore = (*Store)(nil)
