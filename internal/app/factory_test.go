package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	// StyleEnabled should be true by default
	require.True(t, opts.StyleEnabled)
}

func TestNewForTesting(t *testing.T) {
	app := NewForTesting()

	require.NotNil(t, app.Git)
	require.NotNil(t, app.Repo)
	require.NotNil(t, app.Config)
	require.NotNil(t, app.Logger)
	require.NotNil(t, app.Output)
	require.NotNil(t, app.Styler)
	require.NotNil(t, app.Hooks)
}

func TestClose_NilComponents(t *testing.T) {
	app := NewForTesting()
	// Store is intentionally nil in test app
	app.Store = nil

	// Should not panic
	err := Close(app)
	require.NoError(t, err)
}

func TestClose_WithLogger(t *testing.T) {
	app := NewForTesting()
	app.Logger = nil

	// Should not panic with nil logger
	err := Close(app)
	require.NoError(t, err)
}

func TestNew_WithOptions(t *testing.T) {
	// Create temp dir for the test
	dir := t.TempDir()

	// Test with various options
	opts := Options{
		PagerDisabled: true,
		PagerOverride: "less",
		LogEnabled:    false,
		StyleEnabled:  true,
		StyleConfig:   map[string]string{"theme": "dark"},
	}

	// Change to temp dir to avoid polluting the real store
	t.Setenv("XDG_CONFIG_HOME", dir)

	app, err := New(opts)
	require.NoError(t, err)
	require.NotNil(t, app)

	defer func() { _ = Close(app) }()

	require.NotNil(t, app.Git)
	require.NotNil(t, app.Repo)
	require.NotNil(t, app.Store)
	require.NotNil(t, app.Config)
	require.NotNil(t, app.Logger)
	require.NotNil(t, app.Output)
	require.NotNil(t, app.Styler)
	require.NotNil(t, app.Hooks)
}

func TestNew_WithLogEnabled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	opts := Options{
		LogEnabled:   true,
		StyleEnabled: false,
	}

	app, err := New(opts)
	require.NoError(t, err)
	require.NotNil(t, app)

	defer func() { _ = Close(app) }()

	require.NotNil(t, app.Logger)
}
