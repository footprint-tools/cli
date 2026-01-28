package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildTree_ReturnsRoot(t *testing.T) {
	root := BuildTree()

	require.NotNil(t, root)
	require.Equal(t, "fp", root.Name)
}

func TestBuildTree_HasExpectedTopLevelCommands(t *testing.T) {
	root := BuildTree()

	expectedCommands := []string{
		"version",
		"config",
		"theme",
		"repos",
		"record",
		"activity",
		"watch",
		"export",
		"backfill",
		"setup",
		"teardown",
		"logs",
		"help",
	}

	for _, cmd := range expectedCommands {
		_, found := root.Children[cmd]
		require.True(t, found, "expected top-level command '%s' not found", cmd)
	}
}

func TestBuildTree_ConfigHasSubcommands(t *testing.T) {
	root := BuildTree()

	config, found := root.Children["config"]
	require.True(t, found, "config group not found")

	expectedSubcommands := []string{"get", "set", "unset", "list"}
	for _, sub := range expectedSubcommands {
		_, found := config.Children[sub]
		require.True(t, found, "expected config subcommand '%s' not found", sub)
	}
}

func TestBuildTree_ThemeHasSubcommands(t *testing.T) {
	root := BuildTree()

	theme, found := root.Children["theme"]
	require.True(t, found, "theme group not found")

	expectedSubcommands := []string{"list", "set"}
	for _, sub := range expectedSubcommands {
		_, found := theme.Children[sub]
		require.True(t, found, "expected theme subcommand '%s' not found", sub)
	}
}

func TestBuildTree_ReposHasSubcommands(t *testing.T) {
	root := BuildTree()

	repos, found := root.Children["repos"]
	require.True(t, found, "repos group not found")

	expectedSubcommands := []string{"list", "scan", "check"}
	for _, sub := range expectedSubcommands {
		_, found := repos.Children[sub]
		require.True(t, found, "expected repos subcommand '%s' not found", sub)
	}
}

func TestBuildTree_CommandsHaveActions(t *testing.T) {
	root := BuildTree()

	// Commands that should have actions
	commandsWithActions := []string{
		"version",
		"record",
		"activity",
		"watch",
		"export",
		"backfill",
		"setup",
		"teardown",
		"logs",
	}

	for _, cmdName := range commandsWithActions {
		cmd, found := root.Children[cmdName]
		require.True(t, found, "command '%s' not found", cmdName)
		require.NotNil(t, cmd.Action, "command '%s' should have an action", cmdName)
	}
}

func TestBuildTree_RootHasFlags(t *testing.T) {
	root := BuildTree()

	require.NotEmpty(t, root.Flags, "root should have flags")

	// Check for some expected global flags
	flagNames := make(map[string]bool)
	for _, flag := range root.Flags {
		for _, name := range flag.Names {
			flagNames[name] = true
		}
	}

	require.True(t, flagNames["--help"], "root should have --help flag")
	require.True(t, flagNames["--version"], "root should have --version flag")
	require.True(t, flagNames["--no-color"], "root should have --no-color flag")
}

func TestBuildTree_CommandsHaveUsage(t *testing.T) {
	root := BuildTree()

	require.NotEmpty(t, root.Usage, "root should have usage")

	// Check that children have usage
	for name, child := range root.Children {
		if name != "help" { // help is special
			require.NotEmpty(t, child.Usage, "command '%s' should have usage", name)
		}
	}
}

func TestBuildTree_CommandsHaveSummary(t *testing.T) {
	root := BuildTree()

	require.NotEmpty(t, root.Summary, "root should have summary")

	for name, child := range root.Children {
		require.NotEmpty(t, child.Summary, "command '%s' should have summary", name)
	}
}

func TestBuildTree_GroupsWithInteractiveFlag(t *testing.T) {
	root := BuildTree()

	// All groups have -i flag and InteractiveAction (not Action)
	groups := []string{"config", "theme", "repos"}

	for _, groupName := range groups {
		group, found := root.Children[groupName]
		require.True(t, found, "group '%s' not found", groupName)
		require.Nil(t, group.Action, "group '%s' should not have Action (shows help by default)", groupName)
		require.NotNil(t, group.InteractiveAction, "group '%s' should have InteractiveAction for -i flag", groupName)
		require.NotEmpty(t, group.Children, "group '%s' should have children", groupName)
		require.NotEmpty(t, group.Flags, "group '%s' should have flags", groupName)
	}
}

func TestBuildTree_ThemeHasInteractiveFlag(t *testing.T) {
	root := BuildTree()

	theme, found := root.Children["theme"]
	require.True(t, found, "theme group not found")
	require.NotNil(t, theme.InteractiveAction, "theme should have InteractiveAction for -i flag")
	require.NotEmpty(t, theme.Flags, "theme should have flags")

	// Check for -i flag
	flagNames := make(map[string]bool)
	for _, flag := range theme.Flags {
		for _, name := range flag.Names {
			flagNames[name] = true
		}
	}
	require.True(t, flagNames["-i"], "theme should have -i flag")
	require.True(t, flagNames["--interactive"], "theme should have --interactive flag")
}

func TestBuildTree_SubcommandsHaveActions(t *testing.T) {
	root := BuildTree()

	// Check config subcommands
	config := root.Children["config"]
	for name, child := range config.Children {
		require.NotNil(t, child.Action, "config subcommand '%s' should have an action", name)
	}

	// Check theme subcommands
	theme := root.Children["theme"]
	for name, child := range theme.Children {
		require.NotNil(t, child.Action, "theme subcommand '%s' should have an action", name)
	}

	// Check repos subcommands
	repos := root.Children["repos"]
	for name, child := range repos.Children {
		require.NotNil(t, child.Action, "repos subcommand '%s' should have an action", name)
	}
}

func TestBuildTree_HelpHasNoAction(t *testing.T) {
	root := BuildTree()

	help, found := root.Children["help"]
	require.True(t, found, "help command not found")
	require.Nil(t, help.Action, "help should not have an action (handled specially)")
}

func TestBuildTree_SetupHasFlags(t *testing.T) {
	root := BuildTree()

	setup := root.Children["setup"]
	require.NotEmpty(t, setup.Flags, "setup should have flags")

	// Check for --force flag
	flagNames := make(map[string]bool)
	for _, flag := range setup.Flags {
		for _, name := range flag.Names {
			flagNames[name] = true
		}
	}

	require.True(t, flagNames["--force"], "setup should have --force flag")
}

func TestBuildTree_ActivityHasFlags(t *testing.T) {
	root := BuildTree()

	activity := root.Children["activity"]
	require.NotEmpty(t, activity.Flags, "activity should have flags")

	// Check for expected flags
	flagNames := make(map[string]bool)
	for _, flag := range activity.Flags {
		for _, name := range flag.Names {
			flagNames[name] = true
		}
	}

	require.True(t, flagNames["--oneline"], "activity should have --oneline flag")
	require.True(t, flagNames["--since"], "activity should have --since flag")
	require.True(t, flagNames["--until"], "activity should have --until flag")
	require.True(t, flagNames["--limit"], "activity should have --limit flag")
}
