package actions

import "fmt"

func ConfigSet(args []string, flags []string) error {
	// args ya validados por el dispatcher
	key := args[0]
	value := args[1]

	// Aquí iría la lógica real de persistencia
	fmt.Printf("set %s = %s\n", key, value)

	return nil
}
