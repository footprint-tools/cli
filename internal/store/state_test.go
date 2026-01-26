package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore_GetUpdateCache_Empty(t *testing.T) {
	s := newTestStore(t)

	cache, err := s.GetUpdateCache()
	require.NoError(t, err)
	require.Empty(t, cache.LastCheck)
	require.Empty(t, cache.LatestVersion)
}

func TestStore_SetUpdateCache(t *testing.T) {
	s := newTestStore(t)

	err := s.SetUpdateCache("2024-01-15T10:00:00Z", "v1.2.3")
	require.NoError(t, err)

	cache, err := s.GetUpdateCache()
	require.NoError(t, err)
	require.Equal(t, "2024-01-15T10:00:00Z", cache.LastCheck)
	require.Equal(t, "v1.2.3", cache.LatestVersion)
}

func TestStore_SetUpdateCache_Update(t *testing.T) {
	s := newTestStore(t)

	// Set initial values
	err := s.SetUpdateCache("2024-01-15T10:00:00Z", "v1.2.3")
	require.NoError(t, err)

	// Update values
	err = s.SetUpdateCache("2024-01-16T12:00:00Z", "v1.3.0")
	require.NoError(t, err)

	cache, err := s.GetUpdateCache()
	require.NoError(t, err)
	require.Equal(t, "2024-01-16T12:00:00Z", cache.LastCheck)
	require.Equal(t, "v1.3.0", cache.LatestVersion)
}
