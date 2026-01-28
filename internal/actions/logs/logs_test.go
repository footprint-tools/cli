package logs

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/cli/internal/dispatchers"
)

// =========== VIEW TESTS ===========

func TestView_FileNotExists(t *testing.T) {
	var printed string
	deps := Deps{
		LogFilePath: func() string { return "/tmp/nonexistent.log" },
		Stat: func(path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printed = a[0].(string)
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := view([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, printed, "No log file found")
}

func TestView_StatError(t *testing.T) {
	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		Stat: func(path string) (os.FileInfo, error) {
			return nil, errors.New("stat error")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := view([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "stat log file")
}

func TestView_EmptyFile(t *testing.T) {
	var printed string
	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		Stat: func(path string) (os.FileInfo, error) {
			return &mockFileInfo{size: 0}, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printed = a[0].(string)
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := view([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, printed, "Log file is empty")
}

func TestView_ReadFileError(t *testing.T) {
	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		Stat: func(path string) (os.FileInfo, error) {
			return &mockFileInfo{size: 100}, nil
		},
		ReadFile: func(path string) ([]byte, error) {
			return nil, errors.New("read error")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := view([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "read log file")
}

func TestView_Success(t *testing.T) {
	logContent := "line1\nline2\nline3\n"
	var printedLines []string

	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		Stat: func(path string) (os.FileInfo, error) {
			return &mockFileInfo{size: int64(len(logContent))}, nil
		},
		ReadFile: func(path string) ([]byte, error) {
			return []byte(logContent), nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printedLines = append(printedLines, a[0].(string))
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := view([]string{}, flags, deps)

	require.NoError(t, err)
	require.Len(t, printedLines, 3)
}

func TestView_WithLimit(t *testing.T) {
	logContent := "line1\nline2\nline3\nline4\nline5\n"
	var printedLines []string

	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		Stat: func(path string) (os.FileInfo, error) {
			return &mockFileInfo{size: int64(len(logContent))}, nil
		},
		ReadFile: func(path string) ([]byte, error) {
			return []byte(logContent), nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printedLines = append(printedLines, a[0].(string))
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--limit=2"})
	err := view([]string{}, flags, deps)

	require.NoError(t, err)
	require.Len(t, printedLines, 2)
	require.Equal(t, "line4", printedLines[0])
	require.Equal(t, "line5", printedLines[1])
}

func TestView_NegativeLimitDefaultsTo50(t *testing.T) {
	// Create content with 60 lines
	var lines string
	for i := 1; i <= 60; i++ {
		lines += "line\n"
	}
	var printedCount int

	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		Stat: func(path string) (os.FileInfo, error) {
			return &mockFileInfo{size: int64(len(lines))}, nil
		},
		ReadFile: func(path string) ([]byte, error) {
			return []byte(lines), nil
		},
		Println: func(a ...any) (int, error) {
			printedCount++
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--limit=-10"})
	err := view([]string{}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, 50, printedCount) // Defaults to 50 when negative
}

// =========== CLEAR TESTS ===========

func TestClear_Success(t *testing.T) {
	var writtenPath string
	var writtenContent []byte
	var printed string

	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		WriteFile: func(path string, content []byte, perm os.FileMode) error {
			writtenPath = path
			writtenContent = content
			return nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printed = a[0].(string)
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := clear([]string{}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, "/tmp/test.log", writtenPath)
	require.Empty(t, writtenContent)
	require.Contains(t, printed, "cleared")
}

func TestClear_WriteError(t *testing.T) {
	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		WriteFile: func(path string, content []byte, perm os.FileMode) error {
			return errors.New("write error")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := clear([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "clear log file")
}

// =========== COLORIZE LOG LINE TESTS ===========

func TestColorizeLogLine_Error(t *testing.T) {
	input := "[2024-01-01] ERROR: something failed"
	output := colorizeLogLine(input)
	// Output should contain the original text (colorization wraps it)
	require.Contains(t, output, "ERROR")
}

func TestColorizeLogLine_Warn(t *testing.T) {
	input := "[2024-01-01] WARN: something suspicious"
	output := colorizeLogLine(input)
	require.Contains(t, output, "WARN")
}

func TestColorizeLogLine_Info(t *testing.T) {
	input := "[2024-01-01] INFO: something happened"
	output := colorizeLogLine(input)
	require.Contains(t, output, "INFO")
}

func TestColorizeLogLine_Debug(t *testing.T) {
	input := "[2024-01-01] DEBUG: detailed info"
	output := colorizeLogLine(input)
	require.Contains(t, output, "DEBUG")
}

func TestColorizeLogLine_NoLevel(t *testing.T) {
	input := "just a plain line"
	output := colorizeLogLine(input)
	require.Equal(t, input, output)
}

// =========== TAIL TESTS ===========

func TestTail_OpenFileError(t *testing.T) {
	deps := Deps{
		LogFilePath: func() string { return "/tmp/test.log" },
		OpenFile: func(path string, flag int, perm os.FileMode) (*os.File, error) {
			return nil, errors.New("open error")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := tail([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "open log file")
}

// =========== DEFAULT DEPS TESTS ===========

func TestDefaultDeps(t *testing.T) {
	deps := DefaultDeps()

	// Verify all functions are set
	require.NotNil(t, deps.LogFilePath)
	require.NotNil(t, deps.Printf)
	require.NotNil(t, deps.Println)
	require.NotNil(t, deps.ReadFile)
	require.NotNil(t, deps.WriteFile)
	require.NotNil(t, deps.Stat)
	require.NotNil(t, deps.OpenFile)

	// Verify LogFilePath returns a path
	path := deps.LogFilePath()
	require.NotEmpty(t, path)
}

// =========== PARSE LINE TESTS ===========

func TestParseLine_NewFormat(t *testing.T) {
	// Parser splits at first colon after level, so caller is file only
	input := "[2024-01-15 10:30:45] ERROR store.go:123: something failed"
	line := parseLine(input)

	require.Equal(t, input, line.Raw)
	require.Equal(t, "2024-01-15 10:30:45", line.Timestamp)
	require.Equal(t, "ERROR", line.Level)
	require.Equal(t, "store.go", line.Caller) // First colon splits here
	require.Equal(t, "123: something failed", line.Message)
	require.False(t, line.ParsedTime.IsZero())
}

func TestParseLine_OldFormat(t *testing.T) {
	input := "[2024-01-15 10:30:45] WARN: this is a warning"
	line := parseLine(input)

	require.Equal(t, input, line.Raw)
	require.Equal(t, "2024-01-15 10:30:45", line.Timestamp)
	require.Equal(t, "WARN", line.Level)
	require.Equal(t, "", line.Caller)
	require.Equal(t, "this is a warning", line.Message)
}

func TestParseLine_AllLevels(t *testing.T) {
	tests := []struct {
		level string
	}{
		{"ERROR"},
		{"WARN"},
		{"INFO"},
		{"DEBUG"},
	}

	for _, tc := range tests {
		t.Run(tc.level, func(t *testing.T) {
			input := "[2024-01-15 10:30:45] " + tc.level + ": message"
			line := parseLine(input)
			require.Equal(t, tc.level, line.Level)
		})
	}
}

func TestParseLine_NoLevel(t *testing.T) {
	input := "[2024-01-15 10:30:45] just a plain message"
	line := parseLine(input)

	require.Equal(t, "2024-01-15 10:30:45", line.Timestamp)
	require.Equal(t, "", line.Level)
	require.Equal(t, "just a plain message", line.Message)
}

func TestParseLine_NoBracket(t *testing.T) {
	input := "some random text without timestamp"
	line := parseLine(input)

	require.Equal(t, input, line.Raw)
	require.Equal(t, "", line.Timestamp)
	require.Equal(t, "", line.Level)
}

func TestParseLine_ShortLine(t *testing.T) {
	input := "short"
	line := parseLine(input)

	require.Equal(t, input, line.Raw)
	require.Equal(t, "", line.Level)
}

// =========== WRAP TEXT TESTS ===========

func TestWrapText_NoWrapNeeded(t *testing.T) {
	input := "short text"
	output := wrapText(input, 50)
	require.Equal(t, input, output)
}

func TestWrapText_SingleWrap(t *testing.T) {
	input := "this is a longer text that needs wrapping"
	output := wrapText(input, 20)
	require.Contains(t, output, "\n")
}

func TestWrapText_ZeroWidth(t *testing.T) {
	input := "some text"
	output := wrapText(input, 0)
	require.Equal(t, input, output)
}

func TestWrapText_NegativeWidth(t *testing.T) {
	input := "some text"
	output := wrapText(input, -10)
	require.Equal(t, input, output)
}

func TestWrapText_ExactWidth(t *testing.T) {
	input := "hello"
	output := wrapText(input, 5)
	require.Equal(t, input, output)
}

// =========== MODEL TESTS ===========

func TestNewLogsModel(t *testing.T) {
	m := newLogsModel("/tmp/test.log")

	require.Equal(t, "/tmp/test.log", m.logPath)
	require.NotNil(t, m.lines)
	require.NotNil(t, m.byLevel)
	require.NotNil(t, m.byLevelTotal)
	require.True(t, m.autoScroll)
	require.False(t, m.paused)
	require.False(t, m.drawerOpen)
}

func TestLogsModel_FilteredLines_NoFilter(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.lines = []LogLine{
		{Raw: "line1", Level: "ERROR"},
		{Raw: "line2", Level: "INFO"},
		{Raw: "line3", Level: "DEBUG"},
	}

	filtered := m.filteredLines()
	require.Len(t, filtered, 3)
}

func TestLogsModel_FilteredLines_ByLevel(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.lines = []LogLine{
		{Raw: "line1", Level: "ERROR"},
		{Raw: "line2", Level: "INFO"},
		{Raw: "line3", Level: "ERROR"},
	}
	m.filterLevel = "ERROR"

	filtered := m.filteredLines()
	require.Len(t, filtered, 2)
}

func TestLogsModel_FilteredLines_ByQuery(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.lines = []LogLine{
		{Raw: "error in database"},
		{Raw: "info message"},
		{Raw: "another database error"},
	}
	m.filterQuery = "database"

	filtered := m.filteredLines()
	require.Len(t, filtered, 2)
}

func TestLogsModel_FilteredLines_ByLevelAndQuery(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.lines = []LogLine{
		{Raw: "error in database", Level: "ERROR"},
		{Raw: "info database message", Level: "INFO"},
		{Raw: "another database error", Level: "ERROR"},
	}
	m.filterLevel = "ERROR"
	m.filterQuery = "database"

	filtered := m.filteredLines()
	require.Len(t, filtered, 2)
}

func TestLogsModel_FilteredLines_CaseInsensitive(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.lines = []LogLine{
		{Raw: "ERROR in Database"},
		{Raw: "info message"},
	}
	m.filterQuery = "database"

	filtered := m.filteredLines()
	require.Len(t, filtered, 1)
}

func TestLogsModel_AddInitialLines(t *testing.T) {
	m := newLogsModel("/tmp/test.log")

	lines := []LogLine{
		{Raw: "line1", Level: "ERROR"},
		{Raw: "line2", Level: "INFO"},
	}

	m.addInitialLines(lines)

	require.Len(t, m.lines, 2)
	require.Equal(t, 2, m.totalEver)
	require.Equal(t, 0, m.sessionLines) // Initial lines don't count as session
	require.Equal(t, 1, m.byLevelTotal["ERROR"])
	require.Equal(t, 1, m.byLevelTotal["INFO"])
	require.Equal(t, 0, m.byLevel["ERROR"]) // Session counts are 0
}

func TestLogsModel_AddSessionLines(t *testing.T) {
	m := newLogsModel("/tmp/test.log")

	lines := []LogLine{
		{Raw: "line1", Level: "ERROR"},
		{Raw: "line2", Level: "INFO"},
	}

	m.addSessionLines(lines)

	require.Len(t, m.lines, 2)
	require.Equal(t, 2, m.totalEver)
	require.Equal(t, 2, m.sessionLines)
	require.Equal(t, 1, m.byLevelTotal["ERROR"])
	require.Equal(t, 1, m.byLevel["ERROR"]) // Session counts
}

func TestLogsModel_AddLines_TrimBuffer(t *testing.T) {
	m := newLogsModel("/tmp/test.log")

	// Add more than maxLogLines
	var lines []LogLine
	for i := 0; i < maxLogLines+100; i++ {
		lines = append(lines, LogLine{Raw: "line"})
	}

	m.addInitialLines(lines)

	require.LessOrEqual(t, len(m.lines), maxLogLines)
}

func TestLogsModel_MoveCursor(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.lines = []LogLine{
		{Raw: "line1"},
		{Raw: "line2"},
		{Raw: "line3"},
	}

	// Move down
	m.moveCursor(1)
	require.Equal(t, 1, m.cursor)

	// Move up
	m.moveCursor(-1)
	require.Equal(t, 0, m.cursor)

	// Move past beginning
	m.moveCursor(-10)
	require.Equal(t, 0, m.cursor)

	// Move past end
	m.moveCursor(100)
	require.Equal(t, 2, m.cursor)
}

func TestLogsModel_MoveCursor_EmptyLines(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	// No lines

	m.moveCursor(1) // Should not panic
	require.Equal(t, 0, m.cursor)
}

func TestLogsModel_SessionDuration(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.sessionStart = time.Now().Add(-5 * time.Second)

	duration := m.sessionDuration()
	require.GreaterOrEqual(t, duration.Seconds(), float64(5))
}

func TestLogsModel_CalculateStatsWidth(t *testing.T) {
	m := newLogsModel("/tmp/test.log")

	// Zero width
	m.width = 0
	require.Equal(t, 20, m.calculateStatsWidth())

	// Normal width
	m.width = 100
	width := m.calculateStatsWidth()
	require.GreaterOrEqual(t, width, 18)
	require.LessOrEqual(t, width, 24)

	// Very small width
	m.width = 50
	width = m.calculateStatsWidth()
	require.GreaterOrEqual(t, width, 18)
}

func TestLogsModel_CalculateLogsWidth(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.width = 100

	// Without drawer
	m.drawerOpen = false
	width := m.calculateLogsWidth()
	require.Greater(t, width, 0)

	// With drawer
	m.drawerOpen = true
	widthWithDrawer := m.calculateLogsWidth()
	require.Less(t, widthWithDrawer, width)
}

func TestLogsModel_UpdateDrawerDetail(t *testing.T) {
	m := newLogsModel("/tmp/test.log")
	m.lines = []LogLine{
		{Raw: "line1", Level: "ERROR"},
		{Raw: "line2", Level: "INFO"},
	}
	m.cursor = 0

	m.updateDrawerDetail()
	require.NotNil(t, m.drawerDetail)
	require.Equal(t, "line1", m.drawerDetail.Raw)

	// Invalid cursor
	m.cursor = 100
	m.updateDrawerDetail()
	require.Nil(t, m.drawerDetail)
}

// =========== HELPERS ===========

type mockFileInfo struct {
	size int64
}

func (m *mockFileInfo) Name() string       { return "test.log" }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() any           { return nil }
