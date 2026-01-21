package help

import (
	"testing"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/help"
	"github.com/Skryensya/footprint/internal/ui/splitpanel"
	"github.com/Skryensya/footprint/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func mockAction(_ []string, _ *dispatchers.ParsedFlags) error {
	return nil
}

func createTestTree() *dispatchers.DispatchNode {
	return &dispatchers.DispatchNode{
		Name: "fp",
		Path: []string{"fp"},
		Children: map[string]*dispatchers.DispatchNode{
			"track": {
				Name:     "track",
				Path:     []string{"fp", "track"},
				Summary:  "Track a repository",
				Action:   mockAction,
				Category: dispatchers.CategoryGetStarted,
			},
			"status": {
				Name:     "status",
				Path:     []string{"fp", "status"},
				Summary:  "Show tracking status",
				Action:   mockAction,
				Category: dispatchers.CategoryInspectActivity,
			},
			"config": {
				Name: "config",
				Path: []string{"fp", "config"},
				Children: map[string]*dispatchers.DispatchNode{
					"list": {
						Name:     "list",
						Path:     []string{"fp", "config", "list"},
						Summary:  "List configuration",
						Action:   mockAction,
						Category: dispatchers.CategoryConfig,
					},
					"get": {
						Name:     "get",
						Path:     []string{"fp", "config", "get"},
						Summary:  "Get configuration value",
						Action:   mockAction,
						Category: dispatchers.CategoryConfig,
					},
				},
			},
		},
	}
}

func TestBuildSidebarItems_GroupsByCategory(t *testing.T) {
	root := createTestTree()
	topics := []*help.Topic{}

	items := buildSidebarItems(root, topics)

	require.NotEmpty(t, items)

	// Verify we have categories and commands
	var categories, commands int
	for _, item := range items {
		if item.IsCategory {
			categories++
		} else {
			commands++
		}
	}

	require.Greater(t, categories, 0, "Should have category headers")
	require.Greater(t, commands, 0, "Should have commands")
}

func TestBuildSidebarItems_IncludesTopics(t *testing.T) {
	root := &dispatchers.DispatchNode{
		Name:     "fp",
		Path:     []string{"fp"},
		Children: map[string]*dispatchers.DispatchNode{},
	}

	topics := []*help.Topic{
		{Name: "overview", Summary: "Overview of footprint"},
		{Name: "workflow", Summary: "Workflow guide"},
	}

	items := buildSidebarItems(root, topics)

	// Find the conceptual guides category
	var foundCategory bool
	var topicCount int
	for _, item := range items {
		if item.IsCategory && item.Name == "conceptual" {
			foundCategory = true
		}
		if item.IsTopic {
			topicCount++
		}
	}

	require.True(t, foundCategory, "Should have conceptual guides category")
	require.Equal(t, 2, topicCount, "Should have 2 topics")
}

func TestBuildSidebarItems_EmptyRoot(t *testing.T) {
	root := &dispatchers.DispatchNode{
		Name:     "fp",
		Path:     []string{"fp"},
		Children: map[string]*dispatchers.DispatchNode{},
	}

	items := buildSidebarItems(root, nil)

	require.Empty(t, items)
}

func TestCollectLeafCommands_SimpleTree(t *testing.T) {
	node := &dispatchers.DispatchNode{
		Name:   "track",
		Path:   []string{"fp", "track"},
		Action: mockAction,
	}

	var leaves []*dispatchers.DispatchNode
	collectLeafCommands(node, &leaves)

	require.Len(t, leaves, 1)
	require.Equal(t, "track", leaves[0].Name)
}

func TestCollectLeafCommands_NestedTree(t *testing.T) {
	node := &dispatchers.DispatchNode{
		Name: "config",
		Path: []string{"fp", "config"},
		Children: map[string]*dispatchers.DispatchNode{
			"list": {
				Name:   "list",
				Path:   []string{"fp", "config", "list"},
				Action: mockAction,
			},
			"get": {
				Name:   "get",
				Path:   []string{"fp", "config", "get"},
				Action: mockAction,
			},
		},
	}

	var leaves []*dispatchers.DispatchNode
	collectLeafCommands(node, &leaves)

	require.Len(t, leaves, 2)
}

func TestCollectLeafCommands_DeepNesting(t *testing.T) {
	node := &dispatchers.DispatchNode{
		Name: "level1",
		Path: []string{"fp", "level1"},
		Children: map[string]*dispatchers.DispatchNode{
			"level2": {
				Name: "level2",
				Path: []string{"fp", "level1", "level2"},
				Children: map[string]*dispatchers.DispatchNode{
					"leaf": {
						Name:   "leaf",
						Path:   []string{"fp", "level1", "level2", "leaf"},
						Action: mockAction,
					},
				},
			},
		},
	}

	var leaves []*dispatchers.DispatchNode
	collectLeafCommands(node, &leaves)

	require.Len(t, leaves, 1)
	require.Equal(t, "leaf", leaves[0].Name)
}

func TestWrapText_ShortText(t *testing.T) {
	text := "Hello world"
	result := wrapText(text, 80)

	require.Equal(t, "Hello world", result)
}

func TestWrapText_LongText(t *testing.T) {
	text := "This is a very long line of text that should be wrapped at the specified width"
	result := wrapText(text, 40)

	require.Contains(t, result, "\n")
}

func TestWrapText_MultipleLines(t *testing.T) {
	text := "Line one\nLine two"
	result := wrapText(text, 80)

	require.Contains(t, result, "Line one")
	require.Contains(t, result, "Line two")
}

func TestWrapText_ZeroWidth(t *testing.T) {
	text := "Test"
	result := wrapText(text, 0)

	require.NotEmpty(t, result)
}

func TestWrapText_NegativeWidth(t *testing.T) {
	text := "Test"
	result := wrapText(text, -10)

	require.NotEmpty(t, result)
}

func TestWrapText_WordBoundaries(t *testing.T) {
	text := "word1 word2 word3 word4"
	result := wrapText(text, 12)

	// Should not break words in the middle
	require.NotContains(t, result, "wor\n")
}

func TestModel_Init(t *testing.T) {
	m := model{}
	cmd := m.Init()

	require.Nil(t, cmd)
}

func createTestModel() model {
	items := []sidebarItem{
		{Name: "category1", DisplayName: "CATEGORY 1", IsCategory: true},
		{Name: "cmd1", DisplayName: "cmd1", IsCategory: false, Node: &dispatchers.DispatchNode{
			Name:    "cmd1",
			Path:    []string{"fp", "cmd1"},
			Summary: "Command 1",
		}},
		{Name: "cmd2", DisplayName: "cmd2", IsCategory: false, Node: &dispatchers.DispatchNode{
			Name:    "cmd2",
			Path:    []string{"fp", "cmd2"},
			Summary: "Command 2",
		}},
		{Name: "category2", DisplayName: "CATEGORY 2", IsCategory: true},
		{Name: "cmd3", DisplayName: "cmd3", IsCategory: false, Node: &dispatchers.DispatchNode{
			Name:    "cmd3",
			Path:    []string{"fp", "cmd3"},
			Summary: "Command 3",
		}},
	}

	return model{
		allItems:      items,
		items:         items,
		cursor:        1, // Start on first command (skip category)
		colors:        style.GetColors(),
		width:         100,
		height:        30,
		focusSidebar:  true,
		totalCommands: 3,
	}
}

func TestModel_MoveCursor_Down(t *testing.T) {
	m := createTestModel()

	m.moveCursor(1)

	require.Equal(t, 2, m.cursor, "Cursor should move to cmd2")
}

func TestModel_MoveCursor_SkipsCategories(t *testing.T) {
	m := createTestModel()
	m.cursor = 2 // On cmd2

	m.moveCursor(1) // Move down

	require.Equal(t, 4, m.cursor, "Should skip category and land on cmd3")
}

func TestModel_MoveCursor_StopsAtEnd(t *testing.T) {
	m := createTestModel()
	m.cursor = 4 // On last command

	m.moveCursor(1)

	require.Equal(t, 4, m.cursor, "Should stay at last command (no wrap)")
}

func TestModel_MoveCursor_StopsAtStart(t *testing.T) {
	m := createTestModel()
	m.cursor = 1 // On first command

	m.moveCursor(-1)

	require.Equal(t, 1, m.cursor, "Should stay at first command (no wrap)")
}

func TestModel_JumpToFirst(t *testing.T) {
	m := createTestModel()
	m.cursor = 4 // Start on last command

	m.jumpToFirst()

	require.Equal(t, 1, m.cursor, "Should jump to first selectable item")
}

func TestModel_JumpToLast(t *testing.T) {
	m := createTestModel()
	m.cursor = 1 // Start on first command

	m.jumpToLast()

	require.Equal(t, 4, m.cursor, "Should jump to last selectable item")
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := createTestModel()

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	result := newModel.(model)

	require.Equal(t, 120, result.width)
	require.Equal(t, 40, result.height)
}

func TestModel_Update_KeyDown(t *testing.T) {
	m := createTestModel()

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := newModel.(model)

	// After moving down from cursor=1 (cmd1), should be at cursor=2 (cmd2)
	require.Equal(t, 2, result.cursor)
	require.Equal(t, 0, result.contentScroll, "Content scroll should reset on navigation")
}

func TestModel_Update_KeyUp(t *testing.T) {
	m := createTestModel()
	m.cursor = 2 // Start on cmd2

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	result := newModel.(model)

	// After moving up from cursor=2 (cmd2), should be at cursor=1 (cmd1)
	require.Equal(t, 1, result.cursor)
}

func TestModel_Update_KeyQ_Quits(t *testing.T) {
	m := createTestModel()

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	result := newModel.(model)

	require.True(t, result.cancelled)
	require.NotNil(t, cmd)
}

func TestModel_Update_KeyEsc_Quits(t *testing.T) {
	m := createTestModel()

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := newModel.(model)

	require.True(t, result.cancelled)
	require.NotNil(t, cmd)
}

func TestModel_Update_KeyCtrlC_Quits(t *testing.T) {
	m := createTestModel()

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := newModel.(model)

	require.True(t, result.cancelled)
	require.NotNil(t, cmd)
}

func TestModel_Update_KeyJ_MovesDown(t *testing.T) {
	m := createTestModel()
	initialCursor := m.cursor

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	result := newModel.(model)

	require.Greater(t, result.cursor, initialCursor)
}

func TestModel_Update_KeyK_MovesUp(t *testing.T) {
	m := createTestModel()
	m.cursor = 2

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	result := newModel.(model)

	require.Equal(t, 1, result.cursor)
}

func TestModel_Update_KeyG_JumpsToFirst(t *testing.T) {
	m := createTestModel()
	m.cursor = 4

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	result := newModel.(model)

	require.Equal(t, 1, result.cursor)
}

func TestModel_Update_KeyShiftG_JumpsToLast(t *testing.T) {
	m := createTestModel()
	m.cursor = 1

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	result := newModel.(model)

	require.Equal(t, 4, result.cursor)
}

func TestModel_Update_PageDown_ScrollsContent(t *testing.T) {
	m := createTestModel()
	m.contentScroll = 0
	m.focusSidebar = false // Content scroll only when content is focused

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	result := newModel.(model)

	require.Equal(t, 5, result.contentScroll)
}

func TestModel_Update_PageUp_ScrollsContent(t *testing.T) {
	m := createTestModel()
	m.contentScroll = 20
	m.focusSidebar = false // Content scroll only when content is focused

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	result := newModel.(model)

	require.Equal(t, 15, result.contentScroll)
}

func TestModel_Update_PageUp_StopsAtZero(t *testing.T) {
	m := createTestModel()
	m.contentScroll = 3
	m.focusSidebar = false // Content scroll only when content is focused

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	result := newModel.(model)

	require.Equal(t, 0, result.contentScroll)
}

func TestModel_Update_KeyU_ScrollsUp(t *testing.T) {
	m := createTestModel()
	m.contentScroll = 20

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	result := newModel.(model)

	require.Equal(t, 15, result.contentScroll)
}

func TestModel_Update_KeyD_ScrollsDown(t *testing.T) {
	m := createTestModel()
	m.contentScroll = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	result := newModel.(model)

	require.Equal(t, 5, result.contentScroll)
}

func TestModel_Update_Home_JumpsToFirst(t *testing.T) {
	m := createTestModel()
	m.cursor = 4

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	result := newModel.(model)

	require.Equal(t, 1, result.cursor)
}

func TestModel_Update_End_JumpsToLast(t *testing.T) {
	m := createTestModel()
	m.cursor = 1

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	result := newModel.(model)

	require.Equal(t, 4, result.cursor)
}

func TestModel_Update_Navigation_ResetsContentScroll(t *testing.T) {
	m := createTestModel()
	m.contentScroll = 20

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := newModel.(model)

	require.Equal(t, 0, result.contentScroll)
}

func TestModel_View_ReturnsString(t *testing.T) {
	m := createTestModel()

	view := m.View()

	require.NotEmpty(t, view)
}

func TestModel_View_DefaultDimensions(t *testing.T) {
	m := createTestModel()
	m.width = 0
	m.height = 0

	view := m.View()

	require.NotEmpty(t, view, "Should use default dimensions when zero")
}

func TestModel_RenderCommandContent(t *testing.T) {
	m := createTestModel()

	node := &dispatchers.DispatchNode{
		Name:        "track",
		Path:        []string{"fp", "track"},
		Summary:     "Track a repository",
		Description: "Longer description of tracking",
		Usage:       "fp track [path]",
		Flags: []dispatchers.FlagDescriptor{
			{Names: []string{"--help", "-h"}, Description: "Show help"},
		},
		Args: []dispatchers.ArgSpec{
			{Name: "path", Description: "Repository path", Required: false},
		},
	}

	content := m.renderCommandContent(node, 80)

	require.Contains(t, content, "track")
	require.Contains(t, content, "Track a repository")
	require.Contains(t, content, "USAGE")
	require.Contains(t, content, "DESCRIPTION")
	require.Contains(t, content, "FLAGS")
	require.Contains(t, content, "ARGUMENTS")
}

func TestModel_RenderCommandContent_MinimalNode(t *testing.T) {
	m := createTestModel()

	node := &dispatchers.DispatchNode{
		Name:  "cmd",
		Path:  []string{"fp", "cmd"},
		Usage: "fp cmd",
	}

	content := m.renderCommandContent(node, 80)

	require.Contains(t, content, "cmd")
	require.Contains(t, content, "USAGE")
}

func TestModel_RenderTopicContent(t *testing.T) {
	m := createTestModel()

	topic := &help.Topic{
		Name:    "overview",
		Summary: "Overview of footprint",
	}

	content := m.renderTopicContent(topic, 80)

	require.Contains(t, content, "overview")
	require.Contains(t, content, "Overview of footprint")
}

func TestSetBuildTreeFunc(t *testing.T) {
	// Test that SetBuildTreeFunc sets the function
	called := false
	SetBuildTreeFunc(func() *dispatchers.DispatchNode {
		called = true
		return nil
	})

	deps := DefaultDeps()
	deps.BuildTree()

	require.True(t, called)
}

func TestDefaultDeps_AllTopics(t *testing.T) {
	deps := DefaultDeps()

	topics := deps.AllTopics()

	require.NotNil(t, topics)
}

func TestBuildScrollbar_AllFit(t *testing.T) {
	scrollbar := splitpanel.BuildScrollbar(10, 5, 0, lipgloss.Color("14"), lipgloss.Color("238"), true)

	require.Len(t, scrollbar, 10)
	// All items should show blank space since everything fits
	for _, s := range scrollbar {
		require.NotEmpty(t, s)
	}
}

func TestBuildScrollbar_NeedsScroll(t *testing.T) {
	scrollbar := splitpanel.BuildScrollbar(10, 30, 0, lipgloss.Color("14"), lipgloss.Color("238"), true)

	require.Len(t, scrollbar, 10)
	// Should have some thumb indicators
	for _, s := range scrollbar {
		require.NotEmpty(t, s)
	}
}

func TestBuildScrollbar_ScrolledDown(t *testing.T) {
	scrollbar := splitpanel.BuildScrollbar(10, 30, 15, lipgloss.Color("14"), lipgloss.Color("238"), true)

	require.Len(t, scrollbar, 10)
	// Should have thumb in middle/bottom area
	for _, s := range scrollbar {
		require.NotEmpty(t, s)
	}
}

func TestModel_BuildSidebarPanel_WithTopics(t *testing.T) {
	items := []sidebarItem{
		{Name: "category1", DisplayName: "CATEGORY 1", IsCategory: true},
		{Name: "cmd1", DisplayName: "cmd1", IsCategory: false, Node: &dispatchers.DispatchNode{
			Name: "cmd1",
			Path: []string{"fp", "cmd1"},
		}},
		{Name: "conceptual", DisplayName: "CONCEPTUAL GUIDES", IsCategory: true},
		{Name: "topic1", DisplayName: "overview", IsCategory: false, IsTopic: true, Topic: &help.Topic{
			Name:    "overview",
			Summary: "Overview of footprint",
		}},
	}

	m := model{
		allItems:      items,
		items:         items,
		cursor:        1,
		colors:        style.GetColors(),
		width:         80,
		height:        24,
		focusSidebar:  true,
		totalCommands: 2,
	}

	cfg := splitpanel.Config{
		SidebarWidthPercent: 0.25,
		SidebarMinWidth:     24,
		SidebarMaxWidth:     36,
	}
	layout := splitpanel.NewLayout(80, cfg, m.colors)
	panel := m.buildSidebarPanel(layout, 20)

	require.NotEmpty(t, panel.Lines)
}

func TestModel_BuildContentPanel_WithScrollbar(t *testing.T) {
	m := createTestModel()
	m.width = 100
	m.height = 20

	cfg := splitpanel.Config{
		SidebarWidthPercent: 0.25,
		SidebarMinWidth:     24,
		SidebarMaxWidth:     36,
	}
	layout := splitpanel.NewLayout(100, cfg, m.colors)
	panel := m.buildContentPanel(layout, 18)

	require.NotNil(t, panel)
}
