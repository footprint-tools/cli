package actions

import (
	"fmt"

	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/hooks"
	"github.com/Skryensya/footprint/internal/usage"
)

func HooksInstall(_ []string, flags []string) error {
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
		fmt.Println("fp will take control of git hooks")
		fmt.Println("existing hooks will be backed up")
		fmt.Print("continue? [y/N]: ")

		var resp string
		fmt.Scanln(&resp)
		if resp != "y" && resp != "yes" {
			return nil
		}
	}

	if err := hooks.Install(hooksPath); err != nil {
		return err
	}

	fmt.Println("hooks installed")
	return nil
}
