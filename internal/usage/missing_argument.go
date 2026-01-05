package usage

import "fmt"

// MissingArgument is returned when a required argument is not provided.
func MissingArgument(arg string) *Error {
	return &Error{
		Message:  fmt.Sprintf("fp: missing required argument '%s'", arg),
		ExitCode: 2,
	}
}
