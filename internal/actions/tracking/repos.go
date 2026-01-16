package tracking

import (
	"sort"
)

func Repos(args []string, flags []string) error {
	return repos(args, flags, DefaultDeps())
}

func repos(_ []string, _ []string, deps Deps) error {
	repos, err := deps.ListTracked()
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		deps.Println("no tracked repositories")
		return nil
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i] < repos[j]
	})

	for _, r := range repos {
		deps.Println(r)
	}

	return nil
}
