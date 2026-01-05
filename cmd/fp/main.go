package main

import (
	"fmt"
	"footprint/internal/cli"
	"footprint/internal/dispatchers"
	"footprint/internal/usage"
	"os"
)

func main() {
	args := os.Args[1:]

	flags := extractFlags(args)
	commands := extractCommands(args)

	root := cli.BuildTree()

	res, err := dispatchers.Dispatch(root, commands, flags)

	if err != nil {
		if ue, ok := err.(*usage.Error); ok {
			fmt.Fprintln(os.Stderr, ue.Error())
			os.Exit(ue.ExitCode)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err := res.Execute(res.Args, res.Flags); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

}

func extractFlags(args []string) []string {
	var flags []string
	for _, a := range args {
		if len(a) > 0 && a[0] == '-' {
			flags = append(flags, a)
		}
	}
	return flags
}

func extractCommands(args []string) []string {
	var cmds []string
	for _, a := range args {
		if len(a) > 0 && a[0] != '-' {
			cmds = append(cmds, a)
		}
	}
	return cmds
}
