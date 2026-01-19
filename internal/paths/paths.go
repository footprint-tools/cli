package paths

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Skryensya/footprint/internal/app"
)

const appDirName = "footprint"

// AppDataDir returns the application data directory for config/database.
// Uses os.UserConfigDir() which returns:
//   - macOS: ~/Library/Application Support
//   - Linux: $XDG_CONFIG_HOME or ~/.config
//   - Windows: %AppData% (roaming)
func AppDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}

	path := filepath.Join(dir, app.Name)

	_ = os.MkdirAll(path, 0755)

	return path
}

// AppLocalDataDir returns the OS-appropriate local data directory.
// This is where application-managed data (like exports) should live.
//   - macOS: ~/Library/Application Support/footprint
//   - Linux: $XDG_DATA_HOME/footprint or ~/.local/share/footprint
//   - Windows: %LOCALAPPDATA%\footprint
func AppLocalDataDir() string {
	var base string

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support
		home, err := os.UserHomeDir()
		if err != nil {
			return "."
		}
		base = filepath.Join(home, "Library", "Application Support")

	case "windows":
		// Windows: %LOCALAPPDATA%
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "."
			}
			base = filepath.Join(home, "AppData", "Local")
		}

	default:
		// Linux/Unix: $XDG_DATA_HOME or ~/.local/share
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "."
			}
			base = filepath.Join(home, ".local", "share")
		}
	}

	return filepath.Join(base, appDirName)
}

// ExportRepoDir returns the path to the export repository.
// The export repo is internal application data and lives inside AppLocalDataDir.
//   - macOS: ~/Library/Application Support/footprint/export
//   - Linux: $XDG_DATA_HOME/footprint/export or ~/.local/share/footprint/export
//   - Windows: %LOCALAPPDATA%\footprint\export
func ExportRepoDir() string {
	return filepath.Join(AppLocalDataDir(), "export")
}

func ConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".fprc"), nil
}
