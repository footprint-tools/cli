package update

import (
	"io"
	"net/http"
	"os"
	"os/exec"
)

type Dependencies struct {
	Stdout         io.Writer
	Stderr         io.Writer
	HTTPClient     HTTPClient
	CurrentVersion string
	ExecutablePath func() (string, error)
	RunCommand     func(name string, args ...string) error
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

func NewDependencies(version string) Dependencies {
	return Dependencies{
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
		HTTPClient:     http.DefaultClient,
		CurrentVersion: version,
		ExecutablePath: os.Executable,
		RunCommand: func(name string, args ...string) error {
			cmd := exec.Command(name, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}
