package usage

import (
	"fmt"
	"strings"
)

func UnknownCommand(command string, suggestions ...string) *Error {
	var msg strings.Builder
	fmt.Fprintf(&msg, "fp: '%s' is not a fp command. See 'fp --help'.", command)

	if len(suggestions) > 0 {
		msg.WriteString("\n\n")
		if len(suggestions) == 1 {
			msg.WriteString("Did you mean this?\n")
		} else {
			msg.WriteString("Did you mean one of these?\n")
		}
		for _, s := range suggestions {
			fmt.Fprintf(&msg, "    %s\n", s)
		}
	}

	return &Error{
		Message:  strings.TrimSuffix(msg.String(), "\n"),
		ExitCode: 1,
	}
}
