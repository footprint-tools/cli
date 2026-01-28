package completions

import (
	"fmt"
	"os"
	"path/filepath"
)

// SourceInstructions returns shell-specific instructions for loading completions
func SourceInstructions(shell Shell) string {
	bin := GetBinaryPath()
	switch shell {
	case ShellBash:
		return fmt.Sprintf(`eval "$(%s completions --script)"`, bin)
	case ShellZsh:
		return fmt.Sprintf(`eval "$(%s completions --script)"`, bin)
	case ShellFish:
		return fmt.Sprintf(`%s completions --script | source`, bin)
	default:
		return ""
	}
}

// RcFile returns the rc file path for the given shell
func RcFile(shell Shell) string {
	switch shell {
	case ShellBash:
		return "~/.bashrc"
	case ShellZsh:
		return "~/.zshrc"
	case ShellFish:
		return "~/.config/fish/config.fish"
	default:
		return ""
	}
}

// AutoInstallPath returns the path where completions can be auto-loaded from.
// Returns empty string if auto-install is not supported for this shell.
func AutoInstallPath(shell Shell) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	bin := GetBinaryName()

	switch shell {
	case ShellFish:
		// Fish always auto-loads from this directory
		return filepath.Join(home, ".config", "fish", "completions", bin+".fish")
	case ShellBash:
		// Only if bash-completion is installed
		if IsBashCompletionInstalled() {
			return filepath.Join(home, ".local", "share", "bash-completion", "completions", bin)
		}
		return ""
	default:
		return ""
	}
}
