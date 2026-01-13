package hooks

import (
	"os"
	"path/filepath"
)

func Install(hooksPath string) error {
	for _, hook := range ManagedHooks {
		target := filepath.Join(hooksPath, hook)

		if Exists(target) {
			if err := BackupHook(hooksPath, hook); err != nil {
				return err
			}
		}

		if err := os.WriteFile(target, []byte(Script()), 0755); err != nil {
			return err
		}
	}

	return nil
}
