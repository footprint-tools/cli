package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/footprint-tools/cli/internal/domain"
)

// Level represents the logging severity level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger handles thread-safe file logging.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	minLevel Level
}

var (
	defaultLogger   *Logger
	defaultLoggerMu sync.RWMutex
	once            sync.Once
)

// Init initializes the global logger with the specified file.
func Init(logPath string, minLevel Level) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(logPath, minLevel)
	})
	return err
}

// New creates a new logger that writes to the specified file.
func New(logPath string, minLevel Level) (*Logger, error) {
	// Create directory if it doesn't exist with restrictive permissions
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	// Check if file exists and fix permissions if needed (before opening)
	if info, err := os.Stat(logPath); err == nil {
		// File exists - check permissions
		if info.Mode().Perm() != 0600 {
			if err := os.Chmod(logPath, 0600); err != nil {
				return nil, fmt.Errorf("chmod existing log file: %w", err)
			}
		}
	}

	// Open log file with restrictive permissions
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return &Logger{
		file:     file,
		minLevel: minLevel,
	}, nil
}

// Close closes the logger.
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// log writes a message with the specified level.
func (l *Logger) log(level Level, format string, args ...any) {
	l.logWithCaller(level, 3, format, args...)
}

// logWithCaller writes a message with caller information.
func (l *Logger) logWithCaller(level Level, skip int, format string, args ...any) {
	if l == nil || level < l.minLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	// Get caller information
	caller := "unknown"
	if _, file, line, ok := runtime.Caller(skip); ok {
		// Get just the filename, not full path
		file = filepath.Base(file)
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	logLine := fmt.Sprintf("[%s] %s %s: %s\n", timestamp, level.String(), caller, message)

	if _, err := l.file.Write([]byte(logLine)); err != nil {
		// Can't log to file, output to stderr for critical messages
		if level >= LevelError {
			fmt.Fprintf(os.Stderr, "logger: write failed: %v (message: %s)\n", err, message)
		}
	}
}

// Debug writes a debug message.
func (l *Logger) Debug(format string, args ...any) {
	l.log(LevelDebug, format, args...)
}

// Info writes an informational message.
func (l *Logger) Info(format string, args ...any) {
	l.log(LevelInfo, format, args...)
}

// Warn writes a warning message.
func (l *Logger) Warn(format string, args ...any) {
	l.log(LevelWarn, format, args...)
}

// Error writes an error message.
func (l *Logger) Error(format string, args ...any) {
	l.log(LevelError, format, args...)
}

// Package-level convenience functions for the global logger.

// Debug writes a debug message to the global logger.
func Debug(format string, args ...any) {
	defaultLoggerMu.RLock()
	l := defaultLogger
	defaultLoggerMu.RUnlock()
	if l != nil {
		l.Debug(format, args...)
	}
}

// Info writes an informational message to the global logger.
func Info(format string, args ...any) {
	defaultLoggerMu.RLock()
	l := defaultLogger
	defaultLoggerMu.RUnlock()
	if l != nil {
		l.Info(format, args...)
	}
}

// Warn writes a warning message to the global logger.
func Warn(format string, args ...any) {
	defaultLoggerMu.RLock()
	l := defaultLogger
	defaultLoggerMu.RUnlock()
	if l != nil {
		l.Warn(format, args...)
	}
}

// Error writes an error message to the global logger.
func Error(format string, args ...any) {
	defaultLoggerMu.RLock()
	l := defaultLogger
	defaultLoggerMu.RUnlock()
	if l != nil {
		l.Error(format, args...)
	}
}

// Close closes the global logger.
func Close() error {
	defaultLoggerMu.RLock()
	l := defaultLogger
	defaultLoggerMu.RUnlock()
	if l != nil {
		return l.Close()
	}
	return nil
}

// NopLogger is a logger that discards all messages.
// Useful for testing or when logging is disabled.
type NopLogger struct{}

func (NopLogger) Debug(_ string, _ ...any) {}
func (NopLogger) Info(_ string, _ ...any)  {}
func (NopLogger) Warn(_ string, _ ...any)  {}
func (NopLogger) Error(_ string, _ ...any) {}
func (NopLogger) Close() error             { return nil }

// Verify Logger implements domain.Logger
var _ domain.Logger = (*Logger)(nil)
var _ domain.Logger = NopLogger{}
