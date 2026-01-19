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
		{
			Names:       []string{"--no-color"},
			Description: "Disable colored output",
			Scope:       dispatchers.FlagScopeGlobal,
		},
		{
			Names:       []string{"--no-pager"},
			Description: "Do not use pager for output",
			Scope:       dispatchers.FlagScopeGlobal,
		},
		{
			Names:       []string{"--pager"},
			ValueHint:   "<cmd>",
			Description: "Use specified pager for this command",
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
			Names:       []string{"--enrich", "-e"},
			Description: "Show commit message and author from git",
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
			Description: "Filter by source (post-commit, post-rewrite, post-checkout, post-merge, pre-push, manual, backfill)",
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

	WatchFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--oneline"},
			Description: "Show one event per line",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--enrich", "-e"},
			Description: "Show commit message and author from git",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ExportFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--force"},
			Description: "Export even if interval has not passed",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--dry-run"},
			Description: "Show what would be exported without doing it",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--open"},
			Description: "Open the export directory in file manager",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--set-remote"},
			ValueHint:   "<url>",
			Description: "Set the remote URL for the export repository",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	TrackFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--remote"},
			ValueHint:   "<name>",
			Description: "Use specified remote instead of 'origin'",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	UntrackFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--id"},
			ValueHint:   "<repo-id>",
			Description: "Untrack by repository ID instead of path (useful for orphaned repos)",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	BackfillFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--since"},
			ValueHint:   "<date>",
			Description: "Import commits after date (YYYY-MM-DD)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--until"},
			ValueHint:   "<date>",
			Description: "Import commits before date (YYYY-MM-DD)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--limit"},
			ValueHint:   "<n>",
			Description: "Limit number of commits to import",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--branch"},
			ValueHint:   "<name>",
			Description: "Use this branch name for all commits (default: infer)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--dry-run"},
			Description: "Show what would be imported without doing it",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--background"},
			Description: "Run in background mode (internal)",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}
)
