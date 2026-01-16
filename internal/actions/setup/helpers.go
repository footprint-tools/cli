package setup

func hasFlag(flags []string, name string) bool {
	for _, f := range flags {
		if f == name {
			return true
		}
	}
	return false
}
