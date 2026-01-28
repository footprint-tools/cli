package config

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/footprint-tools/cli/internal/paths"
)

func WriteLines(lines []string) error {
	configPath, err := paths.ConfigFilePath()
	if err != nil {
		return err
	}

	// Write to a temporary file first for atomic operation
	dir := filepath.Dir(configPath)
	tmpFile, err := os.CreateTemp(dir, ".fprc.tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on any failure
	success := false
	defer func() {
		if !success {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	// Set proper permissions on temp file
	if err := tmpFile.Chmod(0600); err != nil {
		return err
	}

	writer := bufio.NewWriter(tmpFile)

	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	// Sync to ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, configPath); err != nil {
		return err
	}

	success = true
	return nil
}
