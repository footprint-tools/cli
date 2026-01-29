package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "empty input",
			lines:   []string{},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:  "single key-value",
			lines: []string{"key=value"},
			want: map[string]string{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name: "multiple key-values",
			lines: []string{
				"key1=value1",
				"key2=value2",
				"key3=value3",
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			wantErr: false,
		},
		{
			name: "ignores blank lines",
			lines: []string{
				"key1=value1",
				"",
				"key2=value2",
				"   ",
				"key3=value3",
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			wantErr: false,
		},
		{
			name: "ignores comment lines",
			lines: []string{
				"# This is a comment",
				"key1=value1",
				"# Another comment",
				"key2=value2",
				"  # Indented comment",
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "trims whitespace around key and value",
			lines: []string{
				"  key1  =  value1  ",
				"key2=  value2",
				"  key3=value3  ",
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			wantErr: false,
		},
		{
			name: "handles equals sign in value",
			lines: []string{
				"url=https://example.com?param=value",
				"equation=x=y+z",
				"base64=SGVsbG8=",
			},
			want: map[string]string{
				"url":      "https://example.com?param=value",
				"equation": "x=y+z",
				"base64":   "SGVsbG8=",
			},
			wantErr: false,
		},
		{
			name: "invalid line without equals sign",
			lines: []string{
				"key1=value1",
				"invalid_line",
				"key2=value2",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid empty key",
			lines: []string{
				"=value",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty value is valid",
			lines: []string{
				"key=",
			},
			want: map[string]string{
				"key": "",
			},
			wantErr: false,
		},
		{
			name: "BOM (Byte Order Mark) is stripped from first line",
			lines: []string{
				"\uFEFFkey1=value1",
				"key2=value2",
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "mixed valid content",
			lines: []string{
				"# Config file for footprint",
				"",
				"trackedRepos=github.com/user/repo1,github.com/user/repo2",
				"export_interval_sec=3600",
				"  # Export settings",
				"export_last=2024-01-01T00:00:00Z",
				"",
				"log_level=debug",
			},
			want: map[string]string{
				"trackedRepos":        "github.com/user/repo1,github.com/user/repo2",
				"export_interval_sec": "3600",
				"export_last":         "2024-01-01T00:00:00Z",
				"log_level":           "debug",
			},
			wantErr: false,
		},
		{
			name: "duplicate keys - last one wins",
			lines: []string{
				"key=value1",
				"key=value2",
			},
			want: map[string]string{
				"key": "value2",
			},
			wantErr: false,
		},
		{
			name: "special characters in values",
			lines: []string{
				"path=/path/to/repo",
				"remote=git@github.com:user/repo.git",
				"special=!@#$%^&*()",
			},
			want: map[string]string{
				"path":    "/path/to/repo",
				"remote":  "git@github.com:user/repo.git",
				"special": "!@#$%^&*()",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.lines)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
