package setup

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/usage"
)

func Check(args []string, flags *dispatchers.ParsedFlags) error {
	return check(args, flags, DefaultDeps())
}

func check(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	repo := flags.Has("--repo")

	var (
		hooksPath string
		scope     string
		err       error
	)

	// Default to global behavior unless --repo is explicitly passed
	if repo {
		root, err := deps.RepoRoot(".")
		if err != nil {
			return usage.NotInGitRepo()
		}
		hooksPath, err = deps.RepoHooksPath(root)
		if err != nil {
			return err
		}
		scope = "repository"
	} else {
		hooksPath, err = deps.GlobalHooksPath()
		if err != nil {
			return err
		}
		scope = "global"
	}

	_, _ = deps.Printf("hooks scope: %s\n\n", scope)

	status := deps.HooksStatus(hooksPath)

	for hook, installed := range status {
		if installed {
			_, _ = deps.Printf("%-14s installed\n", hook)
		} else {
			_, _ = deps.Printf("%-14s missing\n", hook)
		}
	}

	return nil
}
