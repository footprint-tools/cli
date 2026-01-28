package completions

import (
	"strings"
	"testing"

	"github.com/footprint-tools/cli/internal/dispatchers"
)

func buildTestTree() *dispatchers.DispatchNode {
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Test CLI",
		Flags: []dispatchers.FlagDescriptor{
			{Names: []string{"--help", "-h"}, Description: "Show help"},
			{Names: []string{"--version", "-v"}, Description: "Show version"},
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

	dispatchers.Command(dispatchers.CommandSpec{
		Name:    "setup",
		Parent:  root,
		Summary: "Start tracking",
		Flags: []dispatchers.FlagDescriptor{
			{Names: []string{"--force", "-f"}, Description: "Force installation"},
		},
	})

	return root
}

func TestGenerateBash(t *testing.T) {
	root := buildTestTree()
	commands := ExtractCommands(root)
	script := GenerateBash(commands)

	// Verify script structure
	checks := []string{
		"_fp_completions()",
		"complete -F _fp_completions fp",
		"config",
		"setup",
	}

	for _, check := range checks {
		if !strings.Contains(script, check) {
			t.Errorf("bash script should contain %q", check)
		}
	}

	// Verify it's valid bash syntax (starts correctly)
	if !strings.HasPrefix(script, "# fp bash completion script") {
		t.Error("bash script should start with comment header")
	}
}

func TestGenerateZsh(t *testing.T) {
	root := buildTestTree()
	commands := ExtractCommands(root)
	script := GenerateZsh(commands)

	// Verify script structure
	checks := []string{
		"#compdef fp",
		"_fp()",
		"_fp_commands()",
		"_describe",
		"config:Manage settings",
		"setup:Start tracking",
	}

	for _, check := range checks {
		if !strings.Contains(script, check) {
			t.Errorf("zsh script should contain %q", check)
		}
	}
}

func TestGenerateFish(t *testing.T) {
	root := buildTestTree()
	commands := ExtractCommands(root)
	script := GenerateFish(commands)

	// Verify script structure
	checks := []string{
		"complete -c fp -f",
		"__fish_use_subcommand",
		"config",
		"setup",
		"-d 'Manage settings'",
		"-d 'Start tracking'",
	}

	for _, check := range checks {
		if !strings.Contains(script, check) {
			t.Errorf("fish script should contain %q", check)
		}
	}
}

func TestGenerateBash_EmptyTree(t *testing.T) {
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Test CLI",
	})

	commands := ExtractCommands(root)
	script := GenerateBash(commands)

	// Should still generate a valid script
	if !strings.Contains(script, "_fp_completions()") {
		t.Error("bash script should contain function definition even for empty tree")
	}
}

func TestGenerateZsh_EmptyTree(t *testing.T) {
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Test CLI",
	})

	commands := ExtractCommands(root)
	script := GenerateZsh(commands)

	// Should still generate a valid script
	if !strings.Contains(script, "#compdef fp") {
		t.Error("zsh script should contain compdef header even for empty tree")
	}
}

func TestGenerateFish_EmptyTree(t *testing.T) {
	root := dispatchers.Root(dispatchers.RootSpec{
		Name:    "fp",
		Summary: "Test CLI",
	})

	commands := ExtractCommands(root)
	script := GenerateFish(commands)

	// Should still generate a valid script
	if !strings.Contains(script, "complete -c fp -f") {
		t.Error("fish script should contain basic completion setup even for empty tree")
	}
}
