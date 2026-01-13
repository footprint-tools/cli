package actions

import (
	"fmt"

	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/hooks"
)

func HooksStatus(_ []string, flags []string) error {
	global := hasFlag(flags, "--global")

	var (
		hooksPath string
		scope     string
		err       error
	)

	if global {
		hooksPath, err = git.GlobalHooksPath()
		scope = "global"
	} else {
		root, err := git.RepoRoot(".")
		if err != nil {
			return nil
		}
		hooksPath, err = git.RepoHooksPath(root)
		scope = "repository"
	}

	if err != nil {
		return err
	}

	fmt.Printf("hooks scope: %s\n\n", scope)

	status := hooks.Status(hooksPath)

	for hook, installed := range status {
		if installed {
			fmt.Printf("%-14s installed\n", hook)
		} else {
			fmt.Printf("%-14s missing\n", hook)
		}
	}

	return nil
}
