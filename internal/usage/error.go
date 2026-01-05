package usage

type Error struct {
	Message  string
	ExitCode int
}

func (e *Error) Error() string {
	return e.Message
}
