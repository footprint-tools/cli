package actions

import (
	"fmt"
	"footprint/internal/config"
)

func ShowVersion(args []string, flags []string) error {
	fmt.Printf("fp version %v\n", config.Version)
	return nil

}
