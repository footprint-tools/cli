package config

import (
	"fmt"

	"github.com/Skryensya/footprint/internal/config"
)

type Deps struct {
	ReadLines  func() ([]string, error)
	WriteLines func([]string) error
	Parse      func([]string) (map[string]string, error)
	Set        func([]string, string, string) ([]string, bool)
	Unset      func([]string, string) ([]string, bool)
	Printf     func(string, ...any) (int, error)
	Println    func(...any) (int, error)
}

func DefaultDeps() Deps {
	return Deps{
		ReadLines:  config.ReadLines,
		WriteLines: config.WriteLines,
		Parse:      config.Parse,
		Set:        config.Set,
		Unset:      config.Unset,
		Printf:     fmt.Printf,
		Println:    fmt.Println,
	}
}
