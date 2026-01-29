package completions

import (
	"fmt"
	"io"
)

// PrintCompletions writes the completion script for the given shell to w
func PrintCompletions(w io.Writer, shell Shell) error {
	root := GetCommandTree()
	if root == nil {
		return fmt.Errorf("command tree not registered")
	}

	commands := ExtractCommands(root)
	script := generateScript(shell, commands)
	if script == "" {
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	_, err := fmt.Fprint(w, script)
	return err
}

func generateScript(shell Shell, commands []CommandInfo) string {
	switch shell {
	case ShellBash:
		return GenerateBash(commands)
	case ShellZsh:
		return GenerateZsh(commands)
	case ShellFish:
		return GenerateFish(commands)
	default:
		return ""
	}
}
