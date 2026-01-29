package tracking

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/format"
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/ui/components"
	"github.com/footprint-tools/cli/internal/ui/splitpanel"
	"github.com/footprint-tools/cli/internal/ui/style"
	"golang.org/x/term"
)


func activityInteractive(_ *dispatchers.ParsedFlags, deps Deps) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("interactive mode requires a terminal")
	}

	// Load data synchronously before starting TUI
	dbPath := deps.DBPath()
	db, err := deps.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.CloseDB(db)

	events, err := deps.ListEvents(db, store.EventFilter{})
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}

	commitMeta := make(map[string]git.CommitMetadata)
	for _, e := range events {
		if _, exists := commitMeta[e.Commit]; !exists {
			commitMeta[e.Commit] = git.GetCommitMetadata(e.RepoPath, e.Commit)
		}
	}

	m := newActivityModel(events, commitMeta)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}


// activityModel is the Bubble Tea model for interactive activity view
type activityModel struct {
	events     []store.RepoEvent
	commitMeta map[string]git.CommitMetadata

	// Stats
	bySource map[store.Source]int
	byRepo   map[string]int

	// UI dimensions
	width  int
	height int

	// State
	cursor       int
	eventScroll  int
	filterQuery  string
	filterSource store.Source // -1 means no filter

	// Focus: 0=events, 1=sidebar, 2=drawer
	focusedPanel  int
	sidebarScroll int

	// Drawer
	drawerOpen     bool
	drawerDetail   *EventDetail
	drawerViewport components.ThemedViewport

	// Styling
	colors style.ColorConfig
}

func newActivityModel(events []store.RepoEvent, commitMeta map[string]git.CommitMetadata) activityModel {
	// Calculate stats
	bySource := make(map[store.Source]int)
	byRepo := make(map[string]int)
	for _, e := range events {
		bySource[e.Source]++
		byRepo[filepath.Base(e.RepoPath)]++
	}

	return activityModel{
		events:         events,
		commitMeta:     commitMeta,
		bySource:       bySource,
		byRepo:         byRepo,
		filterSource:   -1,
		colors:         style.GetColors(),
		drawerViewport: components.NewThemedViewport(40, 20),
	}
}

func (m activityModel) Init() tea.Cmd {
	return nil
}

func (m activityModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}

	return m, nil
}

func (m activityModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyTab:
		if m.drawerOpen {
			m.focusedPanel = (m.focusedPanel + 1) % 3
		} else {
			m.focusedPanel = (m.focusedPanel + 1) % 2
		}
		return m, nil
	}

	// Handle based on focused panel
	if m.drawerOpen && m.focusedPanel == 2 {
		return m.handleDrawerKeys(msg)
	}
	if m.focusedPanel == 1 {
		return m.handleSidebarKeys(msg)
	}
	return m.handleEventsKeys(msg)
}

func (m activityModel) handleDrawerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.drawerOpen = false
		m.drawerDetail = nil
		m.drawerViewport.GotoTop()
		m.focusedPanel = 0
		return m, nil
	case tea.KeyUp:
		m.drawerViewport.LineUp(1)
		return m, nil
	case tea.KeyDown:
		m.drawerViewport.LineDown(1)
		return m, nil
	case tea.KeyPgUp:
		m.drawerViewport.LineUp(10)
		return m, nil
	case tea.KeyPgDown:
		m.drawerViewport.LineDown(10)
		return m, nil
	}

	switch msg.String() {
	case "q":
		m.drawerOpen = false
		m.drawerDetail = nil
		m.drawerViewport.GotoTop()
		m.focusedPanel = 0
		return m, nil
	case "j":
		m.drawerViewport.LineDown(1)
		return m, nil
	case "k":
		m.drawerViewport.LineUp(1)
		return m, nil
	case "g":
		m.drawerViewport.GotoTop()
		return m, nil
	case "G":
		m.drawerViewport.GotoBottom()
		return m, nil
	}
	return m, nil
}

func (m activityModel) handleSidebarKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.filterSource != -1 {
			m.filterSource = -1
			return m, nil
		}
		m.focusedPanel = 0
		return m, nil
	case tea.KeyUp:
		m.sidebarScroll = max(0, m.sidebarScroll-1)
		return m, nil
	case tea.KeyDown:
		m.sidebarScroll++
		return m, nil
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "j":
		m.sidebarScroll++
		return m, nil
	case "k":
		m.sidebarScroll = max(0, m.sidebarScroll-1)
		return m, nil
	case "c":
		m.filterSource = -1
		m.filterQuery = ""
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7":
		return m.toggleSourceFilter(msg.String())
	}
	return m, nil
}

func (m activityModel) handleEventsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.filterQuery != "" {
			m.filterQuery = ""
			return m, nil
		}
		if m.drawerOpen {
			m.drawerOpen = false
			m.drawerDetail = nil
			m.drawerViewport.GotoTop()
			return m, nil
		}
		return m, tea.Quit

	case tea.KeyEnter:
		filtered := m.filteredEvents()
		if len(filtered) > 0 && m.cursor < len(filtered) {
			m.drawerOpen = true
			m.focusedPanel = 2
			m.updateDrawerDetail()
		}
		return m, nil

	case tea.KeyUp:
		m.moveCursor(-1)
		return m, nil

	case tea.KeyDown:
		m.moveCursor(1)
		return m, nil

	case tea.KeyPgUp:
		m.moveCursor(-10)
		return m, nil

	case tea.KeyPgDown:
		m.moveCursor(10)
		return m, nil

	case tea.KeyBackspace:
		if len(m.filterQuery) > 0 {
			m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
		}
		return m, nil

	case tea.KeyRunes:
		return m.handleRunes(msg)
	}

	return m, nil
}

func (m activityModel) toggleSourceFilter(key string) (tea.Model, tea.Cmd) {
	sourceMap := map[string]store.Source{
		"1": store.SourcePostCommit,
		"2": store.SourcePostRewrite,
		"3": store.SourcePostCheckout,
		"4": store.SourcePostMerge,
		"5": store.SourcePrePush,
		"6": store.SourceManual,
		"7": store.SourceBackfill,
	}
	if source, ok := sourceMap[key]; ok {
		if m.filterSource == source {
			m.filterSource = -1
		} else {
			m.filterSource = source
		}
		m.cursor = 0
		m.eventScroll = 0
	}
	return m, nil
}

func (m activityModel) handleRunes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q":
		return m, tea.Quit
	case "j":
		m.moveCursor(1)
		return m, nil
	case "k":
		m.moveCursor(-1)
		return m, nil
	case "g":
		m.cursor = 0
		m.eventScroll = 0
		return m, nil
	case "G":
		filtered := m.filteredEvents()
		if len(filtered) > 0 {
			m.cursor = len(filtered) - 1
		}
		return m, nil
	case "c":
		m.filterQuery = ""
		m.filterSource = -1
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7":
		return m.toggleSourceFilter(key)
	default:
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.filterQuery += key
			m.cursor = 0
			m.eventScroll = 0
		}
	}

	return m, nil
}

func (m activityModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	statsWidth, eventsWidth, _ := m.calculateWidths()
	drawerStart := statsWidth + eventsWidth

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}

		switch {
		case msg.X < statsWidth:
			m.focusedPanel = 1
		case msg.X < drawerStart:
			m.focusedPanel = 0
			headerHeight := 3
			footerHeight := 2
			if msg.Y >= headerHeight && msg.Y < m.height-footerHeight {
				clickedLine := msg.Y - headerHeight
				clickedIdx := m.eventScroll + clickedLine
				filtered := m.filteredEvents()
				if clickedIdx >= 0 && clickedIdx < len(filtered) {
					m.cursor = clickedIdx
				}
			}
		case m.drawerOpen:
			m.focusedPanel = 2
		}

	case tea.MouseButtonWheelUp:
		switch {
		case msg.X < statsWidth:
			m.sidebarScroll = max(0, m.sidebarScroll-1)
		case msg.X < drawerStart:
			m.moveCursor(-1)
		case m.drawerOpen:
			m.drawerViewport.LineUp(1)
		}

	case tea.MouseButtonWheelDown:
		switch {
		case msg.X < statsWidth:
			m.sidebarScroll++
		case msg.X < drawerStart:
			m.moveCursor(1)
		case m.drawerOpen:
			m.drawerViewport.LineDown(1)
		}
	}

	return m, nil
}

func (m *activityModel) moveCursor(delta int) {
	filtered := m.filteredEvents()
	if len(filtered) == 0 {
		return
	}

	m.cursor += delta
	m.cursor = max(0, min(m.cursor, len(filtered)-1))
}

func (m *activityModel) updateDrawerDetail() {
	filtered := m.filteredEvents()
	if m.cursor < 0 || m.cursor >= len(filtered) {
		m.drawerDetail = nil
		return
	}

	event := filtered[m.cursor]
	meta := m.commitMeta[event.Commit]
	m.drawerDetail = &EventDetail{
		Event: event,
		Meta:  meta,
	}
}

func (m activityModel) filteredEvents() []store.RepoEvent {
	if m.filterQuery == "" && m.filterSource == -1 {
		return m.events
	}

	query := strings.ToLower(m.filterQuery)
	var filtered []store.RepoEvent

	for _, e := range m.events {
		if m.filterSource != -1 && e.Source != m.filterSource {
			continue
		}

		if query != "" && !m.matchesQuery(e, query) {
			continue
		}

		filtered = append(filtered, e)
	}

	return filtered
}

func (m activityModel) matchesQuery(e store.RepoEvent, query string) bool {
	repoName := strings.ToLower(filepath.Base(e.RepoPath))
	meta := m.commitMeta[e.Commit]
	return strings.Contains(repoName, query) ||
		strings.Contains(strings.ToLower(e.Branch), query) ||
		strings.Contains(strings.ToLower(e.Commit), query) ||
		strings.Contains(strings.ToLower(meta.Subject), query)
}

func (m activityModel) calculateWidths() (stats, events, drawer int) {
	if m.width == 0 {
		return 25, 75, 0
	}

	stats = max(18, min(26, int(float64(m.width)*0.20)))

	if m.drawerOpen {
		drawer = int(float64(m.width) * 0.35)
		events = m.width - stats - drawer
	} else {
		drawer = 0
		events = m.width - stats
	}
	return
}

// View renders the activity TUI
func (m activityModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	headerHeight := 3
	footerHeight := 2
	mainHeight := max(1, m.height-headerHeight-footerHeight)

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

	statsPanel := m.buildStatsPanel(layout, mainHeight)
	eventsPanel := m.buildEventsPanel(layout, mainHeight)

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

func (m activityModel) renderHeader() string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	uiActiveColor := lipgloss.Color(colors.UIActive)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
	activeStyle := lipgloss.NewStyle().Foreground(uiActiveColor)

	title := titleStyle.Render("fp activity")
	count := mutedStyle.Render(" | ") + mutedStyle.Render("Total: ") + titleStyle.Render(formatCount(len(m.events)))

	filterStr := ""
	if m.filterSource != -1 {
		filterStr = mutedStyle.Render(" | Source: ") + lipgloss.NewStyle().Foreground(m.sourceColor(m.filterSource)).Render(sourceName(m.filterSource))
	}
	if m.filterQuery != "" {
		filterStr += mutedStyle.Render(" | Search: ") + mutedStyle.Render(m.filterQuery)
	}

	filtered := m.filteredEvents()
	if len(filtered) != len(m.events) {
		filterStr += mutedStyle.Render(" | Showing: ") + titleStyle.Render(formatCount(len(filtered)))
	}

	// Add position indicator (paginator-style)
	positionStr := ""
	if len(filtered) > 0 {
		current := m.cursor + 1
		total := len(filtered)
		positionStr = mutedStyle.Render(" | ") + activeStyle.Render(fmt.Sprintf("%d", current)) + mutedStyle.Render("/") + mutedStyle.Render(fmt.Sprintf("%d", total))
	}

	headerContent := title + count + filterStr + positionStr

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}

func (m *activityModel) buildStatsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
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

	lines = append(lines, headerStyle.Render("SUMMARY"))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Events: ")+valueStyle.Render(formatCount(len(m.events))))
	lines = append(lines, "")

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
		count := m.bySource[sf.source]
		indicator := "  "
		if m.filterSource == sf.source {
			indicator = "> "
		}
		sourceNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(sf.color)).Bold(true)
		countDisplay := valueStyle.Render(formatCount(count))
		if count == 0 {
			countDisplay = labelStyle.Render("0")
		}
		lines = append(lines, indicator+keyStyle.Render(sf.key)+" "+sourceNameStyle.Render(sf.name)+" "+countDisplay)
	}

	// height is the panel height, subtract 2 for borders
	visibleHeight := max(1, height-2)
	scrollPos := m.sidebarScroll
	maxScroll := max(0, len(lines)-visibleHeight)
	scrollPos = max(0, min(scrollPos, maxScroll))

	// Save updated scroll position
	m.sidebarScroll = scrollPos

	endIdx := min(scrollPos+visibleHeight, len(lines))
	visibleLines := lines
	if scrollPos < len(lines) {
		visibleLines = lines[scrollPos:endIdx]
	}

	return splitpanel.Panel{
		Lines:      visibleLines,
		ScrollPos:  scrollPos,
		TotalItems: len(lines),
	}
}

func (m *activityModel) buildEventsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	mutedColor := lipgloss.Color(colors.Muted)
	headerColor := lipgloss.Color(colors.Header)
	borderColor := lipgloss.Color(colors.Border)

	filtered := m.filteredEvents()
	// height is the panel height
	// Subtract: 2 for panel borders, 2 for header + separator
	visibleDataRows := max(1, height-4)

	// Calculate scroll to keep cursor visible
	scrollOffset := m.eventScroll

	// Clamp scroll offset to valid range
	maxScroll := max(0, len(filtered)-visibleDataRows)
	scrollOffset = max(0, min(scrollOffset, maxScroll))

	// Ensure cursor is visible
	if m.cursor < scrollOffset {
		scrollOffset = m.cursor
	}
	if m.cursor >= scrollOffset+visibleDataRows {
		scrollOffset = m.cursor - visibleDataRows + 1
	}

	// Re-clamp after cursor adjustment
	scrollOffset = max(0, min(scrollOffset, maxScroll))

	// Save updated scroll position
	m.eventScroll = scrollOffset

	var lines []string
	width := layout.MainContentWidth()

	// Header row
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	header := "  " +
		padRight("DATE", 10) + " " +
		padRight("TIME", 5) + " " +
		padRight("SOURCE", 13) + " " +
		padRight("REPO", 12) + " " +
		padRight("BRANCH", 12) + " " +
		padRight("COMMIT", 7) + " " +
		padLeft("+", 6) + " " +
		padLeft("-", 6) + " " +
		"MESSAGE"
	lines = append(lines, headerStyle.Render(header))

	// Separator line
	sepStyle := lipgloss.NewStyle().Foreground(borderColor)
	lines = append(lines, sepStyle.Render(strings.Repeat("─", min(width, len(header)+10))))

	if len(filtered) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
		if m.filterQuery != "" || m.filterSource != -1 {
			lines = append(lines, emptyStyle.Render("No matching events"))
		} else {
			lines = append(lines, emptyStyle.Render("No events recorded"))
		}
	} else {
		endIdx := min(scrollOffset+visibleDataRows, len(filtered))
		for i := scrollOffset; i < endIdx; i++ {
			event := filtered[i]
			isAlternate := (i-scrollOffset)%2 == 1
			line := m.formatEventLine(event, width, i == m.cursor, isAlternate)
			lines = append(lines, line)
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  scrollOffset,
		TotalItems: len(filtered),
	}
}

func (m activityModel) formatEventLine(event store.RepoEvent, width int, selected bool, isAlternate bool) string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	errorColor := lipgloss.Color(colors.Error)

	meta := m.commitMeta[event.Commit]

	source := sourceName(event.Source)
	sourceColor := m.sourceColor(event.Source)

	dateStr := format.Date(event.Timestamp)
	timeStr := format.Time(event.Timestamp)

	repoName := filepath.Base(event.RepoPath)
	if len(repoName) > 12 {
		repoName = repoName[:9] + "..."
	}

	branch := event.Branch
	if len(branch) > 12 {
		branch = branch[:9] + "..."
	}

	commitShort := event.Commit
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	addStr := ""
	if meta.Insertions > 0 {
		addStr = "+" + formatCount(meta.Insertions)
	}
	delStr := ""
	if meta.Deletions > 0 {
		delStr = "-" + formatCount(meta.Deletions)
	}

	fixedWidth := 80
	msgWidth := max(5, width-fixedWidth)
	message := meta.Subject
	if len(message) > msgWidth {
		message = message[:msgWidth-3] + "..."
	}

	prefix := "  "
	if selected {
		prefix = "▸ "
	}

	// For alternating rows, slightly dim the colors
	if isAlternate && !selected {
		mutedColor = lipgloss.Color(colors.UIDim)
	}

	timeStyle := lipgloss.NewStyle().Foreground(mutedColor)
	sourceStyle := lipgloss.NewStyle().Foreground(sourceColor).Bold(true)
	repoStyle := lipgloss.NewStyle().Foreground(infoColor)
	branchStyle := lipgloss.NewStyle().Foreground(mutedColor)
	commitStyle := lipgloss.NewStyle().Foreground(infoColor)
	addBgStyle := lipgloss.NewStyle().Background(successColor).Foreground(lipgloss.Color("0"))
	delBgStyle := lipgloss.NewStyle().Background(errorColor).Foreground(lipgloss.Color("0"))
	msgStyle := lipgloss.NewStyle().Foreground(mutedColor)

	addRendered := "      "
	if addStr != "" {
		padding := 6 - len(addStr)
		if selected {
			addRendered = strings.Repeat(" ", padding) + addStr
		} else {
			addRendered = strings.Repeat(" ", padding) + addBgStyle.Render(addStr)
		}
	}
	delRendered := "      "
	if delStr != "" {
		padding := 6 - len(delStr)
		if selected {
			delRendered = strings.Repeat(" ", padding) + delStr
		} else {
			delRendered = strings.Repeat(" ", padding) + delBgStyle.Render(delStr)
		}
	}

	if selected {
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(sourceColor)

		line := prefix +
			padRight(dateStr, 10) + " " +
			padRight(timeStr, 5) + " " +
			padRight(source, 13) + " " +
			padRight(repoName, 12) + " " +
			padRight(branch, 12) + " " +
			padRight(commitShort, 7) + " " +
			addRendered + " " +
			delRendered + " " +
			message

		return style.Render(line)
	}

	line := prefix +
		timeStyle.Render(padRight(dateStr, 10)) + " " +
		timeStyle.Render(padRight(timeStr, 5)) + " " +
		sourceStyle.Render(padRight(source, 13)) + " " +
		repoStyle.Render(padRight(repoName, 12)) + " " +
		branchStyle.Render(padRight(branch, 12)) + " " +
		commitStyle.Render(padRight(commitShort, 7)) + " " +
		addRendered + " " +
		delRendered + " " +
		msgStyle.Render(message)

	return line
}

func (m *activityModel) buildDrawerPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	errorColor := lipgloss.Color(colors.Error)
	headerColor := lipgloss.Color(colors.Header)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)
	addStyle := lipgloss.NewStyle().Foreground(successColor)
	delStyle := lipgloss.NewStyle().Foreground(errorColor)

	var lines []string
	width := layout.DrawerContentWidth()

	if m.drawerDetail == nil {
		lines = append(lines, labelStyle.Render("No event selected"))
	} else {
		event := m.drawerDetail.Event
		meta := m.drawerDetail.Meta

		if meta.Subject != "" {
			lines = append(lines, headerStyle.Render("MESSAGE"))
			lines = append(lines, "")
			wrapped := wrapTextSimple(meta.Subject, width-2)
			for _, line := range strings.Split(wrapped, "\n") {
				lines = append(lines, valueStyle.Render(line))
			}
			lines = append(lines, "")
		}

		if meta.FilesChanged > 0 || meta.Insertions > 0 || meta.Deletions > 0 {
			statsLine := ""
			if meta.FilesChanged > 0 {
				statsLine += formatCount(meta.FilesChanged) + " files"
			}
			if meta.Insertions > 0 {
				if statsLine != "" {
					statsLine += "  "
				}
				statsLine += addStyle.Render("+" + formatCount(meta.Insertions))
			}
			if meta.Deletions > 0 {
				if statsLine != "" {
					statsLine += "  "
				}
				statsLine += delStyle.Render("-" + formatCount(meta.Deletions))
			}
			lines = append(lines, statsLine)
			lines = append(lines, "")
		}

		sourceColor := m.sourceColor(event.Source)
		sourceStyle := lipgloss.NewStyle().Foreground(sourceColor).Bold(true)
		lines = append(lines, sourceStyle.Render(sourceName(event.Source))+" "+labelStyle.Render("on")+" "+valueStyle.Render(event.Branch))
		lines = append(lines, "")

		lines = append(lines, headerStyle.Render("DETAILS"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Commit:  ")+valueStyle.Render(event.Commit))
		lines = append(lines, labelStyle.Render("Repo:    ")+valueStyle.Render(filepath.Base(event.RepoPath)))
		lines = append(lines, labelStyle.Render("Branch:  ")+valueStyle.Render(event.Branch))
		lines = append(lines, "")

		if meta.AuthorName != "" {
			lines = append(lines, labelStyle.Render("Author:  ")+valueStyle.Render(meta.AuthorName))
			if meta.AuthorEmail != "" {
				lines = append(lines, labelStyle.Render("Email:   ")+valueStyle.Render(meta.AuthorEmail))
			}
		}
		lines = append(lines, "")

		lines = append(lines, headerStyle.Render("TIMESTAMPS"))
		lines = append(lines, "")
		if meta.AuthoredAt != "" {
			lines = append(lines, labelStyle.Render("Authored: ")+labelStyle.Render(meta.AuthoredAt))
		}
		lines = append(lines, labelStyle.Render("Recorded: ")+labelStyle.Render(format.Full(event.Timestamp)))

		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Path:    ")+labelStyle.Render(event.RepoPath))

		if meta.Body != "" {
			lines = append(lines, "")
			lines = append(lines, headerStyle.Render("DESCRIPTION"))
			lines = append(lines, "")
			for _, bodyLine := range strings.Split(meta.Body, "\n") {
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
	}

	// Set content and dimensions on viewport
	visibleHeight := max(1, height-2)
	contentWidth := layout.DrawerContentWidth()
	m.drawerViewport.SetSize(contentWidth, visibleHeight)
	m.drawerViewport.SetContent(strings.Join(lines, "\n"))

	// Get visible lines from viewport
	scrollPos := m.drawerViewport.YOffset()
	totalLines := len(lines)

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

func (m activityModel) renderFooter() string {
	help := components.NewThemedHelp()

	var bindings []key.Binding
	tabBinding := key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "focus"))

	switch {
	case m.focusedPanel == 2 && m.drawerOpen:
		bindings = []key.Binding{
			tabBinding,
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "close")),
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("jk", "scroll")),
			key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
			key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
		}
	case m.focusedPanel == 1:
		bindings = []key.Binding{
			tabBinding,
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("1", "2", "3", "4", "5", "6", "7"), key.WithHelp("1-7", "filter")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear")),
		}
	default:
		bindings = []key.Binding{
			tabBinding,
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("jk", "nav")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "detail")),
			key.NewBinding(key.WithKeys("1", "2", "3", "4", "5", "6", "7"), key.WithHelp("1-7", "source")),
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

func (m activityModel) sourceColor(source store.Source) lipgloss.Color {
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

func formatCount(n int) string {
	return fmt.Sprintf("%d", n)
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}
