package setup

import (
	"fmt"
	"strings"

	"github.com/footprint-tools/cli/internal/completions"
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/hooks"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/usage"
)

func Setup(args []string, flags *dispatchers.ParsedFlags) error {
	return setup(args, flags, DefaultDeps())
}

func setup(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	if flags.Has("--core-hooks-path") {
		return setupGlobal(flags, deps)
	}
	return setupLocal(args, flags, deps)
}

func setupLocal(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	force := flags.Has("--force")

	// Determine target path
	targetPath := "."
	if len(args) > 0 && args[0] != "" {
		targetPath = args[0]
	}

	root, err := deps.RepoRoot(targetPath)
	if err != nil {
		return usage.NotInGitRepo()
	}

	hooksPath, err := deps.RepoHooksPath(root)
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
		_, _ = deps.Println("fp detected existing git hooks")
		_, _ = deps.Println("they will be backed up and replaced")
		_, _ = deps.Print("continue? [y/N]: ")

		var resp string
		_, _ = deps.Scanln(&resp)
		if resp != "y" && resp != "yes" {
			return nil
		}
	}

	if err := deps.HooksInstall(hooksPath); err != nil {
		return err
	}

	// Register the repo in the store
	addRepoToStore(root)

	if backedUp > 0 {
		_, _ = deps.Printf("installed %d hooks (%d backed up)\n", len(hooks.ManagedHooks), backedUp)
	} else {
		_, _ = deps.Printf("installed %d hooks\n", len(hooks.ManagedHooks))
	}
	_, _ = deps.Printf("  %s\n", strings.Join(hooks.ManagedHooks, ", "))

	// Install shell completions
	if result := completions.InstallSilently(); result != nil {
		if result.Installed {
			_, _ = deps.Printf("\nshell completions installed to %s\n", result.Path)
		} else if result.NeedsManual {
			_, _ = deps.Printf("\n%s\n", result.Instructions)
		}
	}

	return nil
}

func setupGlobal(flags *dispatchers.ParsedFlags, deps Deps) error {
	force := flags.Has("--force")

	// Get global hooks directory
	globalDir, err := hooks.GlobalHooksDir()
	if err != nil {
		return fmt.Errorf("failed to determine global hooks directory: %w", err)
	}

	// Check current global hooks status
	status := hooks.CheckGlobalHooksStatus()

	// Show warning banner
	_, _ = deps.Println("")
	_, _ = deps.Println("┌─────────────────────────────────────────────────────────────────┐")
	_, _ = deps.Println("│                    GLOBAL HOOKS INSTALLATION                    │")
	_, _ = deps.Println("└─────────────────────────────────────────────────────────────────┘")
	_, _ = deps.Println("")
	_, _ = deps.Println("This sets git's global core.hooksPath to use fp hooks.")
	_, _ = deps.Println("")
	_, _ = deps.Println("What this means:")
	_, _ = deps.Println("  - Repos WITHOUT local core.hooksPath will use fp automatically")
	_, _ = deps.Println("  - No need to run 'fp setup' in each of those repositories")
	_, _ = deps.Println("  - Local .git/hooks/ directories will be IGNORED by Git")
	_, _ = deps.Println("")
	_, _ = deps.Println("⚠  LIMITATION:")
	_, _ = deps.Println("  Repos with LOCAL core.hooksPath (Husky, etc.) will NOT be affected.")
	_, _ = deps.Println("  Local config always overrides global. For those repos, integrate")
	_, _ = deps.Println("  manually by adding 'fp record <hook>' to their hook files.")
	_, _ = deps.Println("")

	// Check for existing configuration
	if status.IsSet {
		_, _ = deps.Println("⚠  WARNING: core.hooksPath is already configured")
		_, _ = deps.Printf("   Current path: %s\n", status.Path)
		_, _ = deps.Println("")

		if status.IsFpManaged {
			_, _ = deps.Println("   The existing hooks appear to be fp hooks.")
			_, _ = deps.Println("   This will reinstall/update them.")
		} else if status.HasOtherHooks {
			_, _ = deps.Println("   ⚠  EXISTING HOOKS WILL BE OVERWRITTEN:")
			for _, h := range status.OtherHooks {
				_, _ = deps.Printf("      - %s\n", h)
			}
			_, _ = deps.Println("")
			_, _ = deps.Println("   These hooks will be backed up but may stop working.")
		}
		_, _ = deps.Println("")
	}

	_, _ = deps.Println("Hooks will be installed to:")
	_, _ = deps.Printf("  %s\n", globalDir)
	_, _ = deps.Println("")

	// Require explicit confirmation unless --force
	if !force {
		_, _ = deps.Println("To proceed, type 'yes' (not just 'y'):")
		_, _ = deps.Print("> ")

		var resp string
		_, _ = deps.Scanln(&resp)
		if resp != "yes" {
			_, _ = deps.Println("")
			_, _ = deps.Println("Cancelled. No changes were made.")
			_, _ = deps.Println("")
			_, _ = deps.Println("To install hooks in just the current repository, run:")
			_, _ = deps.Println("  fp setup")
			return nil
		}
	}

	// Install global hooks
	if err := hooks.InstallGlobal(globalDir); err != nil {
		return fmt.Errorf("failed to install global hooks: %w", err)
	}

	_, _ = deps.Println("")
	_, _ = deps.Println("✓ Global hooks installed successfully")
	_, _ = deps.Println("")
	_, _ = deps.Printf("  Hooks directory: %s\n", globalDir)
	_, _ = deps.Printf("  Hooks installed: %s\n", strings.Join(hooks.ManagedHooks, ", "))
	_, _ = deps.Println("")
	_, _ = deps.Println("All git repositories on this computer will now be tracked.")
	_, _ = deps.Println("")
	_, _ = deps.Println("To undo this, run:")
	_, _ = deps.Println("  fp teardown --global")

	// Install shell completions
	if result := completions.InstallSilently(); result != nil {
		if result.Installed {
			_, _ = deps.Printf("\nshell completions installed to %s\n", result.Path)
		} else if result.NeedsManual {
			_, _ = deps.Printf("\n%s\n", result.Instructions)
		}
	}

	return nil
}

func addRepoToStore(repoPath string) {
	s, err := store.New(store.DBPath())
	if err != nil {
		return
	}
	defer func() { _ = s.Close() }()
	_ = s.AddRepo(repoPath)
}
