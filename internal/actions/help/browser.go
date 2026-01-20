package help

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/help"
	"github.com/Skryensya/footprint/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

//
// Public API
//

func Browser(args []string, flags *dispatchers.ParsedFlags) error {
	return browser(args, flags, DefaultDeps())
}

//
// Entrypoint
//

func browser(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("help browser requires an interactive terminal")
	}

	root := deps.BuildTree()
	topics := deps.AllTopics()
	items := buildSidebarItems(root, topics)

	// Find first selectable item (skip category headers)
	cursor := 0
	for i, item := range items {
		if !item.IsCategory {
			cursor = i
			break
		}
	}

	m := model{
		items:  items,
		cursor: cursor,
		colors: style.GetColors(),
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}

//
// Sidebar Item
//

type sidebarItem struct {
	Name        string
	DisplayName string
	IsCategory  bool
	IsTopic     bool
	Node        *dispatchers.DispatchNode
	Topic       *help.Topic
}

//
// Model
//

type model struct {
	items          []sidebarItem
	cursor         int
	sidebarScroll  int
	contentScroll  int
	width          int
	height         int
	cancelled      bool
	colors         style.ColorConfig
}

//
// Build Sidebar
//

func buildSidebarItems(root *dispatchers.DispatchNode, topics []*help.Topic) []sidebarItem {
	var items []sidebarItem

	// Collect leaf commands
	var leaves []*dispatchers.DispatchNode
	for _, child := range root.Children {
		collectLeafCommands(child, &leaves)
	}

	// Group by category
	grouped := make(map[dispatchers.CommandCategory][]*dispatchers.DispatchNode)
	for _, cmd := range leaves {
		grouped[cmd.Category] = append(grouped[cmd.Category], cmd)
	}

	// Sort commands within each category
	for cat := range grouped {
		cmds := grouped[cat]
		sort.Slice(cmds, func(i, j int) bool {
			nameI := strings.Join(cmds[i].Path[1:], " ")
			nameJ := strings.Join(cmds[j].Path[1:], " ")
			return nameI < nameJ
		})
	}

	// Add categories in order
	for _, cat := range dispatchers.CategoryOrder() {
		cmds := grouped[cat]
		if len(cmds) == 0 {
			continue
		}

		// Category header (non-selectable)
		items = append(items, sidebarItem{
			Name:        cat.String(),
			DisplayName: strings.ToUpper(cat.String()),
			IsCategory:  true,
		})

		// Commands in this category
		for _, cmd := range cmds {
			displayName := strings.Join(cmd.Path[1:], " ")
			items = append(items, sidebarItem{
				Name:        displayName,
				DisplayName: displayName,
				IsCategory:  false,
				IsTopic:     false,
				Node:        cmd,
			})
		}
	}

	// Add conceptual guides section
	if len(topics) > 0 {
		items = append(items, sidebarItem{
			Name:        "conceptual",
			DisplayName: "CONCEPTUAL GUIDES",
			IsCategory:  true,
		})

		for _, topic := range topics {
			items = append(items, sidebarItem{
				Name:        topic.Name,
				DisplayName: topic.Name,
				IsCategory:  false,
				IsTopic:     true,
				Topic:       topic,
			})
		}
	}

	return items
}

func collectLeafCommands(node *dispatchers.DispatchNode, out *[]*dispatchers.DispatchNode) {
	if node.Action != nil {
		*out = append(*out, node)
		return
	}

	for _, child := range node.Children {
		collectLeafCommands(child, out)
	}
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

		case tea.KeyUp:
			m.moveCursor(-1)
			m.contentScroll = 0

		case tea.KeyDown:
			m.moveCursor(1)
			m.contentScroll = 0

		case tea.KeyPgUp:
			m.contentScroll -= 10
			if m.contentScroll < 0 {
				m.contentScroll = 0
			}

		case tea.KeyPgDown:
			m.contentScroll += 10

		case tea.KeyHome:
			m.jumpToFirst()
			m.contentScroll = 0

		case tea.KeyEnd:
			m.jumpToLast()
			m.contentScroll = 0

		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.cancelled = true
				return m, tea.Quit
			case "j":
				m.moveCursor(1)
				m.contentScroll = 0
			case "k":
				m.moveCursor(-1)
				m.contentScroll = 0
			case "g":
				m.jumpToFirst()
				m.contentScroll = 0
			case "G":
				m.jumpToLast()
				m.contentScroll = 0
			case "u":
				m.contentScroll -= 10
				if m.contentScroll < 0 {
					m.contentScroll = 0
				}
			case "d":
				m.contentScroll += 10
			}
		}
	}

	return m, nil
}

func (m *model) moveCursor(delta int) {
	newCursor := m.cursor + delta

	// Wrap around
	if newCursor < 0 {
		newCursor = len(m.items) - 1
	} else if newCursor >= len(m.items) {
		newCursor = 0
	}

	// Skip category headers
	for m.items[newCursor].IsCategory {
		newCursor += delta
		if newCursor < 0 {
			newCursor = len(m.items) - 1
		} else if newCursor >= len(m.items) {
			newCursor = 0
		}
	}

	m.cursor = newCursor
}

func (m *model) jumpToFirst() {
	for i, item := range m.items {
		if !item.IsCategory {
			m.cursor = i
			return
		}
	}
}

func (m *model) jumpToLast() {
	for i := len(m.items) - 1; i >= 0; i-- {
		if !m.items[i].IsCategory {
			m.cursor = i
			return
		}
	}
}

//
// View
//

func (m model) View() string {
	// Default dimensions for initial render
	width := m.width
	height := m.height
	if width == 0 {
		width = 100
	}
	if height == 0 {
		height = 30
	}

	// Reserve space for footer
	footerHeight := 2
	mainHeight := height - footerHeight

	// Calculate sidebar and content widths (sidebar ~25% but min 24, max 32)
	sidebarWidth := width / 4
	if sidebarWidth < 24 {
		sidebarWidth = 24
	}
	if sidebarWidth > 32 {
		sidebarWidth = 32
	}
	contentWidth := width - sidebarWidth - 1 // 1 for border

	// Render sidebar
	sidebar := m.renderSidebar(sidebarWidth, mainHeight)

	// Render content
	content := m.renderContent(contentWidth, mainHeight)

	// Join sidebar and content
	main := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		content,
	)

	// Footer
	footer := m.renderFooter(width)

	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
}

func (m model) renderSidebar(width, height int) string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)

	var lines []string
	visibleHeight := height - 2 // Account for padding

	// Calculate scroll offset to keep cursor visible
	scrollOffset := m.sidebarScroll
	if m.cursor < scrollOffset {
		scrollOffset = m.cursor
	}
	if m.cursor >= scrollOffset+visibleHeight {
		scrollOffset = m.cursor - visibleHeight + 1
	}

	// Build visible lines
	for i, item := range m.items {
		// Skip items before scroll offset
		if i < scrollOffset {
			continue
		}
		// Stop when we've filled visible area
		if len(lines) >= visibleHeight {
			break
		}

		var line string
		itemWidth := width - 4 // Account for padding and prefix

		if item.IsCategory {
			// Category header with subtle background
			categoryStyle := lipgloss.NewStyle().
				Foreground(mutedColor).
				Bold(true).
				Width(itemWidth).
				PaddingTop(1)

			// No padding at top for first category
			if i == 0 || (i > 0 && i == scrollOffset) {
				categoryStyle = categoryStyle.PaddingTop(0)
			}

			line = categoryStyle.Render(item.DisplayName)
		} else {
			// Regular item
			prefix := "  "
			if i == m.cursor {
				prefix = "▸ "
			}

			nameStyle := lipgloss.NewStyle().Width(itemWidth - 2)

			if i == m.cursor {
				// Selected item with theme info color
				nameStyle = nameStyle.
					Bold(true).
					Foreground(lipgloss.Color("0")).
					Background(infoColor)
				line = prefix + nameStyle.Render(item.DisplayName)
			} else if item.IsTopic {
				// Topic items slightly muted
				nameStyle = nameStyle.Foreground(mutedColor)
				line = prefix + nameStyle.Render(item.DisplayName)
			} else {
				// Regular command
				line = prefix + nameStyle.Render(item.DisplayName)
			}
		}

		lines = append(lines, line)
	}

	// Pad to fill height
	for len(lines) < visibleHeight {
		lines = append(lines, strings.Repeat(" ", width-2))
	}

	// Scroll indicator
	scrollIndicator := ""
	if len(m.items) > visibleHeight {
		if scrollOffset > 0 {
			scrollIndicator = "▲"
		}
		if scrollOffset+visibleHeight < len(m.items) {
			if scrollIndicator != "" {
				scrollIndicator += " "
			}
			scrollIndicator += "▼"
		}
	}

	sidebarContent := strings.Join(lines, "\n")
	if scrollIndicator != "" {
		indicatorStyle := lipgloss.NewStyle().Foreground(mutedColor)
		sidebarContent += "\n" + indicatorStyle.Render(scrollIndicator)
	}

	sidebarStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1)

	return sidebarStyle.Render(sidebarContent)
}

func (m model) renderContent(width, height int) string {
	item := m.items[m.cursor]

	var content string
	if item.IsTopic {
		content = m.renderTopicContent(item.Topic, width)
	} else if item.Node != nil {
		content = m.renderCommandContent(item.Node, width)
	}

	// Apply scrolling
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	if m.contentScroll > 0 && m.contentScroll < len(lines) {
		lines = lines[m.contentScroll:]
	}

	visibleHeight := height - 2

	// Truncate to fit height
	if len(lines) > visibleHeight {
		lines = lines[:visibleHeight]
	}

	// Add scroll indicator if content is longer than visible area
	scrollInfo := ""
	if totalLines > visibleHeight {
		colors := m.colors
		mutedColor := lipgloss.Color(colors.Muted)
		scrollStyle := lipgloss.NewStyle().Foreground(mutedColor)
		position := m.contentScroll + 1
		if position > totalLines-visibleHeight {
			position = totalLines - visibleHeight + 1
		}
		if position < 1 {
			position = 1
		}
		scrollInfo = scrollStyle.Render(fmt.Sprintf(" [%d/%d]", position, totalLines-visibleHeight+1))
	}

	contentStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 2)

	rendered := contentStyle.Render(strings.Join(lines, "\n"))

	// Append scroll info if present
	if scrollInfo != "" {
		_ = scrollInfo // Scroll info is visual indicator, shown in footer instead
	}

	return rendered
}

func (m model) renderCommandContent(node *dispatchers.DispatchNode, width int) string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor)

	displayName := strings.Join(node.Path[1:], " ")
	b.WriteString(titleStyle.Render(displayName))
	b.WriteString("\n")

	// Summary
	if node.Summary != "" {
		summaryStyle := lipgloss.NewStyle().Foreground(mutedColor)
		b.WriteString(summaryStyle.Render(node.Summary))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Section header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(successColor)

	// Usage
	b.WriteString(headerStyle.Render("USAGE"))
	b.WriteString("\n")

	usageStyle := lipgloss.NewStyle().Foreground(infoColor)
	b.WriteString("   ")
	b.WriteString(usageStyle.Render(node.Usage))
	b.WriteString("\n\n")

	// Description
	if node.Description != "" {
		b.WriteString(headerStyle.Render("DESCRIPTION"))
		b.WriteString("\n")
		b.WriteString(wrapText(node.Description, width-6))
		b.WriteString("\n\n")
	}

	// Flags
	if len(node.Flags) > 0 {
		b.WriteString(headerStyle.Render("FLAGS"))
		b.WriteString("\n")

		flagStyle := lipgloss.NewStyle().Foreground(infoColor)
		descStyle := lipgloss.NewStyle().Foreground(mutedColor)

		for _, f := range node.Flags {
			name := strings.Join(f.Names, ", ")
			if f.ValueHint != "" {
				name = name + " " + f.ValueHint
			}
			b.WriteString("   ")
			b.WriteString(flagStyle.Render(fmt.Sprintf("%-24s", name)))
			b.WriteString("  ")
			b.WriteString(descStyle.Render(f.Description))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Arguments
	if len(node.Args) > 0 {
		b.WriteString(headerStyle.Render("ARGUMENTS"))
		b.WriteString("\n")

		argStyle := lipgloss.NewStyle().Foreground(infoColor)
		descStyle := lipgloss.NewStyle().Foreground(mutedColor)

		for _, a := range node.Args {
			required := ""
			if a.Required {
				required = " (required)"
			}
			b.WriteString("   ")
			b.WriteString(argStyle.Render(fmt.Sprintf("%-16s", a.Name)))
			b.WriteString("  ")
			b.WriteString(descStyle.Render(a.Description + required))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderTopicContent(topic *help.Topic, _ int) string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor)

	b.WriteString(titleStyle.Render(topic.Name))
	b.WriteString("\n")

	summaryStyle := lipgloss.NewStyle().Foreground(mutedColor)
	b.WriteString(summaryStyle.Render(topic.Summary))
	b.WriteString("\n\n")

	b.WriteString(topic.Content())

	return b.String()
}

func wrapText(text string, width int) string {
	if width <= 0 {
		width = 72
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Simple word wrap
		words := strings.Fields(line)
		current := ""
		for _, word := range words {
			if current == "" {
				current = word
			} else if len(current)+1+len(word) <= width {
				current += " " + word
			} else {
				result.WriteString(current)
				result.WriteString("\n")
				current = word
			}
		}
		if current != "" {
			result.WriteString(current)
			result.WriteString("\n")
		}
	}

	return strings.TrimSuffix(result.String(), "\n")
}

func (m model) renderFooter(width int) string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(infoColor).
		Padding(0, 1)

	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("238"))

	labelStyle := lipgloss.NewStyle().
		Foreground(mutedColor)

	sep := sepStyle.Render(" │ ")

	footer := keyStyle.Render("↑↓") + labelStyle.Render(" nav") + sep +
		keyStyle.Render("u/d") + labelStyle.Render(" scroll") + sep +
		keyStyle.Render("g/G") + labelStyle.Render(" jump") + sep +
		keyStyle.Render("q") + labelStyle.Render(" quit")

	footerStyle := lipgloss.NewStyle().
		Width(width).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1)

	return footerStyle.Render(footer)
}
