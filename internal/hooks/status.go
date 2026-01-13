package hooks

import (
	"os"
	"path/filepath"
)

func Status(hooksPath string) map[string]bool {
	out := make(map[string]bool)

	for _, hook := range ManagedHooks {
		_, err := os.Stat(filepath.Join(hooksPath, hook))
		out[hook] = err == nil
	}

	return out
}
