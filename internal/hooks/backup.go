package hooks

import (
	"os"
	"path/filepath"

	"github.com/footprint-tools/cli/internal/log"
)

func BackupDir(hooksPath string) string {
	return filepath.Join(hooksPath, ".fp-backup")
}

func BackupHook(hooksPath, name string) error {
	src := filepath.Join(hooksPath, name)
	dstDir := BackupDir(hooksPath)

	// Create backup directory with restrictive permissions
	if err := os.MkdirAll(dstDir, 0700); err != nil {
		log.Error("hooks: failed to create backup dir %s: %v", dstDir, err)
		return err
	}

	dst := filepath.Join(dstDir, name)
	if err := os.Rename(src, dst); err != nil {
		log.Error("hooks: failed to backup %s to %s: %v", src, dst, err)
		return err
	}

	log.Debug("hooks: backed up %s to %s", name, dst)
	return nil
}
