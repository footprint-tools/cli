package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/footprint-tools/footprint-cli/internal/app"
	"github.com/footprint-tools/footprint-cli/internal/config"
	"github.com/footprint-tools/footprint-cli/internal/ui/style"
)

const (
	// CheckInterval is how often to check for updates (24 hours)
	CheckInterval = 24 * time.Hour
)

// CheckResult contains the result of an update check
type CheckResult struct {
	UpdateAvailable bool
	CurrentVersion  string
	LatestVersion   string
}

// CheckForUpdate checks if a newer version is available.
// Uses cached result if checked recently (within CheckInterval).
func CheckForUpdate() *CheckResult {
	currentVersion := app.Version
	if currentVersion == "" || currentVersion == "dev" {
		return &CheckResult{UpdateAvailable: false, CurrentVersion: currentVersion}
	}

	// Check if we should skip (checked recently)
	lastCheck, _ := config.Get("update_last_check")
	if lastCheck != "" {
		if t, err := time.Parse(time.RFC3339, lastCheck); err == nil {
			if time.Since(t) < CheckInterval {
				// Use cached version
				cachedVersion, _ := config.Get("update_latest_version")
				if cachedVersion != "" && cachedVersion != currentVersion {
					return &CheckResult{
						UpdateAvailable: true,
						CurrentVersion:  currentVersion,
						LatestVersion:   cachedVersion,
					}
				}
				return &CheckResult{UpdateAvailable: false, CurrentVersion: currentVersion}
			}
		}
	}

	// Fetch latest version from GitHub (with short timeout to not block)
	latestVersion, err := fetchLatestVersionQuick()
	if err != nil {
		// On error, don't block - just skip the check
		return &CheckResult{UpdateAvailable: false, CurrentVersion: currentVersion}
	}

	// Cache the result
	saveConfig("update_last_check", time.Now().Format(time.RFC3339))
	saveConfig("update_latest_version", latestVersion)

	// Compare versions
	updateAvailable := latestVersion != currentVersion && latestVersion != ""

	return &CheckResult{
		UpdateAvailable: updateAvailable,
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
	}
}

// fetchLatestVersionQuick fetches the latest release tag from GitHub with a short timeout
func fetchLatestVersionQuick() (string, error) {
	url := apiURL + "/releases/latest"

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch latest version (status %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// saveConfig is a helper to save a config value without error handling
// (we don't want update checks to fail if config can't be written)
func saveConfig(key, value string) {
	lines, err := config.ReadLines()
	if err != nil {
		return
	}
	lines, _ = config.Set(lines, key, value)
	_ = config.WriteLines(lines)
}

// PrintUpdateNotice prints an update notice if one is available.
// This is meant to be called before executing a command.
// It's non-blocking and won't print anything if no update is available.
func PrintUpdateNotice() {
	result := CheckForUpdate()
	if !result.UpdateAvailable {
		return
	}

	// Print colored notice
	notice := fmt.Sprintf("Update available: %s → %s",
		style.Muted(result.CurrentVersion),
		style.Success(result.LatestVersion))
	command := style.Info("fp update")

	fmt.Fprintf(os.Stderr, "\n  %s (run '%s')\n\n", notice, command)
}

// ShouldCheckUpdate returns true if the command should trigger an update check.
// Some commands (like record, export) are automated and should not show notices.
func ShouldCheckUpdate(command string) bool {
	// Commands that should NOT show update notices (automated/plumbing)
	skipCommands := map[string]bool{
		"record":   true, // Called by git hooks
		"export":   true, // Can be automated
		"update":   true, // Already updating
		"backfill": true, // Long-running process
	}
	return !skipCommands[command]
}
