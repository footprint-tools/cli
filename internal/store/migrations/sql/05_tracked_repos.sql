-- Table to track repositories where fp hooks are installed
-- repo_path is unique to allow multiple clones of the same repo (same repo_id) to be tracked independently
CREATE TABLE IF NOT EXISTS tracked_repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id TEXT NOT NULL,
    repo_path TEXT NOT NULL UNIQUE,
    added_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_seen TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_tracked_repos_repo_id ON tracked_repos(repo_id);
