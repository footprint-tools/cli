package config

import (
	"github.com/Skryensya/footprint/internal/usage"
)

func Unset(args []string, flags []string) error {
	return unset(args, flags, DefaultDeps())
}

func unset(args []string, flags []string, deps Deps) error {
	if hasFlag(flags, "--all") {
		if len(args) > 0 {
			return usage.InvalidFlag("--all does not take arguments")
		}

		if err := deps.WriteLines([]string{}); err != nil {
			return err
		}

		deps.Println("all config entries removed")
		return nil
	}

	if len(args) < 1 {
		return usage.MissingArgument("key")
	}

	key := args[0]

	lines, err := deps.ReadLines()
	if err != nil {
		return err
	}

	lines, removed := deps.Unset(lines, key)
	if !removed {
		return usage.InvalidConfigKey(key)
	}

	if err := deps.WriteLines(lines); err != nil {
		return err
	}

	deps.Printf("unset %s\n", key)
	return nil
}
