package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInspectRepo_Clean(t *testing.T) {
	// Create a temporary git repo with no hooks
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0755))

	// Add only sample files
	sampleHook := filepath.Join(hooksDir, "pre-commit.sample")
	require.NoError(t, os.WriteFile(sampleHook, []byte("#!/bin/sh\n# sample"), 0755))

	inspection := InspectRepo(tmpDir)

	require.Equal(t, StatusClean, inspection.Status)
	require.Empty(t, inspection.UnmanagedHooks)
	require.Empty(t, inspection.GlobalHooksPath)
	require.False(t, inspection.FpInstalled)
}

func TestInspectRepo_ManagedPreCommit(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create .pre-commit-config.yaml
	configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("repos:\n  - repo: local\n"), 0644))

	inspection := InspectRepo(tmpDir)

	require.Equal(t, StatusManagedPreCommit, inspection.Status)
	require.False(t, inspection.Status.CanInstall())
}

func TestInspectRepo_ManagedHusky_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create .husky directory
	huskyDir := filepath.Join(tmpDir, ".husky")
	require.NoError(t, os.MkdirAll(huskyDir, 0755))

	inspection := InspectRepo(tmpDir)

	require.Equal(t, StatusManagedHusky, inspection.Status)
	require.False(t, inspection.Status.CanInstall())
}

func TestInspectRepo_ManagedHusky_PackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create package.json with husky section
	pkgPath := filepath.Join(tmpDir, "package.json")
	pkgContent := `{"name": "test", "husky": {"hooks": {"pre-commit": "lint"}}}`
	require.NoError(t, os.WriteFile(pkgPath, []byte(pkgContent), 0644))

	inspection := InspectRepo(tmpDir)

	require.Equal(t, StatusManagedHusky, inspection.Status)
}

func TestInspectRepo_ManagedLefthook(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create lefthook.yml
	configPath := filepath.Join(tmpDir, "lefthook.yml")
	require.NoError(t, os.WriteFile(configPath, []byte("pre-commit:\n  commands:\n"), 0644))

	inspection := InspectRepo(tmpDir)

	require.Equal(t, StatusManagedLefthook, inspection.Status)
	require.False(t, inspection.Status.CanInstall())
}

func TestInspectRepo_UnmanagedHooks(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0755))

	// Create a non-fp hook
	hookPath := filepath.Join(hooksDir, "pre-commit")
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho 'custom hook'"), 0755))

	inspection := InspectRepo(tmpDir)

	require.Equal(t, StatusUnmanagedHooks, inspection.Status)
	require.Contains(t, inspection.UnmanagedHooks, "pre-commit")
	require.False(t, inspection.Status.CanInstall())
}

func TestInspectRepo_FpHooksInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0755))

	// Create an fp hook
	hookPath := filepath.Join(hooksDir, "post-commit")
	fpHookContent := "#!/bin/sh\nfp record post-commit"
	require.NoError(t, os.WriteFile(hookPath, []byte(fpHookContent), 0755))

	inspection := InspectRepo(tmpDir)

	require.Equal(t, StatusClean, inspection.Status)
	require.True(t, inspection.FpInstalled)
	require.True(t, inspection.Status.CanInstall())
}

func TestInspectRepo_StatusStrings(t *testing.T) {
	tests := []struct {
		status RepoHookStatus
		want   string
	}{
		{StatusClean, "Clean"},
		{StatusManagedPreCommit, "Managed: pre-commit"},
		{StatusManagedHusky, "Managed: husky"},
		{StatusManagedLefthook, "Managed: lefthook"},
		{StatusUnmanagedHooks, "Unmanaged hooks"},
		{StatusGlobalHooksActive, "Global hooks active"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			require.Equal(t, tt.want, tt.status.String())
		})
	}
}

func TestInspectRepo_CanInstall(t *testing.T) {
	tests := []struct {
		status     RepoHookStatus
		canInstall bool
	}{
		{StatusClean, true},
		{StatusManagedPreCommit, false},
		{StatusManagedHusky, false},
		{StatusManagedLefthook, false},
		{StatusUnmanagedHooks, false},
		{StatusGlobalHooksActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			require.Equal(t, tt.canInstall, tt.status.CanInstall())
		})
	}
}

func TestGetGuidance(t *testing.T) {
	tests := []struct {
		status   RepoHookStatus
		contains string
	}{
		{StatusManagedPreCommit, "pre-commit-config.yaml"},
		{StatusManagedHusky, ".husky/post-commit"},
		{StatusManagedLefthook, "lefthook.yml"},
		{StatusUnmanagedHooks, ".git/hooks"},
		{StatusGlobalHooksActive, "core.hooksPath"},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			inspection := RepoInspection{Status: tt.status, GlobalHooksPath: "/custom/hooks"}
			guidance := GetGuidance(inspection)
			require.Contains(t, guidance, tt.contains)
		})
	}
}

func TestGetGuidance_Clean(t *testing.T) {
	inspection := RepoInspection{Status: StatusClean}
	guidance := GetGuidance(inspection)
	require.Empty(t, guidance)
}
