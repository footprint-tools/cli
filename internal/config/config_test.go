package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupTempHome creates a temporary HOME directory for testing
func setupTempHome(t *testing.T) string {
	t.Helper()
	tempHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	require.NoError(t, os.Setenv("HOME", tempHome))
	t.Cleanup(func() {
		_ = os.Setenv("HOME", oldHome)
	})
	return tempHome
}

func TestReadLines(t *testing.T) {
	tests := []struct {
		name         string
		setupContent string
		wantLines    []string
	}{
		// Empty file case is tested in TestReadLines_CreatesFileIfNotExists
		{
			name: "single line",
			setupContent: "key=value\n",
			wantLines:    []string{"key=value"},
		},
		{
			name: "multiple lines",
			setupContent: "key1=value1\nkey2=value2\nkey3=value3\n",
			wantLines:    []string{"key1=value1", "key2=value2", "key3=value3"},
		},
		{
			name: "lines with comments",
			setupContent: "# Comment\nkey=value\n",
			wantLines:    []string{"# Comment", "key=value"},
		},
		{
			name: "Windows CRLF line endings",
			setupContent: "key1=value1\r\nkey2=value2\r\n",
			wantLines:    []string{"key1=value1", "key2=value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempHome := setupTempHome(t)

			// Write setup content if provided
			if tt.setupContent != "" {
				configPath := filepath.Join(tempHome, ".fprc")
				err := os.WriteFile(configPath, []byte(tt.setupContent), 0600)
				require.NoError(t, err)
			}

			got, err := ReadLines()
			require.NoError(t, err)
			require.Equal(t, tt.wantLines, got)

			// Verify file was created with correct permissions if it didn't exist
			configPath := filepath.Join(tempHome, ".fprc")
			info, err := os.Stat(configPath)
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0600), info.Mode().Perm())
		})
	}
}

func TestReadLines_CreatesFileIfNotExists(t *testing.T) {
	tempHome := setupTempHome(t)
	configPath := filepath.Join(tempHome, ".fprc")

	// Verify file doesn't exist yet
	_, err := os.Stat(configPath)
	require.True(t, os.IsNotExist(err))

	// ReadLines should create it with defaults
	lines, err := ReadLines()
	require.NoError(t, err)
	require.NotEmpty(t, lines, "should initialize with default config")

	// Verify file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Verify it contains expected defaults
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "pager=")
	require.Contains(t, string(content), "theme=")
	require.Contains(t, string(content), "# color_success=") // commented
}

func TestWriteLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
	}{
		{
			name:  "empty lines",
			lines: []string{},
		},
		{
			name:  "single line",
			lines: []string{"key=value"},
		},
		{
			name:  "multiple lines",
			lines: []string{"key1=value1", "key2=value2", "key3=value3"},
		},
		{
			name:  "lines with comments",
			lines: []string{"# Comment", "key=value", "# Another comment"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempHome := setupTempHome(t)

			err := WriteLines(tt.lines)
			require.NoError(t, err)

			// Verify file was written correctly
			configPath := filepath.Join(tempHome, ".fprc")
			content, err := os.ReadFile(configPath)
			require.NoError(t, err)

			// Verify content
			expected := ""
			for _, line := range tt.lines {
				expected += line + "\n"
			}
			require.Equal(t, expected, string(content))

			// Verify permissions
			info, err := os.Stat(configPath)
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0600), info.Mode().Perm())
		})
	}
}

func TestWriteLines_Overwrites(t *testing.T) {
	tempHome := setupTempHome(t)
	configPath := filepath.Join(tempHome, ".fprc")

	// Write initial content
	err := WriteLines([]string{"key1=value1", "key2=value2"})
	require.NoError(t, err)

	// Overwrite with new content
	err = WriteLines([]string{"key3=value3"})
	require.NoError(t, err)

	// Verify old content was replaced
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Equal(t, "key3=value3\n", string(content))
}

func TestSet(t *testing.T) {
	tests := []struct {
		name        string
		initialLines []string
		key         string
		value       string
		wantLines   []string
		wantUpdated bool
	}{
		{
			name:         "add to empty",
			initialLines: []string{},
			key:          "key",
			value:        "value",
			wantLines:    []string{"key=value"},
			wantUpdated:  false,
		},
		{
			name:         "add new key",
			initialLines: []string{"key1=value1"},
			key:          "key2",
			value:        "value2",
			wantLines:    []string{"key1=value1", "key2=value2"},
			wantUpdated:  false,
		},
		{
			name:         "update existing key",
			initialLines: []string{"key1=value1", "key2=value2"},
			key:          "key1",
			value:        "newvalue",
			wantLines:    []string{"key1=newvalue", "key2=value2"},
			wantUpdated:  true,
		},
		{
			name:         "preserves comments and blank lines",
			initialLines: []string{"# Comment", "", "key1=value1"},
			key:          "key2",
			value:        "value2",
			wantLines:    []string{"# Comment", "", "key1=value1", "key2=value2"},
			wantUpdated:  false,
		},
		{
			name:         "handles whitespace in existing line",
			initialLines: []string{"  key1  =  value1  "},
			key:          "key1",
			value:        "newvalue",
			wantLines:    []string{"key1=newvalue"},
			wantUpdated:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, updated := Set(tt.initialLines, tt.key, tt.value)
			require.Equal(t, tt.wantLines, got)
			require.Equal(t, tt.wantUpdated, updated)
		})
	}
}

func TestUnset(t *testing.T) {
	tests := []struct {
		name        string
		initialLines []string
		key         string
		wantLines   []string
		wantRemoved bool
	}{
		{
			name:         "remove from empty",
			initialLines: []string{},
			key:          "key",
			wantLines:    nil,
			wantRemoved:  false,
		},
		{
			name:         "remove existing key",
			initialLines: []string{"key1=value1", "key2=value2"},
			key:          "key1",
			wantLines:    []string{"key2=value2"},
			wantRemoved:  true,
		},
		{
			name:         "remove non-existent key",
			initialLines: []string{"key1=value1"},
			key:          "key2",
			wantLines:    []string{"key1=value1"},
			wantRemoved:  false,
		},
		{
			name:         "preserves comments and blank lines",
			initialLines: []string{"# Comment", "", "key1=value1", "key2=value2"},
			key:          "key1",
			wantLines:    []string{"# Comment", "", "key2=value2"},
			wantRemoved:  true,
		},
		{
			name:         "handles whitespace in line",
			initialLines: []string{"  key1  =  value1  "},
			key:          "key1",
			wantLines:    nil,
			wantRemoved:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, removed := Unset(tt.initialLines, tt.key)
			require.Equal(t, tt.wantLines, got)
			require.Equal(t, tt.wantRemoved, removed)
		})
	}
}

func TestReadWrite_Integration(t *testing.T) {
	tempHome := setupTempHome(t)
	configPath := filepath.Join(tempHome, ".fprc")

	// Write some lines
	initialLines := []string{"key1=value1", "key2=value2", "# Comment"}
	err := WriteLines(initialLines)
	require.NoError(t, err)

	// Read them back
	gotLines, err := ReadLines()
	require.NoError(t, err)
	require.Equal(t, initialLines, gotLines)

	// Modify using Set
	gotLines, _ = Set(gotLines, "key1", "newvalue")
	err = WriteLines(gotLines)
	require.NoError(t, err)

	// Read and verify
	gotLines, err = ReadLines()
	require.NoError(t, err)
	require.Equal(t, []string{"key1=newvalue", "key2=value2", "# Comment"}, gotLines)

	// Modify using Unset
	gotLines, _ = Unset(gotLines, "key2")
	err = WriteLines(gotLines)
	require.NoError(t, err)

	// Read and verify
	gotLines, err = ReadLines()
	require.NoError(t, err)
	require.Equal(t, []string{"key1=newvalue", "# Comment"}, gotLines)

	// Verify file still has correct permissions
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestAppendArray(t *testing.T) {
	tests := []struct {
		name         string
		initialLines []string
		arrayKey     string
		value        string
		wantLines    []string
		wantAdded    bool
	}{
		{
			name:         "add to empty",
			initialLines: []string{},
			arrayKey:     "repos",
			value:        "/path/to/repo",
			wantLines:    []string{"repos[]=/path/to/repo"},
			wantAdded:    true,
		},
		{
			name:         "add new value",
			initialLines: []string{"repos[]=/path/to/repo1"},
			arrayKey:     "repos",
			value:        "/path/to/repo2",
			wantLines:    []string{"repos[]=/path/to/repo1", "repos[]=/path/to/repo2"},
			wantAdded:    true,
		},
		{
			name:         "value already exists",
			initialLines: []string{"repos[]=/path/to/repo1"},
			arrayKey:     "repos",
			value:        "/path/to/repo1",
			wantLines:    []string{"repos[]=/path/to/repo1"},
			wantAdded:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, added := AppendArray(tt.initialLines, tt.arrayKey, tt.value)
			require.Equal(t, tt.wantLines, got)
			require.Equal(t, tt.wantAdded, added)
		})
	}
}

func TestRemoveFromArray(t *testing.T) {
	tests := []struct {
		name         string
		initialLines []string
		arrayKey     string
		value        string
		wantLines    []string
		wantRemoved  bool
	}{
		{
			name:         "remove from empty",
			initialLines: []string{},
			arrayKey:     "repos",
			value:        "/path/to/repo",
			wantLines:    nil,
			wantRemoved:  false,
		},
		{
			name:         "remove existing value",
			initialLines: []string{"repos[]=/path/to/repo1", "repos[]=/path/to/repo2"},
			arrayKey:     "repos",
			value:        "/path/to/repo1",
			wantLines:    []string{"repos[]=/path/to/repo2"},
			wantRemoved:  true,
		},
		{
			name:         "value not found",
			initialLines: []string{"repos[]=/path/to/repo1"},
			arrayKey:     "repos",
			value:        "/path/to/repo2",
			wantLines:    []string{"repos[]=/path/to/repo1"},
			wantRemoved:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, removed := RemoveFromArray(tt.initialLines, tt.arrayKey, tt.value)
			require.Equal(t, tt.wantLines, got)
			require.Equal(t, tt.wantRemoved, removed)
		})
	}
}

func TestUnsetArray(t *testing.T) {
	tests := []struct {
		name         string
		initialLines []string
		arrayKey     string
		wantLines    []string
		wantRemoved  bool
	}{
		{
			name:         "remove from empty",
			initialLines: []string{},
			arrayKey:     "repos",
			wantLines:    nil,
			wantRemoved:  false,
		},
		{
			name:         "remove all array values",
			initialLines: []string{"repos[]=/path/to/repo1", "repos[]=/path/to/repo2", "key=value"},
			arrayKey:     "repos",
			wantLines:    []string{"key=value"},
			wantRemoved:  true,
		},
		{
			name:         "array not found",
			initialLines: []string{"key=value"},
			arrayKey:     "repos",
			wantLines:    []string{"key=value"},
			wantRemoved:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, removed := UnsetArray(tt.initialLines, tt.arrayKey)
			require.Equal(t, tt.wantLines, got)
			require.Equal(t, tt.wantRemoved, removed)
		})
	}
}

func TestParseArray(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		arrayKey string
		want     []string
	}{
		{
			name:     "empty lines",
			lines:    []string{},
			arrayKey: "repos",
			want:     nil,
		},
		{
			name:     "no array values",
			lines:    []string{"key=value"},
			arrayKey: "repos",
			want:     nil,
		},
		{
			name:     "single array value",
			lines:    []string{"repos[]=/path/to/repo"},
			arrayKey: "repos",
			want:     []string{"/path/to/repo"},
		},
		{
			name:     "multiple array values",
			lines:    []string{"repos[]=/path/to/repo1", "repos[]=/path/to/repo2", "key=value"},
			arrayKey: "repos",
			want:     []string{"/path/to/repo1", "/path/to/repo2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseArray(tt.lines, tt.arrayKey)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestProvider(t *testing.T) {
	tempHome := setupTempHome(t)

	// Write initial config
	initialLines := []string{"key1=value1", "key2=value2"}
	err := WriteLines(initialLines)
	require.NoError(t, err)

	provider := NewProvider()
	require.NotNil(t, provider)

	// Test Get
	val, found := provider.Get("key1")
	require.True(t, found)
	require.Equal(t, "value1", val)

	// Test GetAll
	all, err := provider.GetAll()
	require.NoError(t, err)
	require.Equal(t, "value1", all["key1"])
	require.Equal(t, "value2", all["key2"])

	// Test Set
	err = provider.Set("key3", "value3")
	require.NoError(t, err)

	val, found = provider.Get("key3")
	require.True(t, found)
	require.Equal(t, "value3", val)

	// Test Unset
	err = provider.Unset("key3")
	require.NoError(t, err)

	val, found = provider.Get("key3")
	require.False(t, found)
	require.Empty(t, val)

	_ = tempHome // use tempHome
}
