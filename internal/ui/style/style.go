// Package style provides semantic terminal styling using lipgloss.
//
// This package is the only place where lipgloss is imported. All styling
// is semantic (Success, Warning, Error, etc.) rather than visual (RedBold, etc.).
//
// When disabled, all helpers return the input string unchanged with no ANSI codes.
package style

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	enabled bool

	// Pre-created styles for performance.
	// These are only used when enabled is true.
	successStyle lipgloss.Style
	warningStyle lipgloss.Style
	errorStyle   lipgloss.Style
	infoStyle    lipgloss.Style
	headerStyle  lipgloss.Style
	mutedStyle   lipgloss.Style
)

// Init initializes the style package with the given enabled state.
// It also respects NO_COLOR and FP_NO_COLOR environment variables;
// if either is set (to any non-empty value), styling is disabled
// regardless of the enabled parameter.
//
// This function should be called once from main before any output.
func Init(enable bool) {
	// Respect standard NO_COLOR convention and FP-specific override
	if os.Getenv("NO_COLOR") != "" || os.Getenv("FP_NO_COLOR") != "" {
		enabled = false
		return
	}

	enabled = enable

	if enabled {
		initStyles()
	}
}

// initStyles creates the lipgloss styles.
// Colors are intentionally subtle and readable.
// Uses ANSI 16-color palette for broad terminal compatibility.
func initStyles() {
	// Force lipgloss to use ANSI colors regardless of TTY detection.
	// This is safe because the caller (main) has already decided colors are appropriate.
	lipgloss.SetColorProfile(termenv.ANSI)

	// Green for success - not too bright
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	// Yellow for warnings
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	// Red for errors
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	// Cyan for informational messages
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	// Bold for headers - emphasis without color
	headerStyle = lipgloss.NewStyle().Bold(true)

	// Dim/gray for muted text
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
}

// Enabled returns whether styling is currently enabled.
func Enabled() bool {
	return enabled
}

// Success styles text for successful operations.
func Success(s string) string {
	if !enabled {
		return s
	}
	return successStyle.Render(s)
}

// Warning styles text for warning messages.
func Warning(s string) string {
	if !enabled {
		return s
	}
	return warningStyle.Render(s)
}

// Error styles text for error messages.
func Error(s string) string {
	if !enabled {
		return s
	}
	return errorStyle.Render(s)
}

// Info styles text for informational messages.
func Info(s string) string {
	if !enabled {
		return s
	}
	return infoStyle.Render(s)
}

// Header styles text for section headers or titles.
func Header(s string) string {
	if !enabled {
		return s
	}
	return headerStyle.Render(s)
}

// Muted styles text for less important or secondary information.
func Muted(s string) string {
	if !enabled {
		return s
	}
	return mutedStyle.Render(s)
}

// Color1 through Color6 are neutral colors for visual distinction only.
// They have no semantic meaning.

func Color1(s string) string {
	if !enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(s) // cyan
}

func Color2(s string) string {
	if !enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render(s) // magenta
}

func Color3(s string) string {
	if !enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(s) // blue
}

func Color4(s string) string {
	if !enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(s) // yellow
}

func Color5(s string) string {
	if !enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(s) // green
}

func Color6(s string) string {
	if !enabled {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(s) // red
}
