package usage

import "fmt"

// InvalidFlag is returned when a flag is not valid in the current context.
func InvalidFlag(flag string) *Error {
	return &Error{
		Message:  fmt.Sprintf("fp: invalid flag '%s'", flag),
		ExitCode: 2,
	}
}
