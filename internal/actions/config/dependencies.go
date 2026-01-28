package config

import (
	"fmt"

	"github.com/footprint-tools/cli/internal/config"
)

type Deps struct {
	ReadLines  func() ([]string, error)
	WriteLines func([]string) error
	Parse      func([]string) (map[string]string, error)
	Set        func([]string, string, string) ([]string, bool)
	Unset      func([]string, string) ([]string, bool)
	Get        func(string) (string, bool)
	GetAll     func() (map[string]string, error)
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
		Get:        config.Get,
		GetAll:     config.GetAll,
		Printf:     fmt.Printf,
		Println:    fmt.Println,
	}
}
