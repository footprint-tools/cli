package completions

import (
	"testing"

	"github.com/footprint-tools/cli/internal/dispatchers"
)

func TestExtractCommands(t *testing.T) {
	// Build a simple command tree
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Test CLI",
		Flags: []dispatchers.FlagDescriptor{
			{Names: []string{"--help", "-h"}, Description: "Show help"},
		},
	})

	config := dispatchers.Group(dispatchers.GroupSpec{
		Name:    "config",
		Parent:  root,
		Summary: "Manage settings",
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "get",
		Parent:  config,
		Summary: "Get a setting",
		Flags: []dispatchers.FlagDescriptor{
			{Names: []string{"--json"}, Description: "Output as JSON"},
		},
	})

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "set",
		Parent:  config,
		Summary: "Set a setting",
	})

	commands := ExtractCommands(root)

	if len(commands) != 4 {
		t.Errorf("expected 4 commands, got %d", len(commands))
	}

	// Find root command
	rootCmd := FindCommand(commands, []string{"fp"})
	if rootCmd == nil {
		t.Fatal("root command not found")
	}
	if len(rootCmd.Subcommands) != 1 || rootCmd.Subcommands[0] != "config" {
		t.Errorf("expected 1 subcommand 'config', got %v", rootCmd.Subcommands)
	}

	// Find config command
	configCmd := FindCommand(commands, []string{"fp", "config"})
	if configCmd == nil {
		t.Fatal("config command not found")
	}
	if len(configCmd.Subcommands) != 2 {
		t.Errorf("expected 2 subcommands, got %d", len(configCmd.Subcommands))
	}

	// Find get command
	getCmd := FindCommand(commands, []string{"fp", "config", "get"})
	if getCmd == nil {
		t.Fatal("get command not found")
	}
	if getCmd.Summary != "Get a setting" {
		t.Errorf("expected summary 'Get a setting', got '%s'", getCmd.Summary)
	}
	if len(getCmd.Flags) != 1 {
		t.Errorf("expected 1 flag, got %d", len(getCmd.Flags))
	}
}

func TestFindCommand_NotFound(t *testing.T) {
	commands := []CommandInfo{
		{Name: "fp", Path: []string{"fp"}},
	}

	cmd := FindCommand(commands, []string{"fp", "nonexistent"})
	if cmd != nil {
		t.Error("expected nil for non-existent command")
	}
}
