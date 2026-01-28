package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/log"
	"github.com/footprint-tools/cli/internal/paths"
)

func ReadLines() ([]string, error) {
	configPath, err := paths.ConfigFilePath()
	if err != nil {
		return nil, err
	}

	// Check if file exists and has content
	info, err := os.Stat(configPath)
	isNew := os.IsNotExist(err) || (err == nil && info.Size() == 0)

	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	// Ensure correct permissions if file already existed
	if err := os.Chmod(configPath, 0600); err != nil {
		log.Warn("config: could not set permissions on config file: %v", err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSuffix(line, "\r") // Windows CRLF
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// If file is new/empty, initialize with defaults
	if isNew && len(lines) == 0 {
		lines = initializeDefaults()
		if err := WriteLines(lines); err != nil {
			log.Warn("config: could not write default config: %v", err)
		}
	}

	return lines, nil
}

// initializeDefaults creates config lines with default values for visible keys.
func initializeDefaults() []string {
	var lines []string

	lines = append(lines, "# Footprint configuration")
	lines = append(lines, "# Edit values below or use: fp config set <key> <value>")
	lines = append(lines, "")

	for _, key := range domain.ConfigKeys {
		// Skip hidden keys (internal use only)
		if key.Hidden {
			continue
		}

		// Get the actual default value (some are dynamic)
		value := key.Default
		if fn, ok := Defaults[key.Name]; ok {
			value = fn()
		}

		// Quote values that contain spaces
		if strings.Contains(value, " ") {
			value = "\"" + value + "\""
		}

		// HideIfEmpty keys are commented out (optional overrides)
		if key.HideIfEmpty {
			lines = append(lines, "# "+key.Name+"=")
		} else {
			lines = append(lines, key.Name+"="+value)
		}
	}

	return lines
}
