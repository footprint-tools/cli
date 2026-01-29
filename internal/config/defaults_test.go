package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		configLines []string
		key         string
		wantValue   string
		wantFound   bool
	}{
		{
			name:        "key exists in config file",
			configLines: []string{"export_interval_sec=7200"},
			key:         "export_interval_sec",
			wantValue:   "7200",
			wantFound:   true,
		},
		{
			name:        "key exists in defaults but not in file",
			configLines: []string{},
			key:         "export_interval_sec",
			wantValue:   "3600",
			wantFound:   true,
		},
		{
			name:        "key exists in defaults - enable_log",
			configLines: []string{},
			key:         "enable_log",
			wantValue:   "true",
			wantFound:   true,
		},
		{
			name:        "key exists in defaults - theme",
			configLines: []string{},
			key:         "theme",
			wantValue:   "default",
			wantFound:   true,
		},
		{
			name:        "config overrides default",
			configLines: []string{"export_interval_sec=1800"},
			key:         "export_interval_sec",
			wantValue:   "1800",
			wantFound:   true,
		},
		{
			name:        "key not in config or defaults",
			configLines: []string{},
			key:         "nonexistent_key",
			wantValue:   "",
			wantFound:   false,
		},
		{
			name:        "custom key in config",
			configLines: []string{"custom_key=custom_value"},
			key:         "custom_key",
			wantValue:   "custom_value",
			wantFound:   true,
		},
		{
			name: "config file with multiple keys",
			configLines: []string{
				"export_interval_sec=1800",
				"theme=neon",
				"custom=value",
			},
			key:       "theme",
			wantValue: "neon",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp HOME
			tempHome := t.TempDir()
			configPath := filepath.Join(tempHome, ".fprc")

			// Write config file if needed
			if len(tt.configLines) > 0 {
				content := ""
				for _, line := range tt.configLines {
					content += line + "\n"
				}
				err := os.WriteFile(configPath, []byte(content), 0600)
				require.NoError(t, err)
			}

			// Override HOME
			oldHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tempHome)
			t.Cleanup(func() {
				_ = os.Setenv("HOME", oldHome)
			})

			// Test Get
			gotValue, gotFound := Get(tt.key)
			require.Equal(t, tt.wantFound, gotFound, "found mismatch")
			if tt.wantFound {
				require.Equal(t, tt.wantValue, gotValue, "value mismatch")
			}
		})
	}
}

func TestGet_EmptyConfigFile(t *testing.T) {
	tempHome := t.TempDir()

	// Override HOME but don't create config file
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempHome)
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})

	// Should return default value
	value, found := Get("export_interval_sec")
	require.True(t, found)
	require.Equal(t, "3600", value)
}

func TestGet_AllDefaults(t *testing.T) {
	tempHome := t.TempDir()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempHome)
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})

	// Test all default keys exist
	defaultKeys := []string{
		"export_interval_sec",
		"export_path",
		"export_last",
		"theme",
		"enable_log",
	}

	for _, key := range defaultKeys {
		t.Run(key, func(t *testing.T) {
			value, found := Get(key)
			require.True(t, found, "default key %s should be found", key)
			require.NotEmpty(t, value, "default value for %s should not be empty", key)
		})
	}
}

func TestGetAll(t *testing.T) {
	tests := []struct {
		name         string
		configLines  []string
		wantContains map[string]string
	}{
		{
			name:        "empty config returns all defaults",
			configLines: []string{},
			wantContains: map[string]string{
				"export_interval_sec": "3600",
				"enable_log":          "true",
				"theme":               "default",
			},
		},
		{
			name: "config overrides some defaults",
			configLines: []string{
				"export_interval_sec=7200",
				"theme=neon",
			},
			wantContains: map[string]string{
				"export_interval_sec": "7200",
				"theme":               "neon",
				"enable_log":          "true", // Still default
			},
		},
		{
			name: "config has custom keys",
			configLines: []string{
				"custom_key1=value1",
				"custom_key2=value2",
			},
			wantContains: map[string]string{
				"custom_key1":         "value1",
				"custom_key2":         "value2",
				"export_interval_sec": "3600", // Default still present
			},
		},
		{
			name: "config overrides and adds custom",
			configLines: []string{
				"export_interval_sec=1800",
				"custom=value",
			},
			wantContains: map[string]string{
				"export_interval_sec": "1800",
				"custom":              "value",
				"enable_log":          "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp HOME
			tempHome := t.TempDir()
			configPath := filepath.Join(tempHome, ".fprc")

			// Write config file if needed
			if len(tt.configLines) > 0 {
				content := ""
				for _, line := range tt.configLines {
					content += line + "\n"
				}
				err := os.WriteFile(configPath, []byte(content), 0600)
				require.NoError(t, err)
			}

			// Override HOME
			oldHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tempHome)
			t.Cleanup(func() {
				_ = os.Setenv("HOME", oldHome)
			})

			// Test GetAll
			got, err := GetAll()
			require.NoError(t, err)
			require.NotEmpty(t, got, "should return at least defaults")

			// Verify expected keys and values
			for key, expectedValue := range tt.wantContains {
				actualValue, exists := got[key]
				require.True(t, exists, "key %s should exist", key)
				require.Equal(t, expectedValue, actualValue, "value for key %s mismatch", key)
			}
		})
	}
}

func TestGetAll_NoConfigFile(t *testing.T) {
	tempHome := t.TempDir()

	// Override HOME but don't create config file
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempHome)
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})

	// Should return all defaults
	got, err := GetAll()
	require.NoError(t, err)
	require.NotEmpty(t, got)

	// Verify all defaults are present
	require.Contains(t, got, "export_interval_sec")
	require.Contains(t, got, "enable_log")
	require.Contains(t, got, "theme")
	require.Contains(t, got, "export_last")
	require.Contains(t, got, "export_path")

	// Verify values
	require.Equal(t, "3600", got["export_interval_sec"])
	require.Equal(t, "true", got["enable_log"])
}

func TestGetAll_OnlyReturnsDefaults(t *testing.T) {
	tempHome := t.TempDir()

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempHome)
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})

	got, err := GetAll()
	require.NoError(t, err)

	// Should have exactly the number of default keys (no config file)
	expectedDefaultCount := len(Defaults)
	require.Len(t, got, expectedDefaultCount)
}

func TestGetAll_MergesCorrectly(t *testing.T) {
	tempHome := t.TempDir()
	configPath := filepath.Join(tempHome, ".fprc")

	// Create config with 2 defaults overridden + 1 custom
	configLines := []string{
		"export_interval_sec=9999",
		"theme=neon",
		"my_custom_setting=custom_value",
	}
	content := ""
	for _, line := range configLines {
		content += line + "\n"
	}
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempHome)
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})

	got, err := GetAll()
	require.NoError(t, err)

	// Should have: all defaults + 1 custom = len(Defaults) + 1
	expectedCount := len(Defaults) + 1
	require.Len(t, got, expectedCount)

	// Verify overrides
	require.Equal(t, "9999", got["export_interval_sec"])
	require.Equal(t, "neon", got["theme"])

	// Verify custom
	require.Equal(t, "custom_value", got["my_custom_setting"])

	// Verify non-overridden defaults still present
	require.Equal(t, "true", got["enable_log"])
}
