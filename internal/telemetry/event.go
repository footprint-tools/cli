package telemetry

import "time"

type RepoEvent struct {
	RepoID        string
	RepoPath      string
	Commit        string
	CommitMessage string
	Branch        string
	Timestamp     time.Time
	Status        Status
	Source        Source
}
