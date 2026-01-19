package actions

import "github.com/Skryensya/footprint/internal/dispatchers"

func ShowVersion(args []string, flags *dispatchers.ParsedFlags) error {
	return showVersion(args, flags, defaultDeps())
}

func showVersion(_ []string, _ *dispatchers.ParsedFlags, deps actionDependencies) error {
	deps.Printf("fp version %v\n", deps.Version())
	return nil
}
