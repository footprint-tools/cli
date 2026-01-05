package actions

import "fmt"

func ConfigGet(args []string, flags []string) error {
	key := ""

	if len(args) > 0 {
		key = args[0]
	}

	// LOAD config map from disk
	// IF key not present:
	//     OUTPUT "key is not set"
	//     EXIT successfully
	// ELSE:
	//     OUTPUT value

	if hasFlag(flags, "--json") {
		if key == "" {
			fmt.Println(`{"config":{}}`)
			return nil
		}

		fmt.Printf(`{"%s":"value"}`+"\n", key)
		return nil
	}

	if key == "" {
		fmt.Println("no key provided")
		return nil
	}

	fmt.Printf("%s = value\n", key)
	return nil
}
