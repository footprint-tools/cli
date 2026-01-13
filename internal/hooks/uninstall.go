package hooks

import (
	"os"
	"path/filepath"
)

func Uninstall(hooksPath string) error {
	backupDir := BackupDir(hooksPath)

	for _, hook := range ManagedHooks {
		target := filepath.Join(hooksPath, hook)
		backup := filepath.Join(backupDir, hook)

		if Exists(backup) {
			_ = os.Remove(target)
			if err := os.Rename(backup, target); err != nil {
				return err
			}
			continue
		}

		if Exists(target) {
			if err := os.Remove(target); err != nil {
				return err
			}
		}
	}

	return nil
}
