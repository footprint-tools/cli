package main

import (
	"fmt"
	"os"

	"github.com/Skryensya/footprint/internal/cli"
	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/ui"
	"github.com/Skryensya/footprint/internal/ui/style"
	"github.com/Skryensya/footprint/internal/usage"
	"golang.org/x/term"
)

func main() {
	args := os.Args[1:]

	flags := extractFlags(args)
	commands := extractCommands(args)

	// Enable styling if stdout is a terminal and --no-color is not set
	enableColor := term.IsTerminal(int(os.Stdout.Fd())) && !hasFlag(flags, "--no-color")
	style.Init(enableColor)

	// Disable pager if --no-pager is set
	if hasFlag(flags, "--no-pager") {
		ui.DisablePager()
	}

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

	// Exit with non-zero code if resolution requests it (e.g., fp with no args)
	if res.ExitCode != 0 {
		os.Exit(res.ExitCode)
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

func hasFlag(flags []string, name string) bool {
	for _, f := range flags {
		if f == name {
			return true
		}
	}
	return false
}
