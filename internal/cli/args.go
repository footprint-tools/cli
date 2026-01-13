package cli

import "github.com/Skryensya/footprint/internal/dispatchers"

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
)
