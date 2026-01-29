package hooks

import (
	"os"
	"path/filepath"

	"github.com/footprint-tools/cli/internal/log"
)

func Install(hooksPath string) error {
	log.Debug("hooks: installing to %s", hooksPath)

	// Ensure hooks directory exists
	if err := os.MkdirAll(hooksPath, filePermExecutable); err != nil {
		log.Error("hooks: failed to create directory %s: %v", hooksPath, err)
		return err
	}

	fpPath, err := os.Executable()
	if err != nil {
		log.Error("hooks: failed to get executable path: %v", err)
		return err
	}

	for _, hook := range ManagedHooks {
		target := filepath.Join(hooksPath, hook)

		if exists(target) {
			log.Debug("hooks: backing up existing %s", hook)
			if err := backupHook(hooksPath, hook); err != nil {
				log.Error("hooks: failed to backup %s: %v", hook, err)
				return err
			}
		}

		script := Script(fpPath, hook)

		if err := os.WriteFile(target, []byte(script), filePermExecutable); err != nil {
			log.Error("hooks: failed to write %s: %v", hook, err)
			return err
		}
		log.Debug("hooks: installed %s", hook)
	}

	log.Info("hooks: installed %d hooks to %s", len(ManagedHooks), hooksPath)
	return nil
}
