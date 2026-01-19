package tracking

import (
	"sort"

	"github.com/Skryensya/footprint/internal/dispatchers"
)

func Repos(args []string, flags *dispatchers.ParsedFlags) error {
	return repos(args, flags, DefaultDeps())
}

func repos(_ []string, _ *dispatchers.ParsedFlags, deps Deps) error {
	trackedRepos, err := deps.ListTracked()
	if err != nil {
		return err
	}

	if len(trackedRepos) == 0 {
		deps.Println("no tracked repositories")
		return nil
	}

	sort.Slice(trackedRepos, func(i, j int) bool {
		return trackedRepos[i] < trackedRepos[j]
	})

	for _, repoID := range trackedRepos {
		deps.Println(repoID)
	}

	return nil
}
