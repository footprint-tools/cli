package cli

import (
	"github.com/Skryensya/footprint/internal/actions"
	configactions "github.com/Skryensya/footprint/internal/actions/config"
	setupactions "github.com/Skryensya/footprint/internal/actions/setup"
	trackingactions "github.com/Skryensya/footprint/internal/actions/tracking"
	"github.com/Skryensya/footprint/internal/dispatchers"
)

func BuildTree() *dispatchers.DispatchNode {
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Track and inspect Git activity across repositories",
		Usage:   "fp [--help] [--version] <command> [<args>]",
		Flags:   RootFlags,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "version",
		Parent:  root,
		Summary: "Show current fp version",
		Description: `Prints the version number of the fp binary.

The version includes the git tag, commit count since tag, and commit hash
when built from a non-release commit.`,
		Usage:    "fp version",
		Action:   actions.ShowVersion,
		Category: dispatchers.CategoryInspectActivity,
	})

	// -- config

	config := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "config",
		Parent:  root,
		Summary: "Manage configuration",
		Description: `Read and write fp configuration values.

Configuration is stored in ~/.fprc as simple key=value pairs.
Use 'fp config list' to see all current settings.`,
		Usage: "fp config <command>",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "get",
		Parent:  config,
		Summary: "Get a config value",
		Description: `Prints the value of a configuration key.

If the key does not exist, nothing is printed and the command exits
with a non-zero status.`,
		Usage:    "fp config get <key>",
		Args:     ConfigKeyArg,
		Action:   configactions.Get,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "set",
		Parent:  config,
		Summary: "Set a config value",
		Description: `Sets a configuration key to the specified value.

If the key already exists, its value is overwritten.
The configuration file is created if it does not exist.`,
		Usage:    "fp config set <key> <value>",
		Args:     ConfigKeyValueArgs,
		Action:   configactions.Set,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "unset",
		Parent:  config,
		Summary: "Remove a config value",
		Description: `Removes a configuration key from the config file.

Use --all to remove all configuration values and reset to defaults.`,
		Usage: "fp config unset <key>",
		Flags: ConfigUnsetFlags,
		Args: []dispatchers.ArgSpec{
			{
				Name:        "key",
				Description: "Configuration key to delete",
				Required:    false,
			},
		},
		Action:   configactions.Unset,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "list",
		Parent:  config,
		Summary: "List all the configuration as key=value pairs",
		Description: `Prints all configuration values in key=value format.

This shows the current state of ~/.fprc. If no configuration has been
set, the output will be empty.`,
		Usage:    "fp config list",
		Action:   configactions.List,
		Category: dispatchers.CategoryConfig,
	})

	// -- tracking

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "track",
		Parent:  root,
		Summary: "Start tracking a repository",
		Description: `Marks a git repository for activity tracking.

When a repository is tracked, fp records git events (commits, merges,
checkouts, rebases, pushes) that occur in it. Events are stored locally
in a SQLite database.

The repository is identified by its remote URL (usually 'origin'). If no
remote exists, a hash of the local path is used instead.

If no path is provided, the current directory is used.`,
		Usage:    "fp track [path]",
		Args:     OptionalRepoPathArg,
		Action:   trackingactions.Track,
		Category: dispatchers.CategoryGetStarted,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "untrack",
		Parent:  root,
		Summary: "Stop tracking a repository",
		Description: `Removes a repository from activity tracking.

Future git events in this repository will no longer be recorded.
Existing recorded events are not deleted.

If no path is provided, the current directory is used.`,
		Usage:    "fp untrack [path]",
		Args:     OptionalRepoPathArg,
		Action:   trackingactions.Untrack,
		Category: dispatchers.CategoryManageRepos,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "repos",
		Parent:  root,
		Summary: "Show all tracked repositories",
		Description: `Lists all repositories currently being tracked by fp.

Each entry shows the repository identifier and, when available, the
local path where it was last seen.`,
		Usage:    "fp repos",
		Action:   trackingactions.Repos,
		Category: dispatchers.CategoryInspectActivity,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "list",
		Parent:  root,
		Summary: "Show all tracked repositories (alias for repos)",
		Description: `Lists all repositories currently being tracked by fp.

This is an alias for 'fp repos'.`,
		Usage:    "fp list",
		Action:   trackingactions.Repos,
		Category: dispatchers.CategoryInspectActivity,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "status",
		Parent:  root,
		Summary: "Show repository tracking status",
		Description: `Shows whether a repository is currently being tracked.

Displays the repository identifier, tracking state, and hook installation
status for the given path.

If no path is provided, the current directory is used.`,
		Usage:    "fp status [path]",
		Args:     OptionalRepoPathArg,
		Action:   trackingactions.Status,
		Category: dispatchers.CategoryInspectActivity,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "sync-remote",
		Parent:  root,
		Summary: "Update repository id to reflect new remote",
		Description: `Updates the repository identifier after a remote URL change.

Repository IDs are derived from the remote URL. If you change the remote
(e.g., after transferring a repo to a new host), run this command to
update the tracking ID. This ensures future events are associated with
the new identifier.

If no path is provided, the current directory is used.`,
		Usage:    "fp sync-remote [path]",
		Args:     OptionalRepoPathArg,
		Action:   trackingactions.Adopt,
		Category: dispatchers.CategoryManageRepos,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "record",
		Parent:  root,
		Summary: "Record a git event (invoked automatically by git hooks)",
		Description: `Records a git event to the local database.

This command is normally invoked automatically by git hooks installed
via 'fp setup'. You typically don't need to run it manually.

When called, it checks if the current repository is tracked. If so, it
records the event. If not, it exits silently without error.`,
		Usage:    "fp record",
		Flags:    RecordFlags,
		Action:   trackingactions.Record,
		Category: dispatchers.CategoryPlumbing,
	})

	// -- activity

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "activity",
		Parent:  root,
		Summary: "List recorded activity (newest first)",
		Description: `Shows recorded git events across all tracked repositories.

Events are displayed in reverse chronological order (newest first).
Each entry shows the timestamp, event type, repository, and relevant
details like commit SHA or branch name.

Use -n to limit the number of entries shown.`,
		Usage:    "fp activity",
		Action:   trackingactions.Activity,
		Flags:    ActivityFlags,
		Category: dispatchers.CategoryInspectActivity,
	})

	// -- setup

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "setup",
		Parent:  root,
		Summary: "Install fp git hooks",
		Description: `Installs git hooks that automatically record activity.

By default, hooks are installed in the current repository's .git/hooks/
directory. Use --global to install hooks in git's global hooks directory,
which applies to all repositories.

Installed hooks:
  post-commit     Records commits
  post-merge      Records merges (including pulls)
  post-checkout   Records branch switches
  post-rewrite    Records rebases and amends
  pre-push        Records push attempts

If existing hooks are found, they are backed up before installation.`,
		Usage:    "fp setup [--global]",
		Flags:    SetupFlags,
		Action:   setupactions.Setup,
		Category: dispatchers.CategoryGetStarted,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "teardown",
		Parent:  root,
		Summary: "Remove fp git hooks",
		Description: `Removes git hooks installed by fp.

By default, removes hooks from the current repository. Use --global to
remove hooks from git's global hooks directory.

If hooks were backed up during installation, the original hooks are
restored.`,
		Usage:    "fp teardown [--global]",
		Flags:    TeardownFlags,
		Action:   setupactions.Teardown,
		Category: dispatchers.CategoryManageRepos,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "check",
		Parent:  root,
		Summary: "Show installed fp hooks status",
		Description: `Shows the installation status of fp git hooks.

Displays which hooks are installed, whether they are fp hooks or
third-party hooks, and if any backups exist.

Use --global to check the global hooks directory instead of the
current repository.`,
		Usage:    "fp check [--global]",
		Flags:    CheckFlags,
		Action:   setupactions.Check,
		Category: dispatchers.CategoryInspectActivity,
	})

	// -- help

	dispatchers.NewNode(
		"help",
		root,
		"Show help for a command",
		"", // description not needed for help itself
		"fp help [command]",
		nil,
		nil,
		nil,
	)

	return root
}
