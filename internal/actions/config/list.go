package config

import "github.com/Skryensya/footprint/internal/dispatchers"

func List(args []string, flags *dispatchers.ParsedFlags) error {
	return list(args, flags, DefaultDeps())
}

func list(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	configMap, err := deps.GetAll()
	if err != nil {
		return err
	}

	for key, value := range configMap {
		deps.Printf("%s=%s\n", key, value)
	}

	return nil
}
