package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore_AddRepo(t *testing.T) {
	s := newTestStore(t)

	err := s.AddRepo("/path/to/repo")
	require.NoError(t, err)

	repos, err := s.ListRepos()
	require.NoError(t, err)
	require.Len(t, repos, 1)
	require.Equal(t, "/path/to/repo", repos[0].Path)
}

func TestStore_AddRepo_Upsert(t *testing.T) {
	s := newTestStore(t)

	// Add same repo twice - should upsert
	err := s.AddRepo("/path/to/repo")
	require.NoError(t, err)

	err = s.AddRepo("/path/to/repo")
	require.NoError(t, err)

	repos, err := s.ListRepos()
	require.NoError(t, err)
	require.Len(t, repos, 1, "should upsert, not create duplicate")
}

func TestStore_AddRepo_Multiple(t *testing.T) {
	s := newTestStore(t)

	err := s.AddRepo("/path/to/repo1")
	require.NoError(t, err)

	err = s.AddRepo("/path/to/repo2")
	require.NoError(t, err)

	repos, err := s.ListRepos()
	require.NoError(t, err)
	require.Len(t, repos, 2)
}

func TestStore_RemoveRepo(t *testing.T) {
	s := newTestStore(t)

	// Add then remove
	err := s.AddRepo("/path/to/repo")
	require.NoError(t, err)

	err = s.RemoveRepo("/path/to/repo")
	require.NoError(t, err)

	repos, err := s.ListRepos()
	require.NoError(t, err)
	require.Len(t, repos, 0)
}

func TestStore_RemoveRepo_NotFound(t *testing.T) {
	s := newTestStore(t)

	// Remove non-existent repo - should not error
	err := s.RemoveRepo("/nonexistent/repo")
	require.NoError(t, err)
}

func TestStore_ListRepos_Empty(t *testing.T) {
	s := newTestStore(t)

	repos, err := s.ListRepos()
	require.NoError(t, err)
	require.Len(t, repos, 0)
}

func TestStore_ListRepos_Sorted(t *testing.T) {
	s := newTestStore(t)

	// Add in non-alphabetical order
	err := s.AddRepo("/z/repo")
	require.NoError(t, err)
	err = s.AddRepo("/a/repo")
	require.NoError(t, err)
	err = s.AddRepo("/m/repo")
	require.NoError(t, err)

	repos, err := s.ListRepos()
	require.NoError(t, err)
	require.Len(t, repos, 3)
	require.Equal(t, "/a/repo", repos[0].Path)
	require.Equal(t, "/m/repo", repos[1].Path)
	require.Equal(t, "/z/repo", repos[2].Path)
}

func TestStore_ListRepoPaths(t *testing.T) {
	s := newTestStore(t)

	err := s.AddRepo("/path/to/repo1")
	require.NoError(t, err)
	err = s.AddRepo("/path/to/repo2")
	require.NoError(t, err)

	paths, err := s.ListRepoPaths()
	require.NoError(t, err)
	require.Len(t, paths, 2)
	require.Contains(t, paths, "/path/to/repo1")
	require.Contains(t, paths, "/path/to/repo2")
}

func TestStore_ListRepoPaths_Empty(t *testing.T) {
	s := newTestStore(t)

	paths, err := s.ListRepoPaths()
	require.NoError(t, err)
	require.Len(t, paths, 0)
}
