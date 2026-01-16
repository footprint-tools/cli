package setup

import "github.com/Skryensya/footprint/internal/usage"

func Check(args []string, flags []string) error {
	return check(args, flags, DefaultDeps())
}

func check(_ []string, flags []string, deps Deps) error {
	global := hasFlag(flags, "--global")
	repo := hasFlag(flags, "--repo")

	if global && repo {
		return usage.InvalidFlag("cannot use both --repo and --global")
	}

	var (
		hooksPath string
		scope     string
		err       error
	)

	// Default to repo behavior unless --global is explicitly passed
	if global {
		hooksPath, err = deps.GlobalHooksPath()
		scope = "global"
	} else {
		root, err := deps.RepoRoot(".")
		if err != nil {
			return nil
		}
		hooksPath, err = deps.RepoHooksPath(root)
		scope = "repository"
	}

	if err != nil {
		return err
	}

	deps.Printf("hooks scope: %s\n\n", scope)

	status := deps.HooksStatus(hooksPath)

	for hook, installed := range status {
		if installed {
			deps.Printf("%-14s installed\n", hook)
		} else {
			deps.Printf("%-14s missing\n", hook)
		}
	}

	return nil
}
