package hooks

import (
	"os"
)

func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
