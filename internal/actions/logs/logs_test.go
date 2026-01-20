package logs

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Skryensya/footprint/internal/dispatchers"
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
