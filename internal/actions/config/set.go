package config

import (
	"github.com/Skryensya/footprint/internal/usage"
)

func Set(args []string, flags []string) error {
	return set(args, flags, DefaultDeps())
}

func set(args []string, _ []string, deps Deps) error {
	if len(args) < 2 {
		return usage.MissingArgument("key value")
	}

	key := args[0]
	value := args[1]

	lines, err := deps.ReadLines()
	if err != nil {
		return err
	}

	lines, updated := deps.Set(lines, key, value)

	if err := deps.WriteLines(lines); err != nil {
		return err
	}

	action := "added"
	if updated {
		action = "updated"
	}

	deps.Printf("%s %s=%s\n", action, key, value)

	return nil
}
