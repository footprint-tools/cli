package theme

import (
	"fmt"

	"github.com/Skryensya/footprint/internal/config"
	"github.com/Skryensya/footprint/internal/ui/style"
)

type Deps struct {
	ReadLines  func() ([]string, error)
	WriteLines func([]string) error
	Set        func([]string, string, string) ([]string, bool)
	Get        func(string) (string, bool)
	GetAll     func() (map[string]string, error)
	Printf     func(string, ...any) (int, error)
	Println    func(...any) (int, error)
	ThemeNames []string
	Themes     map[string]style.ColorConfig
}

func DefaultDeps() Deps {
	return Deps{
		ReadLines:  config.ReadLines,
		WriteLines: config.WriteLines,
		Set:        config.Set,
		Get:        config.Get,
		GetAll:     config.GetAll,
		Printf:     fmt.Printf,
		Println:    fmt.Println,
		ThemeNames: style.ThemeNames, // All variants (dark/light) explicitly
		Themes:     style.Themes,
	}
}
