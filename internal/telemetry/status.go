package telemetry

type Status int

const (
	StatusPending  Status = 0
	StatusExported Status = 1
	StatusOrphaned Status = 2
	StatusSkipped  Status = 3
)

func (s Status) String() string {
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
