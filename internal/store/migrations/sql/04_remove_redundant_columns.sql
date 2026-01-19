-- Remove commit_message and author columns (data is fetched from git during export)
ALTER TABLE repo_events DROP COLUMN commit_message;
ALTER TABLE repo_events DROP COLUMN author;
