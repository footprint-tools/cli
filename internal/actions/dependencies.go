package actions

import (
	"fmt"

	"github.com/footprint-tools/cli/internal/app"
)

type actionDependencies struct {
	Printf  func(format string, a ...any) (n int, err error)
	Println func(a ...any) (n int, err error)
	Version func() string
}

func defaultDeps() actionDependencies {
	return actionDependencies{
		Printf:  fmt.Printf,
		Println: fmt.Println,
		Version: func() string { return app.Version },
	}
}
