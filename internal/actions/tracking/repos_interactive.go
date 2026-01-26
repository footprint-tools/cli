package tracking

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/git"
	"github.com/footprint-tools/footprint-cli/internal/hooks"
	"github.com/footprint-tools/footprint-cli/internal/store"
	"github.com/footprint-tools/footprint-cli/internal/ui/splitpanel"
	"github.com/footprint-tools/footprint-cli/internal/ui/style"
	"golang.org/x/term"
)

// RepoEntry represents a discovered git repository.
type RepoEntry struct {
	Path         string            // Absolute path to the repo
	Name         string            // Directory name
	HasHooks     bool              // Whether fp hooks are installed
	HooksChanged bool              // Whether hooks state was changed this session
	Inspection   hooks.RepoInspection // Preflight inspection result
}

// ReposInteractive launches the interactive repository manager.
func ReposInteractive(_ []string, flags *dispatchers.ParsedFlags) error {
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
	fmt.Printf("Scanning for git repositories in %s...\n", root)
	repos, err := scanForRepos(root, maxDepth)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("No git repositories found")
		return nil
	}

	fmt.Printf("Found %d repositories\n", len(repos))

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
		fmt.Println()
		if fm.installed > 0 {
			fmt.Printf("Installed hooks in %d repositories\n", fm.installed)
		}
		if fm.uninstalled > 0 {
			fmt.Printf("Removed hooks from %d repositories\n", fm.uninstalled)
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
		"node_modules": true,
		"vendor":       true,
		"bower_components": true,
		"jspm_packages": true,
		".pnpm":        true,
		// Python
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
		"env":          true,
		".eggs":        true,
		"site-packages": true,
		// Build outputs
		"dist":         true,
		"build":        true,
		"target":       true,
		"out":          true,
		"_build":       true,
		// IDE / tools
		".idea":        true,
		".vscode":      true,
		// System / caches
		".cache":       true,
		".npm":         true,
		".yarn":        true,
		// macOS
		"Library":      true,
		".Trash":       true,
		"Applications": true,
		// Git internals (don't descend)
		".git":         true,
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
	scroll         int
	width          int
	height         int
	installed      int
	uninstalled    int
	message        string
	colors         style.ColorConfig
	focusSidebar   bool
	showingGuidance bool // Whether to show the guidance overlay
}

func newReposModel(repos []RepoEntry) reposModel {
	return reposModel{
		repos:        repos,
		cursor:       0,
		colors:       style.GetColors(),
		focusSidebar: false, // Start with focus on repo list
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
		// If showing guidance, any key closes it
		if m.showingGuidance {
			m.showingGuidance = false
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
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
			if !m.focusSidebar {
				m.toggleHooks()
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
				if !m.focusSidebar && !m.repos[m.cursor].HasHooks {
					m.toggleHooks()
				}
			case "u":
				if !m.focusSidebar && m.repos[m.cursor].HasHooks {
					m.toggleHooks()
				}
			case "a":
				m.installAll()
			case "A":
				m.uninstallAll()
			case "?":
				// Show guidance for non-clean repos
				if len(m.repos) > 0 && !m.repos[m.cursor].Inspection.Status.CanInstall() {
					m.showingGuidance = true
				}
			}
		}
	}

	return m, nil
}

func (m *reposModel) toggleHooks() {
	if len(m.repos) == 0 {
		return
	}

	repo := &m.repos[m.cursor]

	// If already installed, allow uninstall
	if repo.HasHooks {
		hooksPath, err := git.RepoHooksPath(repo.Path)
		if err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
			return
		}
		if err := hooks.Uninstall(hooksPath); err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
			return
		}
		repo.HasHooks = false
		repo.HooksChanged = true
		m.uninstalled++
		m.message = fmt.Sprintf("Removed hooks from %s", repo.Name)
		removeRepoFromStore(repo.Path)
		return
	}

	// Check if installation is allowed
	if !repo.Inspection.Status.CanInstall() {
		m.message = fmt.Sprintf("Cannot install: %s", repo.Inspection.Status)
		return
	}

	hooksPath, err := git.RepoHooksPath(repo.Path)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return
	}

	if err := hooks.Install(hooksPath); err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return
	}
	repo.HasHooks = true
	repo.HooksChanged = true
	repo.Inspection.FpInstalled = true
	m.installed++
	m.message = fmt.Sprintf("Installed hooks in %s", repo.Name)

	// Track the repo in the store
	addRepoToStore(repo.Path)
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

func (m *reposModel) installAll() {
	count := 0
	skipped := 0
	for i := range m.repos {
		r := &m.repos[i]
		if r.HasHooks {
			continue
		}
		// Only install in clean repos
		if !r.Inspection.Status.CanInstall() {
			skipped++
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
		m.installed++
		count++
		addRepoToStore(r.Path)
	}
	if skipped > 0 {
		m.message = fmt.Sprintf("Installed in %d repos, %d skipped (not clean)", count, skipped)
	} else {
		m.message = fmt.Sprintf("Installed hooks in %d repositories", count)
	}
}

func (m *reposModel) uninstallAll() {
	count := 0
	for i := range m.repos {
		r := &m.repos[i]
		if r.HasHooks {
			hooksPath, err := git.RepoHooksPath(r.Path)
			if err != nil {
				continue
			}
			if err := hooks.Uninstall(hooksPath); err != nil {
				continue
			}
			r.HasHooks = false
			r.HooksChanged = true
			m.uninstalled++
			count++
			removeRepoFromStore(r.Path)
		}
	}
	m.message = fmt.Sprintf("Removed hooks from %d repositories", count)
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

	// Show guidance overlay if active
	if m.showingGuidance && len(m.repos) > 0 {
		return m.renderGuidanceOverlay()
	}

	// Calculate dimensions
	headerHeight := 3
	footerHeight := 2
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 1 {
		mainHeight = 1
	}

	// Create layout
	cfg := splitpanel.Config{
		SidebarWidthPercent: 0.22,
		SidebarMinWidth:     18,
		SidebarMaxWidth:     24,
	}
	layout := splitpanel.NewLayout(m.width, cfg, m.colors)
	layout.SetFocus(m.focusSidebar)

	// Build panels
	statsPanel := m.buildStatsPanel(layout, mainHeight)
	reposPanel := m.buildReposPanel(layout, mainHeight)

	// Render
	header := m.renderHeader()
	main := layout.Render(statsPanel, reposPanel, mainHeight)
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func (m reposModel) renderGuidanceOverlay() string {
	colors := m.colors
	headerColor := lipgloss.Color(colors.Header)
	mutedColor := lipgloss.Color(colors.Muted)
	warnColor := lipgloss.Color(colors.Warning)

	repo := m.repos[m.cursor]
	guidance := hooks.GetGuidance(repo.Inspection)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	statusStyle := lipgloss.NewStyle().Foreground(warnColor)
	contentStyle := lipgloss.NewStyle().Foreground(mutedColor)
	hintStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)

	// Build the overlay content
	var lines []string
	lines = append(lines, titleStyle.Render("Integration Guide: "+repo.Name))
	lines = append(lines, "")
	lines = append(lines, statusStyle.Render("Status: "+repo.Inspection.Status.String()))
	lines = append(lines, "")

	// Add guidance text, respecting terminal width
	maxWidth := m.width - 4
	if maxWidth < 40 {
		maxWidth = 40
	}
	for _, line := range strings.Split(guidance, "\n") {
		if len(line) > maxWidth {
			lines = append(lines, contentStyle.Render(line[:maxWidth]))
		} else {
			lines = append(lines, contentStyle.Render(line))
		}
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Render("Press any key to close"))

	content := strings.Join(lines, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors.Info)).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 2)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
}

func (m reposModel) renderHeader() string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	warnColor := lipgloss.Color(colors.Warning)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)

	title := titleStyle.Render("fp repos")
	count := mutedStyle.Render(fmt.Sprintf(" | %d repositories", len(m.repos)))

	// Message if any
	msgStr := ""
	if m.message != "" {
		msgStr = mutedStyle.Render(" | ") + warnStyle.Render(m.message)
	}

	headerContent := title + count + msgStr

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return headerStyle.Render(headerContent)
}

func (m *reposModel) buildStatsPanel(layout *splitpanel.Layout, height int) splitpanel.Panel {
	colors := m.colors
	headerColor := lipgloss.Color(colors.Header)
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	warnColor := lipgloss.Color(colors.Warning)
	errorColor := lipgloss.Color(colors.Error)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(infoColor)
	successStyle := lipgloss.NewStyle().Foreground(successColor)
	warnStyle := lipgloss.NewStyle().Foreground(warnColor)
	errorStyle := lipgloss.NewStyle().Foreground(errorColor)

	var lines []string

	// Stats header
	lines = append(lines, headerStyle.Render("SUMMARY"))
	lines = append(lines, "")

	// Counts
	withHooks := m.countWithHooks()
	clean, blocked := m.countByInstallability()

	lines = append(lines, labelStyle.Render("Total: ")+valueStyle.Render(fmt.Sprintf("%d", len(m.repos))))
	lines = append(lines, labelStyle.Render("Installed: ")+successStyle.Render(fmt.Sprintf("%d", withHooks)))
	lines = append(lines, labelStyle.Render("Clean: ")+warnStyle.Render(fmt.Sprintf("%d", clean-withHooks)))
	if blocked > 0 {
		lines = append(lines, labelStyle.Render("Blocked: ")+errorStyle.Render(fmt.Sprintf("%d", blocked)))
	}
	lines = append(lines, "")

	// Session changes
	if m.installed > 0 || m.uninstalled > 0 {
		lines = append(lines, headerStyle.Render("SESSION"))
		lines = append(lines, "")
		if m.installed > 0 {
			lines = append(lines, labelStyle.Render("Installed: ")+successStyle.Render(fmt.Sprintf("+%d", m.installed)))
		}
		if m.uninstalled > 0 {
			lines = append(lines, labelStyle.Render("Removed: ")+warnStyle.Render(fmt.Sprintf("-%d", m.uninstalled)))
		}
		lines = append(lines, "")
	}

	// Keys section
	lines = append(lines, headerStyle.Render("KEYS"))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("space ")+"toggle")
	lines = append(lines, labelStyle.Render("i ")+"install")
	lines = append(lines, labelStyle.Render("u ")+"uninstall")
	lines = append(lines, labelStyle.Render("a ")+"install all")
	lines = append(lines, labelStyle.Render("A ")+"remove all")
	lines = append(lines, labelStyle.Render("? ")+"show help")

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  0,
		TotalItems: len(lines),
	}
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
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)
	successColor := lipgloss.Color(colors.Success)
	warnColor := lipgloss.Color(colors.Warning)
	errorColor := lipgloss.Color(colors.Error)

	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(infoColor)
	installedStyle := lipgloss.NewStyle().Foreground(successColor)
	notInstalledStyle := lipgloss.NewStyle().Foreground(mutedColor)
	blockedStyle := lipgloss.NewStyle().Foreground(errorColor)
	changedStyle := lipgloss.NewStyle().Foreground(warnColor)
	pathStyle := lipgloss.NewStyle().Foreground(mutedColor)
	statusStyle := lipgloss.NewStyle().Foreground(warnColor).Italic(true)

	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Adjust scroll to keep cursor visible
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	// Each repo takes 3 lines when selected (name, path, status), 1 otherwise
	effectiveHeight := visibleHeight / 3 // Conservative estimate
	if m.cursor >= m.scroll+effectiveHeight {
		m.scroll = m.cursor - effectiveHeight + 1
	}
	if m.scroll < 0 {
		m.scroll = 0
	}

	var lines []string
	contentWidth := layout.MainContentWidth()

	home, _ := os.UserHomeDir()

	lineCount := 0
	for i := m.scroll; i < len(m.repos) && lineCount < visibleHeight; i++ {
		repo := m.repos[i]

		// Cursor indicator
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		// Status indicator based on inspection
		var status string
		canInstall := repo.Inspection.Status.CanInstall()
		if repo.HasHooks {
			status = installedStyle.Render("[✓]")
		} else if canInstall {
			status = notInstalledStyle.Render("[ ]")
		} else {
			status = blockedStyle.Render("[×]")
		}

		// Name
		name := repo.Name
		if i == m.cursor {
			name = cursorStyle.Render(name)
		}

		// Changed indicator
		changed := ""
		if repo.HooksChanged {
			changed = changedStyle.Render(" *")
		}

		line := fmt.Sprintf("%s%s %s%s", prefix, status, name, changed)

		// Truncate if needed
		if lipgloss.Width(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}

		lines = append(lines, line)
		lineCount++

		// Show details when cursor is here
		if i == m.cursor && lineCount < visibleHeight {
			// Path line
			displayPath := repo.Path
			if home != "" {
				if rel, err := filepath.Rel(home, repo.Path); err == nil && !strings.HasPrefix(rel, "..") {
					displayPath = "~/" + rel
				}
			}
			pathLine := "     " + pathStyle.Render(displayPath)
			if lipgloss.Width(pathLine) > contentWidth {
				pathLine = pathLine[:contentWidth-3] + "..."
			}
			lines = append(lines, pathLine)
			lineCount++

			// Status line for non-clean repos
			if !canInstall && lineCount < visibleHeight {
				statusLine := "     " + statusStyle.Render(repo.Inspection.Status.String())
				lines = append(lines, statusLine)
				lineCount++
			}
		}
	}

	return splitpanel.Panel{
		Lines:      lines,
		ScrollPos:  m.scroll,
		TotalItems: len(m.repos),
	}
}

func (m reposModel) renderFooter() string {
	colors := m.colors
	infoColor := lipgloss.Color(colors.Info)
	mutedColor := lipgloss.Color(colors.Muted)

	keyStyle := lipgloss.NewStyle().
		Background(infoColor).
		Foreground(lipgloss.Color("0")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	sepStyle := lipgloss.NewStyle().Foreground(mutedColor)

	sep := sepStyle.Render(" | ")

	footer := keyStyle.Render("Tab") + labelStyle.Render(" switch") + sep +
		keyStyle.Render("jk") + labelStyle.Render(" nav") + sep +
		keyStyle.Render("space") + labelStyle.Render(" toggle") + sep +
		keyStyle.Render("a") + labelStyle.Render(" all") + sep +
		keyStyle.Render("q") + labelStyle.Render(" quit")

	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1)

	return footerStyle.Render(footer)
}
