package cli

import (
	"github.com/footprint-tools/footprint-cli/internal/actions"
	configactions "github.com/footprint-tools/footprint-cli/internal/actions/config"
	logsactions "github.com/footprint-tools/footprint-cli/internal/actions/logs"
	setupactions "github.com/footprint-tools/footprint-cli/internal/actions/setup"
	themeactions "github.com/footprint-tools/footprint-cli/internal/actions/theme"
	trackingactions "github.com/footprint-tools/footprint-cli/internal/actions/tracking"
	updateactions "github.com/footprint-tools/footprint-cli/internal/actions/update"
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
)

func BuildTree() *dispatchers.DispatchNode {
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Track and inspect Git activity across repositories",
		Usage:   "fp [--help] [--version] [--no-color] [--no-pager] <command> [<args>]",
		Flags:   RootFlags,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:        "version",
		Parent:      root,
		Summary:     "Show fp version",
		Description: `Prints the installed version of fp.`,
		Usage:       "fp version",
		Action:      actions.ShowVersion,
		Category:    dispatchers.CategoryInspectActivity,
	})

	addConfigCommands(root)
	addThemeCommands(root)
	addTrackingCommands(root)
	addActivityCommands(root)
	addSetupCommands(root)
	addLogsCommand(root)
	addUpdateCommand(root)
	addHelpCommand(root)

	return root
}

func addConfigCommands(root *dispatchers.DispatchNode) {
	config := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "config",
		Parent:  root,
		Summary: "Manage settings",
		Description: `View and change fp settings.

Settings are stored in ~/.fprc. Use 'fp config list' to see current values.`,
		Usage: "fp config <command>",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:        "get",
		Parent:      config,
		Summary:     "Get a setting value",
		Description: `Prints the value of a setting. Exits with error if not found.`,
		Usage:       "fp config get <key>",
		Args:        ConfigKeyArg,
		Action:      configactions.Get,
		Category:    dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "set",
		Parent:  config,
		Summary: "Change a setting",
		Description: `Sets a configuration value.

Common settings:
  theme               Color theme (e.g., neon-dark, ocean-light)
  export_remote       Git remote for syncing exports
  export_interval_sec Seconds between exports (default: 3600)
  display_date        Date format (dd/mm/yyyy, mm/dd/yyyy, yyyy-mm-dd)
  display_time        Time format (12h, 24h)
  enable_log          Enable logging (true/false)

Example:
  fp config set theme neon-dark`,
		Usage:    "fp config set <key> <value>",
		Args:     ConfigKeyValueArgs,
		Action:   configactions.Set,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "unset",
		Parent:  config,
		Summary: "Remove a setting",
		Description: `Removes a setting from the config file.

Use --all to reset all settings to defaults.`,
		Usage: "fp config unset <key>",
		Flags: ConfigUnsetFlags,
		Args: []dispatchers.ArgSpec{
			{
				Name:        "key",
				Description: "Setting to remove",
				Required:    false,
			},
		},
		Action:   configactions.Unset,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:        "list",
		Parent:      config,
		Summary:     "Show all settings",
		Description: `Prints all current settings from ~/.fprc.`,
		Usage:       "fp config list",
		Action:      configactions.List,
		Category:    dispatchers.CategoryConfig,
	})
}

func addThemeCommands(root *dispatchers.DispatchNode) {
	theme := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "theme",
		Parent:  root,
		Summary: "Change colors",
		Description: `View and change the color theme.

Available themes (add -dark or -light):
  default    Classic terminal colors
  neon       Bright cyberpunk
  aurora     Purple and teal
  mono       Grayscale minimal
  ocean      Cool blues
  sunset     Warm orange to purple
  candy      Soft pastels
  contrast   High readability

Example: fp theme set neon-dark`,
		Usage: "fp theme [command]",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:        "list",
		Parent:      theme,
		Summary:     "Show available themes",
		Description: `Lists all themes. Current theme is marked with *.`,
		Usage:       "fp theme list",
		Action:      themeactions.List,
		Category:    dispatchers.CategoryTheme,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "set",
		Parent:  theme,
		Summary: "Apply a theme",
		Description: `Changes the color theme.

Use -dark for dark terminals, -light for light terminals.

Example: fp theme set ocean-dark`,
		Usage:    "fp theme set <name>",
		Args:     ThemeNameArg,
		Action:   themeactions.Set,
		Category: dispatchers.CategoryTheme,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "pick",
		Parent:  theme,
		Summary: "Browse themes interactively",
		Description: `Opens a visual theme picker with live preview.

Navigate: arrow keys or j/k
Select:   Enter or Space
Cancel:   q or Esc`,
		Usage:    "fp theme pick",
		Action:   themeactions.Pick,
		Category: dispatchers.CategoryTheme,
	})
}

func addTrackingCommands(root *dispatchers.DispatchNode) {
	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "repos",
		Parent:  root,
		Summary: "List tracked repositories",
		Description: `Shows repositories with recorded activity.

Use -i for an interactive view where you can:
  - See all git repos under current directory
  - Check which ones have hooks installed
  - Install or remove hooks easily

Example:
  fp repos              # List repos with activity
  fp repos -i           # Interactive hook manager
  fp repos -i --root ~  # Scan from home directory`,
		Usage:    "fp repos [-i] [--root <path>] [--depth <n>]",
		Flags:    ReposFlags,
		Action:   trackingactions.Repos,
		Category: dispatchers.CategoryInspectActivity,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "record",
		Parent:  root,
		Summary: "Save a git event (internal)",
		Description: `Saves a git event to the database.

This runs automatically via git hooks. You don't need to use it directly.`,
		Usage:    "fp record",
		Flags:    RecordFlags,
		Action:   trackingactions.Record,
		Category: dispatchers.CategoryPlumbing,
	})
}

func addActivityCommands(root *dispatchers.DispatchNode) {
	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "activity",
		Parent:  root,
		Summary: "View your git activity",
		Description: `Shows recent git events across all tracked repositories.

Each entry shows: time, event type, repository, commit/branch info.

Examples:
  fp activity           # Recent events
  fp activity -n 50     # Show 50 events
  fp activity -e        # Include commit messages
  fp activity --repo github.com/user/project  # One repo only`,
		Usage:    "fp activity",
		Action:   trackingactions.Activity,
		Flags:    ActivityFlags,
		Category: dispatchers.CategoryInspectActivity,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "watch",
		Parent:  root,
		Summary: "See events as they happen",
		Description: `Shows new git events in real time.

Events appear as you make commits, switch branches, etc.
Press Ctrl+C to stop.

Examples:
  fp watch       # Stream events live
  fp watch -i    # Interactive dashboard with stats`,
		Usage:    "fp watch",
		Action:   trackingactions.Log,
		Flags:    WatchFlags,
		Category: dispatchers.CategoryInspectActivity,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "export",
		Parent:  root,
		Summary: "Export events to CSV (internal)",
		Description: `Exports events to CSV files. Runs automatically every hour.

You usually don't need this - exports happen in the background.

Use --force to export now (skip the hourly interval).
Use --open to view the export folder.
Use --dry-run to preview without exporting.

Export location: ~/.config/Footprint/exports`,
		Usage:    "fp export [--force] [--dry-run] [--open]",
		Action:   trackingactions.Export,
		Flags:    ExportFlags,
		Category: dispatchers.CategoryPlumbing,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "backfill",
		Parent:  root,
		Summary: "Import past commits",
		Description: `Imports commits that happened before fp was installed.

Scans git history and adds each commit to the database.
Duplicates are skipped automatically.

Examples:
  fp backfill                     # Import all past commits
  fp backfill --since 2024-01-01  # From a specific date
  fp backfill --limit 100         # Only last 100 commits
  fp backfill --dry-run           # Preview without importing`,
		Usage:    "fp backfill [path] [--since=<date>] [--until=<date>] [--limit=<n>]",
		Args:     OptionalRepoPathArg,
		Flags:    BackfillFlags,
		Action:   trackingactions.Backfill,
		Category: dispatchers.CategoryManageRepos,
	})
}

func addSetupCommands(root *dispatchers.DispatchNode) {
	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "setup",
		Parent:  root,
		Summary: "Start tracking this repository",
		Description: `Installs git hooks to record your activity automatically.

After setup, every commit, merge, checkout, rebase, and push in this
repo will be tracked. Run this once per repository.

Existing hooks are backed up before installation.

Example:
  cd /path/to/repo
  fp setup`,
		Usage:    "fp setup [--force]",
		Flags:    SetupFlags,
		Action:   setupactions.Setup,
		Category: dispatchers.CategoryGetStarted,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "teardown",
		Parent:  root,
		Summary: "Stop tracking this repository",
		Description: `Removes fp hooks from the current repository.

If you had hooks before fp, they will be restored from backup.`,
		Usage:    "fp teardown [--force]",
		Flags:    TeardownFlags,
		Action:   setupactions.Teardown,
		Category: dispatchers.CategoryManageRepos,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:        "check",
		Parent:      root,
		Summary:     "Verify hooks are installed",
		Description: `Shows which hooks are installed in the current repository.`,
		Usage:       "fp check",
		Action:      setupactions.Check,
		Category:    dispatchers.CategoryInspectActivity,
	})
}

func addLogsCommand(root *dispatchers.DispatchNode) {
	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "logs",
		Parent:  root,
		Summary: "View fp logs",
		Description: `Shows fp's log file (useful for debugging).

Examples:
  fp logs              # Last 50 lines
  fp logs -n 100       # Last 100 lines
  fp logs -f           # Follow in real time
  fp logs -i           # Interactive viewer
  fp logs --clear      # Delete log file`,
		Usage:    "fp logs [-i] [-n <lines>] [--tail] [--clear]",
		Flags:    LogsFlags,
		Action:   logsAction,
		Category: dispatchers.CategoryInspectActivity,
	})
}

func addUpdateCommand(root *dispatchers.DispatchNode) {
	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "update",
		Parent:  root,
		Summary: "Update to latest version",
		Description: `Downloads and installs a newer version of fp.

Examples:
  fp update          # Install latest release
  fp update v0.1.0   # Install specific version`,
		Usage:    "fp update [version] [--tag]",
		Args:     OptionalVersionArg,
		Flags:    UpdateFlags,
		Action:   updateactions.Update,
		Category: dispatchers.CategoryManageRepos,
	})
}

func addHelpCommand(root *dispatchers.DispatchNode) {
	dispatchers.NewNode(
		"help",
		root,
		"Get help",
		`Shows help for commands and topics.

Examples:
  fp help            # Overview
  fp help setup      # Help for setup command
  fp help -i         # Interactive browser

Interactive mode keys:
  arrows/j/k    Navigate
  /             Search
  Tab           Switch panels
  q/Esc         Exit`,
		"fp help [-i] [command]",
		[]dispatchers.FlagDescriptor{
			{Names: []string{"-i", "--interactive"}, Description: "Browse help interactively"},
		},
		nil,
		nil,
	)
}

// logsAction handles the logs command with its various flags
func logsAction(args []string, flags *dispatchers.ParsedFlags) error {
	if flags.Has("--clear") {
		return logsactions.Clear(args, flags)
	}
	if flags.Has("-i") || flags.Has("--interactive") {
		return logsactions.Interactive(args, flags)
	}
	if flags.Has("--tail") || flags.Has("-f") {
		return logsactions.Tail(args, flags)
	}
	return logsactions.View(args, flags)
}
