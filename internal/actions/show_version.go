package actions

func ShowVersion(args []string, flags []string) error {
	return showVersion(args, flags, defaultDeps())
}

func showVersion(args []string, flags []string, deps actionDependencies) error {
	deps.Printf("fp version %v\n", deps.Version())
	return nil
}
