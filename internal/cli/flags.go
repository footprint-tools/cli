package cli

import "github.com/Skryensya/footprint/internal/dispatchers"

var (
	RootFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--help", "-h"},
			Description: "Show help",
			Scope:       dispatchers.FlagScopeGlobal,
		},
		{
			Names:       []string{"--version", "-v"},
			Description: "Show version",
			Scope:       dispatchers.FlagScopeGlobal,
		},
	}

	ConfigUnsetFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--all"},
			Description: "Delete all the config key=value pairs",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ActivityListFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--oneline"},
			Description: "Show one event per line",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--status"},
			ValueHint:   "<status>",
			Description: "Filter by status (pending, exported, orphaned, skipped)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--source"},
			ValueHint:   "<source>",
			Description: "Filter by source (post-commit, post-rewrite, post-checkout, post-merge, pre-push, manual)",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	RepoRecordFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--verbose"},
			Description: "Print details about the recorded commit",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	HooksInstallFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--repo"},
			Description: "Install hooks in the current repository",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--global"},
			Description: "Install hooks globally",
			Scope:       dispatchers.FlagScopeGlobal,
		},
		{
			Names:       []string{"--force"},
			Description: "Skip confirmation prompt",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	HooksStatusFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--global"},
			Description: "Check global hooks installation",
			Scope:       dispatchers.FlagScopeGlobal,
		},
	}

	HooksUninstallFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--repo"},
			Description: "Uninstall hooks from the current repository",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--global"},
			Description: "Uninstall global hooks",
			Scope:       dispatchers.FlagScopeGlobal,
		},
		{
			Names:       []string{"--force"},
			Description: "Skip confirmation prompt",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}
)
