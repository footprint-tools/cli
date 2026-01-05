package actions

import (
	"fmt"
	"os"
)

func HelloWorld(args []string, flags []string) error {
	fmt.Println("Hello world")
	os.Exit(0)
	return nil
}
