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

	ActivityFlags = []dispatchers.FlagDescriptor{
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
		{
			Names:       []string{"--since"},
			ValueHint:   "<date>",
			Description: "Show events after date (YYYY-MM-DD)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--until"},
			ValueHint:   "<date>",
			Description: "Show events before date (YYYY-MM-DD)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--repo"},
			ValueHint:   "<id>",
			Description: "Filter by repository id",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--limit"},
			ValueHint:   "<n>",
			Description: "Limit number of results",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	RecordFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--verbose"},
			Description: "Print details about the recorded commit",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--manual"},
			Description: "Acknowledge manual execution (suppresses note)",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	SetupFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--repo"},
			Description: "Install hooks in the current repository (default)",
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

	TeardownFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--repo"},
			Description: "Remove hooks from the current repository (default)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--global"},
			Description: "Remove global hooks",
			Scope:       dispatchers.FlagScopeGlobal,
		},
		{
			Names:       []string{"--force"},
			Description: "Skip confirmation prompt",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	CheckFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--repo"},
			Description: "Check hooks in the current repository (default)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--global"},
			Description: "Check global hooks installation",
			Scope:       dispatchers.FlagScopeGlobal,
		},
	}
)
