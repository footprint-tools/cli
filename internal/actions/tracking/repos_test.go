package tracking

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/footprint-tools/footprint-cli/internal/store"
	"github.com/footprint-tools/footprint-cli/internal/store/migrations"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	err = migrations.Run(db)
	require.NoError(t, err)

	return store.NewWithDB(db)
}

func TestReposList_Empty(t *testing.T) {
	s := newTestStore(t)
	var printedLines []string

	deps := reposDeps{
		DBPath: func() string { return ":memory:" },
		OpenStore: func(_ string) (*store.Store, error) {
			return s, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printedLines = append(printedLines, a[0].(string))
			}
			return 0, nil
		},
	}

	err := reposList(deps)
	require.NoError(t, err)
	require.Len(t, printedLines, 2)
	require.Contains(t, printedLines[0], "no tracked repositories")
	require.Contains(t, printedLines[1], "fp setup")
}

func TestReposList_WithRepos(t *testing.T) {
	s := newTestStore(t)

	// Add some repos
	require.NoError(t, s.AddRepo("/path/to/repo1"))
	require.NoError(t, s.AddRepo("/path/to/repo2"))

	var printedLines []string

	deps := reposDeps{
		DBPath: func() string { return ":memory:" },
		OpenStore: func(_ string) (*store.Store, error) {
			return s, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				printedLines = append(printedLines, a[0].(string))
			}
			return 0, nil
		},
	}

	err := reposList(deps)
	require.NoError(t, err)
	require.Len(t, printedLines, 2)
	require.Equal(t, "/path/to/repo1", printedLines[0])
	require.Equal(t, "/path/to/repo2", printedLines[1])
}

func TestReposList_OpenStoreError(t *testing.T) {
	deps := reposDeps{
		DBPath: func() string { return "/invalid/path" },
		OpenStore: func(_ string) (*store.Store, error) {
			return nil, errors.New("failed to open store")
		},
		Println: func(a ...any) (int, error) {
			return 0, nil
		},
	}

	err := reposList(deps)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open store")
}
