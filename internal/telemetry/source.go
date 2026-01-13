package telemetry

type Source int

const (
	SourcePostCommit   Source = 0
	SourcePostRewrite  Source = 1
	SourcePostCheckout Source = 2
	SourcePostMerge    Source = 3
	SourcePrePush      Source = 4
	SourceManual       Source = 5
)

func (s Source) String() string {
	switch s {
	case SourcePostCommit:
		return "POST-COMMIT"
	case SourcePostRewrite:
		return "POST-REWRITE"
	case SourcePostCheckout:
		return "POST-CHECKOUT"
	case SourcePostMerge:
		return "POST-MERGE"
	case SourcePrePush:
		return "PRE-PUSH"
	case SourceManual:
		return "MANUAL"
	default:
		return "UNKNOWN"
	}
}
