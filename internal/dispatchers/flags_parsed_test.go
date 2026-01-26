package dispatchers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParsedFlags_Has(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		checkFor string
		want     bool
	}{
		{
			name:     "flag present",
			flags:    []string{"--verbose", "--debug"},
			checkFor: "--verbose",
			want:     true,
		},
		{
			name:     "flag not present",
			flags:    []string{"--verbose"},
			checkFor: "--debug",
			want:     false,
		},
		{
			name:     "empty flags",
			flags:    []string{},
			checkFor: "--verbose",
			want:     false,
		},
		{
			name:     "flag with value not detected as boolean",
			flags:    []string{"--limit=5"},
			checkFor: "--limit",
			want:     false,
		},
		{
			name:     "multiple flags, check last",
			flags:    []string{"--verbose", "--debug", "--force"},
			checkFor: "--force",
			want:     true,
		},
		{
			name:     "multiple flags, check first",
			flags:    []string{"--verbose", "--debug", "--force"},
			checkFor: "--verbose",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := NewParsedFlags(tt.flags)
			got := pf.Has(tt.checkFor)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParsedFlags_String(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		flagName   string
		defaultVal string
		want       string
	}{
		{
			name:       "flag present with value",
			flags:      []string{"--name=value"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "value",
		},
		{
			name:       "flag not present returns default",
			flags:      []string{"--other=value"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "default",
		},
		{
			name:       "empty flags returns default",
			flags:      []string{},
			flagName:   "--name",
			defaultVal: "default",
			want:       "default",
		},
		{
			name:       "flag with empty value",
			flags:      []string{"--name="},
			flagName:   "--name",
			defaultVal: "default",
			want:       "",
		},
		{
			name:       "flag value with spaces",
			flags:      []string{"--message=hello world"},
			flagName:   "--message",
			defaultVal: "",
			want:       "hello world",
		},
		{
			name:       "flag value with equals sign",
			flags:      []string{"--url=https://example.com?param=value"},
			flagName:   "--url",
			defaultVal: "",
			want:       "https://example.com?param=value",
		},
		{
			name:       "multiple flags, extract correct one",
			flags:      []string{"--first=value1", "--second=value2", "--third=value3"},
			flagName:   "--second",
			defaultVal: "",
			want:       "value2",
		},
		{
			name:       "duplicate flags, first one wins",
			flags:      []string{"--name=first", "--name=second"},
			flagName:   "--name",
			defaultVal: "",
			want:       "first",
		},
		// Space-separated format: --flag value
		{
			name:       "space separated format",
			flags:      []string{"--name", "value"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "value",
		},
		{
			name:       "space separated with other flags",
			flags:      []string{"--verbose", "--name", "value", "--debug"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "value",
		},
		{
			name:       "space separated next value is flag returns default",
			flags:      []string{"--name", "--other"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "default",
		},
		{
			name:       "space separated flag at end returns default",
			flags:      []string{"--verbose", "--name"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "default",
		},
		{
			name:       "equals format takes precedence over space",
			flags:      []string{"--name=equals", "--name", "space"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "equals",
		},
		{
			name:       "space separated with short flag as next",
			flags:      []string{"--name", "-v"},
			flagName:   "--name",
			defaultVal: "default",
			want:       "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := NewParsedFlags(tt.flags)
			got := pf.String(tt.flagName, tt.defaultVal)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParsedFlags_Int(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		flagName   string
		defaultVal int
		want       int
	}{
		{
			name:       "valid integer",
			flags:      []string{"--limit=5"},
			flagName:   "--limit",
			defaultVal: 10,
			want:       5,
		},
		{
			name:       "flag not present returns default",
			flags:      []string{"--other=5"},
			flagName:   "--limit",
			defaultVal: 10,
			want:       10,
		},
		{
			name:       "empty flags returns default",
			flags:      []string{},
			flagName:   "--limit",
			defaultVal: 10,
			want:       10,
		},
		{
			name:       "invalid integer returns default",
			flags:      []string{"--limit=abc"},
			flagName:   "--limit",
			defaultVal: 10,
			want:       10,
		},
		{
			name:       "float returns default",
			flags:      []string{"--limit=5.5"},
			flagName:   "--limit",
			defaultVal: 10,
			want:       10,
		},
		{
			name:       "negative integer",
			flags:      []string{"--limit=-5"},
			flagName:   "--limit",
			defaultVal: 10,
			want:       -5,
		},
		{
			name:       "zero value",
			flags:      []string{"--limit=0"},
			flagName:   "--limit",
			defaultVal: 10,
			want:       0,
		},
		{
			name:       "large integer",
			flags:      []string{"--limit=9999999"},
			flagName:   "--limit",
			defaultVal: 10,
			want:       9999999,
		},
		{
			name:       "empty value returns default",
			flags:      []string{"--limit="},
			flagName:   "--limit",
			defaultVal: 10,
			want:       10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := NewParsedFlags(tt.flags)
			got := pf.Int(tt.flagName, tt.defaultVal)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParsedFlags_Date(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		flagName string
		want     *time.Time
	}{
		{
			name:     "valid date",
			flags:    []string{"--since=2024-01-15"},
			flagName: "--since",
			want:     timePtr(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:     "flag not present returns nil",
			flags:    []string{"--other=2024-01-15"},
			flagName: "--since",
			want:     nil,
		},
		{
			name:     "empty flags returns nil",
			flags:    []string{},
			flagName: "--since",
			want:     nil,
		},
		{
			name:     "invalid date format returns nil",
			flags:    []string{"--since=15-01-2024"},
			flagName: "--since",
			want:     nil,
		},
		{
			name:     "invalid date value returns nil",
			flags:    []string{"--since=2024-13-45"},
			flagName: "--since",
			want:     nil,
		},
		{
			name:     "non-date string returns nil",
			flags:    []string{"--since=yesterday"},
			flagName: "--since",
			want:     nil,
		},
		{
			name:     "empty value returns nil",
			flags:    []string{"--since="},
			flagName: "--since",
			want:     nil,
		},
		{
			name:     "date with time component returns nil",
			flags:    []string{"--since=2024-01-15T12:00:00Z"},
			flagName: "--since",
			want:     nil,
		},
		{
			name:     "leap year date",
			flags:    []string{"--since=2024-02-29"},
			flagName: "--since",
			want:     timePtr(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:     "first day of year",
			flags:    []string{"--since=2024-01-01"},
			flagName: "--since",
			want:     timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:     "last day of year",
			flags:    []string{"--since=2024-12-31"},
			flagName: "--since",
			want:     timePtr(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := NewParsedFlags(tt.flags)
			got := pf.Date(tt.flagName)

			if tt.want == nil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.True(t, tt.want.Equal(*got), "expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestParsedFlags_Raw(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
	}{
		{
			name:  "empty flags",
			flags: []string{},
		},
		{
			name:  "single flag",
			flags: []string{"--verbose"},
		},
		{
			name:  "multiple flags",
			flags: []string{"--verbose", "--limit=5", "--since=2024-01-01"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := NewParsedFlags(tt.flags)
			got := pf.Raw()
			require.Equal(t, tt.flags, got)
		})
	}
}

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
