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

	rawFlags := extractFlags(args)
	commands := extractCommands(args)
	flags := dispatchers.NewParsedFlags(rawFlags)

	// Enable styling if stdout is a terminal and --no-color is not set
	enableColor := term.IsTerminal(int(os.Stdout.Fd())) && !flags.Has("--no-color")
	style.Init(enableColor)

	// Disable pager if --no-pager is set
	if flags.Has("--no-pager") {
		ui.DisablePager()
	}

	// Set pager override if --pager=<cmd> is set
	if pager := flags.String("--pager", ""); pager != "" {
		ui.SetPager(pager)
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
