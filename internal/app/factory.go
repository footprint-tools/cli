package app

import (
	"github.com/footprint-tools/footprint-cli/internal/config"
	"github.com/footprint-tools/footprint-cli/internal/domain"
	"github.com/footprint-tools/footprint-cli/internal/git"
	"github.com/footprint-tools/footprint-cli/internal/hooks"
	"github.com/footprint-tools/footprint-cli/internal/log"
	"github.com/footprint-tools/footprint-cli/internal/paths"
	"github.com/footprint-tools/footprint-cli/internal/repo"
	"github.com/footprint-tools/footprint-cli/internal/store"
	"github.com/footprint-tools/footprint-cli/internal/ui"
	"github.com/footprint-tools/footprint-cli/internal/ui/style"
)

// Options configures the application factory.
type Options struct {
	// Pager options
	PagerDisabled bool
	PagerOverride string

	// Log options
	LogEnabled bool

	// Style options
	StyleEnabled bool
	StyleConfig  map[string]string
}

// DefaultOptions returns the default application options.
func DefaultOptions() Options {
	logEnabled, _ := config.Get("enable_log")
	styleConfig, _ := config.GetAll()

	return Options{
		LogEnabled:   logEnabled == "true",
		StyleEnabled: true,
		StyleConfig:  styleConfig,
	}
}

// New creates a new Application with all dependencies wired up.
func New(opts Options) (*domain.Application, error) {
	// Initialize logger (always at debug level - log everything)
	var logger domain.Logger
	if opts.LogEnabled {
		logPath := paths.LogFilePath()
		l, err := log.New(logPath, log.LevelDebug)
		if err != nil {
			// Fall back to NopLogger on error
			logger = log.NopLogger{}
		} else {
			logger = l
		}
	} else {
		logger = log.NopLogger{}
	}

	// Initialize store
	dbPath := store.DBPath()
	eventStore, err := store.New(dbPath)
	if err != nil {
		return nil, err
	}

	// Initialize style
	style.Init(opts.StyleEnabled, opts.StyleConfig)

	// Create output writer with options
	var writerOpts []ui.WriterOption
	if opts.PagerDisabled {
		writerOpts = append(writerOpts, ui.WithPagerDisabled())
	}
	if opts.PagerOverride != "" {
		writerOpts = append(writerOpts, ui.WithPagerOverride(opts.PagerOverride))
	}
	writerOpts = append(writerOpts, ui.WithConfigGetter(config.Get))

	return &domain.Application{
		Git:    git.NewProvider(),
		Repo:   repo.NewDeriver(),
		Store:  eventStore,
		Config: config.NewProvider(),
		Logger: logger,
		Output: ui.NewWriter(writerOpts...),
		Styler: style.NewStyler(),
		Hooks:  hooks.NewManager(),
	}, nil
}

// NewForTesting creates an Application suitable for testing.
// Uses in-memory store, NopLogger, and no styling.
func NewForTesting() *domain.Application {
	return &domain.Application{
		Git:    git.NewProvider(),
		Repo:   repo.NewDeriver(),
		Store:  store.NewWithDB(nil), // nil DB for testing - callers should provide their own
		Config: config.NewProvider(),
		Logger: log.NopLogger{},
		Output: ui.NewWriter(ui.WithPagerDisabled()),
		Styler: style.NopStyler{},
		Hooks:  hooks.NewManager(),
	}
}

// Close cleans up application resources.
func Close(app *domain.Application) error {
	if app.Logger != nil {
		_ = app.Logger.Close()
	}
	if app.Store != nil {
		_ = app.Store.Close()
	}
	return nil
}
