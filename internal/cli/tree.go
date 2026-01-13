package cli

import (
	"github.com/Skryensya/footprint/internal/actions"
	"github.com/Skryensya/footprint/internal/dispatchers"
)

func BuildTree() *dispatchers.DispatchNode {
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Track your work across repositories",
		Usage:   "fp <command> [flags]",
		Flags:   RootFlags,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "version",
		Parent:   root,
		Summary:  "Show current fp version",
		Usage:    "fp version",
		Action:   actions.ShowVersion,
		Category: dispatchers.CategoryInfo,
	})

	// -- config

	config := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "config",
		Parent:  root,
		Summary: "Manage configuration",
		Usage:   "fp config <command>",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "get",
		Parent:   config,
		Summary:  "Get a config value",
		Usage:    "fp config get <key>",
		Args:     ConfigKeyArg,
		Action:   actions.ConfigGet,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "set",
		Parent:   config,
		Summary:  "Set a config value",
		Usage:    "fp config set <key> <value>",
		Args:     ConfigKeyValueArgs,
		Action:   actions.ConfigSet,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "unset",
		Parent:  config,
		Summary: "Remove a config value",
		Usage:   "fp config unset <key>",
		Flags:   ConfigUnsetFlags,
		Args: []dispatchers.ArgSpec{
			{
				Name:        "key",
				Description: "Configuration key to delete",
				Required:    false,
			},
		},
		Action:   actions.ConfigUnset,
		Category: dispatchers.CategoryConfig,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "list",
		Parent:   config,
		Summary:  "List all the configuration as key=value pairs",
		Usage:    "fp config list",
		Action:   actions.ConfigList,
		Category: dispatchers.CategoryConfig,
	})

	// -- repo

	repo := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "repo",
		Parent:  root,
		Summary: "Manage repository tracking",
		Usage:   "fp repo <command>",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "track",
		Parent:   repo,
		Summary:  "Start tracking a repository",
		Usage:    "fp repo track <path>",
		Args:     OptionalRepoPathArg,
		Action:   actions.RepoTrack,
		Category: dispatchers.CategoryRepo,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "untrack",
		Parent:   repo,
		Summary:  "Stop tracking a repository",
		Usage:    "fp repo untrack <path>",
		Args:     OptionalRepoPathArg,
		Action:   actions.RepoUntrack,
		Category: dispatchers.CategoryRepo,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "list",
		Parent:   repo,
		Summary:  "Show all the repositories being tracked",
		Usage:    "fp repo list",
		Action:   actions.RepoList,
		Category: dispatchers.CategoryRepo,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "status",
		Parent:   repo,
		Summary:  "Show repository tracking status",
		Usage:    "fp repo status <path>",
		Args:     OptionalRepoPathArg,
		Action:   actions.RepoStatus,
		Category: dispatchers.CategoryRepo,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "adopt-remote",
		Parent:   repo,
		Summary:  "Update repository id to reflect new remote",
		Usage:    "fp repo adopt-remote <path>",
		Args:     OptionalRepoPathArg,
		Action:   actions.RepoAdoptRemote,
		Category: dispatchers.CategoryRepo,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "record",
		Parent:   repo,
		Summary:  "Record last commit footprint",
		Usage:    "fp repo record",
		Flags:    RepoRecordFlags,
		Action:   actions.RepoRecord,
		Category: dispatchers.CategoryRepo,
	})

	// -- activity

	activity := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "activity",
		Parent:  root,
		Summary: "Show recorded repository activity",
		Usage:   "fp activity <command>",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "list",
		Parent:   activity,
		Summary:  "List recorded activity (newest first)",
		Usage:    "fp activity list",
		Action:   actions.ActivityList,
		Flags:    ActivityListFlags,
		Category: dispatchers.CategoryInfo,
	})

	// -- hooks

	hooks := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "hooks",
		Parent:  root,
		Summary: "Manage git hook installation",
		Usage:   "fp hooks <command>",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "install",
		Parent:   hooks,
		Summary:  "Install fp git hooks",
		Usage:    "fp hooks install (--repo | --global)",
		Flags:    HooksInstallFlags,
		Action:   actions.HooksInstall,
		Category: dispatchers.CategoryInfo,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "status",
		Parent:   hooks,
		Summary:  "Show installed fp hooks",
		Usage:    "fp hooks status [--global]",
		Flags:    HooksStatusFlags,
		Action:   actions.HooksStatus,
		Category: dispatchers.CategoryInfo,
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:     "uninstall",
		Parent:   hooks,
		Summary:  "Remove fp git hooks",
		Usage:    "fp hooks uninstall (--repo | --global)",
		Flags:    HooksUninstallFlags,
		Action:   actions.HooksUninstall,
		Category: dispatchers.CategoryInfo,
	})

	// -- help

	dispatchers.NewNode(
		"help",
		root,
		"Show help for a command",
		"fp help [command]",
		nil,
		nil,
		nil,
	)

	return root
}
