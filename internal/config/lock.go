package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	lockFileName     = ".fprc.lock"
	lockTimeout      = 5 * time.Second
	staleLockTimeout = 30 * time.Second
	lockPollInterval = 50 * time.Millisecond
)

var (
	// ErrLockTimeout is returned when we cannot acquire the lock within the timeout period.
	ErrLockTimeout = errors.New("config: lock timeout")
)

// WithLock executes the given function while holding a file lock on the config.
// This prevents race conditions when multiple processes try to modify the config simultaneously.
func WithLock(fn func() error) error {
	lockPath, err := getLockPath()
	if err != nil {
		return err
	}

	// Try to acquire the lock
	lockFile, err := acquireLock(lockPath)
	if err != nil {
		return err
	}

	// Ensure we release the lock when done
	defer releaseLock(lockFile, lockPath)

	// Execute the function
	return fn()
}

// getLockPath returns the path to the lock file.
func getLockPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, lockFileName), nil
}

// acquireLock attempts to acquire an exclusive lock on the config file.
// It will retry until the timeout is reached.
func acquireLock(lockPath string) (*os.File, error) {
	deadline := time.Now().Add(lockTimeout)

	for {
		// Check for stale lock first
		if info, err := os.Stat(lockPath); err == nil {
			// Lock file exists - check if it's stale
			if time.Since(info.ModTime()) > staleLockTimeout {
				// Lock is stale, remove it
				_ = os.Remove(lockPath)
			}
		}

		// Try to create the lock file exclusively
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			// Write our PID to the lock file for debugging
			_, _ = f.WriteString(strconv.Itoa(os.Getpid()))
			return f, nil
		}

		// Check if we've exceeded the deadline
		if time.Now().After(deadline) {
			return nil, ErrLockTimeout
		}

		// Wait before retrying
		time.Sleep(lockPollInterval)
	}
}

// releaseLock releases the file lock.
func releaseLock(f *os.File, lockPath string) {
	if f != nil {
		_ = f.Close()
	}
	_ = os.Remove(lockPath)
}
