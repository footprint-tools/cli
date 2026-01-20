package tracking

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Skryensya/footprint/internal/store"
)

func TestResolvePath_CurrentDir(t *testing.T) {
	path, err := resolvePath([]string{})

	require.NoError(t, err)
	require.NotEmpty(t, path)
	// Should be an absolute path
	require.True(t, filepath.IsAbs(path))
}

func TestResolvePath_ExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()

	path, err := resolvePath([]string{tmpDir})

	require.NoError(t, err)
	require.Contains(t, path, filepath.Base(tmpDir))
}

func TestResolvePath_NonexistentPath(t *testing.T) {
	_, err := resolvePath([]string{"/nonexistent/path/that/does/not/exist"})

	require.Error(t, err)
}

func TestResolvePath_FileNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(filePath, []byte("content"), 0644)
	require.NoError(t, err)

	_, err = resolvePath([]string{filePath})

	require.Error(t, err)
}

func TestResolvePath_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	err := os.Mkdir(targetDir, 0755)
	require.NoError(t, err)

	symlinkPath := filepath.Join(tmpDir, "link")
	err = os.Symlink(targetDir, symlinkPath)
	require.NoError(t, err)

	path, err := resolvePath([]string{symlinkPath})

	require.NoError(t, err)
	// Should resolve to the target, not the symlink
	require.Contains(t, path, "target")
}

func TestParseStatus_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected store.Status
	}{
		{"pending", store.StatusPending},
		{"PENDING", store.StatusPending},
		{"Pending", store.StatusPending},
		{"exported", store.StatusExported},
		{"EXPORTED", store.StatusExported},
		{"orphaned", store.StatusOrphaned},
		{"ORPHANED", store.StatusOrphaned},
		{"skipped", store.StatusSkipped},
		{"SKIPPED", store.StatusSkipped},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			status, ok := parseStatus(tt.input)
			require.True(t, ok)
			require.Equal(t, tt.expected, status)
		})
	}
}

func TestParseStatus_Invalid(t *testing.T) {
	invalidInputs := []string{
		"",
		"invalid",
		"unknown",
		"pend",
		"export",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, ok := parseStatus(input)
			require.False(t, ok)
		})
	}
}

func TestParseSource_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected store.Source
	}{
		{"post-commit", store.SourcePostCommit},
		{"POST-COMMIT", store.SourcePostCommit},
		{"Post-Commit", store.SourcePostCommit},
		{"post-rewrite", store.SourcePostRewrite},
		{"post-checkout", store.SourcePostCheckout},
		{"post-merge", store.SourcePostMerge},
		{"pre-push", store.SourcePrePush},
		{"manual", store.SourceManual},
		{"MANUAL", store.SourceManual},
		{"backfill", store.SourceBackfill},
		{"BACKFILL", store.SourceBackfill},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			source, ok := parseSource(tt.input)
			require.True(t, ok)
			require.Equal(t, tt.expected, source)
		})
	}
}

func TestParseSource_Invalid(t *testing.T) {
	invalidInputs := []string{
		"",
		"invalid",
		"unknown",
		"commit",
		"pre-commit",
		"postcommit",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, ok := parseSource(input)
			require.False(t, ok)
		})
	}
}
