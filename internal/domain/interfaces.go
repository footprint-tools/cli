package domain

import (
	"io"
)

// GitProvider defines operations for interacting with git.
type GitProvider interface {
	// IsAvailable checks if git is installed and accessible.
	IsAvailable() bool

	// RepoRoot returns the root directory of the git repository containing the given path.
	RepoRoot(path string) (string, error)

	// OriginURL returns the URL of the 'origin' remote.
	OriginURL(repoRoot string) (string, error)

	// ListRemotes returns a list of configured remote names.
	ListRemotes(repoRoot string) ([]string, error)

	// GetRemoteURL returns the URL for a specific remote.
	GetRemoteURL(repoRoot, remoteName string) (string, error)

	// HeadCommit returns the current HEAD commit hash.
	HeadCommit() (string, error)

	// CurrentBranch returns the current branch name.
	CurrentBranch() (string, error)

	// CommitMessage returns the most recent commit message.
	CommitMessage() (string, error)

	// CommitAuthor returns the author of the most recent commit.
	CommitAuthor() (string, error)
}

// RepoIDDeriver defines operations for deriving repository IDs.
type RepoIDDeriver interface {
	// DeriveID derives a repository ID from a remote URL or local path.
	DeriveID(remoteURL, localPath string) (RepoID, error)
}

// EventStore defines operations for storing and retrieving events.
type EventStore interface {
	// Insert adds a new event to the store.
	Insert(event RepoEvent) error

	// List returns events matching the given filter.
	List(filter EventFilter) ([]RepoEvent, error)

	// GetPending returns all events with pending status.
	GetPending() ([]RepoEvent, error)

	// UpdateStatus updates the status of multiple events.
	UpdateStatus(ids []int64, status EventStatus) error

	// GetMaxID returns the highest event ID.
	GetMaxID() (int64, error)

	// ListSince returns events with ID greater than the given ID.
	ListSince(id int64) ([]RepoEvent, error)

	// MigrateRepoID changes the repo ID for all pending events.
	MigrateRepoID(oldID, newID RepoID) (int64, error)

	// MarkOrphaned marks all pending events for a repo as orphaned.
	MarkOrphaned(repoID RepoID) (int64, error)

	// DeleteOrphaned deletes all orphaned events.
	DeleteOrphaned() (int64, error)

	// CountOrphaned returns the count of orphaned events.
	CountOrphaned() (int64, error)

	// Close closes the store connection.
	Close() error
}

// ConfigProvider defines operations for reading and writing configuration.
type ConfigProvider interface {
	// Get returns the value for a configuration key.
	Get(key string) (string, bool)

	// GetAll returns all configuration values.
	GetAll() (map[string]string, error)

	// Set sets a configuration value.
	Set(key, value string) error

	// Unset removes a configuration value.
	Unset(key string) error
}

// Logger defines logging operations.
type Logger interface {
	// Debug logs a debug message.
	Debug(format string, args ...any)

	// Info logs an info message.
	Info(format string, args ...any)

	// Warn logs a warning message.
	Warn(format string, args ...any)

	// Error logs an error message.
	Error(format string, args ...any)

	// Close closes the logger.
	Close() error
}

// OutputWriter defines output operations.
type OutputWriter interface {
	io.Writer

	// Printf formats and prints to the output.
	Printf(format string, args ...any) (int, error)

	// Println prints a line to the output.
	Println(args ...any) (int, error)

	// Pager displays content through a pager if appropriate.
	Pager(content string)
}

// Styler defines text styling operations.
type Styler interface {
	// Enabled returns true if styling is enabled.
	Enabled() bool

	// Success styles text as success.
	Success(text string) string

	// Warning styles text as warning.
	Warning(text string) string

	// Error styles text as error.
	Error(text string) string

	// Info styles text as info.
	Info(text string) string

	// Muted styles text as muted.
	Muted(text string) string

	// Header styles text as header.
	Header(text string) string
}

// HooksManager defines operations for managing git hooks.
type HooksManager interface {
	// Status returns the installation status of hooks at the given path.
	Status(hooksPath string) (HooksStatus, error)

	// Install installs hooks at the given path.
	Install(hooksPath string) error

	// Uninstall removes hooks from the given path.
	Uninstall(hooksPath string) error
}

// HooksStatus represents the installation status of hooks.
type HooksStatus struct {
	Installed   bool
	HooksPath   string
	ManagedHooks []string
}

// Application represents the main application context with all dependencies.
type Application struct {
	Git     GitProvider
	Repo    RepoIDDeriver
	Store   EventStore
	Config  ConfigProvider
	Logger  Logger
	Output  OutputWriter
	Styler  Styler
	Hooks   HooksManager
}
