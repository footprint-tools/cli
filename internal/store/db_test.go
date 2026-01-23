package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenFresh(t *testing.T) {
	t.Run("creates new database", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := OpenFresh(dbPath)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func() { _ = db.Close() }()

		// Verify database file was created
		_, err = os.Stat(dbPath)
		require.NoError(t, err)

		// Verify migrations were run (check for tables)
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='repo_events'").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "repo_events table should exist")
	})

	t.Run("creates in-memory database", func(t *testing.T) {
		db, err := OpenFresh(":memory:")
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func() { _ = db.Close() }()

		// Verify migrations were run
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='repo_events'").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("sets file permissions", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := OpenFresh(dbPath)
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		// Check file permissions
		info, err := os.Stat(dbPath)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})
}

func TestOpen_Singleton(t *testing.T) {
	// Reset singleton before test
	ResetSingleton()
	defer ResetSingleton()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "singleton.db")

	// First call creates the database
	db1, err := Open(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db1)

	// Second call returns the same instance
	db2, err := Open(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db2)

	// Verify they are the same instance (same pointer)
	require.Same(t, db1, db2)

	// Third call with different path still returns the same instance
	// (singleton pattern ignores the new path)
	db3, err := Open(filepath.Join(dir, "different.db"))
	require.NoError(t, err)
	require.Same(t, db1, db3)
}

func TestResetSingleton(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "reset-test.db")

	// Open database
	db1, err := Open(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db1)

	// Reset singleton
	ResetSingleton()

	// Next call should create a new instance
	db2, err := Open(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db2)

	// Verify they are different instances
	require.NotSame(t, db1, db2)

	// Cleanup
	_ = db1.Close()
	_ = db2.Close()
	ResetSingleton()
}

func TestOpen_RunsMigrations(t *testing.T) {
	ResetSingleton()
	defer ResetSingleton()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "migrations-test.db")

	db, err := Open(dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Verify migrations were run by checking for tables
	tables := []string{"schema_migrations", "event_status", "event_source", "repo_events"}
	for _, table := range tables {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		require.NoError(t, err, "table %s should exist", table)
		require.Equal(t, table, name)
	}
}

func TestOpen_ErrorHandling(t *testing.T) {
	ResetSingleton()
	defer ResetSingleton()

	// Try to open database in non-existent directory
	invalidPath := "/nonexistent/path/to/database.db"

	_, err := Open(invalidPath)
	require.Error(t, err, "should error when path is invalid")
}
