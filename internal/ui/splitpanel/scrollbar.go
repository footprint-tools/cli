package splitpanel

import (
	"github.com/charmbracelet/lipgloss"
)

// Scrollbar characters
const (
	ScrollThumbChar = "\u2588" // Full block for thumb (solid)
	ScrollTrackChar = "\u2502" // Box drawing vertical for track (hollow/border only)
)

// BuildScrollbar creates a visual scrollbar for the given parameters.
// viewHeight: the visible height of the scrollbar track
// totalItems: total number of items/lines in the content
// scrollOffset: current scroll position (0-based)
// activeColor: color for the thumb (when focused)
// trackColor: color for the track
// focused: whether this panel is focused (affects thumb brightness)
func BuildScrollbar(viewHeight, totalItems, scrollOffset int, activeColor, trackColor lipgloss.Color, focused bool) []string {
	scrollbar := make([]string, viewHeight)
	trackStyle := lipgloss.NewStyle().Foreground(trackColor)

	// If all items fit, show blank space (no scrollbar needed)
	if totalItems <= viewHeight {
		for i := range scrollbar {
			scrollbar[i] = " "
		}
		return scrollbar
	}

	// Calculate thumb size proportional to visible content
	// thumbSize = (visible / total) * trackHeight
	thumbSize := (viewHeight * viewHeight) / totalItems

	// Ensure minimum size of 1, maximum of viewHeight-2 (leave room for position indication)
	thumbSize = max(thumbSize, 1)
	maxThumbSize := max(viewHeight-2, 1)
	if thumbSize > maxThumbSize {
		thumbSize = maxThumbSize
	}

	// Calculate thumb position
	// Position is proportional to scroll offset within scrollable range
	maxScroll := max(totalItems-viewHeight, 1)

	// Available track space for thumb movement
	trackSpace := max(viewHeight-thumbSize, 0)

	// Calculate position: (scrollOffset / maxScroll) * trackSpace
	thumbPos := 0
	if maxScroll > 0 && trackSpace > 0 {
		thumbPos = (scrollOffset * trackSpace) / maxScroll
	}

	// Clamp thumb position
	thumbPos = max(thumbPos, 0)
	thumbPos = min(thumbPos, trackSpace)

	// Build scrollbar - use activeColor if focused, trackColor if not
	thumbColor := trackColor
	if focused {
		thumbColor = activeColor
	}
	thumbStyle := lipgloss.NewStyle().Foreground(thumbColor)

	for i := range viewHeight {
		if i >= thumbPos && i < thumbPos+thumbSize {
			scrollbar[i] = thumbStyle.Render(ScrollThumbChar)
		} else {
			scrollbar[i] = trackStyle.Render(ScrollTrackChar)
		}
	}

	return scrollbar
}
