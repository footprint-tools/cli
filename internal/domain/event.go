package domain

import (
	"strings"
	"time"
)

// EventStatus represents the export status of an event.
// STABLE: Integer values are persisted to SQLite. Do not change existing values.
// New statuses should be added with new integer values only.
type EventStatus int

const (
	StatusPending  EventStatus = 0 // stable
	StatusExported EventStatus = 1 // stable
	StatusOrphaned EventStatus = 2 // stable - repo no longer tracked
	StatusSkipped  EventStatus = 3 // stable
)

// String returns the string representation of the status.
func (s EventStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusExported:
		return "EXPORTED"
	case StatusOrphaned:
		return "ORPHANED"
	case StatusSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

// ParseEventStatus parses a string into an EventStatus.
func ParseEventStatus(s string) (EventStatus, bool) {
	switch strings.ToUpper(s) {
	case "PENDING":
		return StatusPending, true
	case "EXPORTED":
		return StatusExported, true
	case "ORPHANED":
		return StatusOrphaned, true
	case "SKIPPED":
		return StatusSkipped, true
	default:
		return 0, false
	}
}

// EventSource represents the origin of an event.
// STABLE: Integer values are persisted to SQLite. Do not change existing values.
// New sources should be added with new integer values only.
type EventSource int

const (
	SourcePostCommit   EventSource = 0 // stable
	SourcePostRewrite  EventSource = 1 // stable
	SourcePostCheckout EventSource = 2 // stable
	SourcePostMerge    EventSource = 3 // stable
	SourcePrePush      EventSource = 4 // stable
	SourceManual       EventSource = 5 // stable
	SourceBackfill     EventSource = 6 // stable
)

// String returns the string representation of the source.
func (s EventSource) String() string {
	switch s {
	case SourcePostCommit:
		return "post-commit"
	case SourcePostRewrite:
		return "post-rewrite"
	case SourcePostCheckout:
		return "post-checkout"
	case SourcePostMerge:
		return "post-merge"
	case SourcePrePush:
		return "pre-push"
	case SourceManual:
		return "manual"
	case SourceBackfill:
		return "backfill"
	default:
		return "unknown"
	}
}

// ParseEventSource parses a string into an EventSource.
func ParseEventSource(s string) (EventSource, bool) {
	switch strings.ToUpper(s) {
	case "POST-COMMIT":
		return SourcePostCommit, true
	case "POST-REWRITE":
		return SourcePostRewrite, true
	case "POST-CHECKOUT":
		return SourcePostCheckout, true
	case "POST-MERGE":
		return SourcePostMerge, true
	case "PRE-PUSH":
		return SourcePrePush, true
	case "MANUAL":
		return SourceManual, true
	case "BACKFILL":
		return SourceBackfill, true
	default:
		return 0, false
	}
}

// RepoEvent represents a tracked git event.
type RepoEvent struct {
	ID        int64
	RepoID    RepoID
	RepoPath  string
	Commit    string
	Branch    string
	Timestamp time.Time
	Status    EventStatus
	Source    EventSource
}

// EventFilter specifies criteria for querying events.
type EventFilter struct {
	RepoID  RepoID
	Status  *EventStatus
	Source  *EventSource
	Since   *time.Time
	Until   *time.Time
	Limit   int
	SinceID int64
}
