package setup

import (
	"strings"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/hooks"
	"github.com/Skryensya/footprint/internal/usage"
)

func Setup(args []string, flags *dispatchers.ParsedFlags) error {
	return setup(args, flags, DefaultDeps())
}

func setup(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	force := flags.Has("--force")
	global := flags.Has("--global")
	repo := flags.Has("--repo")

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

	// Check existing hooks before install
	statusBefore := deps.HooksStatus(hooksPath)
	backedUp := 0
	for _, installed := range statusBefore {
		if installed {
			backedUp++
		}
	}

	if backedUp > 0 && !force {
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

	// Output summary
	location := "repository"
	if global {
		location = "global"
	}

	if backedUp > 0 {
		deps.Printf("Installed %d %s hooks (%d backed up)\n", len(hooks.ManagedHooks), location, backedUp)
	} else {
		deps.Printf("Installed %d %s hooks\n", len(hooks.ManagedHooks), location)
	}
	deps.Printf("  %s\n", strings.Join(hooks.ManagedHooks, ", "))

	if !global {
		deps.Println("")
		deps.Println("Run 'fp track' to start recording activity")
	}

	return nil
}
