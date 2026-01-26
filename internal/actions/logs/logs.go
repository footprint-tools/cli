package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/ui/style"
)

const defaultLogLimit = 50

// View shows the last N lines of the log file
func View(args []string, flags *dispatchers.ParsedFlags) error {
	return view(args, flags, DefaultDeps())
}

func view(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	logPath := deps.LogFilePath()

	// Check if log file exists
	info, err := deps.Stat(logPath)
	if os.IsNotExist(err) {
		_, _ = deps.Println(style.Muted("No log file found at " + logPath))
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat log file: %w", err)
	}

	if info.Size() == 0 {
		_, _ = deps.Println(style.Muted("Log file is empty"))
		return nil
	}

	// Read the entire file
	content, err := deps.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("read log file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Remove empty trailing line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Get limit from flags
	limit := flags.Int("--limit", defaultLogLimit)
	if limit <= 0 {
		limit = defaultLogLimit
	}

	// Take last N lines
	start := 0
	if len(lines) > limit {
		start = len(lines) - limit
	}

	for _, line := range lines[start:] {
		_, _ = deps.Println(colorizeLogLine(line))
	}

	return nil
}

// Tail follows the log file in real time
func Tail(args []string, flags *dispatchers.ParsedFlags) error {
	return tail(args, flags, DefaultDeps())
}

func tail(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	logPath := deps.LogFilePath()

	file, err := deps.OpenFile(logPath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Seek to end of file
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("seek log file: %w", err)
	}

	_, _ = deps.Println(style.Muted("Following logs at " + logPath + " (Ctrl+C to stop)"))
	_, _ = deps.Println("")

	// Setup signal handling for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("read log file: %w", err)
				}
				// EOF - wait for next tick
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					continue
				}
			}

			// Print the line (without extra newline since ReadString includes it)
			fmt.Print(colorizeLogLine(strings.TrimSuffix(line, "\n")) + "\n")
		}
	}
}

// Clear empties the log file
func Clear(args []string, flags *dispatchers.ParsedFlags) error {
	return clear(args, flags, DefaultDeps())
}

func clear(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	logPath := deps.LogFilePath()

	// Truncate the file (or create empty if it doesn't exist)
	err := deps.WriteFile(logPath, []byte{}, 0600)
	if err != nil {
		return fmt.Errorf("clear log file: %w", err)
	}

	_, _ = deps.Println(style.Success("Log file cleared"))
	return nil
}

// colorizeLogLine adds color to log lines based on level
// Supports both old format: [timestamp] LEVEL: message
// And new format: [timestamp] LEVEL file:line: message
func colorizeLogLine(line string) string {
	// Check for level after the timestamp bracket
	if strings.Contains(line, "] ERROR ") || strings.Contains(line, "] ERROR:") {
		return style.Error(line)
	}
	if strings.Contains(line, "] WARN ") || strings.Contains(line, "] WARN:") {
		return style.Warning(line)
	}
	if strings.Contains(line, "] INFO ") || strings.Contains(line, "] INFO:") {
		return style.Info(line)
	}
	if strings.Contains(line, "] DEBUG ") || strings.Contains(line, "] DEBUG:") {
		return style.Muted(line)
	}
	return line
}
