package theme

import (
	"fmt"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/ui/style"
	"github.com/footprint-tools/cli/internal/usage"
)

func Set(args []string, flags *dispatchers.ParsedFlags) error {
	return setTheme(args, flags, DefaultDeps())
}

func setTheme(args []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	if len(args) < 1 {
		return usage.MissingArgument("theme")
	}

	themeName := args[0]

	// Validate theme exists
	if _, ok := deps.Themes[themeName]; !ok {
		_, _ = deps.Printf("%s unknown theme: %s\n", style.Error("error:"), themeName)
		_, _ = deps.Println("")
		_, _ = deps.Println("available themes:")
		for _, name := range deps.ThemeNames {
			_, _ = deps.Printf("  %s\n", name)
		}
		return fmt.Errorf("unknown theme: %s", themeName)
	}

	lines, err := deps.ReadLines()
	if err != nil {
		return err
	}

	lines, _ = deps.Set(lines, "theme", themeName)

	if err := deps.WriteLines(lines); err != nil {
		return err
	}

	_, _ = deps.Printf("theme set to %s\n", style.Success(themeName))

	return nil
}
