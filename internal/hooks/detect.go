package hooks

import (
	"os"
)

func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
