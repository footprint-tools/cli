package cli

import "github.com/footprint-tools/cli/internal/dispatchers"

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
			Names:       []string{"--quiet", "-q"},
			Description: "Suppress non-essential output (useful for scripts)",
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
			Names:       []string{"--json"},
			Description: "Output as JSON",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-e", "--enrich"},
			Description: "Show commit message and author from git",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-s", "--status"},
			ValueHint:   "<status>",
			Description: "Filter by status: pending, exported, orphaned, skipped",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-S", "--source"},
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
			Names:       []string{"-r", "--repo"},
			ValueHint:   "<id>",
			Description: "Filter by repository id",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-n", "--limit"},
			ValueHint:   "<n>",
			Description: "Limit number of results (shorthand: -<n>, e.g., -50)",
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
			Names:       []string{"--core-hooks-path"},
			Description: "Set git core.hooksPath globally (affects ALL repos on this machine)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--force"},
			Description: "Overwrite existing hooks without prompting",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--dry-run"},
			Description: "Show what would be installed without doing it",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	TeardownFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--core-hooks-path"},
			Description: "Unset git core.hooksPath and remove global hooks",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--force"},
			Description: "Remove hooks without prompting",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--dry-run"},
			Description: "Show what would be removed without doing it",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ReposFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"-i", "--interactive"},
			Description: "Interactive mode to manage hooks across repositories",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--json"},
			Description: "Output as JSON",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--root"},
			ValueHint:   "<path>",
			Description: "Root directory to scan (default: current directory)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--depth"},
			ValueHint:   "<n>",
			Description: "Maximum depth to scan (default: 25)",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ReposScanFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--root"},
			ValueHint:   "<path>",
			Description: "Root directory to scan (default: current directory)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--depth"},
			ValueHint:   "<n>",
			Description: "Maximum depth to scan (default: 25)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--json"},
			Description: "Output as JSON",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ReposCheckFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--json"},
			Description: "Output as JSON",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	WatchFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--oneline"},
			Description: "Show one event per line",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--json"},
			Description: "Output as JSON (one object per line)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-e", "--enrich"},
			Description: "Show commit message and author from git",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-i", "--interactive"},
			Description: "Run in interactive TUI mode with stats and detail view",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-s", "--status"},
			ValueHint:   "<status>",
			Description: "Filter by status: pending, exported, orphaned, skipped",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-S", "--source"},
			ValueHint:   "<source>",
			Description: "Filter by source (post-commit, post-rewrite, post-checkout, post-merge, pre-push, manual, backfill)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"-r", "--repo"},
			ValueHint:   "<id>",
			Description: "Filter by repository id",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ExportFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--now"},
			Description: "Export immediately, ignoring the interval",
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
			Names:       []string{"--json"},
			Description: "Output as JSON",
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
			Names:       []string{"--limit", "-n"},
			ValueHint:   "<n>",
			Description: "Limit number of commits to import (shorthand: -<n>)",
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
			Names:       []string{"--json"},
			Description: "Output as JSON",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	LogsFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"-i", "--interactive"},
			Description: "Interactive log viewer with filtering and navigation",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--tail", "-f"},
			Description: "Follow logs in real time",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--clear"},
			Description: "Clear the log file",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--limit", "-n"},
			ValueHint:   "<n>",
			Description: "Number of lines to show (default: 50, shorthand: -<n>)",
			Scope:       dispatchers.FlagScopeLocal,
		},
		{
			Names:       []string{"--json"},
			Description: "Output as JSON array of log entries",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	UpdateFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--tag"},
			Description: "Install from git tag using go install (requires Go)",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	CompletionsFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--script"},
			Description: "Print completion script to stdout (for eval)",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	VersionFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--json"},
			Description: "Output as JSON",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ThemeFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"-i", "--interactive"},
			Description: "Interactive theme picker with live preview",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ConfigFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"-i", "--interactive"},
			Description: "Interactive settings editor",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}

	ConfigListFlags = []dispatchers.FlagDescriptor{
		{
			Names:       []string{"--json"},
			Description: "Output as JSON",
			Scope:       dispatchers.FlagScopeLocal,
		},
	}
)
