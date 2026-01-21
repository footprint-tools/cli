package splitpanel

import (
	"strings"

	"github.com/Skryensya/footprint/internal/ui/style"
	"github.com/charmbracelet/lipgloss"
)

// Panel represents content for one side of the split
type Panel struct {
	Lines      []string // Content lines (already scrolled/visible)
	ScrollPos  int      // Current scroll position (for scrollbar calculation)
	TotalItems int      // Total scrollable items
}

// Config holds layout configuration
type Config struct {
	SidebarWidthPercent float64 // e.g., 0.25 for 25%
	SidebarMinWidth     int     // Minimum sidebar width
	SidebarMaxWidth     int     // Maximum sidebar width
	HasDrawer           bool    // Whether drawer overlay is supported
	DrawerWidthPercent  float64 // Width of drawer when open
}

// Layout holds computed dimensions and renders the split panel
type Layout struct {
	Width        int
	Height       int
	SidebarWidth int
	ContentWidth int
	DrawerWidth  int
	FocusSidebar bool
	DrawerOpen   bool
	Colors       style.ColorConfig
	config       Config
}

// NewLayout creates a new layout with calculated widths
func NewLayout(width int, cfg Config, colors style.ColorConfig) *Layout {
	// Calculate sidebar width
	sidebarWidth := int(float64(width) * cfg.SidebarWidthPercent)
	sidebarWidth = max(sidebarWidth, cfg.SidebarMinWidth)
	sidebarWidth = min(sidebarWidth, cfg.SidebarMaxWidth)

	// Content takes the rest
	contentWidth := width - sidebarWidth

	return &Layout{
		Width:        width,
		SidebarWidth: sidebarWidth,
		ContentWidth: contentWidth,
		Colors:       colors,
		FocusSidebar: true,
		config:       cfg,
	}
}

// SetFocus sets which panel is focused
func (l *Layout) SetFocus(focusSidebar bool) {
	l.FocusSidebar = focusSidebar
}

// SetDrawerOpen sets drawer state and recalculates widths
func (l *Layout) SetDrawerOpen(open bool) {
	l.DrawerOpen = open

	if open && l.config.HasDrawer {
		l.DrawerWidth = int(float64(l.Width) * l.config.DrawerWidthPercent)
		l.ContentWidth = l.Width - l.SidebarWidth - l.DrawerWidth
	} else {
		l.DrawerWidth = 0
		l.ContentWidth = l.Width - l.SidebarWidth
	}
}

// Render renders the split panel
func (l *Layout) Render(sidebar, content Panel, height int) string {
	return l.RenderWithDrawer(sidebar, content, nil, height)
}

// RenderWithDrawer renders the split panel with optional drawer
func (l *Layout) RenderWithDrawer(sidebar, content Panel, drawer *Panel, height int) string {
	l.Height = height
	colors := l.Colors
	uiActiveColor := lipgloss.Color(colors.UIActive)
	uiDimColor := lipgloss.Color(colors.UIDim)

	// Build each panel as simple bordered box
	sidebarStr := l.buildPanel(sidebar, l.SidebarWidth, height, l.FocusSidebar, uiActiveColor, uiDimColor)

	contentFocused := !l.FocusSidebar && !l.DrawerOpen
	contentStr := l.buildPanel(content, l.ContentWidth, height, contentFocused, uiActiveColor, uiDimColor)

	// Join panels directly
	if drawer != nil && l.DrawerOpen && l.DrawerWidth > 0 {
		drawerStr := l.buildPanel(*drawer, l.DrawerWidth, height, true, uiActiveColor, uiDimColor)
		return lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, contentStr, drawerStr)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, contentStr)
}

// buildPanel creates a single panel with border and scrollbar
func (l *Layout) buildPanel(panel Panel, width, height int, focused bool, activeColor, dimColor lipgloss.Color) string {
	// Content width = panel width - border(2) - padding(2) - scrollbar(2)
	contentWidth := max(width-6, 1)

	// Visible height = panel height - border(2)
	visibleHeight := max(height-2, 1)

	// Get lines and pad/truncate to visible height
	lines := panel.Lines
	if len(lines) > visibleHeight {
		lines = lines[:visibleHeight]
	}
	for len(lines) < visibleHeight {
		lines = append(lines, "")
	}

	// Build scrollbar
	totalItems := panel.TotalItems
	if totalItems == 0 {
		totalItems = len(panel.Lines)
	}
	scrollbar := BuildScrollbar(visibleHeight, totalItems, panel.ScrollPos, activeColor, dimColor, focused)

	// Combine lines with scrollbar
	var result []string
	for i, line := range lines {
		// Truncate or pad line to content width
		lineWidth := lipgloss.Width(line)
		if lineWidth > contentWidth {
			// Truncate with ellipsis
			line = truncateString(line, contentWidth)
		} else if lineWidth < contentWidth {
			line = line + strings.Repeat(" ", contentWidth-lineWidth)
		}

		scrollChar := " "
		if i < len(scrollbar) {
			scrollChar = scrollbar[i]
		}
		result = append(result, line+" "+scrollChar)
	}

	content := strings.Join(result, "\n")

	// Border color based on focus
	borderColor := dimColor
	if focused {
		borderColor = activeColor
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	return style.Render(content)
}

// truncateString truncates a string to maxWidth, accounting for ANSI codes
func truncateString(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Simple truncation - might break ANSI codes but works for most cases
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i])
		if lipgloss.Width(candidate) <= maxWidth-3 {
			return candidate + "..."
		}
	}
	return "..."
}

// SidebarContentWidth returns usable width for sidebar content
func (l *Layout) SidebarContentWidth() int {
	return l.SidebarWidth - 6 // border(2) + padding(2) + scrollbar(2)
}

// MainContentWidth returns usable width for main content
func (l *Layout) MainContentWidth() int {
	return l.ContentWidth - 6
}

// DrawerContentWidth returns usable width for drawer content
func (l *Layout) DrawerContentWidth() int {
	if l.DrawerWidth == 0 {
		return 0
	}
	return l.DrawerWidth - 6
}

// VisibleHeight returns visible lines in a panel
func (l *Layout) VisibleHeight() int {
	return l.Height - 2 // inner border(2)
}
