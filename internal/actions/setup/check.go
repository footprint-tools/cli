package setup

import (
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/usage"
)

func Check(args []string, flags *dispatchers.ParsedFlags) error {
	return check(args, flags, DefaultDeps())
}

func check(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	root, err := deps.RepoRoot(".")
	if err != nil {
		return usage.NotInGitRepo()
	}

	hooksPath, err := deps.RepoHooksPath(root)
	if err != nil {
		return err
	}

	status := deps.HooksStatus(hooksPath)

	installed := 0
	for hook, isInstalled := range status {
		if isInstalled {
			_, _ = deps.Printf("%-14s ✓ installed\n", hook)
			installed++
		} else {
			_, _ = deps.Printf("%-14s - not installed\n", hook)
		}
	}

	if installed == len(status) {
		_, _ = deps.Println("\nall hooks installed")
	} else if installed == 0 {
		_, _ = deps.Println("\nno hooks installed - run 'fp setup' to install")
	} else {
		_, _ = deps.Printf("\n%d/%d hooks installed\n", installed, len(status))
	}

	return nil
}
