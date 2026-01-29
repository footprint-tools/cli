package logs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/footprint-tools/cli/internal/format"
	"github.com/footprint-tools/cli/internal/ui/components"
	"github.com/footprint-tools/cli/internal/ui/splitpanel"
)

// View implements tea.Model
func (m logsModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate dimensions
	headerHeight := 3
	footerHeight := 2
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 1 {
		mainHeight = 1
	}

	// Create layout with drawer support
	cfg := splitpanel.Config{
		SidebarWidthPercent: 0.18,
		SidebarMinWidth:     16,
		SidebarMaxWidth:     22,
		HasDrawer:           true,
		DrawerWidthPercent:  0.35,
	}
	layout := splitpanel.NewLayout(m.width, cfg, m.colors)
	layout.SetFocus(false)
	layout.SetDrawerOpen(m.drawerOpen)

	// Build panels
	statsPanel := m.buildStatsPanel(layout, mainHeight)
	logsPanel := m.buildLogsPanel(layout, mainHeight)

	// Render components
	header := m.renderHeader()

	var main string
	if m.drawerOpen {
		drawerPanel := m.buildDrawerPanel(layout, mainHeight)
		main = layout.RenderWithDrawer(statsPanel, logsPanel, &drawerPanel, mainHeight)
	} else {
		main = layout.Render(statsPanel, logsPanel, mainHeight)
	}

	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func (m logsModel) renderHeader() string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	warnColor := lipgloss.Color(colors.Warning)
	successColor := lipgloss.Color(colors.Success)
	uiActiveColor := lipgloss.Color(colors.UIActive)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)
	successStyle := lipgloss.NewStyle().Foreground(successColor)
	activeStyle := lipgloss.NewStyle().Foreground(uiActiveColor)

	// Title
	title := titleStyle.Render("footprint logs")

	// Session duration
	duration := m.sessionDuration()
	hours := int(duration.Hours())
	mins := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)

	// Status indicator
	status := ""
	if m.paused {
		status = warnStyle.Render(" [PAUSED]")
	}
	if m.autoScroll {
		status += successStyle.Render(" [AUTO]")
	}

	// Filter indicators
	filterStr := ""
	if m.filterLevel != "" {
		filterStr = mutedStyle.Render(" | Level: ") + m.levelStyle(m.filterLevel).Render(m.filterLevel)
	}
	if m.filterQuery != "" {
		filterStr += mutedStyle.Render(" | Search: ") + mutedStyle.Render(m.filterQuery)
	}

	// Position indicator
	positionStr := ""
	filtered := m.filteredLines()
	if len(filtered) > 0 {
		current := m.cursor + 1
		total := len(filtered)
		positionStr = mutedStyle.Render(" | ") + activeStyle.Render(fmt.Sprintf("%d", current)) + mutedStyle.Render("/") + mutedStyle.Render(fmt.Sprintf("%d", total))
	}

	headerContent := title + mutedStyle.Render(" | ") +
		mutedStyle.Render("Session: ") + timeStr +
		status + filterStr + positionStr

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}

func (m *logsModel) buildStatsPanel(_ *splitpanel.Layout, _ int) splitpanel.Panel {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(successColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)

	var lines []string

	// Stats header
	lines = append(lines, headerStyle.Render("STATS"))
	lines = append(lines, "")

	// Total ever (all lines in file)
	lines = append(lines, labelStyle.Render("Total: ")+valueStyle.Render(fmt.Sprintf("%d", m.totalEver)))

	// Session lines (new during this session)
	lines = append(lines, labelStyle.Render("Session: ")+valueStyle.Render(fmt.Sprintf("%d", m.sessionLines)))

	// Filtered count (if filtering)
	filtered := m.filteredLines()
	if m.filterLevel != "" || m.filterQuery != "" {
		lines = append(lines, labelStyle.Render("Shown: ")+valueStyle.Render(fmt.Sprintf("%d", len(filtered))))
	}

	lines = append(lines, "")

	// By level (show totals)
	if len(m.byLevelTotal) > 0 {
		lines = append(lines, headerStyle.Render("BY LEVEL"))
		lines = append(lines, "")

		// Sort levels for consistent display
		levelOrder := []string{"ERROR", "WARN", "INFO", "DEBUG"}
		for _, level := range levelOrder {
			total := m.byLevelTotal[level]
			session := m.byLevel[level]
			if total > 0 {
				style := m.levelStyle(level)
				indicator := "  "
				if m.filterLevel == level {
					indicator = "> "
				}
				// Show total and session count if different
				countStr := fmt.Sprintf("%d", total)
				if session > 0 {
					countStr = fmt.Sprintf("%d (+%d)", total, session)
				}
				lines = append(lines, indicator+style.Render(fmt.Sprintf("%-6s", level))+labelStyle.Render(" "+countStr))
			}
		}
	}

	lines = append(lines, "")

	// Key hints - color coded
	lines = append(lines, headerStyle.Render("FILTERS"))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("e ")+m.levelStyle("ERROR").Render("ERROR"))
	lines = append(lines, labelStyle.Render("w ")+m.levelStyle("WARN").Render("WARN"))
	lines = append(lines, labelStyle.Render("i ")+m.levelStyle("INFO").Render("INFO"))
	lines = append(lines, labelStyle.Render("d ")+m.levelStyle("DEBUG").Render("DEBUG"))
	lines = append(lines, labelStyle.Render("c clear"))

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  0,
		TotalItems: len(lines),
	}
}

func (m *logsModel) buildLogsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	mutedColor := lipgloss.Color(colors.Muted)

	filtered := m.filteredLines()
	visibleHeight := height - 2

	// Update viewport dimensions
	contentWidth := layout.MainContentWidth()
	m.logsViewport.SetSize(contentWidth, visibleHeight)

	// Adjust scroll to keep cursor visible
	scrollOffset := m.logsViewport.YOffset()
	if m.cursor < scrollOffset {
		scrollOffset = m.cursor
		m.logsViewport.SetYOffset(scrollOffset)
	}
	if m.cursor >= scrollOffset+visibleHeight {
		scrollOffset = m.cursor - visibleHeight + 1
		m.logsViewport.SetYOffset(scrollOffset)
	}

	var lines []string
	width := layout.MainContentWidth()

	if len(filtered) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
		if m.filterQuery != "" || m.filterLevel != "" {
			lines = append(lines, emptyStyle.Render("No matching log lines"))
		} else {
			lines = append(lines, emptyStyle.Render("No log lines yet..."))
		}
	} else {
		for i := scrollOffset; i < len(filtered) && len(lines) < visibleHeight; i++ {
			logLine := filtered[i]
			line := m.formatLogLine(logLine, width, i == m.cursor)
			lines = append(lines, line)
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  scrollOffset,
		TotalItems: len(filtered),
	}
}

func (m logsModel) formatLogLine(logLine LogLine, width int, selected bool) string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)

	prefix := "  "
	if selected {
		prefix = "> "
	}

	// Build plain text version for selected (inverted) display
	var plainParts []string

	// Time (formatted according to user config)
	ts := ""
	if !logLine.ParsedTime.IsZero() {
		ts = format.TimeFull(logLine.ParsedTime)
		plainParts = append(plainParts, ts)
	} else if logLine.Timestamp != "" {
		// Fallback to raw timestamp if parsing failed
		ts = logLine.Timestamp
		if len(ts) > 11 {
			ts = ts[11:] // Skip date, keep time
		}
		plainParts = append(plainParts, ts)
	}

	// Level
	if logLine.Level != "" {
		plainParts = append(plainParts, fmt.Sprintf("%-5s", logLine.Level))
	}

	// Caller (file:line)
	caller := ""
	if logLine.Caller != "" {
		caller = logLine.Caller
		if len(caller) > 18 {
			caller = caller[:15] + "..."
		}
		plainParts = append(plainParts, fmt.Sprintf("%-18s", caller))
	}

	// Message
	msg := logLine.Message
	if msg == "" && logLine.Raw != "" && logLine.Timestamp == "" {
		msg = logLine.Raw
	}

	// Calculate available width for message
	fixedWidth := 2 + 8 + 1 + 5 + 1 + 18 + 1 // prefix + time + space + level + space + caller + space
	msgWidth := width - fixedWidth
	if msgWidth < 10 {
		msgWidth = 10
	}
	if len(msg) > msgWidth {
		msg = msg[:msgWidth-3] + "..."
	}
	plainParts = append(plainParts, msg)

	plainLine := strings.Join(plainParts, " ")

	if selected {
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(infoColor)
		return style.Render(prefix + plainLine)
	}

	// Styled version
	tsStyle := lipgloss.NewStyle().Foreground(mutedColor)
	callerStyle := lipgloss.NewStyle().Foreground(successColor)
	msgStyle := lipgloss.NewStyle().Foreground(mutedColor)

	var styledParts []string
	if ts != "" {
		styledParts = append(styledParts, tsStyle.Render(ts))
	}
	if logLine.Level != "" {
		levelStyle := m.levelStyle(logLine.Level)
		styledParts = append(styledParts, levelStyle.Render(fmt.Sprintf("%-5s", logLine.Level)))
	}
	if caller != "" {
		styledParts = append(styledParts, callerStyle.Render(fmt.Sprintf("%-18s", caller)))
	}
	styledParts = append(styledParts, msgStyle.Render(msg))

	return prefix + strings.Join(styledParts, " ")
}

func (m *logsModel) buildDrawerPanel(layout *splitpanel.Layout, _ int) splitpanel.Panel {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(successColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)

	var lines []string
	width := layout.DrawerContentWidth()

	if m.drawerDetail == nil {
		lines = append(lines, labelStyle.Render("No line selected"))
	} else {
		logLine := m.drawerDetail

		// Header
		lines = append(lines, headerStyle.Render("LOG DETAIL"))
		lines = append(lines, "")

		// Level
		if logLine.Level != "" {
			levelStyle := m.levelStyle(logLine.Level)
			lines = append(lines, labelStyle.Render("Level: ")+levelStyle.Render(logLine.Level))
			lines = append(lines, "")
		}

		// Timestamp
		if !logLine.ParsedTime.IsZero() {
			lines = append(lines, labelStyle.Render("Timestamp:"))
			lines = append(lines, valueStyle.Render("  "+format.Full(logLine.ParsedTime)))
			lines = append(lines, "")
		} else if logLine.Timestamp != "" {
			// Fallback to raw timestamp if parsing failed
			lines = append(lines, labelStyle.Render("Timestamp:"))
			lines = append(lines, valueStyle.Render("  "+logLine.Timestamp))
			lines = append(lines, "")
		}

		// Caller (source file)
		if logLine.Caller != "" {
			lines = append(lines, labelStyle.Render("Source:"))
			lines = append(lines, valueStyle.Render("  "+logLine.Caller))
			lines = append(lines, "")
		}

		// Message
		if logLine.Message != "" {
			lines = append(lines, headerStyle.Render("MESSAGE"))
			lines = append(lines, "")
			// Wrap message
			wrapped := wrapText(logLine.Message, width-4)
			for _, l := range strings.Split(wrapped, "\n") {
				lines = append(lines, valueStyle.Render("  "+l))
			}
			lines = append(lines, "")
		}

		// Raw line
		lines = append(lines, headerStyle.Render("RAW"))
		lines = append(lines, "")
		wrapped := wrapText(logLine.Raw, width-4)
		for _, l := range strings.Split(wrapped, "\n") {
			lines = append(lines, labelStyle.Render("  "+l))
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  0,
		TotalItems: len(lines),
	}
}

func (m logsModel) renderFooter() string {
	help := components.NewThemedHelp()

	var bindings []key.Binding
	if m.drawerOpen {
		bindings = []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "close")),
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("jk", "navigate")),
		}
	} else {
		bindings = []key.Binding{
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "auto")),
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("jk", "nav")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "detail")),
		}
		if m.filterQuery == "" {
			bindings = append(bindings, key.NewBinding(key.WithKeys(""), key.WithHelp("type", "search")))
		} else {
			bindings = append(bindings, key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear")))
		}
	}

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(help.ShortHelpView(bindings))
}

func (m logsModel) levelStyle(level string) lipgloss.Style {
	colors := m.colors
	switch level {
	case "ERROR":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Error))
	case "WARN":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Warning))
	case "INFO":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Info))
	case "DEBUG":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Muted))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Muted))
	}
}

func wrapText(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		if i > 0 {
			if lineLen+1+len(word) > width {
				result.WriteString("\n")
				lineLen = 0
			} else {
				result.WriteString(" ")
				lineLen++
			}
		}
		result.WriteString(word)
		lineLen += len(word)
	}

	return result.String()
}
