package setup

import (
	"encoding/json"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/usage"
)

func Check(args []string, flags *dispatchers.ParsedFlags) error {
	return check(args, flags, DefaultDeps())
}

func check(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	jsonOutput := flags.Has("--json")

	root, err := deps.RepoRoot(".")
	if err != nil {
		return usage.NotInGitRepo()
	}

	hooksPath, err := deps.RepoHooksPath(root)
	if err != nil {
		return err
	}

	status := deps.HooksStatus(hooksPath)

	if jsonOutput {
		return checkJSON(root, hooksPath, status, deps)
	}

	installed := 0
	for hook, isInstalled := range status {
		if isInstalled {
			_, _ = deps.Printf("%-14s ✓ installed\n", hook)
			installed++
		} else {
			_, _ = deps.Printf("%-14s - not installed\n", hook)
		}
	}

	if installed == len(status) {
		_, _ = deps.Println("\nall hooks installed")
	} else if installed == 0 {
		_, _ = deps.Println("\nno hooks installed - run 'fp setup' to install")
	} else {
		_, _ = deps.Printf("\n%d/%d hooks installed\n", installed, len(status))
	}

	return nil
}

func checkJSON(repoRoot, hooksPath string, status map[string]bool, deps Deps) error {
	type hookStatus struct {
		Name      string `json:"name"`
		Installed bool   `json:"installed"`
	}

	type checkResult struct {
		RepoPath      string       `json:"repo_path"`
		HooksPath     string       `json:"hooks_path"`
		Hooks         []hookStatus `json:"hooks"`
		InstalledCount int         `json:"installed_count"`
		TotalCount    int          `json:"total_count"`
		AllInstalled  bool         `json:"all_installed"`
	}

	hooks := make([]hookStatus, 0, len(status))
	installed := 0
	for hook, isInstalled := range status {
		hooks = append(hooks, hookStatus{Name: hook, Installed: isInstalled})
		if isInstalled {
			installed++
		}
	}

	result := checkResult{
		RepoPath:       repoRoot,
		HooksPath:      hooksPath,
		Hooks:          hooks,
		InstalledCount: installed,
		TotalCount:     len(status),
		AllInstalled:   installed == len(status),
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	_, _ = deps.Println(string(data))
	return nil
}
