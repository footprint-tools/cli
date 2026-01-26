package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/footprint-tools/footprint-cli/internal/app"
	"github.com/footprint-tools/footprint-cli/internal/store"
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

// CheckDependencies contains injectable dependencies for update checking
type CheckDependencies struct {
	CurrentVersion   string
	HTTPClient       HTTPClient
	GetUpdateCache   func() (store.UpdateCache, error)
	SetUpdateCache   func(lastCheck, latestVersion string) error
	Now              func() time.Time
	Stderr           io.Writer
}

// NewCheckDependencies creates CheckDependencies with default implementations
func NewCheckDependencies() CheckDependencies {
	return CheckDependencies{
		CurrentVersion: app.Version,
		HTTPClient:     &http.Client{Timeout: 3 * time.Second},
		GetUpdateCache: func() (store.UpdateCache, error) {
			s, err := store.New(store.DBPath())
			if err != nil {
				return store.UpdateCache{}, err
			}
			defer func() { _ = s.Close() }()
			return s.GetUpdateCache()
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			s, err := store.New(store.DBPath())
			if err != nil {
				return err
			}
			defer func() { _ = s.Close() }()
			return s.SetUpdateCache(lastCheck, latestVersion)
		},
		Now:    time.Now,
		Stderr: os.Stderr,
	}
}

// cleanVersion extracts the base semver from a git describe version.
// For example: "v0.0.10-1-ge69cbeb-dirty" -> "v0.0.10"
// The format from git describe is: v{semver}[-{commits}-g{hash}][-dirty]
func cleanVersion(v string) string {
	// Remove -dirty suffix (uncommitted changes)
	v = strings.TrimSuffix(v, "-dirty")

	// Remove git describe suffix: -{commits}-g{hash}
	// Look for pattern like "-1-g" followed by hex chars
	if idx := strings.LastIndex(v, "-g"); idx > 0 {
		// Find the commit count part before -g
		prefix := v[:idx]
		if dashIdx := strings.LastIndex(prefix, "-"); dashIdx > 0 {
			// Check if the part between dashes is a number (commit count)
			commitPart := prefix[dashIdx+1:]
			isNumber := true
			for _, c := range commitPart {
				if c < '0' || c > '9' {
					isNumber = false
					break
				}
			}
			if isNumber && len(commitPart) > 0 {
				v = prefix[:dashIdx]
			}
		}
	}
	return v
}

// CheckForUpdate checks if a newer version is available.
// Uses cached result if checked recently (within CheckInterval).
func CheckForUpdate() *CheckResult {
	return checkForUpdate(NewCheckDependencies())
}

// checkForUpdate is the internal implementation with dependency injection
func checkForUpdate(deps CheckDependencies) *CheckResult {
	currentVersion := deps.CurrentVersion
	if currentVersion == "" || currentVersion == "dev" {
		return &CheckResult{UpdateAvailable: false, CurrentVersion: currentVersion}
	}

	// Clean version for comparison (strip -dirty, etc.)
	cleanCurrent := cleanVersion(currentVersion)

	// Check if we should skip (checked recently)
	cache, err := deps.GetUpdateCache()
	if err == nil && cache.LastCheck != "" {
		if t, err := time.Parse(time.RFC3339, cache.LastCheck); err == nil {
			if deps.Now().Sub(t) < CheckInterval {
				// Use cached version
				if cache.LatestVersion != "" && cache.LatestVersion != cleanCurrent {
					return &CheckResult{
						UpdateAvailable: true,
						CurrentVersion:  currentVersion,
						LatestVersion:   cache.LatestVersion,
					}
				}
				return &CheckResult{UpdateAvailable: false, CurrentVersion: currentVersion}
			}
		}
	}

	// Fetch latest version from GitHub (with short timeout to not block)
	latestVersion, err := fetchLatestVersionQuick(deps.HTTPClient)
	if err != nil {
		// On error, don't block - just skip the check
		return &CheckResult{UpdateAvailable: false, CurrentVersion: currentVersion}
	}

	// Cache the result
	_ = deps.SetUpdateCache(deps.Now().Format(time.RFC3339), latestVersion)

	// Compare versions (use cleaned current version)
	updateAvailable := latestVersion != cleanCurrent && latestVersion != ""

	return &CheckResult{
		UpdateAvailable: updateAvailable,
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
	}
}

// fetchLatestVersionQuick fetches the latest release tag from GitHub with a short timeout
func fetchLatestVersionQuick(client HTTPClient) (string, error) {
	url := apiURL + "/releases/latest"

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


// PrintUpdateNotice prints an update notice if one is available.
// This is meant to be called before executing a command.
// It's non-blocking and won't print anything if no update is available.
func PrintUpdateNotice() {
	printUpdateNotice(NewCheckDependencies())
}

// printUpdateNotice is the internal implementation with dependency injection
func printUpdateNotice(deps CheckDependencies) {
	result := checkForUpdate(deps)
	if !result.UpdateAvailable {
		return
	}

	// Build notice text
	notice := fmt.Sprintf("Update available: %s → %s (run '%s')",
		style.Muted(result.CurrentVersion),
		style.Success(result.LatestVersion),
		style.Info("fp update"))

	// Print with border
	border := style.Border("─")
	line := strings.Repeat(border, 50)
	_, _ = fmt.Fprintf(deps.Stderr, "\n%s\n  %s\n%s\n\n", line, notice, line)
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
