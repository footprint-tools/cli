package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoID_String(t *testing.T) {
	id := RepoID("github.com/user/repo")

	require.Equal(t, "github.com/user/repo", id.String())
}

func TestRepoID_IsEmpty(t *testing.T) {
	require.True(t, RepoID("").IsEmpty())
	require.False(t, RepoID("test").IsEmpty())
}

func TestRepoID_ToFilesystemSafe(t *testing.T) {
	tests := []struct {
		input    RepoID
		expected string
	}{
		{"github.com/user/repo", "github.com_user_repo"},
		{"git@github.com:user/repo", "git_github.com_user_repo"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		result := tt.input.ToFilesystemSafe()
		require.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestDeriveRepoID_FromRemote(t *testing.T) {
	tests := []struct {
		remoteURL string
		expected  string
	}{
		{"https://github.com/user/repo.git", "github.com/user/repo"},
		{"git@github.com:user/repo.git", "github.com/user/repo"},
		{"https://github.com/user/repo", "github.com/user/repo"},
		{"git://github.com/user/repo.git", "github.com/user/repo"},
	}

	for _, tt := range tests {
		id, err := DeriveRepoID(tt.remoteURL, "")
		require.NoError(t, err, "remoteURL: %s", tt.remoteURL)
		require.Equal(t, RepoID(tt.expected), id)
	}
}

func TestDeriveRepoID_FromPath(t *testing.T) {
	id, err := DeriveRepoID("", "/home/user/project")

	require.NoError(t, err)
	require.Contains(t, id.String(), "local/")
}

func TestDeriveRepoID_EmptyInputs(t *testing.T) {
	_, err := DeriveRepoID("", "")

	require.Error(t, err)
}

func TestEventStatus_String(t *testing.T) {
	tests := []struct {
		status   EventStatus
		expected string
	}{
		{StatusPending, "PENDING"},
		{StatusExported, "EXPORTED"},
		{StatusOrphaned, "ORPHANED"},
		{StatusSkipped, "SKIPPED"},
		{EventStatus(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.status.String())
	}
}

func TestParseEventStatus(t *testing.T) {
	status, ok := ParseEventStatus("pending")
	require.True(t, ok)
	require.Equal(t, StatusPending, status)

	status, ok = ParseEventStatus("EXPORTED")
	require.True(t, ok)
	require.Equal(t, StatusExported, status)

	_, ok = ParseEventStatus("invalid")
	require.False(t, ok)
}

func TestEventSource_String(t *testing.T) {
	tests := []struct {
		source   EventSource
		expected string
	}{
		{SourcePostCommit, "post-commit"},
		{SourceBackfill, "backfill"},
		{SourceManual, "manual"},
		{EventSource(99), "unknown"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.source.String())
	}
}

func TestParseEventSource(t *testing.T) {
	source, ok := ParseEventSource("post-commit")
	require.True(t, ok)
	require.Equal(t, SourcePostCommit, source)

	source, ok = ParseEventSource("BACKFILL")
	require.True(t, ok)
	require.Equal(t, SourceBackfill, source)

	_, ok = ParseEventSource("invalid")
	require.False(t, ok)
}

func TestConfigKey_GetConfigKey(t *testing.T) {
	key, ok := GetConfigKey("export_interval_sec")
	require.True(t, ok)
	require.Equal(t, "export_interval_sec", key.Name)
	require.Equal(t, "3600", key.Default)
}

func TestConfigKey_IsValidConfigKey(t *testing.T) {
	require.True(t, IsValidConfigKey("export_interval_sec"))
	require.False(t, IsValidConfigKey("invalid_key"))
}

func TestConfigKey_GetDefaultValue(t *testing.T) {
	val, ok := GetDefaultValue("export_interval_sec")
	require.True(t, ok)
	require.Equal(t, "3600", val)

	_, ok = GetDefaultValue("invalid")
	require.False(t, ok)
}

func TestVisibleConfigKeys(t *testing.T) {
	visible := VisibleConfigKeys()

	// Should not contain hidden keys
	for _, key := range visible {
		require.False(t, key.Hidden, "key %s should not be hidden", key.Name)
	}

	// Should contain at least some visible keys
	require.Greater(t, len(visible), 0)
}
