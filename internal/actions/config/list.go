package config

import (
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/ui/style"
)

func List(args []string, flags *dispatchers.ParsedFlags) error {
	return list(args, flags, DefaultDeps())
}

func list(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	configMap, err := deps.GetAll()
	if err != nil {
		return err
	}

	// Show all visible keys with their values (or defaults if not set)
	for _, key := range domain.VisibleConfigKeys() {
		value, exists := configMap[key.Name]
		hasValue := exists && value != ""

		// Skip HideIfEmpty keys that have no value
		if key.HideIfEmpty && !hasValue {
			continue
		}

		if hasValue {
			_, _ = deps.Printf("%s=%s\n", style.Info(key.Name), value)
		} else if key.Default != "" {
			_, _ = deps.Printf("%s=%s %s\n", style.Info(key.Name), key.Default, style.Muted("(default)"))
		} else {
			_, _ = deps.Printf("%s= %s\n", style.Info(key.Name), style.Muted("(not set)"))
		}
	}

	return nil
}
