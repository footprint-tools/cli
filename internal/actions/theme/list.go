package theme

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/ui/style"
	"github.com/charmbracelet/lipgloss"
)

func List(args []string, flags *dispatchers.ParsedFlags) error {
	return list(args, flags, DefaultDeps())
}

func list(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	current, _ := deps.Get("color_theme")
	if current == "" {
		current = style.ResolveThemeName("default")
	}

	_, _ = deps.Println("Available themes (* = current)\n")

	for _, name := range deps.ThemeNames {
		marker := "  "
		if name == current {
			marker = style.Success("* ")
		}

		theme := deps.Themes[name]
		preview := renderColorPreview(theme)

		_, _ = deps.Printf("%s%-14s  %s\n", marker, name, preview)
	}

	_, _ = deps.Println("\nUse 'fp theme set <name>' or 'fp theme pick' to change")

	return nil
}

// renderColorPreview returns colored text samples for a theme.
func renderColorPreview(cfg style.ColorConfig) string {
	colorize := func(text, color string) string {
		if color == "" || color == "bold" {
			return lipgloss.NewStyle().Bold(true).Render(text)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
	}

	return colorize("success ", cfg.Success) +
		colorize("error ", cfg.Error) +
		colorize("info ", cfg.Info) +
		colorize("muted", cfg.Muted) +
		"   " +
		colorize("POST-COMMIT ", cfg.Color1) +
		colorize("POST-REWRITE ", cfg.Color2) +
		colorize("POST-CHECKOUT ", cfg.Color3) +
		colorize("POST-MERGE ", cfg.Color4) +
		colorize("PRE-PUSH ", cfg.Color5) +
		colorize("BACKFILL ", cfg.Color6) +
		colorize("MANUAL", cfg.Color7) +
		"   " +
		colorize("UI-active ", cfg.UIActive) +
		colorize("UI-dim", cfg.UIDim)
}
