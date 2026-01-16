package config

func List(args []string, flags []string) error {
	return list(args, flags, DefaultDeps())
}

func list(_ []string, _ []string, deps Deps) error {
	lines, err := deps.ReadLines()
	if err != nil {
		return err
	}

	configMap, err := deps.Parse(lines)
	if err != nil {
		return err
	}

	for key, value := range configMap {
		deps.Printf("%s=%s\n", key, value)
	}

	return nil
}
