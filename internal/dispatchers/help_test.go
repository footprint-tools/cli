package dispatchers

import (
	"testing"

	"github.com/footprint-tools/cli/internal/help"
	"github.com/stretchr/testify/require"
)

func TestFormatUsage_CommandOnly(t *testing.T) {
	result := formatUsage("fp version")
	require.NotEmpty(t, result)
	require.Contains(t, result, "version")
}

func TestFormatUsage_CommandWithBrackets(t *testing.T) {
	result := formatUsage("fp track [path]")
	require.NotEmpty(t, result)
	require.Contains(t, result, "track")
}

func TestFormatUsage_CommandWithAngleBrackets(t *testing.T) {
	result := formatUsage("fp config set <key> <value>")
	require.NotEmpty(t, result)
	require.Contains(t, result, "config set")
}

func TestCollectLeafCommands_SingleCommand(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{Name: "fp"})
	Command(CommandSpec{
		Name:   "version",
		Parent: root,
		Action: action,
	})

	var leaves []*DispatchNode
	collectLeafCommands(root, &leaves)

	require.Len(t, leaves, 1)
	require.Equal(t, "version", leaves[0].Name)
}

func TestCollectLeafCommands_NestedCommands(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{Name: "fp"})

	config := Group(GroupSpec{
		Name:   "config",
		Parent: root,
	})

	Command(CommandSpec{
		Name:   "get",
		Parent: config,
		Action: action,
	})

	Command(CommandSpec{
		Name:   "set",
		Parent: config,
		Action: action,
	})

	var leaves []*DispatchNode
	collectLeafCommands(root, &leaves)

	require.Len(t, leaves, 2)
}

func TestCollectLeafCommands_GroupWithoutAction(t *testing.T) {
	root := Root(RootSpec{Name: "fp"})

	Group(GroupSpec{
		Name:   "emptygroup",
		Parent: root,
	})

	var leaves []*DispatchNode
	collectLeafCommands(root, &leaves)

	// Empty groups don't get collected
	require.Empty(t, leaves)
}

func TestTopicHelpAction(t *testing.T) {
	topic := help.LookupTopic("overview")
	require.NotNil(t, topic)

	action := TopicHelpAction(topic)
	require.NotNil(t, action)

	// Should not panic
	err := action(nil, nil)
	require.NoError(t, err)
}

func TestTopicsListAction(t *testing.T) {
	action := TopicsListAction()
	require.NotNil(t, action)

	// Should not panic
	err := action(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_Root(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	Command(CommandSpec{
		Name:     "version",
		Parent:   root,
		Summary:  "Show version",
		Category: CategoryInspectActivity,
		Action:   action,
	})

	helpAction := HelpAction(root, root)
	require.NotNil(t, helpAction)

	// Should not panic
	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_Subcommand(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	config := Group(GroupSpec{
		Name:        "config",
		Parent:      root,
		Summary:     "Manage configuration",
		Description: "Commands for managing configuration",
		Usage:       "fp config <subcommand>",
	})

	Command(CommandSpec{
		Name:    "get",
		Parent:  config,
		Summary: "Get a config value",
		Action:  action,
	})

	helpAction := HelpAction(config, root)
	require.NotNil(t, helpAction)

	// Should not panic
	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_WithFlags(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	Command(CommandSpec{
		Name:    "track",
		Parent:  root,
		Summary: "Track a repository",
		Usage:   "fp track [path]",
		Flags: []FlagDescriptor{
			{Names: []string{"--remote", "-r"}, ValueHint: "<name>", Description: "Remote name"},
		},
		Action: action,
	})

	track := root.Children["track"]
	helpAction := HelpAction(track, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_SubcommandWithDescription(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	Command(CommandSpec{
		Name:        "version",
		Parent:      root,
		Summary:     "Show version",
		Description: "This is a detailed description of the version command.",
		Usage:       "fp version",
		Action:      action,
	})

	version := root.Children["version"]
	helpAction := HelpAction(version, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_SubcommandNoSummary(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:  "fp",
		Usage: "fp <command>",
	})

	Command(CommandSpec{
		Name:   "cmd",
		Parent: root,
		Usage:  "fp cmd",
		Action: action,
	})

	cmd := root.Children["cmd"]
	helpAction := HelpAction(cmd, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_RootWithMultipleCategories(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	Command(CommandSpec{
		Name:     "setup",
		Parent:   root,
		Summary:  "Setup hooks",
		Category: CategoryGetStarted,
		Action:   action,
	})

	Command(CommandSpec{
		Name:     "activity",
		Parent:   root,
		Summary:  "View activity",
		Category: CategoryInspectActivity,
		Action:   action,
	})

	Command(CommandSpec{
		Name:     "track",
		Parent:   root,
		Summary:  "Track repo",
		Category: CategoryGetStarted,
		Action:   action,
	})

	helpAction := HelpAction(root, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_GroupWithChildren(t *testing.T) {
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	config := Group(GroupSpec{
		Name:        "config",
		Parent:      root,
		Summary:     "Configuration commands",
		Description: "Manage configuration settings",
		Usage:       "fp config <subcommand>",
	})

	Command(CommandSpec{
		Name:    "get",
		Parent:  config,
		Summary: "Get config value",
		Action:  action,
	})

	Command(CommandSpec{
		Name:    "set",
		Parent:  config,
		Summary: "Set config value",
		Action:  action,
	})

	Command(CommandSpec{
		Name:    "list",
		Parent:  config,
		Summary: "List config values",
		Action:  action,
	})

	helpAction := HelpAction(config, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestFormatUsage_EmptyUsage(t *testing.T) {
	result := formatUsage("")
	require.NotNil(t, result)
}

func TestFormatUsage_OnlyCommand(t *testing.T) {
	result := formatUsage("fp")
	require.Contains(t, result, "fp")
}

func TestHelpAction_SortingWithDisplayOrder(t *testing.T) {
	// Create a tree with commands that have explicit display order
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	// Commands with explicit display order in commandDisplayOrder
	Command(CommandSpec{
		Name:     "setup",
		Parent:   root,
		Summary:  "Setup hooks",
		Category: CategoryGetStarted,
		Action:   action,
	})

	Command(CommandSpec{
		Name:     "track",
		Parent:   root,
		Summary:  "Track repo",
		Category: CategoryGetStarted,
		Action:   action,
	})

	helpAction := HelpAction(root, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_SortingWithMixedDisplayOrder(t *testing.T) {
	// Create tree with one command in display order and one not
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	// setup is in commandDisplayOrder
	Command(CommandSpec{
		Name:     "setup",
		Parent:   root,
		Summary:  "Setup hooks",
		Category: CategoryGetStarted,
		Action:   action,
	})

	// "aaa-custom" is not in commandDisplayOrder
	Command(CommandSpec{
		Name:     "aaa-custom",
		Parent:   root,
		Summary:  "Custom command",
		Category: CategoryGetStarted,
		Action:   action,
	})

	helpAction := HelpAction(root, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_SortingAlphabetical(t *testing.T) {
	// Create tree with commands not in display order (sorted alphabetically)
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	// Neither of these are in commandDisplayOrder
	Command(CommandSpec{
		Name:     "zebra-cmd",
		Parent:   root,
		Summary:  "Zebra command",
		Category: CategoryPlumbing,
		Action:   action,
	})

	Command(CommandSpec{
		Name:     "alpha-cmd",
		Parent:   root,
		Summary:  "Alpha command",
		Category: CategoryPlumbing,
		Action:   action,
	})

	helpAction := HelpAction(root, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_GroupChildSortingWithDisplayOrder(t *testing.T) {
	// Test sorting for children within a group (subcommand help)
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	config := Group(GroupSpec{
		Name:    "config",
		Parent:  root,
		Summary: "Configuration commands",
		Usage:   "fp config <subcommand>",
	})

	// These have display order: "config get": 1, "config set": 2
	Command(CommandSpec{
		Name:    "set",
		Parent:  config,
		Summary: "Set config value",
		Action:  action,
	})

	Command(CommandSpec{
		Name:    "get",
		Parent:  config,
		Summary: "Get config value",
		Action:  action,
	})

	// This one doesn't have display order
	Command(CommandSpec{
		Name:    "custom",
		Parent:  config,
		Summary: "Custom subcommand",
		Action:  action,
	})

	helpAction := HelpAction(config, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_SortingOnlySecondHasOrder(t *testing.T) {
	// Create tree where second command (sorted alphabetically) has display order
	// but first doesn't - tests "if hasJ { return false }" branch
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	// "aaa-custom" comes first alphabetically but has no display order
	Command(CommandSpec{
		Name:     "aaa-custom",
		Parent:   root,
		Summary:  "Custom command first alphabetically",
		Category: CategoryGetStarted,
		Action:   action,
	})

	// "track" comes after alphabetically but HAS display order
	Command(CommandSpec{
		Name:     "track",
		Parent:   root,
		Summary:  "Track repo",
		Category: CategoryGetStarted,
		Action:   action,
	})

	helpAction := HelpAction(root, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_GroupChildSortingAlphabetical(t *testing.T) {
	// Test children sorting where neither has display order
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	// Create a custom group with children that have no display order
	custom := Group(GroupSpec{
		Name:    "custom",
		Parent:  root,
		Summary: "Custom commands",
		Usage:   "fp custom <subcommand>",
	})

	// Neither of these are in commandDisplayOrder
	Command(CommandSpec{
		Name:    "zebra",
		Parent:  custom,
		Summary: "Zebra subcommand",
		Action:  action,
	})

	Command(CommandSpec{
		Name:    "alpha",
		Parent:  custom,
		Summary: "Alpha subcommand",
		Action:  action,
	})

	helpAction := HelpAction(custom, root)
	require.NotNil(t, helpAction)

	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_GroupChildOnlyFirstHasOrder(t *testing.T) {
	// Test children sorting where only first has display order (hasI branch)
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	config := Group(GroupSpec{
		Name:    "config",
		Parent:  root,
		Summary: "Config commands",
		Usage:   "fp config <subcommand>",
	})

	// "config get" has display order (1), "zzz" does not
	Command(CommandSpec{
		Name:    "get",
		Parent:  config,
		Summary: "Get value",
		Action:  action,
	})

	Command(CommandSpec{
		Name:    "zzz",
		Parent:  config,
		Summary: "ZZZ subcommand",
		Action:  action,
	})

	helpAction := HelpAction(config, root)
	err := helpAction(nil, nil)
	require.NoError(t, err)
}

func TestHelpAction_GroupChildOnlySecondHasOrder(t *testing.T) {
	// Test children sorting where only second has display order (hasJ branch)
	action := func(args []string, flags *ParsedFlags) error { return nil }

	root := Root(RootSpec{
		Name:    "fp",
		Summary: "Footprint CLI",
		Usage:   "fp <command>",
	})

	config := Group(GroupSpec{
		Name:    "config",
		Parent:  root,
		Summary: "Config commands",
		Usage:   "fp config <subcommand>",
	})

	// "aaa" has no display order, "config set" has display order (2)
	Command(CommandSpec{
		Name:    "aaa",
		Parent:  config,
		Summary: "AAA subcommand",
		Action:  action,
	})

	Command(CommandSpec{
		Name:    "set",
		Parent:  config,
		Summary: "Set value",
		Action:  action,
	})

	helpAction := HelpAction(config, root)
	err := helpAction(nil, nil)
	require.NoError(t, err)
}
