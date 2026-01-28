package setup

import (
	"fmt"

	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/hooks"
)

type Deps struct {
	// git
	RepoRoot        func(string) (string, error)
	RepoHooksPath   func(string) (string, error)
	GlobalHooksPath func() (string, error)

	// hooks
	HooksStatus    func(string) map[string]bool
	HooksInstall   func(string) error
	HooksUninstall func(string) error

	// io
	Printf  func(string, ...any) (int, error)
	Println func(...any) (int, error)
	Print   func(...any) (int, error)
	Scanln  func(...any) (int, error)
}

func DefaultDeps() Deps {
	return Deps{
		RepoRoot:        git.RepoRoot,
		RepoHooksPath:   git.RepoHooksPath,
		GlobalHooksPath: git.GlobalHooksPath,

		HooksStatus:    hooks.Status,
		HooksInstall:   hooks.Install,
		HooksUninstall: hooks.Uninstall,

		Printf:  fmt.Printf,
		Println: fmt.Println,
		Print:   fmt.Print,
		Scanln:  fmt.Scanln,
	}
}
