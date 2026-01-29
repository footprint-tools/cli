package completions

import (
	"fmt"
	"os"

	"github.com/footprint-tools/cli/internal/completions"
	"github.com/footprint-tools/cli/internal/dispatchers"
)

type Deps struct {
	Printf  func(string, ...any) (int, error)
	Println func(...any) (int, error)
}

func DefaultDeps() Deps {
	return Deps{
		Printf:  fmt.Printf,
		Println: fmt.Println,
	}
}

// Completions shows instructions for installing shell completions
func Completions(args []string, flags *dispatchers.ParsedFlags) error {
	return completionsCmd(args, flags, DefaultDeps())
}

func completionsCmd(args []string, flags *dispatchers.ParsedFlags, deps Deps) error {
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

	// Show installation instructions
	printInstructions(shell, deps)
	return nil
}

func printInstructions(shell completions.Shell, deps Deps) {
	evalLine := completions.SourceInstructions(shell)
	rcFile := completions.RcFile(shell)
	autoPath := completions.AutoInstallPath(shell)

	_, _ = deps.Println("To enable completions, choose one of the following:")
	_, _ = deps.Println()

	optionNum := 1

	// Option 1: Auto-load path (if available)
	if autoPath != "" {
		_, _ = deps.Printf("%d. Write to auto-load directory:\n", optionNum)
		_, _ = deps.Printf("   fp completions --script > %s\n", autoPath)
		_, _ = deps.Println()
		optionNum++
	}

	// Option 2: Add to rc file
	_, _ = deps.Printf("%d. Add to %s:\n", optionNum, rcFile)
	_, _ = deps.Printf("   %s\n", evalLine)
	_, _ = deps.Println()

	_, _ = deps.Println("Then restart your shell or run: exec $SHELL")
}
