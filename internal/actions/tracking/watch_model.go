package tracking

import (
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/store"
	"github.com/Skryensya/footprint/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxEvents       = 100
	pollIntervalTUI = 300 * time.Millisecond
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

	// Session stats
	sessionStart time.Time
	totalEvents  int
	byType       map[string]int
	byRepo       map[string]int

	// UI dimensions
	width  int
	height int

	// State
	paused      bool
	cursor      int
	eventScroll int
	filterQuery string

	// Drawer
	drawerOpen   bool
	drawerDetail *EventDetail

	// Styling
	colors style.ColorConfig
}

func newWatchModel(db *sql.DB, lastID int64) watchModel {
	return watchModel{
		db:           db,
		lastID:       lastID,
		events:       make([]store.RepoEvent, 0, maxEvents),
		sessionStart: time.Now(),
		byType:       make(map[string]int),
		byRepo:       make(map[string]int),
		colors:       style.GetColors(),
	}
}

// Init implements tea.Model
func (m watchModel) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
	)
}

// tickCmd returns a command that ticks at the poll interval
func tickCmd() tea.Cmd {
	return tea.Tick(pollIntervalTUI, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
			return m, tickCmd()
		}
		return m, tea.Batch(
			m.pollEvents(),
			tickCmd(),
		)

	case newEventsMsg:
		m.addEvents([]store.RepoEvent(msg))
		// Update drawer if open
		if m.drawerOpen && m.cursor < len(m.events) {
			m.updateDrawerDetail()
		}
		return m, nil
	}

	return m, nil
}

func (m watchModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If drawer is open, handle drawer-specific keys
	if m.drawerOpen {
		switch msg.Type {
		case tea.KeyEsc:
			m.drawerOpen = false
			m.drawerDetail = nil
			return m, nil
		}

		switch msg.String() {
		case "q":
			m.drawerOpen = false
			m.drawerDetail = nil
			return m, nil
		case "j", "down":
			m.moveCursor(1)
			m.updateDrawerDetail()
			return m, nil
		case "k", "up":
			m.moveCursor(-1)
			m.updateDrawerDetail()
			return m, nil
		}
		return m, nil
	}

	// Normal mode keys
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		if m.filterQuery != "" {
			m.filterQuery = ""
			return m, nil
		}
		return m, tea.Quit

	case tea.KeyEnter:
		if len(m.events) > 0 && m.cursor < len(m.events) {
			m.drawerOpen = true
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
		// Clear filter
		m.filterQuery = ""
		return m, nil
	case "/":
		// Start filter mode - just append to filter
		return m, nil
	default:
		// Any other character goes to filter
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
	statsWidth, eventsWidth, drawerWidth := m.calculateWidths()

	switch msg.Type {
	case tea.MouseLeft:
		// If drawer is open and click is outside drawer, close it
		if m.drawerOpen {
			drawerStart := statsWidth + eventsWidth
			if msg.X < drawerStart {
				m.drawerOpen = false
				m.drawerDetail = nil
				return m, nil
			}
		}

		// Click in events area - select event
		if msg.X >= statsWidth && msg.X < statsWidth+eventsWidth {
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
		}

	case tea.MouseWheelUp:
		if msg.X >= statsWidth && msg.X < statsWidth+eventsWidth+drawerWidth {
			m.moveCursor(-1)
		}

	case tea.MouseWheelDown:
		if msg.X >= statsWidth && msg.X < statsWidth+eventsWidth+drawerWidth {
			m.moveCursor(1)
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

		// Add to buffer (prepend for newest-first)
		// More efficient prepend: shift in place instead of allocating new slice
		if len(m.events) < maxEvents {
			m.events = append(m.events, store.RepoEvent{})
		}
		copy(m.events[1:], m.events)
		m.events[0] = e
	}
}

func (m *watchModel) updateDrawerDetail() {
	filtered := m.filteredEvents()
	if m.cursor < 0 || m.cursor >= len(filtered) {
		m.drawerDetail = nil
		return
	}

	event := filtered[m.cursor]
	meta := git.GetCommitMetadata(event.RepoPath, event.Commit)
	m.drawerDetail = &EventDetail{
		Event: event,
		Meta:  meta,
	}
}

func (m watchModel) filteredEvents() []store.RepoEvent {
	if m.filterQuery == "" {
		return m.events
	}

	query := strings.ToLower(m.filterQuery)
	var filtered []store.RepoEvent

	for _, e := range m.events {
		repoName := strings.ToLower(filepath.Base(e.RepoPath))
		branch := strings.ToLower(e.Branch)
		commit := strings.ToLower(e.Commit)

		if strings.Contains(repoName, query) ||
			strings.Contains(branch, query) ||
			strings.Contains(commit, query) {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

func (m watchModel) calculateWidths() (stats, events, drawer int) {
	if m.width == 0 {
		return 25, 75, 0
	}

	if m.drawerOpen {
		// 20% stats, 35% events, 45% drawer
		stats = m.width * 20 / 100
		if stats < 20 {
			stats = 20
		}
		drawer = m.width * 45 / 100
		if drawer < 30 {
			drawer = 30
		}
		events = m.width - stats - drawer - 2 // borders
		if events < 20 {
			events = 20
		}
	} else {
		// 25% stats, 75% events
		stats = m.width * 25 / 100
		if stats < 20 {
			stats = 20
		}
		if stats > 30 {
			stats = 30
		}
		events = m.width - stats - 1 // border
		drawer = 0
	}
	return
}

func (m watchModel) sessionDuration() time.Duration {
	return time.Since(m.sessionStart)
}

func (m watchModel) eventsPerMinute() float64 {
	duration := m.sessionDuration()
	if duration < time.Second {
		return 0
	}
	minutes := duration.Minutes()
	if minutes < 0.1 {
		minutes = 0.1
	}
	return float64(m.totalEvents) / minutes
}
