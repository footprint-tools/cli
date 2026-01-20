package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	helpactions "github.com/Skryensya/footprint/internal/actions/help"
	"github.com/Skryensya/footprint/internal/cli"
	"github.com/Skryensya/footprint/internal/config"
	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/log"
	"github.com/Skryensya/footprint/internal/paths"
	"github.com/Skryensya/footprint/internal/ui"
	"github.com/Skryensya/footprint/internal/ui/style"
	"github.com/Skryensya/footprint/internal/usage"
	"golang.org/x/term"
)

func main() {
	// Initialize logger based on config (must read config before CLI setup)
	initLogger()
	defer log.Close()

	args := os.Args[1:]

	rawFlags, commands := extractFlagsAndCommands(args)
	flags := dispatchers.NewParsedFlags(rawFlags)

	// Enable styling if stdout is a terminal and --no-color is not set
	enableColor := term.IsTerminal(int(os.Stdout.Fd())) && !flags.Has("--no-color")
	cfg, _ := config.GetAll()
	style.Init(enableColor, cfg)

	// Disable pager if --no-pager is set
	if flags.Has("--no-pager") {
		ui.DisablePager()
	}

	// Set pager override if --pager=<cmd> is set
	if pager := flags.String("--pager", ""); pager != "" {
		ui.SetPager(pager)
	}

	// Set BuildTree function for help browser (avoids import cycle)
	helpactions.SetBuildTreeFunc(cli.BuildTree)

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

// initLogger initializes the logger based on config settings.
// Reads log_enabled and log_level from config.
func initLogger() {
	// Check if logging is enabled
	enabled, _ := config.Get("log_enabled")
	if enabled == "false" {
		return
	}

	// Get log level from config
	levelStr, _ := config.Get("log_level")
	level := log.ParseLevel(levelStr)

	// Initialize logger (ignore errors - logging is optional)
	_ = log.Init(paths.LogFilePath(), level)
}

// extractFlagsAndCommands parses command-line arguments into flags and commands.
// It handles special cases:
// - Flags with values like --limit=5, --limit 5, -n 5
// - Numeric shortcuts like -5 (converted to --limit=5)
// - Boolean flags like --help, -h
func extractFlagsAndCommands(args []string) ([]string, []string) {
	flags := []string{}
	commands := []string{}

	// Flags that require a value (short form)
	valueFlagsShort := map[string]string{
		"-n": "--limit",
	}

	// Flags that require a value (long form prefix)
	valueFlagsLong := []string{"--limit", "--pager", "--status", "--source", "--since", "--until", "--repo"}

	i := 0
	for i < len(args) {
		arg := args[i]

		if len(arg) == 0 || arg[0] != '-' {
			// Not a flag, it's a command
			commands = append(commands, arg)
			i++
			continue
		}

		// Handle -<number> shorthand for --limit=<number>
		if len(arg) > 1 && arg[1] != '-' {
			numStr := arg[1:]
			if n, err := strconv.Atoi(numStr); err == nil {
				if n > 0 {
					flags = append(flags, fmt.Sprintf("--limit=%d", n))
					i++
					continue
				}
				// Invalid limit (0 or negative) - let it through for error handling
			}
		}

		// Check if it's a flag with = separator (--limit=5)
		if strings.Contains(arg, "=") {
			flags = append(flags, arg)
			i++
			continue
		}

		// Check if it's a short flag that needs a value (-n)
		if targetFlag, ok := valueFlagsShort[arg]; ok {
			if i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
				// Next arg is the value
				flags = append(flags, fmt.Sprintf("%s=%s", targetFlag, args[i+1]))
				i += 2
				continue
			}
			// No value provided - let it through for error handling
			flags = append(flags, arg)
			i++
			continue
		}

		// Check if it's a long flag that needs a value (--limit)
		isValueFlag := false
		for _, prefix := range valueFlagsLong {
			if arg == prefix {
				isValueFlag = true
				break
			}
		}

		if isValueFlag {
			if i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
				// Next arg is the value
				flags = append(flags, fmt.Sprintf("%s=%s", arg, args[i+1]))
				i += 2
				continue
			}
			// No value provided - let it through for error handling
			flags = append(flags, arg)
			i++
			continue
		}

		// Boolean flag or unknown flag
		flags = append(flags, arg)
		i++
	}

	return flags, commands
}
