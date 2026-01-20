package tracking

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Skryensya/footprint/internal/dispatchers"
	"github.com/Skryensya/footprint/internal/repo"
)

func TestRepos_EmptyList(t *testing.T) {
	var capturedOutput []string

	deps := Deps{
		ListTracked: func() ([]repo.RepoID, error) {
			return []repo.RepoID{}, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				// Can be either string or repo.RepoID
				switch v := a[0].(type) {
				case string:
					capturedOutput = append(capturedOutput, v)
				case repo.RepoID:
					capturedOutput = append(capturedOutput, string(v))
				}
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := repos([]string{}, flags, deps)

	require.NoError(t, err)
	require.Len(t, capturedOutput, 1)
	require.Equal(t, "no tracked repositories", capturedOutput[0])
}

func TestRepos_SingleRepo(t *testing.T) {
	var capturedOutput []string

	deps := Deps{
		ListTracked: func() ([]repo.RepoID, error) {
			return []repo.RepoID{"github.com/user/repo"}, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				// Can be either string or repo.RepoID
				switch v := a[0].(type) {
				case string:
					capturedOutput = append(capturedOutput, v)
				case repo.RepoID:
					capturedOutput = append(capturedOutput, string(v))
				}
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := repos([]string{}, flags, deps)

	require.NoError(t, err)
	require.Len(t, capturedOutput, 1)
	require.Equal(t, "github.com/user/repo", capturedOutput[0])
}

func TestRepos_MultipleReposSorted(t *testing.T) {
	var capturedOutput []string

	deps := Deps{
		ListTracked: func() ([]repo.RepoID, error) {
			// Return unsorted list
			return []repo.RepoID{
				"github.com/zebra/repo",
				"github.com/alpha/repo",
				"github.com/beta/repo",
			}, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				// Can be either string or repo.RepoID
				switch v := a[0].(type) {
				case string:
					capturedOutput = append(capturedOutput, v)
				case repo.RepoID:
					capturedOutput = append(capturedOutput, string(v))
				}
			}
			return 0, nil
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := repos([]string{}, flags, deps)

	require.NoError(t, err)
	require.Len(t, capturedOutput, 3)
	// Should be sorted alphabetically
	require.Equal(t, "github.com/alpha/repo", capturedOutput[0])
	require.Equal(t, "github.com/beta/repo", capturedOutput[1])
	require.Equal(t, "github.com/zebra/repo", capturedOutput[2])
}

func TestRepos_ListTrackedError(t *testing.T) {
	deps := Deps{
		ListTracked: func() ([]repo.RepoID, error) {
			return nil, errors.New("failed to list tracked repos")
		},
	}

	flags := dispatchers.NewParsedFlags([]string{})
	err := repos([]string{}, flags, deps)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list tracked repos")
}

func TestRepos_IgnoresArgsAndFlags(t *testing.T) {
	var capturedOutput []string

	deps := Deps{
		ListTracked: func() ([]repo.RepoID, error) {
			return []repo.RepoID{"github.com/user/repo"}, nil
		},
		Println: func(a ...any) (int, error) {
			if len(a) > 0 {
				// Can be either string or repo.RepoID
				switch v := a[0].(type) {
				case string:
					capturedOutput = append(capturedOutput, v)
				case repo.RepoID:
					capturedOutput = append(capturedOutput, string(v))
				}
			}
			return 0, nil
		},
	}

	// Should ignore args and flags
	flags := dispatchers.NewParsedFlags([]string{"--some-flag"})
	err := repos([]string{"arg1", "arg2"}, flags, deps)

	require.NoError(t, err)
	require.Len(t, capturedOutput, 1)
	require.Equal(t, "github.com/user/repo", capturedOutput[0])
}
