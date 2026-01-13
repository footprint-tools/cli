package hooks

func Script() string {
	return `#!/bin/sh
fp repo record >/dev/null 2>&1
`
}
