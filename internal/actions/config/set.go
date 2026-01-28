package config

import (
	"github.com/footprint-tools/cli/internal/actions/tracking"
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/usage"
)

func Set(args []string, flags *dispatchers.ParsedFlags) error {
	return set(args, flags, DefaultDeps())
}

func set(args []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	if len(args) < 2 {
		return usage.MissingArgument("key value")
	}

	key := args[0]
	value := args[1]

	// Warn if key is not a recognized config key
	if !domain.IsValidConfigKey(key) {
		_, _ = deps.Printf("warning: '%s' is not a recognized config key\n", key)
	}

	lines, err := deps.ReadLines()
	if err != nil {
		return err
	}

	lines, updated := deps.Set(lines, key, value)

	if err := deps.WriteLines(lines); err != nil {
		return err
	}

	// Special handling for export_remote: configure git remote
	if key == "export_remote" {
		if err := tracking.SetupExportRemote(value); err != nil {
			return err
		}
	}

	action := "added"
	if updated {
		action = "updated"
	}

	_, _ = deps.Printf("%s %s=%s\n", action, key, value)

	return nil
}
