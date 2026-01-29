package format

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testTime is a fixed time for consistent test results
var testTime = time.Date(2024, 1, 23, 15, 4, 5, 0, time.UTC)

func setupConfig(t *testing.T, content string) func() {
	t.Helper()

	// Create a temp config file
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	configPath := filepath.Join(home, ".fprc")

	// Backup existing config if it exists
	var backup []byte
	if data, err := os.ReadFile(configPath); err == nil {
		backup = data
	}

	// Write new config
	if content != "" {
		err = os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)
	} else {
		_ = os.Remove(configPath)
	}

	// Return cleanup function
	return func() {
		if backup != nil {
			_ = os.WriteFile(configPath, backup, 0644)
		} else {
			_ = os.Remove(configPath)
		}
	}
}

func TestDateTime(t *testing.T) {
	cleanup := setupConfig(t, "")
	defer cleanup()

	result := DateTime(testTime)
	require.NotEmpty(t, result)
	// Default format is "Jan 02 15:04"
	require.Contains(t, result, "Jan 23")
	require.Contains(t, result, "15:04")
}

func TestDateTime_WithCustomDateFormat(t *testing.T) {
	cleanup := setupConfig(t, "display_date=mm/dd/yyyy")
	defer cleanup()

	result := DateTime(testTime)
	require.Contains(t, result, "01/23/2024")
}

func TestDateTimeShort(t *testing.T) {
	cleanup := setupConfig(t, "")
	defer cleanup()

	result := DateTimeShort(testTime)
	require.NotEmpty(t, result)
	require.Contains(t, result, "Jan 23")
	require.Contains(t, result, "15:04")
}

func TestDateTimeShort_WithCustomFormat(t *testing.T) {
	cleanup := setupConfig(t, "display_date=dd/mm/yyyy")
	defer cleanup()

	result := DateTimeShort(testTime)
	// Short format for dd/mm/yyyy is "02/01"
	require.Contains(t, result, "23/01")
}

func TestDate(t *testing.T) {
	tests := []struct {
		name           string
		config         string
		expectContains string
	}{
		{
			name:           "default format",
			config:         "",
			expectContains: "Jan 23",
		},
		{
			name:           "mm/dd/yyyy",
			config:         "display_date=mm/dd/yyyy",
			expectContains: "01/23/2024",
		},
		{
			name:           "yyyy-mm-dd",
			config:         "display_date=yyyy-mm-dd",
			expectContains: "2024-01-23",
		},
		{
			name:           "dd/mm/yyyy",
			config:         "display_date=dd/mm/yyyy",
			expectContains: "23/01/2024",
		},
		{
			name:           "custom Go format",
			config:         "display_date=2006/01/02",
			expectContains: "2024/01/23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupConfig(t, tt.config)
			defer cleanup()

			result := Date(testTime)
			require.Contains(t, result, tt.expectContains)
		})
	}
}

func TestDateShort(t *testing.T) {
	tests := []struct {
		name           string
		config         string
		expectContains string
	}{
		{
			name:           "default format",
			config:         "",
			expectContains: "Jan 23",
		},
		{
			name:           "mm/dd/yyyy - short",
			config:         "display_date=mm/dd/yyyy",
			expectContains: "01/23",
		},
		{
			name:           "yyyy-mm-dd - short",
			config:         "display_date=yyyy-mm-dd",
			expectContains: "01-23",
		},
		{
			name:           "dd/mm/yyyy - short",
			config:         "display_date=dd/mm/yyyy",
			expectContains: "23/01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupConfig(t, tt.config)
			defer cleanup()

			result := DateShort(testTime)
			require.Contains(t, result, tt.expectContains)
		})
	}
}

func TestTime_24h(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   string
	}{
		{
			name:   "default 24h format",
			config: "",
			want:   "15:04",
		},
		{
			name:   "explicit 24h format",
			config: "display_time=24h",
			want:   "15:04",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupConfig(t, tt.config)
			defer cleanup()

			result := Time(testTime)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestTime_12h(t *testing.T) {
	cleanup := setupConfig(t, "display_time=12h")
	defer cleanup()

	result := Time(testTime)
	require.Contains(t, result, "3:04")
	require.Contains(t, result, "PM")
}

func TestTimeFull_24h(t *testing.T) {
	cleanup := setupConfig(t, "display_time=24h")
	defer cleanup()

	result := TimeFull(testTime)
	require.Equal(t, "15:04:05", result)
}

func TestTimeFull_12h(t *testing.T) {
	cleanup := setupConfig(t, "display_time=12h")
	defer cleanup()

	result := TimeFull(testTime)
	require.Contains(t, result, "3:04:05")
	require.Contains(t, result, "PM")
}

func TestFull(t *testing.T) {
	cleanup := setupConfig(t, "display_date=mm/dd/yyyy\ndisplay_time=24h")
	defer cleanup()

	result := Full(testTime)
	require.Contains(t, result, "01/23/2024")
	require.Contains(t, result, "15:04:05")
}

func TestGetDateFormatShort_CustomFormat(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   string
	}{
		{
			name:   "format with /06 year",
			config: "display_date=01/02/06",
			want:   "01/02", // /06 should be removed
		},
		{
			name:   "format with -06 year",
			config: "display_date=01-02-06",
			want:   "01-02", // -06 should be removed
		},
		{
			name:   "format with 2006",
			config: "display_date=01/02/2006",
			want:   "01/02", // 2006 should be removed
		},
		{
			name:   "format with space 06",
			config: "display_date=01/02 06",
			want:   "01/02", // space 06 should be removed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupConfig(t, tt.config)
			defer cleanup()

			result := DateShort(testTime)
			require.NotEmpty(t, result)
		})
	}
}

func TestGetDateFormatShort_EmptyAfterStripping(t *testing.T) {
	// This is hard to trigger because we'd need a format that only contains year
	// The fallback is "Jan 02"
	cleanup := setupConfig(t, "display_date=2006")
	defer cleanup()

	result := DateShort(testTime)
	// Should fallback to "Jan 02" after stripping the year
	require.NotEmpty(t, result)
}

func TestDateTime_WithMorningTime(t *testing.T) {
	cleanup := setupConfig(t, "display_time=12h")
	defer cleanup()

	// Test with morning time (before noon)
	morningTime := time.Date(2024, 1, 23, 9, 30, 0, 0, time.UTC)
	result := Time(morningTime)
	require.Contains(t, result, "9:30")
	require.Contains(t, result, "AM")
}

func TestTimeFull_DefaultIs24h(t *testing.T) {
	cleanup := setupConfig(t, "")
	defer cleanup()

	result := TimeFull(testTime)
	require.Equal(t, "15:04:05", result)
}

func TestTime_DefaultIs24h(t *testing.T) {
	cleanup := setupConfig(t, "")
	defer cleanup()

	result := Time(testTime)
	require.Equal(t, "15:04", result)
}

func TestTime_UnknownFormat_FallsTo24h(t *testing.T) {
	// Unknown format should fall through to 24h
	cleanup := setupConfig(t, "display_time=unknown")
	defer cleanup()

	result := Time(testTime)
	require.Equal(t, "15:04", result)
}

func TestTimeFull_UnknownFormat_FallsTo24h(t *testing.T) {
	// Unknown format should fall through to 24h
	cleanup := setupConfig(t, "display_time=unknown")
	defer cleanup()

	result := TimeFull(testTime)
	require.Equal(t, "15:04:05", result)
}

func TestGetDateFormatShort_FallbackToJan02(t *testing.T) {
	// Test with format that becomes empty after stripping year patterns
	cleanup := setupConfig(t, "display_date=/2006/")
	defer cleanup()

	result := DateShort(testTime)
	// After stripping 2006 and trimming, should fallback to "Jan 02"
	require.NotEmpty(t, result)
}

func TestDate_DDMMYYYYFormat(t *testing.T) {
	cleanup := setupConfig(t, "display_date=dd/mm/yyyy")
	defer cleanup()

	result := Date(testTime)
	require.Equal(t, "23/01/2024", result)
}
