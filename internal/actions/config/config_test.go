package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/cli/internal/dispatchers"
)

// =========== GET TESTS ===========

func TestGet_Success(t *testing.T) {
	var capturedValue string
	deps := Deps{
		Get: func(key string) (string, bool) {
			if key == "theme" {
				return "dark", true
			}
			return "", false
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				capturedValue, _ = a[0].(string)
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := get([]string{"theme"}, flags, deps)

	require.NoError(t, err)
	require.Equal(t, "dark", capturedValue)
}

func TestGet_MissingKey(t *testing.T) {
	deps := Deps{}

	flags := dispatchers.NewParsedFlags([]string{})
	err := get([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "key")
}

func TestGet_KeyNotFound(t *testing.T) {
	deps := Deps{
		Get: func(key string) (string, bool) {
			return "", false
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := get([]string{"nonexistent"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "nonexistent")
}

// =========== SET TESTS ===========

func TestSet_AddNew(t *testing.T) {
	var capturedPrintf string
	var writtenLines []string
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{}, nil
		},
		Set: func(lines []string, key, value string) ([]string, bool) {
			return append(lines, key+"="+value), false // Not updated (new)
		},
		WriteLines: func(lines []string) error {
			writtenLines = lines
			return nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = fmt.Sprintf(format, a...)
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := set([]string{"theme", "dark"}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "added")
	require.Len(t, writtenLines, 1)
	require.Equal(t, "theme=dark", writtenLines[0])
}

func TestSet_UpdateExisting(t *testing.T) {
	var capturedPrintf string
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{"theme=light"}, nil
		},
		Set: func(lines []string, key, value string) ([]string, bool) {
			return []string{"theme=" + value}, true // Updated
		},
		WriteLines: func(lines []string) error {
			return nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = fmt.Sprintf(format, a...)
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := set([]string{"theme", "dark"}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "updated")
}

func TestSet_MissingArguments(t *testing.T) {
	deps := Deps{}

	flags := dispatchers.NewParsedFlags([]string{})

	// No arguments
	err := set([]string{}, flags, deps)
	require.Error(t, err)

	// Only key, no value
	err = set([]string{"theme"}, flags, deps)
	require.Error(t, err)
}

func TestSet_ReadLinesError(t *testing.T) {
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return nil, errors.New("cannot read config")
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := set([]string{"theme", "dark"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read config")
}

func TestSet_WriteLinesError(t *testing.T) {
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{}, nil
		},
		Set: func(lines []string, key, value string) ([]string, bool) {
			return append(lines, key+"="+value), false
		},
		WriteLines: func(lines []string) error {
			return errors.New("cannot write config")
		},
		Printf: func(format string, a ...any) (int, error) {
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := set([]string{"theme", "dark"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot write config")
}

// =========== UNSET TESTS ===========

func TestUnset_Success(t *testing.T) {
	var capturedPrintf string
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{"theme=dark", "other=value"}, nil
		},
		Unset: func(lines []string, key string) ([]string, bool) {
			return []string{"other=value"}, true // Removed
		},
		WriteLines: func(lines []string) error {
			return nil
		},
		Printf: func(format string, a ...any) (int, error) {
			capturedPrintf = format
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := unset([]string{"theme"}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintf, "unset")
}

func TestUnset_KeyNotFound(t *testing.T) {
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{"other=value"}, nil
		},
		Unset: func(lines []string, key string) ([]string, bool) {
			return lines, false // Not removed (doesn't exist)
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := unset([]string{"theme"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "theme")
}

func TestUnset_MissingKey(t *testing.T) {
	deps := Deps{}

	flags := dispatchers.NewParsedFlags([]string{})
	err := unset([]string{}, flags, deps)

	require.Error(t, err)
}

func TestUnset_AllFlag(t *testing.T) {
	var capturedPrintln string
	var writtenLines []string
	deps := Deps{
		WriteLines: func(lines []string) error {
			writtenLines = lines
			return nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				capturedPrintln, _ = a[0].(string)
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--all"})
	err := unset([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, capturedPrintln, "all config entries removed")
	require.Empty(t, writtenLines)
}

func TestUnset_AllFlagWithArgs(t *testing.T) {
	deps := Deps{}

	flags := dispatchers.NewParsedFlags([]string{"--all"})
	err := unset([]string{"theme"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "--all does not take arguments")
}

func TestUnset_AllFlagWriteError(t *testing.T) {
	deps := Deps{
		WriteLines: func(lines []string) error {
			return errors.New("cannot write config")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--all"})
	err := unset([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot write config")
}

func TestUnset_ReadLinesError(t *testing.T) {
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return nil, errors.New("cannot read config")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := unset([]string{"theme"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read config")
}

func TestUnset_WriteLinesError(t *testing.T) {
	deps := Deps{
		ReadLines: func() ([]string, error) {
			return []string{"theme=dark"}, nil
		},
		Unset: func(lines []string, key string) ([]string, bool) {
			return []string{}, true
		},
		WriteLines: func(lines []string) error {
			return errors.New("cannot write config")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := unset([]string{"theme"}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot write config")
}

// =========== LIST TESTS ===========

func TestList_Success(t *testing.T) {
	var printedLines []string
	deps := Deps{
		GetAll: func() (map[string]string, error) {
			return map[string]string{
				"theme":     "neon",
				"log_level": "info",
			}, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			printedLines = append(printedLines, fmt.Sprintf(format, a...))
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := list([]string{}, flags, deps)

	require.NoError(t, err)
	// Should show visible keys (HideIfEmpty keys are hidden when not set)
	require.Len(t, printedLines, 8) // 8 always-visible keys
}

func TestList_ShowsDefaults(t *testing.T) {
	var printedLines []string
	deps := Deps{
		GetAll: func() (map[string]string, error) {
			return map[string]string{}, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			printedLines = append(printedLines, fmt.Sprintf(format, a...))
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := list([]string{}, flags, deps)

	require.NoError(t, err)
	// Should show visible keys with defaults (HideIfEmpty keys are hidden)
	require.Len(t, printedLines, 8)
}

func TestList_ShowsColorOverridesWhenSet(t *testing.T) {
	var printedLines []string
	deps := Deps{
		GetAll: func() (map[string]string, error) {
			return map[string]string{
				"color_success": "46",
				"color_error":   "196",
			}, nil
		},
		Printf: func(format string, a ...any) (int, error) {
			printedLines = append(printedLines, fmt.Sprintf(format, a...))
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := list([]string{}, flags, deps)

	require.NoError(t, err)
	// 8 always-visible + 2 color overrides that are set
	require.Len(t, printedLines, 10)
}

func TestList_GetAllError(t *testing.T) {
	deps := Deps{
		GetAll: func() (map[string]string, error) {
			return nil, errors.New("cannot read config")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := list([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read config")
}

func TestList_JSON(t *testing.T) {
	var printedOutput string
	deps := Deps{
		GetAll: func() (map[string]string, error) {
			return map[string]string{
				"theme": "neon",
			}, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printedOutput, _ = a[0].(string)
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{"--json"})
	err := list([]string{}, flags, deps)

	require.NoError(t, err)
	require.Contains(t, printedOutput, `"key":`)
	require.Contains(t, printedOutput, `"value":`)
}
