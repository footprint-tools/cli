package tracking

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Skryensya/footprint/internal/store"
	"github.com/Skryensya/footprint/internal/ui/splitpanel"
	"github.com/charmbracelet/lipgloss"
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

	// Create layout with drawer support
	cfg := splitpanel.Config{
		SidebarWidthPercent: 0.20,
		SidebarMinWidth:     18,
		SidebarMaxWidth:     24,
		HasDrawer:           true,
		DrawerWidthPercent:  0.35,
	}
	layout := splitpanel.NewLayout(m.width, cfg, m.colors)
	layout.SetFocus(false) // Stats sidebar is never focused
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

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)

	// Title
	title := titleStyle.Render("fp watch")

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

	// Filter indicator
	filterStr := ""
	if m.filterQuery != "" {
		filterStr = mutedStyle.Render(fmt.Sprintf(" | Filter: %s", m.filterQuery))
	}

	headerContent := title + mutedStyle.Render(" | ") +
		mutedStyle.Render("Session: ") + timeStr +
		status + filterStr

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}


func (m *watchModel) buildStatsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(successColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)

	var lines []string

	// Session stats header
	lines = append(lines, headerStyle.Render("SESSION"))
	lines = append(lines, "")

	// Total events
	lines = append(lines, labelStyle.Render("Events: ")+valueStyle.Render(fmt.Sprintf("%d", m.totalEvents)))

	// Events per minute
	rate := m.eventsPerMinute()
	lines = append(lines, labelStyle.Render("Rate: ")+valueStyle.Render(fmt.Sprintf("%.1f/min", rate)))

	// Buffer info
	lines = append(lines, labelStyle.Render("Buffer: ")+valueStyle.Render(fmt.Sprintf("%d/%d", len(m.events), maxEvents)))

	lines = append(lines, "")

	// By type
	if len(m.byType) > 0 {
		lines = append(lines, headerStyle.Render("BY TYPE"))
		lines = append(lines, "")

		// Sort types for consistent display
		var types []string
		for t := range m.byType {
			types = append(types, t)
		}
		sort.Strings(types)

		for _, t := range types {
			count := m.byType[t]
			line := fmt.Sprintf("  %-10s %d", t, count)
			lines = append(lines, labelStyle.Render(line))
		}
		lines = append(lines, "")
	}

	// By repo
	if len(m.byRepo) > 0 {
		lines = append(lines, headerStyle.Render("BY REPO"))
		lines = append(lines, "")

		// Sort repos by count (descending)
		type repoCount struct {
			name  string
			count int
		}
		var repos []repoCount
		for name, count := range m.byRepo {
			repos = append(repos, repoCount{name, count})
		}
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].count > repos[j].count
		})

		// Show top repos (limit to fit)
		width := layout.SidebarContentWidth()
		maxRepos := max((height-12)/1, 3)
		maxRepos = min(maxRepos, len(repos))

		for i := 0; i < maxRepos; i++ {
			r := repos[i]
			name := r.name
			if len(name) > width-8 {
				name = name[:width-11] + "..."
			}
			line := fmt.Sprintf("  %-*s %d", width-8, name, r.count)
			lines = append(lines, labelStyle.Render(line))
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  0,
		TotalItems: len(lines),
	}
}

func (m *watchModel) buildEventsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	mutedColor := lipgloss.Color(colors.Muted)

	filtered := m.filteredEvents()
	visibleHeight := height - 2 // Account for panel border

	// Adjust scroll to keep cursor visible
	scrollOffset := m.eventScroll
	if m.cursor < scrollOffset {
		scrollOffset = m.cursor
	}
	if m.cursor >= scrollOffset+visibleHeight {
		scrollOffset = m.cursor - visibleHeight + 1
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

	// Time
	timeStr := event.Timestamp.Format("15:04:05")

	// Type (for now all are commits)
	typeStr := "commit"

	// Repo name (basename)
	repoName := filepath.Base(event.RepoPath)
	if len(repoName) > 12 {
		repoName = repoName[:9] + "..."
	}

	// Commit (short)
	commitShort := event.Commit
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	// Branch
	branch := event.Branch
	if len(branch) > 15 {
		branch = branch[:12] + "..."
	}

	// Format line
	prefix := "  "
	if selected {
		prefix = "> "
	}

	line := fmt.Sprintf("%s%-8s %-6s %-12s %-7s %s",
		prefix, timeStr, typeStr, repoName, commitShort, branch)

	// Truncate if needed
	if len(line) > width {
		line = line[:width-3] + "..."
	}

	// Style
	if selected {
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(infoColor)
		return style.Render(line)
	}

	style := lipgloss.NewStyle().Foreground(mutedColor)
	return style.Render(line)
}

func (m *watchModel) buildDrawerPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
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
		lines = append(lines, labelStyle.Render("No event selected"))
	} else {
		event := m.drawerDetail.Event
		meta := m.drawerDetail.Meta

		// Header
		lines = append(lines, headerStyle.Render("EVENT DETAIL"))
		lines = append(lines, "")

		// Type and commit
		lines = append(lines, labelStyle.Render("Type: ")+valueStyle.Render("commit"))
		commitShort := event.Commit
		if len(commitShort) > 10 {
			commitShort = commitShort[:10]
		}
		lines = append(lines, labelStyle.Render("Commit: ")+valueStyle.Render(commitShort))
		lines = append(lines, "")

		// Repository
		lines = append(lines, labelStyle.Render("Repository:"))
		repoName := filepath.Base(event.RepoPath)
		lines = append(lines, valueStyle.Render("  "+repoName))
		lines = append(lines, "")

		// Branch
		lines = append(lines, labelStyle.Render("Branch: ")+valueStyle.Render(event.Branch))
		lines = append(lines, "")

		// Timestamps
		lines = append(lines, headerStyle.Render("TIMESTAMPS"))
		lines = append(lines, "")
		if meta.AuthoredAt != "" {
			lines = append(lines, labelStyle.Render("  Authored: ")+valueStyle.Render(meta.AuthoredAt))
		}
		lines = append(lines, labelStyle.Render("  Recorded: ")+valueStyle.Render(event.Timestamp.Format("2006-01-02 15:04:05")))
		lines = append(lines, "")

		// Author
		if meta.AuthorName != "" {
			lines = append(lines, headerStyle.Render("AUTHOR"))
			lines = append(lines, "")
			lines = append(lines, valueStyle.Render("  "+meta.AuthorName))
			if meta.AuthorEmail != "" {
				lines = append(lines, labelStyle.Render("  "+meta.AuthorEmail))
			}
			lines = append(lines, "")
		}

		// Committer (only show if different from author)
		if meta.CommitterName != "" && (meta.CommitterName != meta.AuthorName || meta.CommitterEmail != meta.AuthorEmail) {
			lines = append(lines, headerStyle.Render("COMMITTER"))
			lines = append(lines, "")
			lines = append(lines, valueStyle.Render("  "+meta.CommitterName))
			if meta.CommitterEmail != "" {
				lines = append(lines, labelStyle.Render("  "+meta.CommitterEmail))
			}
			lines = append(lines, "")
		}

		// Message
		if meta.Subject != "" {
			lines = append(lines, headerStyle.Render("MESSAGE"))
			lines = append(lines, "")
			// Wrap message
			wrapped := wrapTextSimple(meta.Subject, width-4)
			for _, line := range strings.Split(wrapped, "\n") {
				lines = append(lines, valueStyle.Render("  "+line))
			}
			lines = append(lines, "")
		}

		// Stats
		if meta.FilesChanged > 0 || meta.Insertions > 0 || meta.Deletions > 0 {
			lines = append(lines, headerStyle.Render("STATS"))
			lines = append(lines, "")

			statsLine := fmt.Sprintf("  %d files", meta.FilesChanged)
			if meta.Insertions > 0 {
				statsLine += fmt.Sprintf(" +%d", meta.Insertions)
			}
			if meta.Deletions > 0 {
				statsLine += fmt.Sprintf(" -%d", meta.Deletions)
			}
			lines = append(lines, valueStyle.Render(statsLine))
			lines = append(lines, "")
		}

		// Status
		lines = append(lines, headerStyle.Render("STATUS"))
		lines = append(lines, "")
		lines = append(lines, valueStyle.Render("  "+event.Status.String()))

		// Parents (shows for merges)
		if meta.ParentCommits != "" {
			isMerge := strings.Contains(meta.ParentCommits, " ")
			if isMerge {
				lines = append(lines, "")
				lines = append(lines, headerStyle.Render("MERGE"))
				lines = append(lines, "")
				parents := strings.Split(meta.ParentCommits, " ")
				for i, parent := range parents {
					if len(parent) > 10 {
						parent = parent[:10]
					}
					lines = append(lines, labelStyle.Render(fmt.Sprintf("  Parent %d: ", i+1))+valueStyle.Render(parent))
				}
			}
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  0,
		TotalItems: len(lines),
	}
}

func (m watchModel) renderFooter() string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	borderColor := lipgloss.Color(colors.Border)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(infoColor).
		Padding(0, 1)

	sepStyle := lipgloss.NewStyle().Foreground(borderColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)

	sep := sepStyle.Render(" | ")

	var footer string
	if m.drawerOpen {
		footer = keyStyle.Render("Esc") + labelStyle.Render(" close") + sep +
			keyStyle.Render("jk") + labelStyle.Render(" navigate") + sep +
			labelStyle.Render("click outside to close")
	} else {
		footer = keyStyle.Render("q") + labelStyle.Render(" quit") + sep +
			keyStyle.Render("p") + labelStyle.Render(" pause") + sep +
			keyStyle.Render("jk") + labelStyle.Render(" nav") + sep +
			keyStyle.Render("Enter") + labelStyle.Render(" detail") + sep +
			labelStyle.Render("type to filter")
	}

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(footer)
}

func wrapTextSimple(text string, width int) string {
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
