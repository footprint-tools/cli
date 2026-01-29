package tracking

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/git"
	"github.com/footprint-tools/cli/internal/hooks"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/ui/components"
	"github.com/footprint-tools/cli/internal/ui/splitpanel"
	"github.com/footprint-tools/cli/internal/ui/style"
	"golang.org/x/term"
)

// RepoEntry represents a discovered git repository.
type RepoEntry struct {
	Path         string               // Absolute path to the repo
	Name         string               // Directory name
	HasHooks     bool                 // Whether fp hooks are installed
	HooksChanged bool                 // Whether hooks state was changed this session
	Selected     bool                 // Whether selected for batch operation
	Inspection   hooks.RepoInspection // Preflight inspection result
}

// ReposInteractive launches the interactive repository manager.
func ReposInteractive(args []string, flags *dispatchers.ParsedFlags) error {
	return reposInteractive(args, flags, DefaultDeps())
}

func reposInteractive(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("interactive mode requires a terminal")
	}

	// Get root directory - default to current directory
	root := flags.String("--root", ".")

	// Make path absolute
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("invalid path %s: %w", root, err)
	}
	root = absRoot

	// Get max depth (default: no practical limit)
	maxDepth := flags.Int("--depth", 25)

	// Scan for repos
	_, _ = deps.Printf("Scanning for git repositories in %s...\n", root)
	repos, err := scanForRepos(root, maxDepth)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		_, _ = deps.Println("No git repositories found")
		return nil
	}

	_, _ = deps.Printf("Found %d repositories\n", len(repos))

	// Launch TUI
	m := newReposModel(repos)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	final, err := p.Run()
	if err != nil {
		return err
	}

	fm := final.(reposModel)

	// Show summary of changes
	if fm.installed > 0 || fm.uninstalled > 0 {
		_, _ = deps.Println()
		if fm.installed > 0 {
			_, _ = deps.Printf("Installed hooks in %d repositories\n", fm.installed)
		}
		if fm.uninstalled > 0 {
			_, _ = deps.Printf("Removed hooks from %d repositories\n", fm.uninstalled)
		}
	}

	return nil
}

// scanForRepos finds git repositories under the given root.
func scanForRepos(root string, maxDepth int) ([]RepoEntry, error) {
	var repos []RepoEntry
	seen := make(map[string]bool)

	// Directories to skip (optimization)
	skipDirs := map[string]bool{
		// Package managers / dependencies
		"node_modules":     true,
		"vendor":           true,
		"bower_components": true,
		"jspm_packages":    true,
		".pnpm":            true,
		// Python
		"__pycache__":   true,
		".venv":         true,
		"venv":          true,
		"env":           true,
		".eggs":         true,
		"site-packages": true,
		// Build outputs
		"dist":   true,
		"build":  true,
		"target": true,
		"out":    true,
		"_build": true,
		// IDE / tools
		".idea":   true,
		".vscode": true,
		// System / caches
		".cache": true,
		".npm":   true,
		".yarn":  true,
		// macOS
		"Library":      true,
		".Trash":       true,
		"Applications": true,
		// Git internals (don't descend)
		".git": true,
	}

	rootDepth := strings.Count(root, string(os.PathSeparator))

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible directories
		}

		if !d.IsDir() {
			return nil
		}

		// Check depth
		currentDepth := strings.Count(path, string(os.PathSeparator)) - rootDepth
		if currentDepth > maxDepth {
			return fs.SkipDir
		}

		// Skip certain directories
		name := d.Name()
		if skipDirs[name] || (strings.HasPrefix(name, ".") && name != ".") {
			return fs.SkipDir
		}

		// Check if this is a git repo
		gitDir := filepath.Join(path, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			if !seen[path] {
				seen[path] = true
				inspection := hooks.InspectRepo(path)
				entry := RepoEntry{
					Path:       path,
					Name:       filepath.Base(path),
					HasHooks:   inspection.FpInstalled,
					Inspection: inspection,
				}
				repos = append(repos, entry)
			}
			return fs.SkipDir // Don't descend into git repos
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by path
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Path < repos[j].Path
	})

	return repos, nil
}

// reposModel is the Bubble Tea model for the interactive repos view.
type reposModel struct {
	repos          []RepoEntry
	cursor         int
	reposViewport  components.ThemedViewport
	width          int
	height         int
	installed      int
	uninstalled    int
	message        string
	colors         style.ColorConfig
	focusSidebar   bool
	drawerOpen     bool // Whether the details drawer is open
	drawerViewport components.ThemedViewport
}

func newReposModel(repos []RepoEntry) reposModel {
	return reposModel{
		repos:          repos,
		cursor:         0,
		colors:         style.GetColors(),
		focusSidebar:   false, // Start with focus on repo list
		reposViewport:  components.NewThemedViewport(60, 20),
		drawerViewport: components.NewThemedViewport(40, 20),
	}
}

func (m reposModel) Init() tea.Cmd {
	return nil
}

func (m reposModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// If drawer is open, handle drawer-specific keys
		if m.drawerOpen {
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc, tea.KeyEnter:
				m.drawerOpen = false
				m.drawerViewport.GotoTop()
				return m, nil
			case tea.KeyUp:
				m.drawerViewport.LineUp(1)
				return m, nil
			case tea.KeyDown:
				m.drawerViewport.LineDown(1)
				return m, nil
			case tea.KeyPgUp:
				m.drawerViewport.LineUp(5)
				return m, nil
			case tea.KeyPgDown:
				m.drawerViewport.LineDown(5)
				return m, nil
			case tea.KeyHome:
				m.drawerViewport.GotoTop()
				return m, nil
			case tea.KeyRunes:
				switch string(msg.Runes) {
				case "q":
					m.drawerOpen = false
					m.drawerViewport.GotoTop()
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
				case "?":
					m.drawerOpen = false
					m.drawerViewport.GotoTop()
					return m, nil
				}
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyTab:
			m.focusSidebar = !m.focusSidebar
			return m, nil

		case tea.KeyUp:
			if !m.focusSidebar && m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if !m.focusSidebar && m.cursor < len(m.repos)-1 {
				m.cursor++
			}

		case tea.KeyPgUp:
			if !m.focusSidebar {
				m.cursor -= 10
				if m.cursor < 0 {
					m.cursor = 0
				}
			}

		case tea.KeyPgDown:
			if !m.focusSidebar {
				m.cursor += 10
				if m.cursor >= len(m.repos) {
					m.cursor = len(m.repos) - 1
				}
			}

		case tea.KeyHome:
			if !m.focusSidebar {
				m.cursor = 0
			}

		case tea.KeyEnd:
			if !m.focusSidebar {
				m.cursor = len(m.repos) - 1
			}

		case tea.KeyEnter, tea.KeySpace:
			// Both toggle selection
			if !m.focusSidebar && len(m.repos) > 0 {
				m.toggleSelection()
			}

		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				return m, tea.Quit
			case "j":
				if !m.focusSidebar && m.cursor < len(m.repos)-1 {
					m.cursor++
				}
			case "k":
				if !m.focusSidebar && m.cursor > 0 {
					m.cursor--
				}
			case "g":
				if !m.focusSidebar {
					m.cursor = 0
				}
			case "G":
				if !m.focusSidebar {
					m.cursor = len(m.repos) - 1
				}
			case "h":
				m.focusSidebar = true
			case "l":
				m.focusSidebar = false
			case "i":
				m.installSelected()
			case "u":
				m.uninstallSelected()
			case "a":
				m.selectAll()
			case "A":
				m.deselectAll()
			case "?", "d":
				// Open drawer for any repo (d = details)
				if len(m.repos) > 0 {
					m.drawerOpen = true
					m.drawerViewport.GotoTop()
				}
			}
		}
	}

	return m, nil
}

func (m *reposModel) toggleSelection() {
	if len(m.repos) == 0 {
		return
	}

	repo := &m.repos[m.cursor]

	// Blocked repos can't be selected, show drawer instead
	if !repo.Inspection.Status.CanInstall() && !repo.HasHooks {
		m.drawerOpen = true
		return
	}

	// Toggle selection for clean repos or repos with hooks installed
	repo.Selected = !repo.Selected
}

func (m *reposModel) installSelected() {
	count := 0
	for i := range m.repos {
		r := &m.repos[i]
		if !r.Selected || r.HasHooks {
			continue
		}
		if !r.Inspection.Status.CanInstall() {
			continue
		}
		hooksPath, err := git.RepoHooksPath(r.Path)
		if err != nil {
			continue
		}
		if err := hooks.Install(hooksPath); err != nil {
			continue
		}
		r.HasHooks = true
		r.HooksChanged = true
		r.Inspection.FpInstalled = true
		r.Selected = false
		m.installed++
		count++
		addRepoToStore(r.Path)
	}
	if count > 0 {
		m.message = fmt.Sprintf("Installed hooks in %d repos", count)
	} else {
		m.message = "No repos selected for installation"
	}
}

func (m *reposModel) uninstallSelected() {
	count := 0
	for i := range m.repos {
		r := &m.repos[i]
		if !r.Selected || !r.HasHooks {
			continue
		}
		hooksPath, err := git.RepoHooksPath(r.Path)
		if err != nil {
			continue
		}
		if err := hooks.Uninstall(hooksPath); err != nil {
			continue
		}
		r.HasHooks = false
		r.HooksChanged = true
		r.Selected = false
		m.uninstalled++
		count++
		removeRepoFromStore(r.Path)
	}
	if count > 0 {
		m.message = fmt.Sprintf("Removed hooks from %d repos", count)
	} else {
		m.message = "No repos selected for removal"
	}
}

func addRepoToStore(repoPath string) {
	s, err := store.New(store.DBPath())
	if err != nil {
		return
	}
	defer func() { _ = s.Close() }()
	_ = s.AddRepo(repoPath)
}

func removeRepoFromStore(repoPath string) {
	s, err := store.New(store.DBPath())
	if err != nil {
		return
	}
	defer func() { _ = s.Close() }()
	_ = s.RemoveRepo(repoPath)
}

func (m *reposModel) selectAll() {
	count := 0
	for i := range m.repos {
		r := &m.repos[i]
		// Only select repos that can be installed or already have hooks
		if r.Inspection.Status.CanInstall() || r.HasHooks {
			if !r.Selected {
				r.Selected = true
				count++
			}
		}
	}
	m.message = fmt.Sprintf("Selected %d repos", count)
}

func (m *reposModel) deselectAll() {
	count := 0
	for i := range m.repos {
		r := &m.repos[i]
		if r.Selected {
			r.Selected = false
			count++
		}
	}
	m.message = fmt.Sprintf("Deselected %d repos", count)
}

func (m *reposModel) countWithHooks() int {
	count := 0
	for _, r := range m.repos {
		if r.HasHooks {
			count++
		}
	}
	return count
}

func (m reposModel) View() string {
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
		SidebarWidthPercent: 0.22,
		SidebarMinWidth:     18,
		SidebarMaxWidth:     24,
		HasDrawer:           true,
		DrawerWidthPercent:  0.40,
	}
	layout := splitpanel.NewLayout(m.width, cfg, m.colors)
	layout.SetDrawerOpen(m.drawerOpen)

	// Set focus: drawer (2) > sidebar (1) > content (0)
	switch {
	case m.drawerOpen:
		layout.SetFocusedPanel(2)
	case m.focusSidebar:
		layout.SetFocusedPanel(1)
	default:
		layout.SetFocusedPanel(0)
	}

	// Build panels
	statsPanel := m.buildStatsPanel(layout, mainHeight)
	reposPanel := m.buildReposPanel(layout, mainHeight)

	// Render
	header := m.renderHeader()

	var main string
	if m.drawerOpen && len(m.repos) > 0 {
		drawerPanel := m.buildDrawerPanel(layout, mainHeight)
		main = layout.RenderWithDrawer(statsPanel, reposPanel, &drawerPanel, mainHeight)
	} else {
		main = layout.Render(statsPanel, reposPanel, mainHeight)
	}

	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func (m *reposModel) buildDrawerPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	headerColor := lipgloss.Color(colors.Header)
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	warnColor := lipgloss.Color(colors.Warning)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)
	successStyle := lipgloss.NewStyle().Foreground(successColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)

	var lines []string
	width := layout.DrawerContentWidth()

	repo := m.repos[m.cursor]

	// Header with repo name
	lines = append(lines, headerStyle.Render(repo.Name))
	lines = append(lines, "")

	// Path
	lines = append(lines, labelStyle.Render(repo.Path))
	lines = append(lines, "")

	// Status section based on repo state
	switch {
	case repo.HasHooks:
		// Installed repo
		lines = append(lines, successStyle.Render("Tracking active"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Footprint hooks are installed"))
		lines = append(lines, labelStyle.Render("and recording your activity."))
		lines = append(lines, "")
		lines = append(lines, headerStyle.Render("ACTIONS"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Select and press ")+valueStyle.Render("u")+labelStyle.Render(" to remove"))
	case repo.Inspection.Status.CanInstall():
		// Ready to install
		lines = append(lines, valueStyle.Render("Ready to track"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("This repo has no conflicts."))
		lines = append(lines, labelStyle.Render("You can install hooks now."))
		lines = append(lines, "")
		lines = append(lines, headerStyle.Render("ACTIONS"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Select and press ")+valueStyle.Render("i")+labelStyle.Render(" to install"))
	default:
		// Blocked - needs setup
		lines = append(lines, warnStyle.Render("Needs setup"))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Status: ")+warnStyle.Render(repo.Inspection.Status.String()))
		lines = append(lines, "")

		// Show details about conflicts
		if repo.Inspection.GlobalHooksPath != "" {
			lines = append(lines, headerStyle.Render("DETECTED"))
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render("Global hooks path:"))
			lines = append(lines, labelStyle.Render("  "+repo.Inspection.GlobalHooksPath))
			lines = append(lines, "")
		}
		if len(repo.Inspection.UnmanagedHooks) > 0 {
			lines = append(lines, headerStyle.Render("EXISTING HOOKS"))
			lines = append(lines, "")
			for _, h := range repo.Inspection.UnmanagedHooks {
				lines = append(lines, labelStyle.Render("  "+h))
			}
			lines = append(lines, "")
		}

		// Show guidance
		guidance := hooks.GetGuidance(repo.Inspection)
		lines = append(lines, headerStyle.Render("HOW TO FIX"))
		lines = append(lines, "")
		for _, line := range strings.Split(guidance, "\n") {
			if line == "" {
				lines = append(lines, "")
				continue
			}
			wrapped := wrapTextSimple(line, width-2)
			for _, wl := range strings.Split(wrapped, "\n") {
				lines = append(lines, labelStyle.Render(wl))
			}
		}
	}

	// Set content and dimensions on viewport
	visibleLines := height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}
	totalLines := len(lines)
	contentWidth := layout.DrawerContentWidth()
	m.drawerViewport.SetSize(contentWidth, visibleLines)
	m.drawerViewport.SetContent(strings.Join(lines, "\n"))

	// Get visible lines from viewport
	scrollPos := m.drawerViewport.YOffset()
	endLine := min(scrollPos+visibleLines, totalLines)
	startLine := min(scrollPos, totalLines)
	visibleContent := lines
	if startLine < totalLines {
		visibleContent = lines[startLine:endLine]
	}

	return splitpanel.Panel{
		Lines:      visibleContent,
		ScrollPos:  scrollPos,
		TotalItems: totalLines,
	}
}

// wrapTextSimple wraps text to the specified width.
func wrapTextSimple(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if lineLen+wordLen+1 > width && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}
		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += wordLen
		_ = i
	}

	return result.String()
}

func (m reposModel) renderHeader() string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	warnColor := lipgloss.Color(colors.Warning)
	uiActiveColor := lipgloss.Color(colors.UIActive)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)
	activeStyle := lipgloss.NewStyle().Foreground(uiActiveColor)

	title := titleStyle.Render("fp repos")
	count := mutedStyle.Render(fmt.Sprintf(" | %d repositories", len(m.repos)))

	// Position indicator
	positionStr := ""
	if len(m.repos) > 0 {
		current := m.cursor + 1
		total := len(m.repos)
		positionStr = mutedStyle.Render(" | ") + activeStyle.Render(fmt.Sprintf("%d", current)) + mutedStyle.Render("/") + mutedStyle.Render(fmt.Sprintf("%d", total))
	}

	// Message if any
	msgStr := ""
	if m.message != "" {
		msgStr = mutedStyle.Render(" | ") + warnStyle.Render(m.message)
	}

	headerContent := title + count + positionStr + msgStr

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}

func (m *reposModel) buildStatsPanel(_ *splitpanel.Layout, _ int) splitpanel.Panel {
	colors := m.colors
	headerColor := lipgloss.Color(colors.Header)
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	warnColor := lipgloss.Color(colors.Warning)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)
	successStyle := lipgloss.NewStyle().Foreground(successColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)

	var lines []string

	// Counts - simplified
	withHooks := m.countWithHooks()
	clean, blocked := m.countByInstallability()
	ready := clean - withHooks
	selected := m.countSelected()

	// Overview with non-color identifiers
	lines = append(lines, headerStyle.Render("REPOS"))
	lines = append(lines, "")
	lines = append(lines, successStyle.Render(fmt.Sprintf("%d ✓", withHooks))+labelStyle.Render(" tracking"))
	lines = append(lines, valueStyle.Render(fmt.Sprintf("%d", ready))+labelStyle.Render(" ready"))
	if blocked > 0 {
		lines = append(lines, warnStyle.Render(fmt.Sprintf("%d !", blocked))+labelStyle.Render(" need setup"))
	}
	lines = append(lines, "")

	// Selection - only show when relevant
	if selected > 0 {
		lines = append(lines, headerStyle.Render("SELECTED"))
		lines = append(lines, valueStyle.Render(fmt.Sprintf("%d repos", selected)))
		lines = append(lines, "")
	}

	// Session changes - only show when there are changes
	if m.installed > 0 || m.uninstalled > 0 {
		lines = append(lines, headerStyle.Render("CHANGES"))
		if m.installed > 0 {
			lines = append(lines, successStyle.Render(fmt.Sprintf("+%d", m.installed))+labelStyle.Render(" added"))
		}
		if m.uninstalled > 0 {
			lines = append(lines, warnStyle.Render(fmt.Sprintf("-%d", m.uninstalled))+labelStyle.Render(" removed"))
		}
		lines = append(lines, "")
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  0,
		TotalItems: len(lines),
	}
}

func (m *reposModel) countSelected() int {
	count := 0
	for _, r := range m.repos {
		if r.Selected {
			count++
		}
	}
	return count
}

func (m *reposModel) countByInstallability() (clean, blocked int) {
	for _, r := range m.repos {
		if r.Inspection.Status.CanInstall() {
			clean++
		} else {
			blocked++
		}
	}
	return
}

func (m *reposModel) buildReposPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	warnColor := lipgloss.Color(colors.Warning)
	uiActiveColor := lipgloss.Color(colors.UIActive)

	// Simplified styling - state shown through color, not icons
	installedNameStyle := lipgloss.NewStyle().Foreground(successColor)
	readyNameStyle := lipgloss.NewStyle() // Default terminal color for "ready" repos
	blockedNameStyle := lipgloss.NewStyle().Foreground(warnColor)
	changedStyle := lipgloss.NewStyle().Foreground(warnColor)
	pathStyle := lipgloss.NewStyle().Foreground(mutedColor)

	// Style for selected rows - background highlight
	selectedBg := lipgloss.NewStyle().Background(uiActiveColor).Foreground(lipgloss.Color("0"))

	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Update viewport size
	contentWidth := layout.MainContentWidth()
	m.reposViewport.SetSize(contentWidth, visibleHeight)

	// Adjust scroll to keep cursor visible
	scroll := m.reposViewport.YOffset()
	if m.cursor < scroll {
		scroll = m.cursor
		m.reposViewport.SetYOffset(scroll)
	}
	// Each repo takes 3 lines when selected (name, path, status), 1 otherwise
	effectiveHeight := visibleHeight / 3 // Conservative estimate
	if m.cursor >= scroll+effectiveHeight {
		scroll = m.cursor - effectiveHeight + 1
		m.reposViewport.SetYOffset(scroll)
	}

	var lines []string

	home, _ := os.UserHomeDir()

	lineCount := 0
	for i := scroll; i < len(m.repos) && lineCount < visibleHeight; i++ {
		repo := m.repos[i]

		// Build prefix: selected items are less indented to stand out
		var prefix string
		if repo.Selected {
			prefix = "* "
		} else {
			prefix = "    "
		}

		canInstall := repo.Inspection.Status.CanInstall()

		// Name styling based on state - color + non-color identifier
		var name string
		var marker string
		switch {
		case repo.HasHooks:
			// Installed: green with checkmark
			name = installedNameStyle.Render(repo.Name)
			marker = installedNameStyle.Render(" ✓")
		case canInstall:
			// Ready to install: normal, no marker
			name = readyNameStyle.Render(repo.Name)
			marker = ""
		default:
			// Blocked: yellow with "!" indicator
			name = blockedNameStyle.Render(repo.Name)
			marker = blockedNameStyle.Render(" !")
		}

		// Add cursor indicator without changing color
		cursorIndicator := "  "
		if i == m.cursor {
			cursorIndicator = "> "
		}

		// Changed indicator (this session)
		changed := ""
		if repo.HooksChanged {
			changed = changedStyle.Render(" ~")
		}

		line := fmt.Sprintf("%s%s%s%s%s", cursorIndicator, prefix, name, marker, changed)

		// Pad line to full width for selection background
		lineWidth := lipgloss.Width(line)
		if lineWidth < contentWidth {
			line += strings.Repeat(" ", contentWidth-lineWidth)
		} else if lineWidth > contentWidth {
			line = line[:contentWidth-3] + "..."
		}

		// Apply selection background to entire line width
		if repo.Selected {
			line = selectedBg.Render(line)
		}

		lines = append(lines, line)
		lineCount++

		// Show path when cursor is here
		if i == m.cursor && lineCount < visibleHeight {
			displayPath := repo.Path
			if home != "" {
				if rel, err := filepath.Rel(home, repo.Path); err == nil && !strings.HasPrefix(rel, "..") {
					displayPath = "~/" + rel
				}
			}
			// Indent to align with name (cursor 2 + prefix 2-4)
			indent := "      "
			if repo.Selected {
				indent = "    "
			}
			pathLine := indent + pathStyle.Render(displayPath)
			pathLineWidth := lipgloss.Width(pathLine)
			if pathLineWidth > contentWidth {
				pathLine = pathLine[:contentWidth-3] + "..."
			}
			lines = append(lines, pathLine)
			lineCount++
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  m.reposViewport.YOffset(),
		TotalItems: len(m.repos),
	}
}

func (m reposModel) renderFooter() string {
	help := components.NewThemedHelp()

	var bindings []key.Binding
	if m.drawerOpen {
		bindings = []key.Binding{
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("j/k", "scroll")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "close")),
		}
	} else {
		selected := m.countSelected()
		if selected > 0 {
			bindings = []key.Binding{
				key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "install")),
				key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "remove")),
				key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "clear")),
				key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			}
		} else {
			bindings = []key.Binding{
				key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
				key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "all")),
				key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "details")),
				key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			}
		}
	}

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(help.ShortHelpView(bindings))
}
