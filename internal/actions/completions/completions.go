package completions

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/footprint-tools/cli/internal/completions"
	"github.com/footprint-tools/cli/internal/dispatchers"
)

// Completions installs shell completions interactively
func Completions(args []string, flags *dispatchers.ParsedFlags) error {
	var shell completions.Shell

	if len(args) > 0 {
		shell = completions.Shell(args[0])
	} else {
		shell = completions.RunningShell()
		if shell == "" {
			return fmt.Errorf("could not detect shell, specify one: fp completions <bash|zsh|fish>")
		}
	}

	// Validate shell
	switch shell {
	case completions.ShellBash, completions.ShellZsh, completions.ShellFish:
		// valid
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}

	// --script flag: print script to stdout (for eval)
	if flags.Has("--script") {
		return completions.PrintCompletions(os.Stdout, shell)
	}

	result := completions.InstallForShell(shell)
	if result == nil {
		return fmt.Errorf("could not install completions")
	}

	if result.Installed {
		fmt.Printf("Completions installed to %s\n", result.Path)
		if shell == completions.ShellFish {
			fmt.Println("Completions are active immediately.")
		} else {
			fmt.Println("Restart your shell or run: exec $SHELL")
		}
		return nil
	}

	// Need manual installation - ask to add to rc file
	rcFile := completions.RcFile(shell)
	evalLine := completions.SourceInstructions(shell)

	fmt.Printf("Add to %s?\n", rcFile)
	fmt.Printf("  %s\n", evalLine)
	fmt.Print("[y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	resp, _ := reader.ReadString('\n')
	resp = strings.TrimSpace(strings.ToLower(resp))

	if resp != "y" && resp != "yes" {
		fmt.Println("\nTo install manually:")
		fmt.Printf("  echo '%s' >> %s\n", evalLine, rcFile)
		return nil
	}

	// Add to rc file
	if err := completions.AppendToRcFile(shell, evalLine); err != nil {
		return fmt.Errorf("could not write to %s: %w", rcFile, err)
	}

	fmt.Printf("\nAdded to %s\n", rcFile)
	fmt.Println("Restart your shell or run: exec $SHELL")
	return nil
}
