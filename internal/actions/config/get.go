package config

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/usage"
)

func Get(args []string, flags *dispatchers.ParsedFlags) error {
	return get(args, flags, DefaultDeps())
}

func get(args []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	if len(args) < 1 {
		return usage.MissingArgument("key")
	}

	key := args[0]

	value, found := deps.Get(key)
	if !found {
		return usage.InvalidConfigKey(key)
	}

	_, _ = deps.Println(value)
	return nil
}
