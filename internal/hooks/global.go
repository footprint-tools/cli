package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/footprint-tools/cli/internal/log"
)

// GlobalHooksDir returns the default directory for global hooks.
// On Unix: ~/.config/git/hooks
// On macOS: ~/Library/Application Support/git/hooks (but we use ~/.config for consistency)
func GlobalHooksDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "git", "hooks"), nil
}

// GetCurrentGlobalHooksPath returns the current value of core.hooksPath, if set.
func GetCurrentGlobalHooksPath() string {
	cmd := exec.Command("git", "config", "--global", "--get", "core.hooksPath")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// SetGlobalHooksPath sets the global core.hooksPath configuration.
func SetGlobalHooksPath(path string) error {
	cmd := exec.Command("git", "config", "--global", "core.hooksPath", path)
	return cmd.Run()
}

// UnsetGlobalHooksPath removes the global core.hooksPath configuration.
func UnsetGlobalHooksPath() error {
	cmd := exec.Command("git", "config", "--global", "--unset", "core.hooksPath")
	return cmd.Run()
}

// InstallGlobal installs fp hooks in the global hooks directory and sets core.hooksPath.
func InstallGlobal(hooksDir string) error {
	log.Debug("hooks: installing globally to %s", hooksDir)

	// Ensure hooks directory exists with restrictive permissions
	if err := os.MkdirAll(hooksDir, 0700); err != nil {
		log.Error("hooks: failed to create global directory %s: %v", hooksDir, err)
		return err
	}

	// Install the hooks
	if err := Install(hooksDir); err != nil {
		return err
	}

	// Set core.hooksPath
	if err := SetGlobalHooksPath(hooksDir); err != nil {
		log.Error("hooks: failed to set core.hooksPath: %v", err)
		return err
	}

	log.Info("hooks: set core.hooksPath to %s", hooksDir)
	return nil
}

// UninstallGlobal removes fp hooks from the global directory and unsets core.hooksPath.
func UninstallGlobal(hooksDir string) error {
	log.Debug("hooks: uninstalling globally from %s", hooksDir)

	// Uninstall the hooks
	if err := Uninstall(hooksDir); err != nil {
		return err
	}

	// Unset core.hooksPath
	if err := UnsetGlobalHooksPath(); err != nil {
		// Ignore error if not set
		log.Debug("hooks: core.hooksPath was not set or failed to unset")
	}

	return nil
}

// GlobalHooksStatus describes the current state of global hooks.
type GlobalHooksStatus struct {
	// IsSet is true if core.hooksPath is configured
	IsSet bool
	// Path is the current value of core.hooksPath (empty if not set)
	Path string
	// IsFpManaged is true if the hooks in that path are fp hooks
	IsFpManaged bool
	// HasOtherHooks is true if there are non-fp hooks in the directory
	HasOtherHooks bool
	// OtherHooks lists any non-fp hooks found
	OtherHooks []string
}

// CheckGlobalHooksStatus returns the current state of global hooks.
func CheckGlobalHooksStatus() GlobalHooksStatus {
	status := GlobalHooksStatus{}

	path := GetCurrentGlobalHooksPath()
	if path == "" {
		return status
	}

	status.IsSet = true
	status.Path = path

	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return status
	}

	// Check hooks in the directory
	hooks := findUnmanagedHooks(path)
	if len(hooks) == 0 {
		return status
	}

	// Check if all hooks are fp hooks
	status.IsFpManaged = areAllFpHooks(path, hooks)
	if !status.IsFpManaged {
		status.HasOtherHooks = true
		// Find which hooks are not fp
		for _, hook := range hooks {
			hookPath := filepath.Join(path, hook)
			if !isFpHook(hookPath) {
				status.OtherHooks = append(status.OtherHooks, hook)
			}
		}
	}

	return status
}
