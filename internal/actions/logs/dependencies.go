package logs

import (
	"fmt"
	"os"

	"github.com/footprint-tools/cli/internal/paths"
)

type Deps struct {
	LogFilePath func() string
	Printf      func(string, ...any) (int, error)
	Println     func(...any) (int, error)
	ReadFile    func(string) ([]byte, error)
	WriteFile   func(string, []byte, os.FileMode) error
	Stat        func(string) (os.FileInfo, error)
	OpenFile    func(string, int, os.FileMode) (*os.File, error)
}

func DefaultDeps() Deps {
	return Deps{
		LogFilePath: paths.LogFilePath,
		Printf:      fmt.Printf,
		Println:     fmt.Println,
		ReadFile:    os.ReadFile,
		WriteFile:   os.WriteFile,
		Stat:        os.Stat,
		OpenFile:    os.OpenFile,
	}
}
