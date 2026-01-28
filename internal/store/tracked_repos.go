package store

import (
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/repo"
)

// RegisteredRepo represents a repository where fp hooks are installed.
type RegisteredRepo struct {
	Path     string
	AddedAt  string
	LastSeen string
}

// AddRepo registers a repository when hooks are installed.
func (s *Store) AddRepo(repoPath string) error {
	remoteURL, _ := git.OriginURL(repoPath)
	repoID, _ := repo.DeriveID(remoteURL, repoPath)

	_, err := s.db.Exec(`
		INSERT INTO tracked_repos (repo_id, repo_path, last_seen)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(repo_path) DO UPDATE SET
			repo_id = excluded.repo_id,
			last_seen = datetime('now')
	`, repoID, repoPath)
	return err
}

// RemoveRepo removes a repository when hooks are uninstalled.
func (s *Store) RemoveRepo(repoPath string) error {
	_, err := s.db.Exec(`DELETE FROM tracked_repos WHERE repo_path = ?`, repoPath)
	return err
}

// ListRepos returns all repositories with hooks installed.
func (s *Store) ListRepos() ([]RegisteredRepo, error) {
	rows, err := s.db.Query(`
		SELECT repo_path, added_at, last_seen
		FROM tracked_repos
		ORDER BY repo_path
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var repos []RegisteredRepo
	for rows.Next() {
		var r RegisteredRepo
		if err := rows.Scan(&r.Path, &r.AddedAt, &r.LastSeen); err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, rows.Err()
}

// ListRepoPaths returns just the paths of all repositories with hooks installed.
func (s *Store) ListRepoPaths() ([]string, error) {
	rows, err := s.db.Query(`SELECT repo_path FROM tracked_repos ORDER BY repo_path`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}
