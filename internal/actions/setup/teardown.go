package setup

import (
	"fmt"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/hooks"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/usage"
)

func Teardown(args []string, flags *dispatchers.ParsedFlags) error {
	return teardown(args, flags, DefaultDeps())
}

func teardown(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	if flags.Has("--core-hooks-path") {
		return teardownGlobal(flags, deps)
	}
	return teardownLocal(args, flags, deps)
}

func teardownLocal(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	force := flags.Has("--force")
	dryRun := flags.Has("--dry-run")

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

	if dryRun {
		_, _ = deps.Println("dry-run: would remove hooks from:")
		_, _ = deps.Printf("  %s\n", hooksPath)
		_, _ = deps.Println("  previous hooks would be restored if available")
		return nil
	}

	if !force {
		_, _ = deps.Println("fp will remove its git hooks from this repository")
		_, _ = deps.Println("previous hooks will be restored if available")
		_, _ = deps.Print("continue? [y/N]: ")

		var resp string
		_, _ = deps.Scanln(&resp)
		if resp != "y" && resp != "yes" {
			return nil
		}
	}

	if err := deps.HooksUninstall(hooksPath); err != nil {
		return err
	}

	// Remove the repo from the store
	removeRepoFromStore(root)

	_, _ = deps.Println("hooks removed")
	return nil
}

func teardownGlobal(flags *dispatchers.ParsedFlags, deps Deps) error {
	force := flags.Has("--force")
	dryRun := flags.Has("--dry-run")

	// Check current global hooks status
	status := hooks.CheckGlobalHooksStatus()

	if !status.IsSet {
		_, _ = deps.Println("No global hooks are configured (core.hooksPath is not set)")
		return nil
	}

	if dryRun {
		_, _ = deps.Println("dry-run: would remove global hooks from:")
		_, _ = deps.Printf("  %s\n", status.Path)
		_, _ = deps.Println("  core.hooksPath would be unset")
		return nil
	}

	_, _ = deps.Println("")
	_, _ = deps.Println("This will remove global fp hooks and unset core.hooksPath.")
	_, _ = deps.Println("")
	_, _ = deps.Printf("Current hooks path: %s\n", status.Path)
	_, _ = deps.Println("")

	if status.IsFpManaged {
		_, _ = deps.Println("The hooks appear to be managed by fp.")
	} else if status.HasOtherHooks {
		_, _ = deps.Println("⚠  WARNING: Some hooks may not be fp hooks:")
		for _, h := range status.OtherHooks {
			_, _ = deps.Printf("   - %s\n", h)
		}
		_, _ = deps.Println("")
	}

	_, _ = deps.Println("After removal:")
	_, _ = deps.Println("  - Git will use local .git/hooks/ directories again")
	_, _ = deps.Println("  - New commits will NOT be tracked automatically")
	_, _ = deps.Println("  - You'll need to run 'fp setup' in each repo to track it")
	_, _ = deps.Println("")

	if !force {
		_, _ = deps.Print("Remove global hooks? [y/N]: ")

		var resp string
		_, _ = deps.Scanln(&resp)
		if resp != "y" && resp != "yes" {
			_, _ = deps.Println("Cancelled.")
			return nil
		}
	}

	// Uninstall global hooks
	if err := hooks.UninstallGlobal(status.Path); err != nil {
		return fmt.Errorf("failed to remove global hooks: %w", err)
	}

	_, _ = deps.Println("")
	_, _ = deps.Println("✓ Global hooks removed")
	_, _ = deps.Println("  core.hooksPath has been unset")

	return nil
}

func removeRepoFromStore(repoPath string) {
	s, err := store.New(store.DBPath())
	if err != nil {
		return
	}
	defer func() { _ = s.Close() }()
	_ = s.RemoveRepo(repoPath)
}
