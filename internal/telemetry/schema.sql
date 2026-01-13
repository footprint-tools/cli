CREATE TABLE IF NOT EXISTS event_status (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE
);

INSERT OR IGNORE INTO event_status (id, name) VALUES
	(0, 'pending'),
	(1, 'exported'),
	(2, 'orphaned'),
	(3, 'skipped');

CREATE TABLE IF NOT EXISTS event_source (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE
);

INSERT OR IGNORE INTO event_source (id, name) VALUES
	(0, 'post-commit'),
	(1, 'post-rewrite'),
	(2, 'post-checkout'),
	(3, 'post-merge'),
	(4, 'pre-push'),
	(5, 'manual');

CREATE TABLE IF NOT EXISTS repo_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repo_id TEXT NOT NULL,
	repo_path TEXT,
	commit_hash TEXT,
	commit_message TEXT,
	branch TEXT,
	timestamp TEXT NOT NULL,
	status_id INTEGER NOT NULL,
	source_id INTEGER NOT NULL,
	FOREIGN KEY(status_id) REFERENCES event_status(id),
	FOREIGN KEY(source_id) REFERENCES event_source(id),
	UNIQUE(repo_id, commit_hash, source_id)
);

CREATE INDEX IF NOT EXISTS idx_repo_events_timestamp
	ON repo_events(timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_repo_events_repo
	ON repo_events(repo_id);

CREATE INDEX IF NOT EXISTS idx_repo_events_status
	ON repo_events(status_id);

CREATE INDEX IF NOT EXISTS idx_repo_events_source
	ON repo_events(source_id);
