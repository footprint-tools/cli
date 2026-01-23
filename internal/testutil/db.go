package testutil

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/footprint-cli/internal/store"
	"github.com/footprint-tools/footprint-cli/internal/store/migrations"
)

// NewTestDB creates an in-memory SQLite database with migrations applied.
// The database is automatically closed when the test finishes.
func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "failed to open in-memory database")

	t.Cleanup(func() {
		_ = db.Close()
	})

	err = migrations.Run(db)
	require.NoError(t, err, "failed to run migrations")

	return db
}

// SeedEvents inserts a slice of events into the test database.
func SeedEvents(t *testing.T, db *sql.DB, events []store.RepoEvent) {
	t.Helper()

	for _, event := range events {
		err := store.InsertEvent(db, event)
		require.NoError(t, err, "failed to seed event: %+v", event)
	}
}
