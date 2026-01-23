package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/footprint-tools/footprint-cli/internal/log"
	"github.com/footprint-tools/footprint-cli/internal/paths"
)

func ReadLines() ([]string, error) {
	configPath, err := paths.ConfigFilePath()
	if err != nil {
		return nil, err
	}

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

	return lines, scanner.Err()
}
