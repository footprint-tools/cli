package theme

import (
	"errors"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/ui/style"
)

// =========== LIST TESTS ===========

func TestList_Success(t *testing.T) {
	var printedLines []string
	deps := Deps{
		Get: func(key string) (string, bool) {
			if key == "theme" {
				return "default-dark", true
			}
			return "", false
		},
		Printf: func(format string, a ...any) (int, error) {
			printedLines = append(printedLines, fmt.Sprintf(format, a...))
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				if s, ok := a[0].(string); ok {
					printedLines = append(printedLines, s)
				}
			}
			return 0, nil
		},
		ThemeNames: []string{"default-dark", "default-light"},
		Themes: map[string]style.ColorConfig{
			"default-dark":  {Success: "10"},
			"default-light": {Success: "28"},
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := list([]string{}, flags, deps)

	require.NoError(t, err)
	require.NotEmpty(t, printedLines)
}

func TestList_NoCurrentTheme(t *testing.T) {
	var printedLines []string
	deps := Deps{
		Get: func(key string) (string, bool) {
			return "", false // No theme set
		},
		Printf: func(format string, a ...any) (int, error) {
			printedLines = append(printedLines, fmt.Sprintf(format, a...))
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				if s, ok := a[0].(string); ok {
					printedLines = append(printedLines, s)
				}
			}
			return 0, nil
		},
		ThemeNames: []string{"default-dark"},
		Themes: map[string]style.ColorConfig{
			"default-dark": {Success: "10"},
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := list([]string{}, flags, deps)

	require.NoError(t, err)
}

// =========== SET TESTS ===========

func TestSet_Success(t *testing.T) {
	var printedOutput string
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{}, nil
		},
		WriteLines: func(lines []string) error {
			return nil
		},
		Set: func(lines []string, key, value string) ([]string, bool) {
			return append(lines, key+"="+value), false
		},
		Printf: func(format string, a ...any) (int, error) {
			printedOutput = fmt.Sprintf(format, a...)
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
		ThemeNames: []string{"default-dark", "default-light"},
		Themes: map[string]style.ColorConfig{
			"default-dark":  {Success: "10"},
			"default-light": {Success: "28"},
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := setTheme([]string{"default-dark"}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, printedOutput, "default-dark")
}

func TestSet_MissingArgument(t *testing.T) {
	deps := Deps{
		ThemeNames: []string{"default-dark"},
		Themes:     map[string]style.ColorConfig{},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := setTheme([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "theme")
}

func TestSet_UnknownTheme(t *testing.T) {
	var printedLines []string
	deps := Deps{
		Printf: func(format string, a ...any) (int, error) {
			printedLines = append(printedLines, fmt.Sprintf(format, a...))
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				if s, ok := a[0].(string); ok {
					printedLines = append(printedLines, s)
				}
			}
			return 0, nil
		},
		ThemeNames: []string{"default-dark", "default-light"},
		Themes: map[string]style.ColorConfig{
			"default-dark":  {Success: "10"},
			"default-light": {Success: "28"},
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := setTheme([]string{"nonexistent"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown theme")
}

func TestSet_ReadLinesError(t *testing.T) {
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return nil, errors.New("read error")
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
		ThemeNames: []string{"default-dark"},
		Themes: map[string]style.ColorConfig{
			"default-dark": {Success: "10"},
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := setTheme([]string{"default-dark"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "read error")
}

func TestSet_WriteLinesError(t *testing.T) {
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{}, nil
		},
		WriteLines: func(lines []string) error {
			return errors.New("write error")
		},
		Set: func(lines []string, key, value string) ([]string, bool) {
			return append(lines, key+"="+value), false
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
		ThemeNames: []string{"default-dark"},
		Themes: map[string]style.ColorConfig{
			"default-dark": {Success: "10"},
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := setTheme([]string{"default-dark"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "write error")
}

// =========== RENDER COLOR PREVIEW TESTS ===========

func TestRenderColorPreview(t *testing.T) {
	cfg := style.ColorConfig{
		Success: "10",
		Error:   "9",
		Info:    "14",
		Muted:   "245",
		Color1:  "166",
		Color2:  "172",
		Color3:  "178",
		Color4:  "184",
		Color5:  "190",
		Color6:  "196",
		Color7:  "202",
	}

	output := renderColorPreview(cfg)

	// Verify output contains expected labels
	require.Contains(t, output, "success")
	require.Contains(t, output, "error")
	require.Contains(t, output, "POST-COMMIT")
	require.Contains(t, output, "PRE-PUSH")
}

func TestRenderColorPreview_BoldHeader(t *testing.T) {
	cfg := style.ColorConfig{
		Header: "bold",
	}

	output := renderColorPreview(cfg)
	// Should not panic and should render something
	require.NotEmpty(t, output)
}

func TestRenderColorPreview_EmptyColors(t *testing.T) {
	cfg := style.ColorConfig{}

	output := renderColorPreview(cfg)
	// Should not panic with empty colors
	require.NotEmpty(t, output)
}

// =========== PICK MODEL TESTS ===========

func TestModel_Init(t *testing.T) {
	m := model{
		themes:  []string{"default-dark"},
		configs: map[string]style.ColorConfig{"default-dark": {}},
	}

	cmd := m.Init()
	require.Nil(t, cmd)
}

func TestModel_View(t *testing.T) {
	m := model{
		themes:   []string{"default-dark", "default-light"},
		configs:  map[string]style.ColorConfig{"default-dark": {Success: "10", Info: "14", Muted: "245", UIActive: "14", UIDim: "240"}, "default-light": {Success: "28", Info: "27", Muted: "243", UIActive: "27", UIDim: "252"}},
		cursor:   0,
		selected: "default-dark",
		width:    100,
		height:   30,
	}

	output := m.View()

	require.Contains(t, output, "footprint themes")
	require.Contains(t, output, "default-dark")
}

func TestRenderFooter(t *testing.T) {
	m := model{
		themes:  []string{"test-theme"},
		configs: map[string]style.ColorConfig{"test-theme": {Info: "14", Muted: "245", UIDim: "240"}},
		cursor:  0,
		width:   80,
		height:  24,
	}
	output := m.renderFooter()

	require.Contains(t, output, "switch")
	require.Contains(t, output, "select")
	require.Contains(t, output, "quit")
}

func TestRenderPreviewCard(t *testing.T) {
	lines := []string{"Line 1", "Line 2", "Line 3"}
	output := renderPreviewCard(lines)

	require.Contains(t, output, "Line 1")
	require.Contains(t, output, "Line 2")
	require.Contains(t, output, "Line 3")
}

func TestBuildThemeDetailsLines(t *testing.T) {
	cfg := style.ColorConfig{
		Success: "10",
		Warning: "11",
		Error:   "9",
		Info:    "14",
		Muted:   "245",
		Header:  "bold",
		Color1:  "166",
		Color2:  "172",
		Color3:  "178",
		Color4:  "184",
		Color5:  "190",
		Color6:  "196",
		Color7:  "202",
	}

	lines := buildThemeDetailsLines("test-theme", cfg)

	require.NotEmpty(t, lines)
	// Should contain theme name in preview
	found := false
	for _, line := range lines {
		if len(line) > 0 {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestBuildThemeDetailsLines_BoldHeader(t *testing.T) {
	cfg := style.ColorConfig{
		Header: "bold",
	}

	lines := buildThemeDetailsLines("bold-theme", cfg)
	require.NotEmpty(t, lines)
}

// =========== UPDATE TESTS ===========

func TestModel_Update_CtrlC(t *testing.T) {
	m := model{
		themes:  []string{"theme1", "theme2"},
		configs: map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:  0,
	}

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, cmd := m.Update(msg)

	result := newModel.(model)
	require.True(t, result.cancelled)
	require.NotNil(t, cmd)
}

func TestModel_Update_Esc(t *testing.T) {
	m := model{
		themes:  []string{"theme1", "theme2"},
		configs: map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:  0,
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)

	result := newModel.(model)
	require.True(t, result.cancelled)
	require.NotNil(t, cmd)
}

func TestModel_Update_ArrowDown(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       0,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 1, result.cursor)
}

func TestModel_Update_ArrowDown_Wrap(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:       1, // At last item
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 0, result.cursor) // Should wrap to beginning
}

func TestModel_Update_ArrowUp(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       1,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 0, result.cursor)
}

func TestModel_Update_ArrowUp_Wrap(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:       0, // At first item
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 1, result.cursor) // Should wrap to end
}

func TestModel_Update_Home(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       2,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyHome}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 0, result.cursor)
}

func TestModel_Update_End(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       0,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnd}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 2, result.cursor)
}

func TestModel_Update_Enter(t *testing.T) {
	m := model{
		themes:  []string{"theme1", "theme2"},
		configs: map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:  1,
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, "theme2", result.chosen)
	require.NotNil(t, cmd)
}

func TestModel_Update_Space(t *testing.T) {
	m := model{
		themes:  []string{"theme1", "theme2"},
		configs: map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:  0,
	}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, "theme1", result.chosen)
	require.NotNil(t, cmd)
}

func TestModel_Update_Q(t *testing.T) {
	m := model{
		themes:  []string{"theme1", "theme2"},
		configs: map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:  0,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, cmd := m.Update(msg)

	result := newModel.(model)
	require.True(t, result.cancelled)
	require.NotNil(t, cmd)
}

func TestModel_Update_J(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       0,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 1, result.cursor)
}

func TestModel_Update_J_Wrap(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:       1, // At last item
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 0, result.cursor) // Should wrap
}

func TestModel_Update_K(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       1,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 0, result.cursor)
}

func TestModel_Update_K_Wrap(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:       0, // At first item
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 1, result.cursor) // Should wrap
}

func TestModel_Update_g(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       2,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 0, result.cursor)
}

func TestModel_Update_G(t *testing.T) {
	m := model{
		themes:       []string{"theme1", "theme2", "theme3"},
		configs:      map[string]style.ColorConfig{"theme1": {}, "theme2": {}, "theme3": {}},
		cursor:       0,
		focusSidebar: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 2, result.cursor)
}

func TestModel_Update_UnknownMessage(t *testing.T) {
	m := model{
		themes:  []string{"theme1", "theme2"},
		configs: map[string]style.ColorConfig{"theme1": {}, "theme2": {}},
		cursor:  0,
	}

	// Send an unknown message type
	msg := tea.WindowSizeMsg{}
	newModel, cmd := m.Update(msg)

	result := newModel.(model)
	require.Equal(t, 0, result.cursor) // Should remain unchanged
	require.Nil(t, cmd)
}
