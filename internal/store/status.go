package store

import "github.com/footprint-tools/cli/internal/domain"

// Status is an alias for domain.EventStatus.
// Kept for backward compatibility with existing code.
type Status = domain.EventStatus

// Status constants - aliases for domain constants.
const (
	StatusPending  = domain.StatusPending
	StatusExported = domain.StatusExported
	StatusOrphaned = domain.StatusOrphaned
	StatusSkipped  = domain.StatusSkipped
)
