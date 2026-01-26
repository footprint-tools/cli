package tracking

import (
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
	"github.com/footprint-tools/footprint-cli/internal/store"
)

// Repos lists tracked repositories, or launches interactive mode.
func Repos(args []string, flags *dispatchers.ParsedFlags) error {
	if flags.Has("-i") || flags.Has("--interactive") {
		return ReposInteractive(args, flags)
	}
	return reposList(reposDeps{
		DBPath:    store.DBPath,
		OpenStore: store.New,
		Println:   defaultPrintln,
	})
}

type reposDeps struct {
	DBPath    func() string
	OpenStore func(string) (*store.Store, error)
	Println   func(...any) (int, error)
}

func reposList(deps reposDeps) error {
	s, err := deps.OpenStore(deps.DBPath())
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	repos, err := s.ListRepos()
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		_, _ = deps.Println("no tracked repositories")
		_, _ = deps.Println("run 'fp setup' in a repo to install hooks")
		return nil
	}

	for _, r := range repos {
		_, _ = deps.Println(r.Path)
	}

	return nil
}

func defaultPrintln(args ...any) (int, error) {
	return DefaultDeps().Println(args...)
}
