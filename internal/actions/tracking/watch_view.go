package tracking

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/footprint-tools/cli/internal/format"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/ui/components"
	"github.com/footprint-tools/cli/internal/ui/splitpanel"
)

// View implements tea.Model
func (m watchModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate dimensions
	headerHeight := 3
	footerHeight := 2
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 1 {
		mainHeight = 1 // Ensure minimum height to prevent layout issues
	}

	// Create layout with drawer support
	cfg := splitpanel.Config{
		SidebarWidthPercent: 0.20,
		SidebarMinWidth:     18,
		SidebarMaxWidth:     26,
		HasDrawer:           true,
		DrawerWidthPercent:  0.35,
	}
	layout := splitpanel.NewLayout(m.width, cfg, m.colors)
	layout.SetFocusedPanel(m.focusedPanel)
	layout.SetDrawerOpen(m.drawerOpen)

	// Build panels
	statsPanel := m.buildStatsPanel(layout, mainHeight)
	eventsPanel := m.buildEventsPanel(layout, mainHeight)

	// Render components
	header := m.renderHeader()

	var main string
	if m.drawerOpen {
		drawerPanel := m.buildDrawerPanel(layout, mainHeight)
		main = layout.RenderWithDrawer(statsPanel, eventsPanel, &drawerPanel, mainHeight)
	} else {
		main = layout.Render(statsPanel, eventsPanel, mainHeight)
	}

	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func (m watchModel) renderHeader() string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	warnColor := lipgloss.Color(colors.Warning)
	uiActiveColor := lipgloss.Color(colors.UIActive)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)
	activeStyle := lipgloss.NewStyle().Foreground(uiActiveColor)

	// Title
	title := titleStyle.Render("footprint watch")

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

	// Filter indicators
	filterStr := ""
	if m.filterSource != -1 {
		filterStr = mutedStyle.Render(" | Source: ") + warnStyle.Render(sourceName(m.filterSource))
	}
	if m.filterQuery != "" {
		filterStr += mutedStyle.Render(" | Search: ") + mutedStyle.Render(m.filterQuery)
	}

	// Position indicator
	positionStr := ""
	filtered := m.filteredEvents()
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

func (m *watchModel) buildStatsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	headerColor := lipgloss.Color(colors.Header)
	warnColor := lipgloss.Color(colors.Warning)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)
	keyStyle := lipgloss.NewStyle().Foreground(warnColor).Bold(true)

	var lines []string

	// Session stats header
	lines = append(lines, headerStyle.Render("SESSION"))
	lines = append(lines, "")

	// Total events
	lines = append(lines, labelStyle.Render("Events: ")+valueStyle.Render(fmt.Sprintf("%d", m.totalEvents)))

	lines = append(lines, "")

	// Source type filters with individual colors
	lines = append(lines, headerStyle.Render("BY SOURCE"))
	lines = append(lines, "")

	sourceFilters := []struct {
		key    string
		source store.Source
		name   string
		color  string
	}{
		{"1", store.SourcePostCommit, "POST-COMMIT", colors.Color1},
		{"2", store.SourcePostRewrite, "POST-REWRITE", colors.Color2},
		{"3", store.SourcePostCheckout, "POST-CHECKOUT", colors.Color3},
		{"4", store.SourcePostMerge, "POST-MERGE", colors.Color4},
		{"5", store.SourcePrePush, "PRE-PUSH", colors.Color5},
		{"6", store.SourceManual, "MANUAL", colors.Color7},
		{"7", store.SourceBackfill, "BACKFILL", colors.Color6},
	}

	for _, sf := range sourceFilters {
		count := m.countBySource(sf.source)
		indicator := "  "
		if m.filterSource == sf.source {
			indicator = "> "
		}
		sourceNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(sf.color)).Bold(true)
		// Key, name, and count on same line
		countDisplay := valueStyle.Render(fmt.Sprintf("%d", count))
		if count == 0 {
			countDisplay = labelStyle.Render("0")
		}
		lines = append(lines, indicator+keyStyle.Render(sf.key)+" "+sourceNameStyle.Render(sf.name)+" "+countDisplay)
	}

	// By repo - compact format
	if len(m.byRepo) > 0 {
		lines = append(lines, "") // Margin before BY REPO
		lines = append(lines, headerStyle.Render("BY REPO"))
		lines = append(lines, "")

		// Sort repos by count (descending), then by name (ascending) for stability
		type repoCount struct {
			name  string
			count int
		}
		repos := make([]repoCount, 0, len(m.byRepo))
		for name, count := range m.byRepo {
			repos = append(repos, repoCount{name, count})
		}
		sort.Slice(repos, func(i, j int) bool {
			if repos[i].count != repos[j].count {
				return repos[i].count > repos[j].count
			}
			return repos[i].name < repos[j].name
		})

		// Show top repos - name and count on same line
		width := layout.SidebarContentWidth()
		maxRepos := max((height - 12), 3) // 1 line per repo now
		maxRepos = min(maxRepos, len(repos))

		for i := 0; i < maxRepos; i++ {
			r := repos[i]
			countStr := fmt.Sprintf(" %d", r.count)
			maxNameWidth := width - len(countStr) - 2 // -2 for indent
			name := r.name
			if len(name) > maxNameWidth {
				name = name[:maxNameWidth-3] + "..."
			}
			lines = append(lines, "  "+valueStyle.Render(name)+labelStyle.Render(countStr))
		}
	}

	// Set content and dimensions on viewport
	totalLines := len(lines)
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	contentWidth := layout.SidebarContentWidth()
	m.sidebarViewport.SetSize(contentWidth, visibleHeight)
	m.sidebarViewport.SetContent(strings.Join(lines, "\n"))

	// Get visible lines from viewport
	scrollPos := m.sidebarViewport.YOffset()
	endIdx := min(scrollPos+visibleHeight, totalLines)
	startIdx := min(scrollPos, totalLines)
	visibleLines := lines
	if startIdx < totalLines {
		visibleLines = lines[startIdx:endIdx]
	}

	return splitpanel.Panel{
		Lines:      visibleLines,
		ScrollPos:  scrollPos,
		TotalItems: totalLines,
	}
}

// countBySource counts events by source type
func (m *watchModel) countBySource(source store.Source) int {
	count := 0
	for _, e := range m.events {
		if e.Source == source {
			count++
		}
	}
	return count
}

func (m *watchModel) buildEventsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	mutedColor := lipgloss.Color(colors.Muted)

	filtered := m.filteredEvents()
	visibleHeight := height - 2 // Account for panel border

	// Scroll position logic:
	// - When drawer is open: stay at top (0) to show newest, selection may be off-screen
	// - When drawer is closed: follow cursor
	scrollOffset := 0
	if !m.drawerOpen {
		scrollOffset = m.eventScroll
		if m.cursor < scrollOffset {
			scrollOffset = m.cursor
		}
		if m.cursor >= scrollOffset+visibleHeight {
			scrollOffset = m.cursor - visibleHeight + 1
		}
	}

	var lines []string
	width := layout.MainContentWidth()

	if len(filtered) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
		if m.filterQuery != "" {
			lines = append(lines, emptyStyle.Render("No matching events"))
		} else {
			lines = append(lines, emptyStyle.Render("Waiting for events..."))
		}
	} else {
		for i := scrollOffset; i < len(filtered) && len(lines) < visibleHeight; i++ {
			event := filtered[i]
			// Show selection highlight even if drawer is open
			line := m.formatEventLine(event, width, i == m.cursor)
			lines = append(lines, line)
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  scrollOffset,
		TotalItems: len(filtered),
	}
}

func (m watchModel) formatEventLine(event store.RepoEvent, width int, selected bool) string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	errorColor := lipgloss.Color(colors.Error)

	// Get cached metadata
	meta := m.getCommitMeta(event.RepoPath, event.Commit)

	// Column definitions (predefined widths)
	const (
		colTime   = 5  // "15:04"
		colSource = 13 // "POST-CHECKOUT"
		colRepo   = 12 // repo name
		colBranch = 12 // branch name
		colCommit = 7  // short hash
		colAdd    = 6  // "+1234"
		colDel    = 6  // "-1234"
		colFiles  = 4  // "12F"
	)

	// Source type (full name) and color
	source := sourceName(event.Source)
	sourceColor := m.sourceColor(event.Source)

	// Time only (e.g., "15:04")
	timeStr := format.Time(event.Timestamp)

	// Repo name (basename)
	repoName := filepath.Base(event.RepoPath)
	if len(repoName) > colRepo {
		repoName = repoName[:colRepo-3] + "..."
	}

	// Branch name
	branch := event.Branch
	if len(branch) > colBranch {
		branch = branch[:colBranch-3] + "..."
	}

	// Commit (short)
	commitShort := event.Commit
	if len(commitShort) > colCommit {
		commitShort = commitShort[:colCommit]
	}

	// Stats columns
	addStr := ""
	if meta.Insertions > 0 {
		addStr = fmt.Sprintf("+%d", meta.Insertions)
	}
	delStr := ""
	if meta.Deletions > 0 {
		delStr = fmt.Sprintf("-%d", meta.Deletions)
	}
	filesStr := ""
	if meta.FilesChanged > 0 {
		filesStr = fmt.Sprintf("%dF", meta.FilesChanged)
	}

	// Commit message - remaining space
	// Fixed: 2 (prefix) + 5 + 1 + 13 + 1 + 12 + 1 + 12 + 1 + 7 + 1 + 6 + 1 + 6 + 1 + 4 + 1 = 75
	fixedWidth := 75
	msgWidth := width - fixedWidth
	if msgWidth < 5 {
		msgWidth = 5
	}
	message := meta.Subject
	if len(message) > msgWidth {
		message = message[:msgWidth-3] + "..."
	}

	// Format line
	prefix := "  "
	if selected {
		prefix = "> "
	}

	// Style components
	timeStyle := lipgloss.NewStyle().Foreground(mutedColor)
	sourceStyle := lipgloss.NewStyle().Foreground(sourceColor).Bold(true)
	repoStyle := lipgloss.NewStyle().Foreground(infoColor)
	branchStyle := lipgloss.NewStyle().Foreground(mutedColor)
	commitStyle := lipgloss.NewStyle().Foreground(infoColor)
	// Additions/deletions with background color and contrasting text
	addBgStyle := lipgloss.NewStyle().Background(successColor).Foreground(lipgloss.Color("0"))
	delBgStyle := lipgloss.NewStyle().Background(errorColor).Foreground(lipgloss.Color("0"))
	filesStyle := lipgloss.NewStyle().Foreground(infoColor)
	msgStyle := lipgloss.NewStyle().Foreground(mutedColor)

	// Format stats - right aligned in column
	addRendered := "      " // 6 spaces
	if addStr != "" {
		padding := 6 - len(addStr)
		if selected {
			addRendered = strings.Repeat(" ", padding) + addStr
		} else {
			addRendered = strings.Repeat(" ", padding) + addBgStyle.Render(addStr)
		}
	}
	delRendered := "      " // 6 spaces
	if delStr != "" {
		padding := 6 - len(delStr)
		if selected {
			delRendered = strings.Repeat(" ", padding) + delStr
		} else {
			delRendered = strings.Repeat(" ", padding) + delBgStyle.Render(delStr)
		}
	}

	if selected {
		// Selected row - use inverted colors with source color as background
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(sourceColor)

		// Same column structure as non-selected
		line := prefix +
			fmt.Sprintf("%-5s", timeStr) + " " +
			fmt.Sprintf("%-13s", source) + " " +
			fmt.Sprintf("%-12s", repoName) + " " +
			fmt.Sprintf("%-12s", branch) + " " +
			fmt.Sprintf("%-7s", commitShort) + " " +
			addRendered + " " +
			delRendered + " " +
			fmt.Sprintf("%4s", filesStr) + " " +
			message

		return style.Render(line)
	}

	// Normal row - columns: time | source | repo | branch | commit | +add | -del | files | message
	line := prefix +
		timeStyle.Render(fmt.Sprintf("%-5s", timeStr)) + " " +
		sourceStyle.Render(fmt.Sprintf("%-13s", source)) + " " +
		repoStyle.Render(fmt.Sprintf("%-12s", repoName)) + " " +
		branchStyle.Render(fmt.Sprintf("%-12s", branch)) + " " +
		commitStyle.Render(fmt.Sprintf("%-7s", commitShort)) + " " +
		addRendered + " " +
		delRendered + " " +
		filesStyle.Render(fmt.Sprintf("%4s", filesStr)) + " " +
		msgStyle.Render(message)

	return line
}

// sourceColor returns the color for the given source type
func (m watchModel) sourceColor(source store.Source) lipgloss.Color {
	colors := m.colors
	switch source {
	case store.SourcePostCommit:
		return lipgloss.Color(colors.Color1)
	case store.SourcePostRewrite:
		return lipgloss.Color(colors.Color2)
	case store.SourcePostCheckout:
		return lipgloss.Color(colors.Color3)
	case store.SourcePostMerge:
		return lipgloss.Color(colors.Color4)
	case store.SourcePrePush:
		return lipgloss.Color(colors.Color5)
	case store.SourceBackfill:
		return lipgloss.Color(colors.Color6)
	case store.SourceManual:
		return lipgloss.Color(colors.Color7)
	default:
		return lipgloss.Color(colors.Muted)
	}
}

// sourceName returns the full hook name for the event source
func sourceName(source store.Source) string {
	switch source {
	case store.SourcePostCommit:
		return "POST-COMMIT"
	case store.SourcePostRewrite:
		return "POST-REWRITE"
	case store.SourcePostCheckout:
		return "POST-CHECKOUT"
	case store.SourcePostMerge:
		return "POST-MERGE"
	case store.SourcePrePush:
		return "PRE-PUSH"
	case store.SourceManual:
		return "MANUAL"
	case store.SourceBackfill:
		return "BACKFILL"
	default:
		return "UNKNOWN"
	}
}

func (m *watchModel) buildDrawerPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	errorColor := lipgloss.Color(colors.Error)
	headerColor := lipgloss.Color(colors.Header)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)
	clickableStyle := lipgloss.NewStyle().Foreground(infoColor).Underline(true)
	addStyle := lipgloss.NewStyle().Foreground(successColor)
	delStyle := lipgloss.NewStyle().Foreground(errorColor)

	var lines []string
	width := layout.DrawerContentWidth()

	addClickable := func(label, value string) {
		labelWidth := len(label)
		valueWidth := width - labelWidth - 1 // -1 for safety margin
		if valueWidth < 10 {
			valueWidth = 10
		}

		if len(value) <= valueWidth {
			// Value fits on one line
			lines = append(lines, labelStyle.Render(label)+clickableStyle.Render(value))
		} else {
			// Value needs multiple lines
			lines = append(lines, labelStyle.Render(label)+clickableStyle.Render(value[:valueWidth]))

			// Continuation lines: indented to align with value
			indent := strings.Repeat(" ", labelWidth)
			remaining := value[valueWidth:]
			for len(remaining) > 0 {
				chunkSize := valueWidth
				if chunkSize > len(remaining) {
					chunkSize = len(remaining)
				}
				lines = append(lines, indent+clickableStyle.Render(remaining[:chunkSize]))
				remaining = remaining[chunkSize:]
			}
		}
	}

	if m.drawerDetail == nil {
		lines = append(lines, labelStyle.Render("No event selected"))
	} else {
		event := m.drawerDetail.Event
		meta := m.drawerDetail.Meta

		// === IMPORTANT INFO AT TOP ===

		// Commit message (most important)
		if meta.Subject != "" {
			lines = append(lines, headerStyle.Render("MESSAGE"))
			lines = append(lines, "")
			wrapped := wrapTextSimple(meta.Subject, width-2)
			for _, line := range strings.Split(wrapped, "\n") {
				lines = append(lines, valueStyle.Render(line))
			}
			lines = append(lines, "")
		}

		// Stats (quick overview)
		if meta.FilesChanged > 0 || meta.Insertions > 0 || meta.Deletions > 0 {
			statsLine := ""
			if meta.FilesChanged > 0 {
				statsLine += fmt.Sprintf("%d files", meta.FilesChanged)
			}
			if meta.Insertions > 0 {
				if statsLine != "" {
					statsLine += "  "
				}
				statsLine += addStyle.Render(fmt.Sprintf("+%d", meta.Insertions))
			}
			if meta.Deletions > 0 {
				if statsLine != "" {
					statsLine += "  "
				}
				statsLine += delStyle.Render(fmt.Sprintf("-%d", meta.Deletions))
			}
			lines = append(lines, statsLine)
			lines = append(lines, "")
		}

		// Source and branch
		sourceColor := m.sourceColor(event.Source)
		sourceStyle := lipgloss.NewStyle().Foreground(sourceColor).Bold(true)
		lines = append(lines, sourceStyle.Render(sourceName(event.Source))+" "+labelStyle.Render("on")+" "+valueStyle.Render(event.Branch))
		lines = append(lines, "")

		// === CLICKABLE DETAILS ===
		lines = append(lines, headerStyle.Render("DETAILS"))
		lines = append(lines, "")

		// Commit hash (clickable)
		addClickable("Commit:  ", event.Commit)

		// Repo name (clickable)
		repoName := filepath.Base(event.RepoPath)
		addClickable("Repo:    ", repoName)

		// Branch (clickable)
		addClickable("Branch:  ", event.Branch)

		lines = append(lines, "")

		// Author (clickable)
		if meta.AuthorName != "" {
			addClickable("Author:  ", meta.AuthorName)
			if meta.AuthorEmail != "" {
				addClickable("Email:   ", meta.AuthorEmail)
			}
		}

		lines = append(lines, "")

		// === MORE INFO ===
		lines = append(lines, headerStyle.Render("MORE"))
		lines = append(lines, "")

		// Path (clickable)
		addClickable("Path:    ", event.RepoPath)

		// ID (clickable)
		addClickable("ID:      ", event.RepoID)

		// Timestamps
		if meta.AuthoredAt != "" {
			lines = append(lines, labelStyle.Render("Authored: ")+labelStyle.Render(meta.AuthoredAt))
		}
		lines = append(lines, labelStyle.Render("Recorded: ")+labelStyle.Render(format.Full(event.Timestamp)))

		// Body (if any)
		if meta.Body != "" {
			lines = append(lines, "")
			lines = append(lines, headerStyle.Render("DESCRIPTION"))
			lines = append(lines, "")
			bodyLines := strings.Split(meta.Body, "\n")
			for _, bodyLine := range bodyLines {
				if strings.TrimSpace(bodyLine) == "" {
					lines = append(lines, "")
				} else {
					wrapped := wrapTextSimple(bodyLine, width-2)
					for _, wl := range strings.Split(wrapped, "\n") {
						lines = append(lines, labelStyle.Render(wl))
					}
				}
			}
		}

		// Parents (for merges)
		if meta.ParentCommits != "" && strings.Contains(meta.ParentCommits, " ") {
			lines = append(lines, "")
			lines = append(lines, headerStyle.Render("MERGE PARENTS"))
			lines = append(lines, "")
			parents := strings.Split(meta.ParentCommits, " ")
			for i, parent := range parents {
				if len(parent) > 10 {
					parent = parent[:10]
				}
				lines = append(lines, labelStyle.Render(fmt.Sprintf("  %d: ", i+1))+valueStyle.Render(parent))
			}
		}
	}

	// Set content and dimensions on viewport
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	totalLines := len(lines)
	contentWidth := layout.DrawerContentWidth()
	m.drawerViewport.SetSize(contentWidth, visibleHeight)
	m.drawerViewport.SetContent(strings.Join(lines, "\n"))

	// Get visible lines from viewport
	scrollPos := m.drawerViewport.YOffset()
	endIdx := min(scrollPos+visibleHeight, totalLines)
	startIdx := min(scrollPos, totalLines)
	visibleLines := lines
	if startIdx < totalLines {
		visibleLines = lines[startIdx:endIdx]
	}

	return splitpanel.Panel{
		Lines:      visibleLines,
		ScrollPos:  scrollPos,
		TotalItems: totalLines,
	}
}

func (m watchModel) renderFooter() string {
	help := components.NewThemedHelp()

	var bindings []key.Binding

	tabBinding := key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "focus"))

	switch {
	case m.focusedPanel == 2 && m.drawerOpen:
		// Drawer focused
		bindings = []key.Binding{
			tabBinding,
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "close")),
			key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
			key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
			key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
			key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
		}
	case m.focusedPanel == 1:
		// Sidebar focused
		bindings = []key.Binding{
			tabBinding,
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "scroll")),
			key.NewBinding(key.WithKeys("1", "2", "3", "4", "5", "6", "7"), key.WithHelp("1-7", "filter")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear")),
		}
	default:
		// Events focused
		bindings = []key.Binding{
			tabBinding,
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause")),
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("jk", "nav")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "detail")),
			key.NewBinding(key.WithKeys("1", "2", "3", "4", "5", "6", "7"), key.WithHelp("1-7", "source")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear")),
		}
	}

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(help.ShortHelpView(bindings))
}
