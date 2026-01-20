package theme

import (
	"errors"
	"os"
	"strings"

	"github.com/Skryensya/footprint/internal/dispatchers"
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
		themes:   deps.ThemeNames,
		configs:  deps.Themes,
		cursor:   cursor,
		selected: current,
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
	themes    []string
	configs   map[string]style.ColorConfig
	cursor    int
	selected  string
	chosen    string
	cancelled bool
}

//
// Bubble Tea lifecycle
//

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.themes) - 1
			}

		case tea.KeyDown:
			if m.cursor < len(m.themes)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case tea.KeyHome:
			m.cursor = 0

		case tea.KeyEnd:
			m.cursor = len(m.themes) - 1

		case tea.KeyEnter, tea.KeySpace:
			m.chosen = m.themes[m.cursor]
			return m, tea.Quit

		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.cancelled = true
				return m, tea.Quit
			case "j":
				if m.cursor < len(m.themes)-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			case "k":
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(m.themes) - 1
				}
			case "g":
				m.cursor = 0
			case "G":
				m.cursor = len(m.themes) - 1
			}
		}
	}

	return m, nil
}

//
// View
//

func (m model) View() string {
	var b strings.Builder

	b.WriteString("Select a theme:\n\n")

	// Left column
	left := make([]string, len(m.themes))
	for i, name := range m.themes {
		cursor := "   "
		if i == m.cursor {
			cursor = " → "
		}

		selected := "  "
		if name == m.selected {
			selected = "✓ "
		}

		styleName := lipgloss.NewStyle().Width(14)
		if i == m.cursor {
			styleName = styleName.Bold(true).Background(lipgloss.Color("237"))
		}

		left[i] = cursor + selected + styleName.Render(name)
	}

	// Right column (preview)
	themeName := m.themes[m.cursor]
	cfg := m.configs[themeName]

	right := renderPreviewCard(
		buildThemeDetailsLines(themeName, cfg),
	)

	b.WriteString(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			strings.Join(left, "\n"),
			"    ",
			right,
		),
	)

	b.WriteString("\n\n")
	b.WriteString(renderFooter())

	return b.String()
}

func renderFooter() string {
	key := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("238")).
		Padding(0, 1)

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		Render(" │ ")

	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	return key.Render("↑↓") + label.Render(" move") + sep +
		key.Render("g/G") + label.Render(" jump") + sep +
		key.Render("enter") + label.Render(" select") + sep +
		key.Render("q") + label.Render(" cancel")
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
