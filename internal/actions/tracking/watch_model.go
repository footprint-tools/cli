package tracking

import (
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/ui/components"
	"github.com/footprint-tools/cli/internal/ui/style"
)

const (
	maxEvents = 100

	// Adaptive polling intervals
	pollFast    = 50 * time.Millisecond  // When there's recent activity
	pollNormal  = 200 * time.Millisecond // Default polling
	pollSlow    = 500 * time.Millisecond // When idle for a while
	idleTimeout = 3 * time.Second        // Time without events before slowing down
)

// Messages

type tickMsg time.Time

type newEventsMsg []store.RepoEvent

// EventDetail contains enriched event information for the drawer
type EventDetail struct {
	Event store.RepoEvent
	Meta  git.CommitMetadata
}

// watchModel is the Bubble Tea model for interactive watch
type watchModel struct {
	// Database
	db     *sql.DB
	lastID int64

	// Event buffer (circular, newest first for display)
	events []store.RepoEvent

	// Cached commit metadata (commit hash -> metadata)
	commitMeta map[string]git.CommitMetadata

	// Session stats
	sessionStart time.Time
	totalEvents  int
	byType       map[string]int
	byRepo       map[string]int

	// Adaptive polling
	lastEventTime time.Time // When we last received new events

	// UI dimensions
	width  int
	height int

	// State
	paused       bool
	cursor       int
	eventScroll  int
	filterQuery  string
	filterSource store.Source // -1 means no filter
	filterRepo   string       // "" means no filter

	// Focus: 0=events, 1=sidebar, 2=drawer
	focusedPanel    int
	sidebarViewport components.ThemedViewport

	// Drawer
	drawerOpen     bool
	drawerDetail   *EventDetail
	drawerViewport components.ThemedViewport

	// Styling
	colors style.ColorConfig
}

func newWatchModel(db *sql.DB, lastID int64) watchModel {
	return watchModel{
		db:              db,
		lastID:          lastID,
		events:          make([]store.RepoEvent, 0, maxEvents),
		commitMeta:      make(map[string]git.CommitMetadata),
		sessionStart:    time.Now(),
		byType:          make(map[string]int),
		byRepo:          make(map[string]int),
		filterSource:    -1, // No filter
		colors:          style.GetColors(),
		sidebarViewport: components.NewThemedViewport(20, 20),
		drawerViewport:  components.NewThemedViewport(40, 20),
	}
}

// Init implements tea.Model
func (m watchModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnableMouseCellMotion,
		tickCmd(pollFast), // Start fast to catch any immediate events
	)
}

// tickCmd returns a command that ticks at the given poll interval
func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// getPollInterval returns the appropriate polling interval based on recent activity
func (m watchModel) getPollInterval() time.Duration {
	if m.lastEventTime.IsZero() {
		return pollNormal
	}

	timeSinceEvent := time.Since(m.lastEventTime)

	if timeSinceEvent < 500*time.Millisecond {
		// Very recent activity - poll fast
		return pollFast
	} else if timeSinceEvent < idleTimeout {
		// Some recent activity - poll normal
		return pollNormal
	}
	// Idle - poll slow
	return pollSlow
}

// Update implements tea.Model
func (m watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tickMsg:
		if m.paused {
			return m, tickCmd(pollSlow) // Slow poll when paused
		}
		return m, tea.Batch(
			m.pollEvents(),
			tickCmd(m.getPollInterval()),
		)

	case newEventsMsg:
		m.addEvents([]store.RepoEvent(msg))
		// Don't update drawer - cursor is adjusted in addEvents to keep same event selected
		return m, nil
	}

	return m, nil
}

func (m watchModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyTab:
		// Cycle focus: events -> sidebar -> (drawer if open) -> events
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

func (m watchModel) handleDrawerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m watchModel) handleSidebarKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Clear filters or switch focus
		if m.filterSource != -1 || m.filterRepo != "" {
			m.filterSource = -1
			m.filterRepo = ""
			return m, nil
		}
		m.focusedPanel = 0
		return m, nil
	case tea.KeyUp:
		m.sidebarViewport.LineUp(1)
		return m, nil
	case tea.KeyDown:
		m.sidebarViewport.LineDown(1)
		return m, nil
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "j":
		m.sidebarViewport.LineDown(1)
		return m, nil
	case "k":
		m.sidebarViewport.LineUp(1)
		return m, nil
	case "c":
		m.filterSource = -1
		m.filterRepo = ""
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7":
		return m.toggleSourceFilter(msg.String())
	}
	return m, nil
}

func (m watchModel) handleEventsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if len(m.events) > 0 && m.cursor < len(m.events) {
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

func (m watchModel) toggleSourceFilter(key string) (tea.Model, tea.Cmd) {
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

func (m watchModel) handleRunes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q":
		return m, tea.Quit
	case "p":
		m.paused = !m.paused
		return m, nil
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
		if len(m.filteredEvents()) > 0 {
			m.cursor = len(m.filteredEvents()) - 1
		}
		return m, nil
	case "c":
		// Clear all filters
		m.filterQuery = ""
		m.filterSource = -1
		m.filterRepo = ""
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7":
		return m.toggleSourceFilter(key)
	case "/":
		// Start filter mode
		return m, nil
	default:
		// Any other character goes to text filter
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.filterQuery += key
			m.cursor = 0
			m.eventScroll = 0
		}
	}

	return m, nil
}

func (m watchModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Calculate regions
	statsWidth, eventsWidth, _ := m.calculateWidths()
	drawerStart := statsWidth + eventsWidth

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}

		// Determine which panel was clicked and set focus
		switch {
		case msg.X < statsWidth:
			// Click in sidebar
			m.focusedPanel = 1
		case msg.X < drawerStart:
			// Click in events area
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
			// Click in drawer
			m.focusedPanel = 2
		}

	case tea.MouseButtonWheelUp:
		// Scroll based on mouse position
		switch {
		case msg.X < statsWidth:
			m.sidebarViewport.LineUp(1)
		case msg.X < drawerStart:
			m.moveCursor(-1)
		case m.drawerOpen:
			m.drawerViewport.LineUp(1)
		}

	case tea.MouseButtonWheelDown:
		// Scroll based on mouse position
		switch {
		case msg.X < statsWidth:
			m.sidebarViewport.LineDown(1)
		case msg.X < drawerStart:
			m.moveCursor(1)
		case m.drawerOpen:
			m.drawerViewport.LineDown(1)
		}
	}

	return m, nil
}

func (m *watchModel) moveCursor(delta int) {
	filtered := m.filteredEvents()
	if len(filtered) == 0 {
		return
	}

	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(filtered) {
		m.cursor = len(filtered) - 1
	}
}

func (m watchModel) pollEvents() tea.Cmd {
	return func() tea.Msg {
		events, err := store.ListEventsSince(m.db, m.lastID)
		if err != nil {
			return nil
		}
		return newEventsMsg(events)
	}
}

func (m *watchModel) addEvents(events []store.RepoEvent) {
	if len(events) == 0 {
		return
	}

	// Mark that we received events (for adaptive polling)
	m.lastEventTime = time.Now()

	// Count how many events we'll actually add (to adjust cursor)
	eventsAdded := 0

	for _, e := range events {
		// Update lastID
		if e.ID > m.lastID {
			m.lastID = e.ID
		}

		// Update stats
		m.totalEvents++
		m.byType["commit"]++ // For now all events are commits
		repoName := filepath.Base(e.RepoPath)
		m.byRepo[repoName]++

		// Fetch and cache commit metadata
		if _, exists := m.commitMeta[e.Commit]; !exists {
			meta := git.GetCommitMetadata(e.RepoPath, e.Commit)
			m.commitMeta[e.Commit] = meta
		}

		// Add to buffer (prepend for newest-first)
		// More efficient prepend: shift in place instead of allocating new slice
		if len(m.events) < maxEvents {
			m.events = append(m.events, store.RepoEvent{})
		}
		copy(m.events[1:], m.events)
		m.events[0] = e
		eventsAdded++
	}

	// When drawer is open, adjust cursor to keep the same event selected
	// When drawer is closed, cursor stays at 0 (newest event)
	if eventsAdded > 0 && m.drawerOpen && len(m.events) > eventsAdded {
		m.cursor += eventsAdded
		// Clamp to valid range
		filtered := m.filteredEvents()
		if m.cursor >= len(filtered) {
			m.cursor = len(filtered) - 1
		}
	}
	// eventScroll stays at 0 to show newest events at top
}

func (m *watchModel) updateDrawerDetail() {
	filtered := m.filteredEvents()
	if m.cursor < 0 || m.cursor >= len(filtered) {
		m.drawerDetail = nil
		return
	}

	event := filtered[m.cursor]
	meta := m.getCommitMeta(event.RepoPath, event.Commit)
	m.drawerDetail = &EventDetail{
		Event: event,
		Meta:  meta,
	}
}

func (m watchModel) filteredEvents() []store.RepoEvent {
	// No filters active
	if m.filterQuery == "" && m.filterSource == -1 {
		return m.events
	}

	query := strings.ToLower(m.filterQuery)
	var filtered []store.RepoEvent

	for _, e := range m.events {
		// Filter by source type
		if m.filterSource != -1 && e.Source != m.filterSource {
			continue
		}

		// Filter by text query
		if query != "" && !m.matchesQuery(e, query) {
			continue
		}

		filtered = append(filtered, e)
	}

	return filtered
}

func (m watchModel) calculateWidths() (stats, events, drawer int) {
	if m.width == 0 {
		return 25, 75, 0
	}

	// Match splitpanel.Config values from View()
	// Clamp stats width between 18 and 26
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

func (m watchModel) sessionDuration() time.Duration {
	return time.Since(m.sessionStart)
}

func (m watchModel) getCommitMessage(commit string) string {
	if meta, exists := m.commitMeta[commit]; exists {
		return meta.Subject
	}
	return ""
}

// matchesQuery checks if an event matches the search query.
func (m watchModel) matchesQuery(e store.RepoEvent, query string) bool {
	repoName := strings.ToLower(filepath.Base(e.RepoPath))
	return strings.Contains(repoName, query) ||
		strings.Contains(strings.ToLower(e.Branch), query) ||
		strings.Contains(strings.ToLower(e.Commit), query) ||
		strings.Contains(strings.ToLower(m.getCommitMessage(e.Commit)), query)
}

func (m watchModel) getCommitMeta(repoPath, commit string) git.CommitMetadata {
	if meta, exists := m.commitMeta[commit]; exists {
		return meta
	}
	// Fetch if not cached (shouldn't happen normally)
	meta := git.GetCommitMetadata(repoPath, commit)
	return meta
}
