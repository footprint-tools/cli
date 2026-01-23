package config

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/domain"
)

func List(args []string, flags *dispatchers.ParsedFlags) error {
	return list(args, flags, DefaultDeps())
}

func list(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	configMap, err := deps.GetAll()
	if err != nil {
		return err
	}

	// Only show visible (non-hidden) keys
	for _, key := range domain.VisibleConfigKeys() {
		if value, exists := configMap[key.Name]; exists {
			_, _ = deps.Printf("%s=%s\n", key.Name, value)
		}
	}

	return nil
}
