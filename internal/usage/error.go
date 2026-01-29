package usage

// ErrorKind represents the type of usage error.
type ErrorKind int

const (
	ErrUnknown ErrorKind = iota
	ErrInvalidFlag
	ErrMissingArgument
	ErrUnknownCommand
	ErrNotInGitRepo
	ErrInvalidRepo
	ErrInvalidPath
	ErrMissingRemote
	ErrAmbiguousRemote
	ErrGitNotInstalled
	ErrInvalidConfigKey
	ErrFailedConfigPath
)

// Exit codes:
//
//	Exit 1: Environment/system errors
//	  - Unknown errors
//	  - Unknown command
//	  - Not in git repo
//	  - Invalid repository
//	  - Invalid path
//	  - Git not installed
//	  - Invalid config key
//	  - Failed config path
//
//	Exit 2: User input errors
//	  - Invalid flag
//	  - Missing argument
//	  - Missing remote
//	  - Ambiguous remote
var exitCodes = map[ErrorKind]int{
	ErrUnknown:          1,
	ErrInvalidFlag:      2,
	ErrMissingArgument:  2,
	ErrUnknownCommand:   1,
	ErrNotInGitRepo:     1,
	ErrInvalidRepo:      1,
	ErrInvalidPath:      1,
	ErrMissingRemote:    2,
	ErrAmbiguousRemote:  2,
	ErrGitNotInstalled:  1,
	ErrInvalidConfigKey: 1,
	ErrFailedConfigPath: 1,
}

// Error represents a user-facing usage error with semantic type information.
type Error struct {
	Kind     ErrorKind
	Message  string
	ExitCode int // kept for backward compatibility, computed from Kind if zero
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// GetExitCode returns the appropriate exit code for this error.
// If ExitCode is explicitly set, it is returned; otherwise, the code is derived from Kind.
func (e *Error) GetExitCode() int {
	if e.ExitCode != 0 {
		return e.ExitCode
	}
	if code, ok := exitCodes[e.Kind]; ok {
		return code
	}
	return 1
}

// Verify Error implements the error interface.
var _ error = (*Error)(nil)
