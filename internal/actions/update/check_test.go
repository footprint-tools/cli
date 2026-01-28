package update

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/footprint-tools/cli/internal/store"
	"github.com/stretchr/testify/require"
)

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Clean versions should stay the same
		{"v0.0.10", "v0.0.10"},
		{"v1.2.3", "v1.2.3"},

		// Dirty suffix should be removed
		{"v0.0.10-dirty", "v0.0.10"},

		// Git describe format: v{version}-{commits}-g{hash}
		{"v0.0.10-1-ge69cbeb", "v0.0.10"},
		{"v0.0.10-5-gabcdef0", "v0.0.10"},
		{"v1.2.3-42-g1234567", "v1.2.3"},

		// Git describe with dirty
		{"v0.0.10-1-ge69cbeb-dirty", "v0.0.10"},
		{"v1.2.3-10-gabc1234-dirty", "v1.2.3"},

		// Edge cases
		{"v0.0.10-beta", "v0.0.10-beta"},     // Pre-release tag, not git describe
		{"v0.0.10-rc1-dirty", "v0.0.10-rc1"}, // Pre-release with dirty
		{"v0.0.10-0-ge69cbeb", "v0.0.10"},    // Zero commits after tag

		// More edge cases for full coverage
		{"", ""},                             // Empty string
		{"v1", "v1"},                         // Simple version
		{"-dirty", ""},                       // Just dirty
		{"v1.0.0-g", "v1.0.0-g"},             // Incomplete git describe
		{"v1.0.0-alpha-gabc123", "v1.0.0-alpha-gabc123"}, // Non-numeric before -g
		{"v1.0.0-rc-g1234567", "v1.0.0-rc-g1234567"},     // Pre-release with -g pattern
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := cleanVersion(tc.input)
			if result != tc.expected {
				t.Errorf("cleanVersion(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestShouldCheckUpdate(t *testing.T) {
	tests := []struct {
		command  string
		expected bool
	}{
		// Commands that should NOT check for updates
		{"record", false},
		{"export", false},
		{"update", false},
		{"backfill", false},

		// Commands that SHOULD check for updates
		{"status", true},
		{"activity", true},
		{"setup", true},
		{"teardown", true},
		{"repos", true},
		{"config", true},
		{"version", true},
		{"help", true},
		{"", true}, // Empty command
	}

	for _, tc := range tests {
		t.Run(tc.command, func(t *testing.T) {
			result := ShouldCheckUpdate(tc.command)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckResult(t *testing.T) {
	// Test CheckResult struct
	result := &CheckResult{
		UpdateAvailable: true,
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
	}

	require.True(t, result.UpdateAvailable)
	require.Equal(t, "v1.0.0", result.CurrentVersion)
	require.Equal(t, "v1.1.0", result.LatestVersion)
}

func TestCheckInterval(t *testing.T) {
	// Verify the check interval is 24 hours
	require.Equal(t, 24*time.Hour, CheckInterval)
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	deps := CheckDependencies{
		CurrentVersion: "dev",
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable)
	require.Equal(t, "dev", result.CurrentVersion)
}

func TestCheckForUpdate_EmptyVersion(t *testing.T) {
	deps := CheckDependencies{
		CurrentVersion: "",
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable)
	require.Equal(t, "", result.CurrentVersion)
}

func TestCheckForUpdate_CachedVersionRecent(t *testing.T) {
	now := time.Now()

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     now.Add(-1 * time.Hour).Format(time.RFC3339), // 1 hour ago
				LatestVersion: "v1.1.0",
			}, nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.True(t, result.UpdateAvailable)
	require.Equal(t, "v1.0.0", result.CurrentVersion)
	require.Equal(t, "v1.1.0", result.LatestVersion)
}

func TestCheckForUpdate_CachedVersionSame(t *testing.T) {
	now := time.Now()

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     now.Add(-1 * time.Hour).Format(time.RFC3339),
				LatestVersion: "v1.0.0", // Same as current
			}, nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable)
	require.Equal(t, "v1.0.0", result.CurrentVersion)
}

func TestCheckForUpdate_CacheExpired(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": "v1.2.0"}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	setCacheCalled := false
	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     now.Add(-25 * time.Hour).Format(time.RFC3339), // 25 hours ago (expired)
				LatestVersion: "v1.1.0",
			}, nil
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			setCacheCalled = true
			require.Equal(t, "v1.2.0", latestVersion)
			return nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.True(t, result.UpdateAvailable)
	require.Equal(t, "v1.0.0", result.CurrentVersion)
	require.Equal(t, "v1.2.0", result.LatestVersion)
	require.True(t, setCacheCalled)
}

func TestCheckForUpdate_CacheError(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": "v1.2.0"}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{}, errors.New("cache error")
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			return nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.True(t, result.UpdateAvailable)
	require.Equal(t, "v1.2.0", result.LatestVersion)
}

func TestCheckForUpdate_HTTPError(t *testing.T) {
	now := time.Now()

	client := &mockHTTPClient{
		errors: map[string]error{
			apiURL + "/releases/latest": errors.New("network error"),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     now.Add(-25 * time.Hour).Format(time.RFC3339), // expired
				LatestVersion: "",
			}, nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable) // On error, return false
	require.Equal(t, "v1.0.0", result.CurrentVersion)
}

func TestCheckForUpdate_DirtyVersion(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": "v1.0.0"}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0-dirty",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{}, errors.New("no cache")
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			return nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable) // cleaned v1.0.0 == v1.0.0
	require.Equal(t, "v1.0.0-dirty", result.CurrentVersion)
}

func TestCheckForUpdate_GitDescribeVersion(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": "v1.0.0"}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0-5-gabcdef0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{}, errors.New("no cache")
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			return nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable) // cleaned v1.0.0 == v1.0.0
}

func TestFetchLatestVersionQuick_Success(t *testing.T) {
	releaseJSON := `{"tag_name": "v2.0.0"}`
	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	version, err := fetchLatestVersionQuick(client)
	require.NoError(t, err)
	require.Equal(t, "v2.0.0", version)
}

func TestFetchLatestVersionQuick_HTTPError(t *testing.T) {
	client := &mockHTTPClient{
		errors: map[string]error{
			apiURL + "/releases/latest": errors.New("network error"),
		},
	}

	_, err := fetchLatestVersionQuick(client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "network error")
}

func TestFetchLatestVersionQuick_NotFound(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(404, ""),
		},
	}

	_, err := fetchLatestVersionQuick(client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "status 404")
}

func TestFetchLatestVersionQuick_InvalidJSON(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, "not json"),
		},
	}

	_, err := fetchLatestVersionQuick(client)
	require.Error(t, err)
}

func TestFetchLatestVersionQuick_ReadError(t *testing.T) {
	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": {
				StatusCode: 200,
				Body:       io.NopCloser(&errorReader{}),
			},
		},
	}

	_, err := fetchLatestVersionQuick(client)
	require.Error(t, err)
}

// errorReader always returns an error on Read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestPrintUpdateNotice_NoUpdate(t *testing.T) {
	var stderr bytes.Buffer

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				LatestVersion: "v1.0.0", // Same version
			}, nil
		},
		Now:    time.Now,
		Stderr: &stderr,
	}

	printUpdateNotice(deps)
	require.Empty(t, stderr.String())
}

func TestPrintUpdateNotice_WithUpdate(t *testing.T) {
	var stderr bytes.Buffer

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				LatestVersion: "v1.1.0",
			}, nil
		},
		Now:    time.Now,
		Stderr: &stderr,
	}

	printUpdateNotice(deps)
	output := stderr.String()
	require.Contains(t, output, "Update available")
	require.Contains(t, output, "v1.0.0")
	require.Contains(t, output, "v1.1.0")
	require.Contains(t, output, "fp update")
}

func TestNewCheckDependencies(t *testing.T) {
	deps := NewCheckDependencies()

	require.NotNil(t, deps.HTTPClient)
	require.NotNil(t, deps.GetUpdateCache)
	require.NotNil(t, deps.SetUpdateCache)
	require.NotNil(t, deps.Now)
	require.NotNil(t, deps.Stderr)
}

func TestCheckForUpdate_AlreadyAtLatest(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": "v1.0.0"}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{}, errors.New("no cache")
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			return nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable)
	require.Equal(t, "v1.0.0", result.CurrentVersion)
	require.Equal(t, "v1.0.0", result.LatestVersion)
}

func TestCheckForUpdate_InvalidCacheTime(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": "v1.1.0"}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     "invalid-time-format",
				LatestVersion: "v1.0.5",
			}, nil
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			return nil
		},
		Now: func() time.Time { return now },
	}

	// With invalid time format, it should fetch fresh
	result := checkForUpdate(deps)
	require.True(t, result.UpdateAvailable)
	require.Equal(t, "v1.1.0", result.LatestVersion)
}

func TestCheckForUpdate_EmptyLatestVersion(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": ""}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{}, errors.New("no cache")
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			return nil
		},
		Now: func() time.Time { return now },
	}

	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable) // Empty latest version = no update
}

func TestCheckForUpdate_CachedEmptyLastCheck(t *testing.T) {
	now := time.Now()
	releaseJSON := `{"tag_name": "v1.1.0"}`

	client := &mockHTTPClient{
		responses: map[string]*http.Response{
			apiURL + "/releases/latest": newMockResponse(200, releaseJSON),
		},
	}

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		HTTPClient:     client,
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     "", // Empty
				LatestVersion: "v1.0.5",
			}, nil
		},
		SetUpdateCache: func(lastCheck, latestVersion string) error {
			return nil
		},
		Now: func() time.Time { return now },
	}

	// With empty last check, it should fetch fresh
	result := checkForUpdate(deps)
	require.True(t, result.UpdateAvailable)
	require.Equal(t, "v1.1.0", result.LatestVersion)
}

func TestCheckForUpdate_CachedEmptyLatestVersion(t *testing.T) {
	now := time.Now()

	deps := CheckDependencies{
		CurrentVersion: "v1.0.0",
		GetUpdateCache: func() (store.UpdateCache, error) {
			return store.UpdateCache{
				LastCheck:     now.Add(-1 * time.Hour).Format(time.RFC3339),
				LatestVersion: "", // Empty cached version
			}, nil
		},
		Now: func() time.Time { return now },
	}

	// With empty cached latest version, no update available (within cache period)
	result := checkForUpdate(deps)
	require.False(t, result.UpdateAvailable)
}

func TestPrintUpdateNotice_DevVersion(t *testing.T) {
	var stderr bytes.Buffer

	deps := CheckDependencies{
		CurrentVersion: "dev",
		Stderr:         &stderr,
	}

	printUpdateNotice(deps)
	require.Empty(t, stderr.String()) // Dev version never shows updates
}

func TestNewCheckDependencies_Closures(t *testing.T) {
	deps := NewCheckDependencies()

	// Test GetUpdateCache - will use real store
	// This may return an error if store doesn't exist, but that's OK
	_, _ = deps.GetUpdateCache()

	// Test SetUpdateCache - will use real store
	_ = deps.SetUpdateCache("2024-01-01T00:00:00Z", "v1.0.0")

	// Test Now
	now := deps.Now()
	require.False(t, now.IsZero())

	// Test Stderr is set
	require.NotNil(t, deps.Stderr)

	// Test CurrentVersion is set (might be empty in test context)
	// Just verify it doesn't panic
	_ = deps.CurrentVersion

	// Test HTTPClient is set
	require.NotNil(t, deps.HTTPClient)
}
