package cli

import "github.com/footprint-tools/cli/internal/dispatchers"

var (
	ConfigKeyArg = []dispatchers.ArgSpec{
		{
			Name:        "key",
			Description: "Configuration key",
			Required:    true,
		},
	}

	ConfigKeyValueArgs = []dispatchers.ArgSpec{
		{
			Name:        "key",
			Description: "Configuration key",
			Required:    true,
		},
		{
			Name:        "value",
			Description: "Value to assign",
			Required:    true,
		},
	}

	OptionalRepoPathArg = []dispatchers.ArgSpec{
		{
			Name:        "path",
			Description: "Path to a git repository (defaults to current directory)",
			Required:    false,
		},
	}

	ThemeNameArg = []dispatchers.ArgSpec{
		{
			Name:        "name",
			Description: "Theme name (e.g., default-dark, neon-light)",
			Required:    true,
		},
	}

	OptionalVersionArg = []dispatchers.ArgSpec{
		{
			Name:        "version",
			Description: "Version to install (e.g., v0.1.0)",
			Required:    false,
		},
	}
)
