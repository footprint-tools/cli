package hooks

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoHookStatus represents the classification of a repository's hook state.
type RepoHookStatus int

const (
	StatusClean RepoHookStatus = iota
	StatusManagedPreCommit
	StatusManagedHusky
	StatusManagedLefthook
	StatusUnmanagedHooks
	StatusGlobalHooksActive
)

func (s RepoHookStatus) String() string {
	switch s {
	case StatusClean:
		return "Clean"
	case StatusManagedPreCommit:
		return "Managed: pre-commit"
	case StatusManagedHusky:
		return "Managed: husky"
	case StatusManagedLefthook:
		return "Managed: lefthook"
	case StatusUnmanagedHooks:
		return "Unmanaged hooks"
	case StatusGlobalHooksActive:
		return "Global hooks active"
	default:
		return "Unknown"
	}
}

// CanInstall returns true if Footprint can safely install hooks.
func (s RepoHookStatus) CanInstall() bool {
	return s == StatusClean
}

// RepoInspection contains detailed information about a repository's hook state.
type RepoInspection struct {
	Status          RepoHookStatus
	GlobalHooksPath string   // Value of core.hooksPath if set
	UnmanagedHooks  []string // List of unmanaged hook files found
	FpInstalled     bool     // Whether fp hooks are already installed
}

// InspectRepo performs a preflight inspection of a repository's hook state.
// It classifies the repo according to the detection rules in this exact order:
// 1. Global hooks override (core.hooksPath)
// 2. Known hook managers (pre-commit, husky, lefthook)
// 3. Unmanaged local hooks
// 4. Clean (none of the above)
func InspectRepo(repoPath string) RepoInspection {
	inspection := RepoInspection{}

	// 1. Check for global hooks override first (highest priority)
	globalPath := getGlobalHooksPath(repoPath)
	if globalPath != "" {
		inspection.Status = StatusGlobalHooksActive
		inspection.GlobalHooksPath = globalPath
		return inspection
	}

	// 2. Detect known hook managers
	if hasPreCommit(repoPath) {
		inspection.Status = StatusManagedPreCommit
		return inspection
	}

	if hasHusky(repoPath) {
		inspection.Status = StatusManagedHusky
		return inspection
	}

	if hasLefthook(repoPath) {
		inspection.Status = StatusManagedLefthook
		return inspection
	}

	// 3. Detect unmanaged local hooks
	hooksPath := filepath.Join(repoPath, ".git", "hooks")
	unmanaged := findUnmanagedHooks(hooksPath)
	if len(unmanaged) > 0 {
		// Check if these are fp hooks
		fpInstalled := areAllFpHooks(hooksPath, unmanaged)
		if fpInstalled {
			inspection.Status = StatusClean
			inspection.FpInstalled = true
			return inspection
		}
		inspection.Status = StatusUnmanagedHooks
		inspection.UnmanagedHooks = unmanaged
		return inspection
	}

	// 4. Clean - no hooks, no managers, no global override
	inspection.Status = StatusClean
	return inspection
}

// getGlobalHooksPath returns the value of core.hooksPath if set, empty string otherwise.
func getGlobalHooksPath(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "config", "--get", "core.hooksPath")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// hasPreCommit checks for .pre-commit-config.yaml in the repo root.
func hasPreCommit(repoPath string) bool {
	configPath := filepath.Join(repoPath, ".pre-commit-config.yaml")
	_, err := os.Stat(configPath)
	return err == nil
}

// hasHusky checks for .husky/ directory OR package.json with husky section.
func hasHusky(repoPath string) bool {
	// Check for .husky directory
	huskyDir := filepath.Join(repoPath, ".husky")
	if info, err := os.Stat(huskyDir); err == nil && info.IsDir() {
		return true
	}

	// Check package.json for husky section
	pkgPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}

	var pkg map[string]any
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	_, hasHuskyKey := pkg["husky"]
	return hasHuskyKey
}

// hasLefthook checks for lefthook.yml or lefthook.yaml in the repo root.
func hasLefthook(repoPath string) bool {
	for _, name := range []string{"lefthook.yml", "lefthook.yaml"} {
		configPath := filepath.Join(repoPath, name)
		if _, err := os.Stat(configPath); err == nil {
			return true
		}
	}
	return false
}

// findUnmanagedHooks scans .git/hooks/ for executable or non-empty hook files.
// It ignores *.sample files.
func findUnmanagedHooks(hooksPath string) []string {
	var hooks []string

	entries, err := os.ReadDir(hooksPath)
	if err != nil {
		return hooks
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip sample files
		if strings.HasSuffix(name, ".sample") {
			continue
		}

		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Skip fp backup directory marker or other non-hook files
		if name == ".fp-backup" {
			continue
		}

		hookPath := filepath.Join(hooksPath, name)
		info, err := os.Stat(hookPath)
		if err != nil {
			continue
		}

		// Check if executable or non-empty
		isExecutable := info.Mode()&0111 != 0
		isNonEmpty := info.Size() > 0

		if isExecutable || isNonEmpty {
			hooks = append(hooks, name)
		}
	}

	return hooks
}

// areAllFpHooks checks if all the hooks in the list are fp-managed hooks.
func areAllFpHooks(hooksPath string, hookNames []string) bool {
	if len(hookNames) == 0 {
		return false
	}

	for _, name := range hookNames {
		hookPath := filepath.Join(hooksPath, name)
		if !isFpHook(hookPath) {
			return false
		}
	}
	return true
}

// isFpHook checks if a hook file was installed by footprint.
func isFpHook(hookPath string) bool {
	data, err := os.ReadFile(hookPath)
	if err != nil {
		return false
	}
	content := string(data)
	// fp hooks contain this marker
	return strings.Contains(content, "fp record") || strings.Contains(content, "footprint")
}

// Guidance messages for manual integration

const GuidancePreCommit = `Footprint detected pre-commit in this repository.

To integrate Footprint with pre-commit, add a local hook to your
.pre-commit-config.yaml:

  - repo: local
    hooks:
      - id: footprint
        name: footprint
        entry: fp record post-commit
        language: system
        always_run: true
        pass_filenames: false
        stages: [post-commit]

Then run: pre-commit install --hook-type post-commit`

const GuidanceHusky = `Footprint detected Husky in this repository.

To integrate Footprint with Husky, add to your Husky hook files:

  # In .husky/post-commit (create if needed):
  fp record post-commit

  # In .husky/post-merge (create if needed):
  fp record post-merge

  # In .husky/post-checkout (create if needed):
  fp record post-checkout`

const GuidanceLefthook = `Footprint detected Lefthook in this repository.

To integrate Footprint with Lefthook, add to your lefthook.yml:

  post-commit:
    commands:
      footprint:
        run: fp record post-commit

  post-merge:
    commands:
      footprint:
        run: fp record post-merge

  post-checkout:
    commands:
      footprint:
        run: fp record post-checkout`

const GuidanceUnmanagedHooks = `Footprint found existing hooks in this repository that are not
managed by a known tool.

To install Footprint, you have these options:

  1. Remove or rename the existing hooks in .git/hooks/
     Then run: fp setup

  2. Manually add Footprint to your existing hooks by adding:
     fp record <hook-name>

     For example, in .git/hooks/post-commit add:
     fp record post-commit`

const GuidanceGlobalHooks = `This repository has core.hooksPath set, which means local hooks
in .git/hooks/ are ignored by Git.

Footprint will not install hooks because they would have no effect.

To use Footprint, you can either:

  1. Disable the global hooks path for this repo:
     git config --unset core.hooksPath

  2. Add Footprint to your global hooks directory:
     %s

     Add to each hook file:
     fp record <hook-name>`

// GetGuidance returns the appropriate guidance message for a status.
func GetGuidance(inspection RepoInspection) string {
	switch inspection.Status {
	case StatusManagedPreCommit:
		return GuidancePreCommit
	case StatusManagedHusky:
		return GuidanceHusky
	case StatusManagedLefthook:
		return GuidanceLefthook
	case StatusUnmanagedHooks:
		return GuidanceUnmanagedHooks
	case StatusGlobalHooksActive:
		return strings.Replace(GuidanceGlobalHooks, "%s", inspection.GlobalHooksPath, 1)
	default:
		return ""
	}
}
