package domain

// ConfigKey defines a configuration key with its metadata.
type ConfigKey struct {
	Name        string
	Default     string
	Description string
	Section     string // Section for grouping in UI (Display, Colors, Export, etc.)
	Hidden      bool   // Hidden keys are not shown in help or config list
	HideIfEmpty bool   // Only show in config list if explicitly set
}

// ConfigKeys defines all available configuration keys.
// This is the single source of truth for configuration.
// Order determines display order in `fp config list`.
var ConfigKeys = []ConfigKey{
	// Display
	{
		Name:        "pager",
		Default:     "less -FRSX",
		Description: "Pager command for long output",
		Section:     "Display",
	},
	{
		Name:        "theme",
		Default:     "default",
		Description: "Color theme: default, neon, aurora, mono, ocean, sunset, candy, contrast",
		Section:     "Display",
	},
	{
		Name:        "display_date",
		Default:     "Jan 02",
		Description: "Date format: dd/mm/yyyy, mm/dd/yyyy, yyyy-mm-dd, or Go format",
		Section:     "Display",
	},
	{
		Name:        "display_time",
		Default:     "24h",
		Description: "Time format: 12h, 24h",
		Section:     "Display",
	},
	// Logging
	{
		Name:        "enable_log",
		Default:     "true",
		Description: "Enable logging to file (true/false)",
		Section:     "Logging",
	},
	// Export
	{
		Name:        "export_interval_sec",
		Default:     "3600",
		Description: "Seconds between automatic exports",
		Section:     "Export",
	},
	{
		Name:        "export_path",
		Default:     "", // Set dynamically to paths.ExportRepoDir()
		Description: "Path to the export repository",
		Section:     "Export",
	},
	{
		Name:        "export_remote",
		Default:     "",
		Description: "Remote URL for syncing exports",
		Section:     "Export",
	},
	// Hidden (internal)
	{
		Name:        "export_last",
		Default:     "0",
		Description: "Unix timestamp of last export",
		Section:     "Export",
		Hidden:      true,
	},
	// Color Overrides - override specific colors from the current theme (ANSI 0-255)
	{
		Name:        "color_success",
		Description: "Override success color from current theme (ANSI 0-255)",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
	{
		Name:        "color_warning",
		Description: "Override warning color from current theme (ANSI 0-255)",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
	{
		Name:        "color_error",
		Description: "Override error color from current theme (ANSI 0-255)",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
	{
		Name:        "color_info",
		Description: "Override info color from current theme (ANSI 0-255)",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
	{
		Name:        "color_muted",
		Description: "Override muted text color from current theme (ANSI 0-255)",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
	{
		Name:        "color_header",
		Description: "Override header style from current theme (ANSI 0-255 or 'bold')",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
	{
		Name:        "color_ui_active",
		Description: "Override focused/active UI color from current theme (ANSI 0-255)",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
	{
		Name:        "color_ui_dim",
		Description: "Override unfocused/inactive UI color from current theme (ANSI 0-255)",
		Section:     "Color Overrides",
		HideIfEmpty: true,
	},
}

// configKeyMap is a lookup map for configuration keys.
var configKeyMap map[string]ConfigKey

func init() {
	configKeyMap = make(map[string]ConfigKey, len(ConfigKeys))
	for _, key := range ConfigKeys {
		configKeyMap[key.Name] = key
	}
}

// GetConfigKey returns the ConfigKey for a given name.
func GetConfigKey(name string) (ConfigKey, bool) {
	key, ok := configKeyMap[name]
	return key, ok
}

// IsValidConfigKey checks if a key name is valid.
func IsValidConfigKey(name string) bool {
	_, ok := configKeyMap[name]
	return ok
}

// GetDefaultValue returns the default value for a config key.
func GetDefaultValue(name string) (string, bool) {
	if key, ok := configKeyMap[name]; ok {
		return key.Default, true
	}
	return "", false
}

// VisibleConfigKeys returns all non-hidden configuration keys.
func VisibleConfigKeys() []ConfigKey {
	var visible []ConfigKey
	for _, key := range ConfigKeys {
		if !key.Hidden {
			visible = append(visible, key)
		}
	}
	return visible
}

// ConfigSections returns the ordered list of section names.
func ConfigSections() []string {
	return []string{"Display", "Logging", "Export", "Color Overrides"}
}

// ConfigKeysBySection returns visible config keys grouped by section.
func ConfigKeysBySection() map[string][]ConfigKey {
	result := make(map[string][]ConfigKey)
	for _, key := range ConfigKeys {
		if !key.Hidden {
			result[key.Section] = append(result[key.Section], key)
		}
	}
	return result
}
