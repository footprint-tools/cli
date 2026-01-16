package setup

import (
	"github.com/Skryensya/footprint/internal/usage"
)

func Teardown(args []string, flags []string) error {
	return teardown(args, flags, DefaultDeps())
}

func teardown(_ []string, flags []string, deps Deps) error {
	force := hasFlag(flags, "--force")
	global := hasFlag(flags, "--global")
	repo := hasFlag(flags, "--repo")

	if global && repo {
		return usage.InvalidFlag("cannot use both --repo and --global")
	}

	var hooksPath string
	var err error

	// Default to repo behavior unless --global is explicitly passed
	if global {
		hooksPath, err = deps.GlobalHooksPath()
	} else {
		root, err := deps.RepoRoot(".")
		if err != nil {
			return usage.NotInGitRepo()
		}
		hooksPath, err = deps.RepoHooksPath(root)
	}

	if err != nil {
		return err
	}

	if !force {
		deps.Println("fp will remove its git hooks")
		deps.Println("previous hooks will be restored if available")
		deps.Print("continue? [y/N]: ")

		var resp string
		deps.Scanln(&resp)
		if resp != "y" && resp != "yes" {
			return nil
		}
	}

	if err := deps.HooksUninstall(hooksPath); err != nil {
		return err
	}

	deps.Println("fp teardown complete")
	return nil
}
