package setup

import (
	"github.com/Skryensya/footprint/internal/usage"
)

func Setup(args []string, flags []string) error {
	return setup(args, flags, DefaultDeps())
}

func setup(_ []string, flags []string, deps Deps) error {
	force := hasFlag(flags, "--force")
	global := hasFlag(flags, "--global")
	repo := hasFlag(flags, "--repo")

	if global && repo {
		return usage.InvalidFlag("cannot use both --repo and --global")
	}

	var (
		hooksPath string
		err       error
	)

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

	needsConfirm := false
	status := deps.HooksStatus(hooksPath)

	for _, installed := range status {
		if installed {
			needsConfirm = true
			break
		}
	}

	if needsConfirm && !force {
		deps.Println("fp detected existing git hooks")
		deps.Println("they will be backed up and replaced")
		deps.Print("continue? [y/N]: ")

		var resp string
		deps.Scanln(&resp)
		if resp != "y" && resp != "yes" {
			return nil
		}
	}

	if err := deps.HooksInstall(hooksPath); err != nil {
		return err
	}

	deps.Println("fp setup complete")
	return nil
}
