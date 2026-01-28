package config

import (
	"github.com/footprint-tools/cli/internal/paths"
)

// Default configuration values (in code, not persisted)
var Defaults = map[string]func() string{
	"export_interval_sec": func() string { return "3600" },
	"export_path":         func() string { return paths.ExportRepoDir() },
	"export_last":         func() string { return "0" },
	"export_remote":       func() string { return "" },
	"theme":               func() string { return "default" }, // auto-detects -dark/-light
	"display_date":        func() string { return "Jan 02" },
	"display_time":        func() string { return "24h" },
	"color_success":       func() string { return "" }, // uses theme default
	"color_warning":       func() string { return "" }, // uses theme default
	"color_error":         func() string { return "" }, // uses theme default
	"color_info":          func() string { return "" }, // uses theme default
	"color_muted":         func() string { return "" }, // uses theme default
	"color_header":        func() string { return "" }, // uses theme default
	"enable_log":          func() string { return "true" },
	"pager":               func() string { return "less -FRSX" },
}

// Get returns the value for a config key.
// It checks the config file first, then falls back to the default.
// Returns the value and whether it was found (in file or defaults).
func Get(key string) (string, bool) {
	lines, err := ReadLines()
	if err != nil {
		// On error, try defaults
		if defaultFn, ok := Defaults[key]; ok {
			return defaultFn(), true
		}
		return "", false
	}

	cfg, err := Parse(lines)
	if err != nil {
		if defaultFn, ok := Defaults[key]; ok {
			return defaultFn(), true
		}
		return "", false
	}

	// Check config file first
	if value, exists := cfg[key]; exists {
		return value, true
	}

	// Fall back to default
	if defaultFn, ok := Defaults[key]; ok {
		return defaultFn(), true
	}

	return "", false
}

// GetAll returns all config values (user overrides merged with defaults).
func GetAll() (map[string]string, error) {
	result := make(map[string]string)

	// Start with defaults
	for key, valueFn := range Defaults {
		result[key] = valueFn()
	}

	// Override with user config
	lines, err := ReadLines()
	if err != nil {
		return result, nil // Return defaults on error
	}

	cfg, err := Parse(lines)
	if err != nil {
		return result, nil
	}

	for key, value := range cfg {
		result[key] = value
	}

	return result, nil
}
