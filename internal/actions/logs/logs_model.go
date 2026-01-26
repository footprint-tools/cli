package logs

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/footprint-tools/footprint-cli/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxLogLines     = 1000
	pollIntervalLog = 500 * time.Millisecond
)

// LogLine represents a parsed log line
type LogLine struct {
	Raw       string
	Timestamp string
	Level     string // ERROR, WARN, INFO, DEBUG
	Caller    string // file:line
	Message   string
}

// Messages

type logTickMsg time.Time

type newLinesMsg []LogLine

type fileStatMsg struct {
	size    int64
	modTime time.Time
}

type initialLoadMsg struct {
	lines   []LogLine
	size    int64
	modTime time.Time
}

type newDataMsg struct {
	lines   []LogLine
	size    int64
	modTime time.Time
}

// logsModel is the Bubble Tea model for interactive logs
type logsModel struct {
	// File state
	logPath     string
	lastSize    int64
	lastModTime time.Time

	// Log buffer (newest last for display)
	lines []LogLine

	// Stats
	sessionStart  time.Time
	totalEver     int            // Total lines in file (all time)
	sessionLines  int            // Lines received during this session
	byLevel       map[string]int // Counts by log level (session only)
	byLevelTotal  map[string]int // Counts by log level (all time)

	// UI dimensions
	width  int
	height int

	// State
	paused      bool
	autoScroll  bool
	cursor      int
	scrollPos   int
	filterQuery string
	filterLevel string // Filter by level: "", "ERROR", "WARN", "INFO", "DEBUG"

	// Drawer
	drawerOpen   bool
	drawerDetail *LogLine

	// Styling
	colors style.ColorConfig
}

func newLogsModel(logPath string) logsModel {
	return logsModel{
		logPath:      logPath,
		lines:        make([]LogLine, 0, maxLogLines),
		sessionStart: time.Now(),
		byLevel:      make(map[string]int),
		byLevelTotal: make(map[string]int),
		autoScroll:   true,
		colors:       style.GetColors(),
	}
}

// Init implements tea.Model
func (m logsModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadInitialLines(),
		logTickCmd(),
	)
}

// logTickCmd returns a command that ticks at the poll interval
func logTickCmd() tea.Cmd {
	return tea.Tick(pollIntervalLog, func(t time.Time) tea.Msg {
		return logTickMsg(t)
	})
}

// Update implements tea.Model
func (m logsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case logTickMsg:
		if m.paused {
			return m, logTickCmd()
		}
		return m, tea.Batch(
			m.checkForNewLines(),
			logTickCmd(),
		)

	case newLinesMsg:
		m.addSessionLines([]LogLine(msg))
		if m.autoScroll && len(m.filteredLines()) > 0 {
			m.cursor = len(m.filteredLines()) - 1
		}
		return m, nil

	case fileStatMsg:
		m.lastSize = msg.size
		m.lastModTime = msg.modTime
		return m, nil

	case initialLoadMsg:
		m.addInitialLines(msg.lines)
		m.lastSize = msg.size
		m.lastModTime = msg.modTime
		if m.autoScroll && len(m.filteredLines()) > 0 {
			m.cursor = len(m.filteredLines()) - 1
		}
		return m, nil

	case newDataMsg:
		m.addSessionLines(msg.lines)
		m.lastSize = msg.size
		m.lastModTime = msg.modTime
		if m.autoScroll && len(m.filteredLines()) > 0 {
			m.cursor = len(m.filteredLines()) - 1
		}
		return m, nil
	}

	return m, nil
}

func (m logsModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.filterQuery != "" || m.filterLevel != "" {
			m.filterQuery = ""
			m.filterLevel = ""
			return m, nil
		}
		return m, tea.Quit

	case tea.KeyEnter:
		if len(m.filteredLines()) > 0 && m.cursor < len(m.filteredLines()) {
			m.drawerOpen = true
			m.updateDrawerDetail()
		}
		return m, nil

	case tea.KeyUp:
		m.moveCursor(-1)
		m.autoScroll = false
		return m, nil

	case tea.KeyDown:
		m.moveCursor(1)
		return m, nil

	case tea.KeyPgUp:
		m.moveCursor(-10)
		m.autoScroll = false
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

func (m logsModel) handleRunes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.autoScroll = false
		return m, nil
	case "g":
		m.cursor = 0
		m.scrollPos = 0
		m.autoScroll = false
		return m, nil
	case "G":
		if len(m.filteredLines()) > 0 {
			m.cursor = len(m.filteredLines()) - 1
			m.autoScroll = true
		}
		return m, nil
	case "c":
		// Clear filter
		m.filterQuery = ""
		m.filterLevel = ""
		return m, nil
	case "e":
		// Filter by ERROR
		if m.filterLevel == "ERROR" {
			m.filterLevel = ""
		} else {
			m.filterLevel = "ERROR"
		}
		m.cursor = 0
		m.scrollPos = 0
		return m, nil
	case "w":
		// Filter by WARN
		if m.filterLevel == "WARN" {
			m.filterLevel = ""
		} else {
			m.filterLevel = "WARN"
		}
		m.cursor = 0
		m.scrollPos = 0
		return m, nil
	case "i":
		// Filter by INFO
		if m.filterLevel == "INFO" {
			m.filterLevel = ""
		} else {
			m.filterLevel = "INFO"
		}
		m.cursor = 0
		m.scrollPos = 0
		return m, nil
	case "d":
		// Filter by DEBUG
		if m.filterLevel == "DEBUG" {
			m.filterLevel = ""
		} else {
			m.filterLevel = "DEBUG"
		}
		m.cursor = 0
		m.scrollPos = 0
		return m, nil
	case "a":
		// Toggle auto-scroll
		m.autoScroll = !m.autoScroll
		if m.autoScroll && len(m.filteredLines()) > 0 {
			m.cursor = len(m.filteredLines()) - 1
		}
		return m, nil
	default:
		// Any other character goes to filter
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.filterQuery += key
			m.cursor = 0
			m.scrollPos = 0
		}
	}

	return m, nil
}

func (m logsModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Calculate regions
	statsWidth := m.calculateStatsWidth()

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		// If drawer is open and click is outside drawer, close it
		if m.drawerOpen {
			drawerStart := statsWidth + m.calculateLogsWidth()
			if msg.X < drawerStart {
				m.drawerOpen = false
				m.drawerDetail = nil
				return m, nil
			}
		}

		// Click in logs area - select line
		if msg.X >= statsWidth {
			headerHeight := 3
			footerHeight := 2
			if msg.Y >= headerHeight && msg.Y < m.height-footerHeight {
				clickedLine := msg.Y - headerHeight
				clickedIdx := m.scrollPos + clickedLine
				filtered := m.filteredLines()
				if clickedIdx >= 0 && clickedIdx < len(filtered) {
					m.cursor = clickedIdx
					m.autoScroll = false
				}
			}
		}

	case tea.MouseButtonWheelUp:
		m.moveCursor(-3)
		m.autoScroll = false

	case tea.MouseButtonWheelDown:
		m.moveCursor(3)
	}

	return m, nil
}

func (m *logsModel) moveCursor(delta int) {
	filtered := m.filteredLines()
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

	// Check if we're at the end
	if m.cursor == len(filtered)-1 {
		m.autoScroll = true
	}
}

func (m logsModel) loadInitialLines() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(m.logPath)
		if err != nil {
			return nil
		}
		defer func() { _ = file.Close() }()

		var lines []LogLine
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := parseLine(scanner.Text())
			lines = append(lines, line)
		}

		// Get file stats
		info, err := file.Stat()
		if err == nil {
			// Return a combined message
			return initialLoadMsg{
				lines:   lines,
				size:    info.Size(),
				modTime: info.ModTime(),
			}
		}

		return newLinesMsg(lines)
	}
}

func (m logsModel) checkForNewLines() tea.Cmd {
	return func() tea.Msg {
		info, err := os.Stat(m.logPath)
		if err != nil {
			return nil
		}

		// Check if file has grown
		if info.Size() <= m.lastSize && !info.ModTime().After(m.lastModTime) {
			return nil
		}

		// Read new lines from lastSize
		file, err := os.Open(m.logPath)
		if err != nil {
			return nil
		}
		defer func() { _ = file.Close() }()

		// Seek to last known position
		if m.lastSize > 0 {
			_, err = file.Seek(m.lastSize, 0)
			if err != nil {
				return nil
			}
		}

		var lines []LogLine
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := parseLine(scanner.Text())
			lines = append(lines, line)
		}

		if len(lines) > 0 {
			return newDataMsg{
				lines:   lines,
				size:    info.Size(),
				modTime: info.ModTime(),
			}
		}

		return fileStatMsg{size: info.Size(), modTime: info.ModTime()}
	}
}

// addInitialLines adds lines from the initial file load (not counted as session)
func (m *logsModel) addInitialLines(lines []LogLine) {
	for _, line := range lines {
		// Update total stats only
		m.totalEver++
		if line.Level != "" {
			m.byLevelTotal[line.Level]++
		}

		// Add to buffer
		m.lines = append(m.lines, line)

		// Trim if too many
		if len(m.lines) > maxLogLines {
			m.lines = m.lines[1:]
		}
	}
}

// addSessionLines adds lines that arrived during this session
func (m *logsModel) addSessionLines(lines []LogLine) {
	for _, line := range lines {
		// Update both total and session stats
		m.totalEver++
		m.sessionLines++
		if line.Level != "" {
			m.byLevelTotal[line.Level]++
			m.byLevel[line.Level]++
		}

		// Add to buffer
		m.lines = append(m.lines, line)

		// Trim if too many
		if len(m.lines) > maxLogLines {
			m.lines = m.lines[1:]
		}
	}
}

func (m *logsModel) updateDrawerDetail() {
	filtered := m.filteredLines()
	if m.cursor < 0 || m.cursor >= len(filtered) {
		m.drawerDetail = nil
		return
	}
	line := filtered[m.cursor]
	m.drawerDetail = &line
}

func (m logsModel) filteredLines() []LogLine {
	var filtered []LogLine

	query := strings.ToLower(m.filterQuery)

	for _, line := range m.lines {
		// Filter by level
		if m.filterLevel != "" && line.Level != m.filterLevel {
			continue
		}

		// Filter by query
		if query != "" {
			if !strings.Contains(strings.ToLower(line.Raw), query) {
				continue
			}
		}

		filtered = append(filtered, line)
	}

	return filtered
}

func (m logsModel) calculateStatsWidth() int {
	if m.width == 0 {
		return 20
	}

	statsWidth := m.width * 20 / 100
	if statsWidth < 18 {
		statsWidth = 18
	}
	if statsWidth > 24 {
		statsWidth = 24
	}
	return statsWidth
}

func (m logsModel) calculateLogsWidth() int {
	statsWidth := m.calculateStatsWidth()
	if m.drawerOpen {
		drawerWidth := m.width * 35 / 100
		if drawerWidth < 30 {
			drawerWidth = 30
		}
		return m.width - statsWidth - drawerWidth - 2
	}
	return m.width - statsWidth - 1
}

func (m logsModel) sessionDuration() time.Duration {
	return time.Since(m.sessionStart)
}

// parseLine parses a log line into its components
func parseLine(raw string) LogLine {
	line := LogLine{Raw: raw}

	// Try to parse timestamp, level, caller, and message
	// New format: [2024-01-15 10:30:45] LEVEL file:line: message
	// Old format: [2024-01-15 10:30:45] LEVEL: message
	if len(raw) > 22 && raw[0] == '[' {
		closeBracket := strings.Index(raw, "]")
		if closeBracket > 0 {
			line.Timestamp = raw[1:closeBracket]
			rest := strings.TrimSpace(raw[closeBracket+1:])

			// Check for level (with or without caller)
			for _, level := range []string{"ERROR", "WARN", "INFO", "DEBUG"} {
				if strings.HasPrefix(rest, level+" ") || strings.HasPrefix(rest, level+":") {
					line.Level = level

					// Skip level
					afterLevel := strings.TrimSpace(rest[len(level):])

					// Check if next part is caller (file.go:123:) or old format (:)
					if len(afterLevel) > 0 && afterLevel[0] == ':' {
						// Old format: LEVEL: message
						line.Message = strings.TrimSpace(afterLevel[1:])
					} else {
						// New format: LEVEL file:line: message
						// Find the colon after file:line
						colonIdx := strings.Index(afterLevel, ":")
						if colonIdx > 0 {
							line.Caller = strings.TrimSpace(afterLevel[:colonIdx])
							line.Message = strings.TrimSpace(afterLevel[colonIdx+1:])
						} else {
							line.Message = afterLevel
						}
					}
					return line
				}
			}
			line.Message = rest
		}
	}

	return line
}
