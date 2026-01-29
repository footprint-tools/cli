package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/ui/components"
	"github.com/footprint-tools/cli/internal/ui/splitpanel"
	"github.com/footprint-tools/cli/internal/ui/style"
	overlay "github.com/rmhubbert/bubbletea-overlay"
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
		editInput:    components.NewThemedInput(""),
		help:         components.NewThemedHelp(),
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
	keys          []domain.ConfigKey
	values        map[string]string
	deps          Deps
	cursor        int
	sidebarScroll int
	width         int
	height        int
	focusSidebar  bool
	editing       bool
	editInput     components.ThemedInput
	// Overlay confirmation
	showConfirm   bool
	confirmDialog components.ThemedConfirm
	cancelled     bool
	changesMade   bool
	message       string
	messageIsError bool
	help          components.ThemedHelp
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

	case components.ConfirmResult:
		m.showConfirm = false
		if msg.Confirmed {
			configKey := m.keys[m.cursor]
			if err := m.unsetValue(configKey.Name); err == nil {
				delete(m.values, configKey.Name)
				m.message = "Value cleared"
				m.messageIsError = false
				m.changesMade = true
			} else {
				m.message = "Error: " + err.Error()
				m.messageIsError = true
			}
		} else {
			m.message = "Cancelled"
			m.messageIsError = false
		}
		return m, nil

	case tea.KeyMsg:
		if m.showConfirm {
			var cmd tea.Cmd
			newConfirm, cmd := m.confirmDialog.Update(msg)
			m.confirmDialog = newConfirm.(components.ThemedConfirm)
			return m, cmd
		}
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
		configKey := m.keys[m.cursor]
		currentValue := m.values[configKey.Name]
		if currentValue == "" {
			currentValue = configKey.Default
		}
		m.editing = true
		m.editInput.SetValue(currentValue)
		m.editInput.CursorEnd()
		_ = m.editInput.Focus()
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
			// Show confirmation overlay for unset
			configKey := m.keys[m.cursor]
			if m.values[configKey.Name] != "" {
				m.confirmDialog = components.NewThemedConfirm(
					"Unset value?",
					fmt.Sprintf("Remove custom value for %s?", configKey.Name),
				)
				m.showConfirm = true
				m.message = ""
			} else {
				m.message = "No value to unset"
				m.messageIsError = false
			}
		}
	}

	return m, nil
}


func (m configModel) handleEditingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editing = false
		m.editInput.Reset()
		m.editInput.Blur()
		m.message = ""
		return m, nil

	case tea.KeyEnter:
		configKey := m.keys[m.cursor]
		value := m.editInput.Value()
		if err := m.saveValue(configKey.Name, value); err == nil {
			m.values[configKey.Name] = value
			m.message = "Saved"
			m.messageIsError = false
			m.changesMade = true
		} else {
			m.message = "Error: " + err.Error()
			m.messageIsError = true
		}
		m.editing = false
		m.editInput.Reset()
		m.editInput.Blur()
		return m, nil
	}

	// Delegate all other input to the textinput component
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
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

	baseView := lipgloss.JoinVertical(lipgloss.Left, header, main, footer)

	// Overlay confirmation dialog if active
	if m.showConfirm {
		return overlay.Composite(
			m.confirmDialog.View(),
			baseView,
			overlay.Center,
			overlay.Center,
			0, 0,
		)
	}

	return baseView
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

func (m configModel) renderFooter(_ style.ColorConfig) string {
	var bindings []key.Binding

	switch {
	case m.showConfirm:
		bindings = components.ConfirmKeyBindings()
	case m.editing:
		bindings = []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "save")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "cancel")),
		}
	default:
		bindings = []key.Binding{
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "switch")),
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("jk", "nav")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "edit")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "default")),
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unset")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		}
	}

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(m.help.ShortHelpView(bindings))
}

func (m *configModel) buildSidebarPanel(_ *splitpanel.Layout, height int) splitpanel.Panel {
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
	scrollOffset := min(m.sidebarScroll, cursorVisualPos)
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
		lines = append(lines, "  "+m.editInput.View())
	} else {
		lines = append(lines, labelStyle.Render("Value"))
		switch {
		case isDefault && key.Default != "":
			lines = append(lines, "  "+defaultStyle.Render(displayValue+" (default)"))
		case displayValue != "":
			lines = append(lines, "  "+valueStyle.Render(displayValue))
		default:
			lines = append(lines, "  "+mutedStyle.Render("(not set)"))
		}
	}

	// Default value (if different from current)
	if key.Default != "" && !m.editing {
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Default"))
		lines = append(lines, "  "+mutedStyle.Render(key.Default))
	}

	// Message (only show when not in overlay mode)
	if m.message != "" && !m.showConfirm {
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
