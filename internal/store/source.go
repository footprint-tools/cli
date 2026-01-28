package store

import "github.com/footprint-tools/cli/internal/domain"

// Source is an alias for domain.EventSource.
// Kept for backward compatibility with existing code.
type Source = domain.EventSource

// Source constants - aliases for domain constants.
const (
	SourcePostCommit   = domain.SourcePostCommit
	SourcePostRewrite  = domain.SourcePostRewrite
	SourcePostCheckout = domain.SourcePostCheckout
	SourcePostMerge    = domain.SourcePostMerge
	SourcePrePush      = domain.SourcePrePush
	SourceManual       = domain.SourceManual
	SourceBackfill     = domain.SourceBackfill
)
