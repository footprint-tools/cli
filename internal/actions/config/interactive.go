package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/ui/splitpanel"
	"github.com/footprint-tools/cli/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Interactive launches the config editor TUI.
func Interactive(args []string, flags *dispatchers.ParsedFlags) error {
	return interactive(args, flags, DefaultDeps())
}

func interactive(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("config editor requires an interactive terminal")
	}

	configMap, err := deps.GetAll()
	if err != nil {
		return err
	}

	keys := domain.VisibleConfigKeys()

	m := configModel{
		keys:         keys,
		values:       configMap,
		deps:         deps,
		cursor:       0,
		focusSidebar: true,
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	final, err := p.Run()
	if err != nil {
		return err
	}

	fm := final.(configModel)

	if fm.cancelled {
		_, _ = deps.Println("\nCancelled")
	} else if fm.changesMade {
		_, _ = deps.Println("\nSettings updated")
	}

	return nil
}

type configModel struct {
	keys           []domain.ConfigKey
	values         map[string]string
	deps           Deps
	cursor         int
	sidebarScroll  int
	width          int
	height         int
	focusSidebar   bool
	editing        bool
	editValue      string
	editCursor     int
	cancelled      bool
	changesMade    bool
	message        string
	messageIsError bool
}

func (m configModel) Init() tea.Cmd {
	return nil
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingInput(msg)
		}
		return m.handleNavigationInput(msg)
	}

	return m, nil
}

func (m configModel) handleNavigationInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.cancelled = true
		return m, tea.Quit

	case tea.KeyTab:
		m.focusSidebar = !m.focusSidebar
		m.message = ""
		return m, nil

	case tea.KeyUp:
		if m.focusSidebar {
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.keys) - 1
			}
			m.message = ""
		}

	case tea.KeyDown:
		if m.focusSidebar {
			if m.cursor < len(m.keys)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
			m.message = ""
		}

	case tea.KeyEnter:
		key := m.keys[m.cursor]
		currentValue := m.values[key.Name]
		if currentValue == "" {
			currentValue = key.Default
		}
		m.editing = true
		m.editValue = currentValue
		m.editCursor = len(m.editValue)
		m.message = ""

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "q":
			m.cancelled = true
			return m, tea.Quit
		case "j":
			if m.focusSidebar {
				if m.cursor < len(m.keys)-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
				m.message = ""
			}
		case "k":
			if m.focusSidebar {
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(m.keys) - 1
				}
				m.message = ""
			}
		case "g":
			if m.focusSidebar {
				m.cursor = 0
				m.message = ""
			}
		case "G":
			if m.focusSidebar {
				m.cursor = len(m.keys) - 1
				m.message = ""
			}
		case "h":
			m.focusSidebar = true
			m.message = ""
		case "l":
			m.focusSidebar = false
			m.message = ""
		case "d", "D":
			// Reset to default
			key := m.keys[m.cursor]
			if key.Default != "" {
				if err := m.saveValue(key.Name, key.Default); err == nil {
					m.values[key.Name] = key.Default
					m.message = "Reset to default"
					m.messageIsError = false
					m.changesMade = true
				} else {
					m.message = "Error: " + err.Error()
					m.messageIsError = true
				}
			}
		case "u", "U":
			// Unset value
			key := m.keys[m.cursor]
			if err := m.unsetValue(key.Name); err == nil {
				delete(m.values, key.Name)
				m.message = "Value cleared"
				m.messageIsError = false
				m.changesMade = true
			} else {
				m.message = "Error: " + err.Error()
				m.messageIsError = true
			}
		}
	}

	return m, nil
}

func (m configModel) handleEditingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editing = false
		m.editValue = ""
		m.message = ""

	case tea.KeyEnter:
		key := m.keys[m.cursor]
		if err := m.saveValue(key.Name, m.editValue); err == nil {
			m.values[key.Name] = m.editValue
			m.message = "Saved"
			m.messageIsError = false
			m.changesMade = true
		} else {
			m.message = "Error: " + err.Error()
			m.messageIsError = true
		}
		m.editing = false
		m.editValue = ""

	case tea.KeyBackspace:
		if m.editCursor > 0 {
			m.editValue = m.editValue[:m.editCursor-1] + m.editValue[m.editCursor:]
			m.editCursor--
		}

	case tea.KeyDelete:
		if m.editCursor < len(m.editValue) {
			m.editValue = m.editValue[:m.editCursor] + m.editValue[m.editCursor+1:]
		}

	case tea.KeyLeft:
		if m.editCursor > 0 {
			m.editCursor--
		}

	case tea.KeyRight:
		if m.editCursor < len(m.editValue) {
			m.editCursor++
		}

	case tea.KeyHome, tea.KeyCtrlA:
		m.editCursor = 0

	case tea.KeyEnd, tea.KeyCtrlE:
		m.editCursor = len(m.editValue)

	case tea.KeyRunes:
		m.editValue = m.editValue[:m.editCursor] + string(msg.Runes) + m.editValue[m.editCursor:]
		m.editCursor += len(msg.Runes)
	}

	return m, nil
}

func (m *configModel) saveValue(key, value string) error {
	lines, err := m.deps.ReadLines()
	if err != nil {
		return err
	}
	lines, _ = m.deps.Set(lines, key, value)
	return m.deps.WriteLines(lines)
}

func (m *configModel) unsetValue(key string) error {
	lines, err := m.deps.ReadLines()
	if err != nil {
		return err
	}
	lines, _ = m.deps.Unset(lines, key)
	return m.deps.WriteLines(lines)
}

func (m configModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if len(m.keys) == 0 {
		return "No settings available"
	}

	// Calculate dimensions
	headerHeight := 2
	footerHeight := 2
	mainHeight := m.height - headerHeight - footerHeight

	// Get current color config
	cfg := style.GetColors()

	// Create layout
	layoutCfg := splitpanel.Config{
		SidebarWidthPercent: 0.30,
		SidebarMinWidth:     22,
		SidebarMaxWidth:     35,
	}
	layout := splitpanel.NewLayout(m.width, layoutCfg, cfg)
	layout.SetFocus(m.focusSidebar)

	// Build panels
	sidebarPanel := m.buildSidebarPanel(layout, mainHeight)
	detailPanel := m.buildDetailPanel(layout, mainHeight)

	// Render components
	header := m.renderHeader(cfg)
	main := layout.Render(sidebarPanel, detailPanel, mainHeight)
	footer := m.renderFooter(cfg)

	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func (m configModel) renderHeader(cfg style.ColorConfig) string {
	infoColor := lipgloss.Color(cfg.Info)
	mutedColor := lipgloss.Color(cfg.Muted)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)

	title := titleStyle.Render("fp config")
	subtitle := mutedStyle.Render(fmt.Sprintf(" (%d settings)", len(m.keys)))

	headerContent := title + subtitle

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}

func (m configModel) renderFooter(cfg style.ColorConfig) string {
	infoColor := lipgloss.Color(cfg.Info)
	mutedColor := lipgloss.Color(cfg.Muted)
	borderColor := lipgloss.Color(cfg.UIDim)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(infoColor).
		Padding(0, 1)

	sepStyle := lipgloss.NewStyle().Foreground(borderColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)

	sep := sepStyle.Render(" | ")

	var footer string
	if m.editing {
		footer = keyStyle.Render("Enter") + labelStyle.Render(" save") + sep +
			keyStyle.Render("Esc") + labelStyle.Render(" cancel")
	} else {
		footer = keyStyle.Render("Tab") + labelStyle.Render(" switch") + sep +
			keyStyle.Render("jk") + labelStyle.Render(" nav") + sep +
			keyStyle.Render("Enter") + labelStyle.Render(" edit") + sep +
			keyStyle.Render("d") + labelStyle.Render(" default") + sep +
			keyStyle.Render("u") + labelStyle.Render(" unset") + sep +
			keyStyle.Render("q") + labelStyle.Render(" quit")
	}

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(footer)
}

func (m *configModel) buildSidebarPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	cfg := style.GetColors()
	mutedColor := lipgloss.Color(cfg.Muted)
	successColor := lipgloss.Color(cfg.Success)
	uiActiveColor := lipgloss.Color(cfg.UIActive)
	infoColor := lipgloss.Color(cfg.Info)

	visibleHeight := height - 2 // Account for panel border

	// Build all lines with sections
	type sidebarItem struct {
		isSection bool
		keyIndex  int    // index in m.keys (only valid if !isSection)
		text      string // section name or key name
	}

	var allItems []sidebarItem
	currentSection := ""

	for i, key := range m.keys {
		// Add section header if section changed
		if key.Section != currentSection {
			currentSection = key.Section
			allItems = append(allItems, sidebarItem{isSection: true, text: key.Section})
		}
		allItems = append(allItems, sidebarItem{isSection: false, keyIndex: i, text: key.Name})
	}

	// Find visual position of cursor
	cursorVisualPos := 0
	for i, item := range allItems {
		if !item.isSection && item.keyIndex == m.cursor {
			cursorVisualPos = i
			break
		}
	}

	// Calculate scroll offset to keep cursor visible
	scrollOffset := m.sidebarScroll
	if cursorVisualPos < scrollOffset {
		scrollOffset = cursorVisualPos
	}
	if cursorVisualPos >= scrollOffset+visibleHeight {
		scrollOffset = cursorVisualPos - visibleHeight + 1
	}
	m.sidebarScroll = scrollOffset

	// Build visible lines
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)

	var lines []string
	for i := scrollOffset; i < len(allItems) && len(lines) < visibleHeight; i++ {
		item := allItems[i]

		if item.isSection {
			// Section header
			lines = append(lines, sectionStyle.Render("─ "+item.text))
			continue
		}

		// Regular key
		key := m.keys[item.keyIndex]
		prefix := "  "

		// Check if has custom value (not default)
		value := m.values[key.Name]
		hasCustomValue := value != "" && value != key.Default

		isCursor := item.keyIndex == m.cursor

		if isCursor {
			if m.focusSidebar {
				prefix = "> "
			} else {
				prefix = "* "
			}
		}

		// Add indicator for custom values (but not if it's the cursor)
		if hasCustomValue && !isCursor {
			prefix = "• "
		}

		nameStyle := lipgloss.NewStyle()

		if isCursor {
			if m.focusSidebar {
				nameStyle = nameStyle.
					Bold(true).
					Foreground(lipgloss.Color("0")).
					Background(uiActiveColor)
			} else {
				nameStyle = nameStyle.
					Bold(true).
					Foreground(successColor)
			}
		} else {
			nameStyle = nameStyle.Foreground(mutedColor)
		}

		line := prefix + nameStyle.Render(key.Name)
		lines = append(lines, line)
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  scrollOffset,
		TotalItems: len(allItems),
	}
}

func (m *configModel) buildDetailPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	cfg := style.GetColors()
	infoColor := lipgloss.Color(cfg.Info)
	mutedColor := lipgloss.Color(cfg.Muted)
	successColor := lipgloss.Color(cfg.Success)
	warningColor := lipgloss.Color(cfg.Warning)
	errorColor := lipgloss.Color(cfg.Error)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	labelStyle := lipgloss.NewStyle().Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(successColor)
	defaultStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
	editStyle := lipgloss.NewStyle().Foreground(warningColor)

	key := m.keys[m.cursor]
	value := m.values[key.Name]
	isDefault := value == "" || value == key.Default
	displayValue := value
	if displayValue == "" {
		displayValue = key.Default
	}

	contentWidth := layout.MainContentWidth()

	var lines []string

	// Setting name with section
	lines = append(lines, titleStyle.Render(key.Name))
	lines = append(lines, mutedStyle.Render(key.Section))
	lines = append(lines, "")

	// Description (word-wrapped)
	lines = append(lines, labelStyle.Render("Description"))
	descLines := wrapText(key.Description, contentWidth-2)
	for _, dl := range descLines {
		lines = append(lines, mutedStyle.Render("  "+dl))
	}
	lines = append(lines, "")

	// Current value
	if m.editing {
		lines = append(lines, labelStyle.Render("Value ")+editStyle.Render("(editing)"))
		// Show edit field with cursor
		editDisplay := m.editValue[:m.editCursor] + "█"
		if m.editCursor < len(m.editValue) {
			editDisplay += m.editValue[m.editCursor:]
		}
		lines = append(lines, "  "+editStyle.Render(editDisplay))
	} else {
		lines = append(lines, labelStyle.Render("Value"))
		if isDefault && key.Default != "" {
			lines = append(lines, "  "+defaultStyle.Render(displayValue+" (default)"))
		} else if displayValue != "" {
			lines = append(lines, "  "+valueStyle.Render(displayValue))
		} else {
			lines = append(lines, "  "+mutedStyle.Render("(not set)"))
		}
	}

	// Default value (if different from current)
	if key.Default != "" && !m.editing {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Default"))
		lines = append(lines, "  "+mutedStyle.Render(key.Default))
	}

	// Message
	if m.message != "" {
		lines = append(lines, "")
		if m.messageIsError {
			msgStyle := lipgloss.NewStyle().Foreground(errorColor)
			lines = append(lines, msgStyle.Render("✗ "+m.message))
		} else {
			msgStyle := lipgloss.NewStyle().Foreground(successColor)
			lines = append(lines, msgStyle.Render("✓ "+m.message))
		}
	}

	visibleHeight := height - 2
	// Pad to fill height
	for len(lines) < visibleHeight {
		lines = append(lines, "")
	}
	// Truncate if too many
	if len(lines) > visibleHeight {
		lines = lines[:visibleHeight]
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  0,
		TotalItems: len(lines),
	}
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return lines
}
