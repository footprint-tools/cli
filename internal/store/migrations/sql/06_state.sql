-- App state (singleton row)
CREATE TABLE IF NOT EXISTS state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    update_last_check TEXT,
    update_latest_version TEXT
);

-- Insert the singleton row
INSERT OR IGNORE INTO state (id) VALUES (1);
