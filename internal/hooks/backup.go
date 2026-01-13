package hooks

import (
	"os"
	"path/filepath"
)

func BackupDir(hooksPath string) string {
	return filepath.Join(hooksPath, ".fp-backup")
}

func BackupHook(hooksPath, name string) error {
	src := filepath.Join(hooksPath, name)
	dstDir := BackupDir(hooksPath)

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	dst := filepath.Join(dstDir, name)
	return os.Rename(src, dst)
}
