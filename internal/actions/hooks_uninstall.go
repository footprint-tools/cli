package actions

import (
	"fmt"

	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/hooks"
	"github.com/Skryensya/footprint/internal/usage"
)

func HooksUninstall(_ []string, flags []string) error {
	force := hasFlag(flags, "--force")
	global := hasFlag(flags, "--global")
	repo := hasFlag(flags, "--repo")

	if global == repo {
		return usage.InvalidFlag("--repo | --global")
	}

	var hooksPath string
	var err error

	if repo {
		root, err := git.RepoRoot(".")
		if err != nil {
			return usage.NotInGitRepo()
		}
		hooksPath, err = git.RepoHooksPath(root)
	} else {
		hooksPath, err = git.GlobalHooksPath()
	}

	if err != nil {
		return err
	}

	if !force {
		fmt.Println("fp will remove its git hooks")
		fmt.Println("previous hooks will be restored if available")
		fmt.Print("continue? [y/N]: ")

		var resp string
		fmt.Scanln(&resp)
		if resp != "y" && resp != "yes" {
			return nil
		}
	}

	if err := hooks.Uninstall(hooksPath); err != nil {
		return err
	}

	fmt.Println("hooks uninstalled")
	return nil
}
