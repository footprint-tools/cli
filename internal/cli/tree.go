package cli

import (
	"footprint/internal/actions"
	"footprint/internal/dispatchers"
)

func BuildTree() *dispatchers.DispatchNode {
	root := dispatchers.NewNode(
		"fp",
		nil,
		"Track your work across repositories",
		"fp <command> [flags]",
		[]dispatchers.FlagDescriptor{
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
		},
		nil,
		nil,
	)

	dispatchers.NewNode(
		"version",
		root,
		"Show fp version",
		"fp version",
		nil,
		nil,
		actions.ShowVersion,
	).Category = dispatchers.CategoryInfo

	dispatchers.NewNode(
		"hello-world",
		root,
		"Run a basic sanity check",
		"fp hello-world",
		nil,
		nil,
		actions.HelloWorld,
	).Category = dispatchers.CategoryInfo

	config := dispatchers.NewNode(
		"config",
		root,
		"Manage configuration",
		"fp config <command>",
		nil,
		nil,
		nil,
	)

	dispatchers.NewNode(
		"get",
		config,
		"Get a config value",
		"fp config get <key>",
		[]dispatchers.FlagDescriptor{
			{
				Names:       []string{"--json"},
				Description: "Output result as JSON",
				Scope:       dispatchers.FlagScopeLocal,
			},
		},
		[]dispatchers.ArgSpec{
			{
				Name:        "key",
				Description: "Configuration key to read",
				Required:    true,
			},
		},
		actions.ConfigGet,
	).Category = dispatchers.CategoryConfig

	dispatchers.NewNode(
		"set",
		config,
		"Set a config value",
		"fp config set <key> <value>",
		nil,
		[]dispatchers.ArgSpec{
			{
				Name:        "key",
				Description: "Configuration key to write",
				Required:    true,
			},
			{
				Name:        "value",
				Description: "Value to assign",
				Required:    true,
			},
		},
		actions.ConfigSet,
	).Category = dispatchers.CategoryConfig

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
