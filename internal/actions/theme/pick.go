package theme

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/ui/splitpanel"
	"github.com/Skryensya/footprint/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

//
// Public API
//

func Pick(args []string, flags *dispatchers.ParsedFlags) error {
	return pick(args, flags, DefaultDeps())
}

//
// Entrypoint
//

func pick(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	// Hard guard: Bubble Tea REQUIRES a real terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("theme picker requires an interactive terminal")
	}

	current, _ := deps.Get("color_theme")
	if current == "" {
		current = "default-dark"
	}

	cursor := 0
	for i, name := range deps.ThemeNames {
		if name == current {
			cursor = i
			break
		}
	}

	m := model{
		themes:       deps.ThemeNames,
		configs:      deps.Themes,
		cursor:       cursor,
		selected:     current,
		focusSidebar: true,
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // harmless, but stabilizes input
	)

	final, err := p.Run()
	if err != nil {
		return err
	}

	fm := final.(model)

	if fm.chosen != "" {
		if fm.chosen == current {
			deps.Printf("\nTheme %s is already active\n", style.Info(fm.chosen))
			return nil
		}
		lines, err := deps.ReadLines()
		if err != nil {
			return err
		}
		lines, _ = deps.Set(lines, "color_theme", fm.chosen)
		if err := deps.WriteLines(lines); err != nil {
			return err
		}
		deps.Printf("\nTheme set to %s\n", style.Success(fm.chosen))
		return nil
	}

	if fm.cancelled {
		deps.Println("\nCancelled")
	}

	return nil
}

//
// Model
//

type model struct {
	themes        []string
	configs       map[string]style.ColorConfig
	cursor        int
	sidebarScroll int
	previewScroll int
	width         int
	height        int
	focusSidebar  bool
	selected      string
	chosen        string
	cancelled     bool
}

//
// Bubble Tea lifecycle
//

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyTab:
			m.focusSidebar = !m.focusSidebar
			return m, nil

		case tea.KeyUp:
			if m.focusSidebar {
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(m.themes) - 1
				}
				m.previewScroll = 0
			} else {
				m.previewScroll--
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
			}

		case tea.KeyDown:
			if m.focusSidebar {
				if m.cursor < len(m.themes)-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
				m.previewScroll = 0
			} else {
				m.previewScroll++
			}

		case tea.KeyPgUp:
			if m.focusSidebar {
				m.cursor -= 5
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.previewScroll = 0
			} else {
				m.previewScroll -= 5
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
			}

		case tea.KeyPgDown:
			if m.focusSidebar {
				m.cursor += 5
				if m.cursor >= len(m.themes) {
					m.cursor = len(m.themes) - 1
				}
				m.previewScroll = 0
			} else {
				m.previewScroll += 5
			}

		case tea.KeyHome:
			if m.focusSidebar {
				m.cursor = 0
				m.previewScroll = 0
			} else {
				m.previewScroll = 0
			}

		case tea.KeyEnd:
			if m.focusSidebar {
				m.cursor = len(m.themes) - 1
				m.previewScroll = 0
			}

		case tea.KeyEnter, tea.KeySpace:
			m.chosen = m.themes[m.cursor]
			return m, tea.Quit

		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.cancelled = true
				return m, tea.Quit
			case "j":
				if m.focusSidebar {
					if m.cursor < len(m.themes)-1 {
						m.cursor++
					} else {
						m.cursor = 0
					}
					m.previewScroll = 0
				} else {
					m.previewScroll++
				}
			case "k":
				if m.focusSidebar {
					if m.cursor > 0 {
						m.cursor--
					} else {
						m.cursor = len(m.themes) - 1
					}
					m.previewScroll = 0
				} else {
					m.previewScroll--
					if m.previewScroll < 0 {
						m.previewScroll = 0
					}
				}
			case "g":
				if m.focusSidebar {
					m.cursor = 0
					m.previewScroll = 0
				} else {
					m.previewScroll = 0
				}
			case "G":
				if m.focusSidebar {
					m.cursor = len(m.themes) - 1
					m.previewScroll = 0
				}
			case "h":
				m.focusSidebar = true
			case "l":
				m.focusSidebar = false
			case "u":
				m.previewScroll -= 5
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
			case "d":
				m.previewScroll += 5
			}
		}
	}

	return m, nil
}

//
// View
//

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate dimensions
	headerHeight := 3
	footerHeight := 2
	mainHeight := m.height - headerHeight - footerHeight

	// Get current theme config for layout colors
	currentCfg := m.configs[m.themes[m.cursor]]

	// Create layout
	cfg := splitpanel.Config{
		SidebarWidthPercent: 0.25,
		SidebarMinWidth:     20,
		SidebarMaxWidth:     30,
	}
	layout := splitpanel.NewLayout(m.width, cfg, currentCfg)
	layout.SetFocus(m.focusSidebar)

	// Build panels
	sidebarPanel := m.buildSidebarPanel(layout, mainHeight)
	previewPanel := m.buildPreviewPanel(layout, mainHeight)

	// Render components
	header := m.renderHeader()
	main := layout.Render(sidebarPanel, previewPanel, mainHeight)
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func (m model) renderHeader() string {
	// Get current theme config for styling
	currentCfg := m.configs[m.themes[m.cursor]]
	infoColor := lipgloss.Color(currentCfg.Info)
	mutedColor := lipgloss.Color(currentCfg.Muted)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)

	title := titleStyle.Render("fp theme pick")
	count := mutedStyle.Render(" (" + strings.Join([]string{
		strings.Join([]string{fmt.Sprintf("%d", len(m.themes))}, ""),
		" themes)",
	}, "") + " themes)")

	headerContent := title + count

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}

func (m model) renderFooter() string {
	// Get current theme config for styling
	currentCfg := m.configs[m.themes[m.cursor]]
	infoColor := lipgloss.Color(currentCfg.Info)
	mutedColor := lipgloss.Color(currentCfg.Muted)
	borderColor := lipgloss.Color(currentCfg.UIDim)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(infoColor).
		Padding(0, 1)

	sepStyle := lipgloss.NewStyle().Foreground(borderColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)

	sep := sepStyle.Render(" | ")

	footer := keyStyle.Render("Tab") + labelStyle.Render(" switch") + sep +
		keyStyle.Render("jk") + labelStyle.Render(" nav") + sep +
		keyStyle.Render("Enter") + labelStyle.Render(" select") + sep +
		keyStyle.Render("q") + labelStyle.Render(" quit")

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(footer)
}

func (m *model) buildSidebarPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	// Get current theme config for styling
	currentCfg := m.configs[m.themes[m.cursor]]
	mutedColor := lipgloss.Color(currentCfg.Muted)
	successColor := lipgloss.Color(currentCfg.Success)
	uiActiveColor := lipgloss.Color(currentCfg.UIActive)

	visibleHeight := height - 2 // Account for panel border

	// Calculate scroll offset to keep cursor visible
	scrollOffset := m.sidebarScroll
	if m.cursor < scrollOffset {
		scrollOffset = m.cursor
	}
	if m.cursor >= scrollOffset+visibleHeight {
		scrollOffset = m.cursor - visibleHeight + 1
	}
	m.sidebarScroll = scrollOffset

	// Build visible lines
	var lines []string
	for i := scrollOffset; i < len(m.themes) && len(lines) < visibleHeight; i++ {
		themeName := m.themes[i]
		prefix := "  "
		if i == m.cursor {
			if m.focusSidebar {
				prefix = "> "
			} else {
				prefix = "* "
			}
		}

		// Add checkmark for selected theme
		if themeName == m.selected {
			prefix = "✓ "
		}

		nameStyle := lipgloss.NewStyle()

		if i == m.cursor {
			if m.focusSidebar {
				// Focused: bold with background
				nameStyle = nameStyle.
					Bold(true).
					Foreground(lipgloss.Color("0")).
					Background(uiActiveColor)
			} else {
				// Not focused: just highlighted
				nameStyle = nameStyle.
					Bold(true).
					Foreground(successColor)
			}
		} else {
			nameStyle = nameStyle.Foreground(mutedColor)
		}

		line := prefix + nameStyle.Render(themeName)
		lines = append(lines, line)
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  scrollOffset,
		TotalItems: len(m.themes),
	}
}


func (m *model) buildPreviewPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	// Get current theme
	themeName := m.themes[m.cursor]
	cfg := m.configs[themeName]

	contentWidth := layout.MainContentWidth()

	// Build preview content
	content := buildDetailedPreview(themeName, cfg, contentWidth)

	// Split into lines
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	visibleHeight := height - 2 // Account for panel border

	// Clamp scroll
	maxScroll := max(totalLines-visibleHeight, 0)
	scrollOffset := min(m.previewScroll, maxScroll)
	scrollOffset = max(scrollOffset, 0)

	// Apply scrolling
	if scrollOffset > 0 && scrollOffset < len(lines) {
		lines = lines[scrollOffset:]
	}

	// Truncate to fit height
	if len(lines) > visibleHeight {
		lines = lines[:visibleHeight]
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  scrollOffset,
		TotalItems: totalLines,
	}
}

// buildDetailedPreview creates a detailed preview with UI examples
func buildDetailedPreview(name string, cfg style.ColorConfig, width int) string {
	// Color helpers
	colorize := func(text, color string) string {
		if color == "" || color == "bold" {
			return lipgloss.NewStyle().Bold(true).Render(text)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cfg.Success))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Muted))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Info))

	var b strings.Builder

	// Title
	b.WriteString(infoStyle.Render(name))
	b.WriteString("\n\n")

	// Status colors section
	b.WriteString(headerStyle.Render("STATUS COLORS"))
	b.WriteString("\n")
	b.WriteString(colorize("success", cfg.Success) + " " +
		colorize("warning", cfg.Warning) + " " +
		colorize("error", cfg.Error) + " " +
		colorize("info", cfg.Info) + " " +
		colorize("muted", cfg.Muted))
	b.WriteString("\n\n")

	// UI Interactive colors section
	b.WriteString(headerStyle.Render("UI INTERACTIVE"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Active/Focused: ") + colorize("█████", cfg.UIActive))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Dim/Unfocused:  ") + colorize("█████", cfg.UIDim))
	b.WriteString("\n\n")

	// Event source colors
	b.WriteString(headerStyle.Render("EVENT SOURCES"))
	b.WriteString("\n")
	b.WriteString(colorize("POST-COMMIT ", cfg.Color1) + mutedStyle.Render("• commit events"))
	b.WriteString("\n")
	b.WriteString(colorize("POST-REWRITE ", cfg.Color2) + mutedStyle.Render("• rebase, amend"))
	b.WriteString("\n")
	b.WriteString(colorize("POST-CHECKOUT ", cfg.Color3) + mutedStyle.Render("• branch switches"))
	b.WriteString("\n")
	b.WriteString(colorize("POST-MERGE ", cfg.Color4) + mutedStyle.Render("• merge operations"))
	b.WriteString("\n")
	b.WriteString(colorize("PRE-PUSH ", cfg.Color5) + mutedStyle.Render("• push events"))
	b.WriteString("\n")
	b.WriteString(colorize("BACKFILL ", cfg.Color6) + mutedStyle.Render("• imported events"))
	b.WriteString("\n")
	b.WriteString(colorize("MANUAL ", cfg.Color7) + mutedStyle.Render("• manual records"))
	b.WriteString("\n\n")

	// UI Examples section
	b.WriteString(headerStyle.Render("UI EXAMPLES"))
	b.WriteString("\n\n")

	// Example: Scrollbar
	b.WriteString(mutedStyle.Render("Scrollbar (focused):"))
	b.WriteString("\n")
	scrollFocused := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.UIActive)).Render("█")
	b.WriteString("  " + scrollFocused + " " + mutedStyle.Render("← active thumb"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Scrollbar (unfocused):"))
	b.WriteString("\n")
	scrollUnfocused := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.UIDim)).Render("│")
	b.WriteString("  " + scrollUnfocused + " " + mutedStyle.Render("← dim track"))
	b.WriteString("\n\n")

	// Example: Panel borders
	b.WriteString(mutedStyle.Render("Panel borders:"))
	b.WriteString("\n")
	borderFocused := lipgloss.NewStyle().
		BorderLeft(true).
		BorderForeground(lipgloss.Color(cfg.UIActive)).
		Padding(0, 1).
		Render("Focused")
	borderUnfocused := lipgloss.NewStyle().
		BorderLeft(true).
		BorderForeground(lipgloss.Color(cfg.UIDim)).
		Padding(0, 1).
		Render("Unfocused")
	b.WriteString("  " + borderFocused + "   " + borderUnfocused)
	b.WriteString("\n\n")

	// Example: Help/Watch UI simulation
	b.WriteString(headerStyle.Render("SIMULATED UI"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Interactive help/watch mode:"))
	b.WriteString("\n\n")

	// Simulated header
	headerSim := lipgloss.NewStyle().
		BorderBottom(true).
		BorderForeground(lipgloss.Color(cfg.UIDim)).
		Padding(0, 1).
		Foreground(lipgloss.Color(cfg.Info)).
		Bold(true).
		Render("fp command")
	b.WriteString(headerSim)
	b.WriteString("\n")

	// Simulated content with borders
	leftPanel := lipgloss.NewStyle().
		Width(12).
		BorderRight(true).
		BorderForeground(lipgloss.Color(cfg.UIActive)).
		Padding(0, 1).
		Render(
			colorize("> item 1\n", cfg.Info) +
				mutedStyle.Render("  item 2\n") +
				mutedStyle.Render("  item 3"),
		)

	rightPanel := lipgloss.NewStyle().
		Width(20).
		BorderLeft(true).
		BorderForeground(lipgloss.Color(cfg.UIDim)).
		Padding(0, 1).
		Render(
			mutedStyle.Render("Details panel\n") +
				infoStyle.Render("Active content\n") +
				mutedStyle.Render("More info..."),
		)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	b.WriteString(panels)
	b.WriteString("\n")

	// Simulated footer
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color(cfg.Info)).
		Padding(0, 1)
	footerSim := lipgloss.NewStyle().
		BorderTop(true).
		BorderForeground(lipgloss.Color(cfg.UIDim)).
		Padding(0, 1).
		Render(keyStyle.Render("q") + mutedStyle.Render(" quit"))
	b.WriteString(footerSim)

	return b.String()
}

//
// Preview rendering
//

func renderPreviewCard(lines []string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

func buildThemeDetailsLines(name string, cfg style.ColorConfig) []string {
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	color := func(label, c string) string {
		if c == "" {
			return label
		}
		if c == "bold" {
			return lipgloss.NewStyle().Bold(true).Render(label)
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(c)).
			Render(label)
	}

	source := func(label, c string) string {
		s := lipgloss.NewStyle().
			Width(18).
			Align(lipgloss.Left)
		if c != "" {
			s = s.Foreground(lipgloss.Color(c))
		}
		return s.Render(label)
	}

	var lines []string

	lines = append(
		lines,
		muted.Render("Preview: ")+
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(cfg.Info)).
				Render(name),
	)

	lines = append(
		lines,
		color("success", cfg.Success)+"  "+
			color("warning", cfg.Warning)+"  "+
			color("error", cfg.Error)+"  "+
			color("info", cfg.Info)+"  "+
			color("muted", cfg.Muted)+"  "+
			color("header", cfg.Header),
	)

	lines = append(lines, "")
	lines = append(lines, muted.Render("UI:"))
	lines = append(lines, source("UI-active", cfg.UIActive))
	lines = append(lines, source("UI-dim", cfg.UIDim))

	lines = append(lines, "")
	lines = append(lines, muted.Render("Sources:"))

	lines = append(lines, source("POST-COMMIT", cfg.Color1))
	lines = append(lines, source("POST-REWRITE", cfg.Color2))
	lines = append(lines, source("POST-CHECKOUT", cfg.Color3))
	lines = append(lines, source("POST-MERGE", cfg.Color4))
	lines = append(lines, source("PRE-PUSH", cfg.Color5))
	lines = append(lines, source("BACKFILL", cfg.Color6))
	lines = append(lines, source("MANUAL", cfg.Color7))

	return lines
}
