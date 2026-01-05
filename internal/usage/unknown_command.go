package usage

import "fmt"

func UnknownCommand(command string) *Error {
	return &Error{
		Message:  fmt.Sprintf("fp: '%s' is not a fp command. See 'fp --help'.", command),
		ExitCode: 1,
	}
}
