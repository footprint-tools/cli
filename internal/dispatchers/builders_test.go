package dispatchers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNode_NoParent(t *testing.T) {
	node := NewNode(
		"test",
		nil,
		"summary",
		"description",
		"usage",
		nil,
		nil,
		nil,
	)

	require.NotNil(t, node)
	require.Equal(t, "test", node.Name)
	require.Equal(t, "summary", node.Summary)
	require.Equal(t, "description", node.Description)
	require.Equal(t, "usage", node.Usage)
	require.Equal(t, []string{"test"}, node.Path)
	require.NotNil(t, node.Children)
}

func TestNewNode_WithParent(t *testing.T) {
	parent := NewNode("parent", nil, "", "", "", nil, nil, nil)
	child := NewNode("child", parent, "child summary", "", "", nil, nil, nil)

	require.Equal(t, []string{"parent", "child"}, child.Path)
	require.Contains(t, parent.Children, "child")
	require.Equal(t, child, parent.Children["child"])
}

func TestNewNode_WithFlags(t *testing.T) {
	flags := []FlagDescriptor{
		{Names: []string{"--verbose", "-v"}},
	}

	node := NewNode("test", nil, "", "", "", flags, nil, nil)

	require.Len(t, node.Flags, 1)
	require.Contains(t, node.Flags[0].Names, "--verbose")
}

func TestNewNode_WithArgs(t *testing.T) {
	args := []ArgSpec{
		{Name: "path", Required: true},
	}

	node := NewNode("test", nil, "", "", "", nil, args, nil)

	require.Len(t, node.Args, 1)
	require.Equal(t, "path", node.Args[0].Name)
}

func TestNewNode_WithAction(t *testing.T) {
	called := false
	action := func(args []string, flags *ParsedFlags) error {
		called = true
		return nil
	}

	node := NewNode("test", nil, "", "", "", nil, nil, action)

	require.NotNil(t, node.Action)
	err := node.Action(nil, nil)
	require.NoError(t, err)
	require.True(t, called)
}

func TestRoot(t *testing.T) {
	root := Root(RootSpec{
		Name:        "fp",
		Summary:     "Footprint CLI",
		Description: "A tracking tool",
		Usage:       "fp <command>",
		Flags:       []FlagDescriptor{{Names: []string{"--version"}}},
	})

	require.NotNil(t, root)
	require.Equal(t, "fp", root.Name)
	require.Equal(t, "Footprint CLI", root.Summary)
	require.Equal(t, []string{"fp"}, root.Path)
	require.Len(t, root.Flags, 1)
}

func TestGroup(t *testing.T) {
	parent := Root(RootSpec{Name: "fp"})
	group := Group(GroupSpec{
		Name:        "config",
		Parent:      parent,
		Summary:     "Configuration commands",
		Description: "Manage configuration",
		Usage:       "fp config <subcommand>",
	})

	require.NotNil(t, group)
	require.Equal(t, "config", group.Name)
	require.Equal(t, []string{"fp", "config"}, group.Path)
	require.Contains(t, parent.Children, "config")
}

func TestCommand(t *testing.T) {
	parent := Root(RootSpec{Name: "fp"})
	action := func(args []string, flags *ParsedFlags) error { return nil }

	cmd := Command(CommandSpec{
		Name:        "track",
		Parent:      parent,
		Summary:     "Track a repository",
		Description: "Start tracking a git repository",
		Usage:       "fp track [path]",
		Category:    CategoryGetStarted,
		Flags:       []FlagDescriptor{{Names: []string{"--all"}}},
		Args:        []ArgSpec{{Name: "path"}},
		Action:      action,
	})

	require.NotNil(t, cmd)
	require.Equal(t, "track", cmd.Name)
	require.Equal(t, CategoryGetStarted, cmd.Category)
	require.Equal(t, []string{"fp", "track"}, cmd.Path)
	require.Contains(t, parent.Children, "track")
	require.NotNil(t, cmd.Action)
}
