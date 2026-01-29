package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManagedHooks_ContainsExpectedHooks(t *testing.T) {
	expectedHooks := []string{
		"post-commit",
		"post-merge",
		"post-checkout",
		"post-rewrite",
		"pre-push",
	}

	require.Equal(t, expectedHooks, ManagedHooks)
}

func TestExists_Fileexists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-file")

	err := os.WriteFile(filePath, []byte("content"), 0644)
	require.NoError(t, err)

	require.True(t, exists(filePath))
}

func TestExists_FileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent")

	require.False(t, exists(filePath))
}

func TestExists_IsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// tmpDir is a directory, not a file
	require.False(t, exists(tmpDir))
}

func TestStatus_AllHooksExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create all managed hooks
	for _, hook := range ManagedHooks {
		hookPath := filepath.Join(tmpDir, hook)
		err := os.WriteFile(hookPath, []byte("#!/bin/sh"), 0755)
		require.NoError(t, err)
	}

	status := Status(tmpDir)

	for _, hook := range ManagedHooks {
		require.True(t, status[hook], "hook '%s' should be marked as existing", hook)
	}
}

func TestStatus_NoHooksExist(t *testing.T) {
	tmpDir := t.TempDir()

	status := Status(tmpDir)

	for _, hook := range ManagedHooks {
		require.False(t, status[hook], "hook '%s' should be marked as not existing", hook)
	}
}

func TestStatus_SomeHooksExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only some hooks
	err := os.WriteFile(filepath.Join(tmpDir, "post-commit"), []byte("#!/bin/sh"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "pre-push"), []byte("#!/bin/sh"), 0755)
	require.NoError(t, err)

	status := Status(tmpDir)

	require.True(t, status["post-commit"])
	require.False(t, status["post-merge"])
	require.False(t, status["post-checkout"])
	require.False(t, status["post-rewrite"])
	require.True(t, status["pre-push"])
}

func TestScript_GeneratesValidScript(t *testing.T) {
	fpPath := "/usr/local/bin/fp"
	source := "post-commit"

	script := Script(fpPath, source)

	require.Contains(t, script, "#!/bin/sh")
	require.Contains(t, script, "FP_SOURCE='post-commit'")
	require.Contains(t, script, "'/usr/local/bin/fp' record")
	require.Contains(t, script, ">/dev/null")
}

func TestScript_DifferentSources(t *testing.T) {
	fpPath := "/opt/fp"

	for _, source := range ManagedHooks {
		script := Script(fpPath, source)
		require.Contains(t, script, "FP_SOURCE='"+source+"'")
	}
}

func TestBackupDir_ReturnsCorrectPath(t *testing.T) {
	hooksPath := "/home/user/.config/git/hooks"

	backupDir := backupDir(hooksPath)

	require.Equal(t, "/home/user/.config/git/hooks/.fp-backup", backupDir)
}

func TestBackupHook_Success(t *testing.T) {
	tmpDir := t.TempDir()
	hookName := "post-commit"
	hookPath := filepath.Join(tmpDir, hookName)

	// Create the hook file
	originalContent := "#!/bin/sh\necho 'original'"
	err := os.WriteFile(hookPath, []byte(originalContent), 0755)
	require.NoError(t, err)

	// Backup the hook
	err = backupHook(tmpDir, hookName)
	require.NoError(t, err)

	// Verify original is gone
	require.False(t, exists(hookPath))

	// Verify backup exists
	backupPath := filepath.Join(backupDir(tmpDir), hookName)
	require.True(t, exists(backupPath))

	// Verify content is preserved
	content, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	require.Equal(t, originalContent, string(content))
}

func TestBackupHook_CreatesDirIfNeeded(t *testing.T) {
	tmpDir := t.TempDir()
	hookName := "post-merge"
	hookPath := filepath.Join(tmpDir, hookName)

	err := os.WriteFile(hookPath, []byte("#!/bin/sh"), 0755)
	require.NoError(t, err)

	// Backup dir doesn't exist yet
	backupDir := backupDir(tmpDir)
	_, err = os.Stat(backupDir)
	require.True(t, os.IsNotExist(err))

	// Backup should create the dir
	err = backupHook(tmpDir, hookName)
	require.NoError(t, err)

	// Backup dir now exists
	info, err := os.Stat(backupDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestInstall_Success(t *testing.T) {
	tmpDir := t.TempDir()

	err := Install(tmpDir)
	require.NoError(t, err)

	// Verify all hooks were installed
	for _, hook := range ManagedHooks {
		hookPath := filepath.Join(tmpDir, hook)
		require.True(t, exists(hookPath), "hook '%s' should be installed", hook)

		// Verify the hook is executable
		info, err := os.Stat(hookPath)
		require.NoError(t, err)
		require.True(t, info.Mode().Perm()&0100 != 0, "hook '%s' should be executable", hook)
	}
}

func TestInstall_BacksUpExistingHooks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an existing hook
	existingHook := filepath.Join(tmpDir, "post-commit")
	originalContent := "#!/bin/sh\necho 'existing hook'"
	err := os.WriteFile(existingHook, []byte(originalContent), 0755)
	require.NoError(t, err)

	err = Install(tmpDir)
	require.NoError(t, err)

	// Verify backup was created
	backupPath := filepath.Join(backupDir(tmpDir), "post-commit")
	require.True(t, exists(backupPath))

	// Verify backup content is correct
	content, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	require.Equal(t, originalContent, string(content))

	// Verify new hook was installed
	require.True(t, exists(existingHook))
	newContent, err := os.ReadFile(existingHook)
	require.NoError(t, err)
	require.Contains(t, string(newContent), "record")
	require.Contains(t, string(newContent), "FP_SOURCE='post-commit'")
}

func TestUninstall_RemovesHooks(t *testing.T) {
	tmpDir := t.TempDir()

	// First install hooks
	err := Install(tmpDir)
	require.NoError(t, err)

	// Then uninstall
	err = Uninstall(tmpDir)
	require.NoError(t, err)

	// Verify all hooks were removed
	for _, hook := range ManagedHooks {
		hookPath := filepath.Join(tmpDir, hook)
		require.False(t, exists(hookPath), "hook '%s' should be removed", hook)
	}
}

func TestUninstall_RestoresBackups(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing hooks
	originalContent := "#!/bin/sh\necho 'original'"
	for _, hook := range ManagedHooks {
		hookPath := filepath.Join(tmpDir, hook)
		err := os.WriteFile(hookPath, []byte(originalContent), 0755)
		require.NoError(t, err)
	}

	// Install (which backs up existing hooks)
	err := Install(tmpDir)
	require.NoError(t, err)

	// Uninstall (which should restore backups)
	err = Uninstall(tmpDir)
	require.NoError(t, err)

	// Verify original hooks were restored
	for _, hook := range ManagedHooks {
		hookPath := filepath.Join(tmpDir, hook)
		require.True(t, exists(hookPath), "hook '%s' should be restored", hook)

		content, err := os.ReadFile(hookPath)
		require.NoError(t, err)
		require.Equal(t, originalContent, string(content), "hook '%s' content should be original", hook)
	}
}

func TestUninstall_NoHooksPresent(t *testing.T) {
	tmpDir := t.TempDir()

	// Uninstall on empty dir should not error
	err := Uninstall(tmpDir)
	require.NoError(t, err)
}

func TestInstall_InvalidPath(t *testing.T) {
	// Try to install to a non-existent path
	err := Install("/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
}
