package actions

import (
	"fmt"
	"time"

	"github.com/Skryensya/footprint/internal/app"
	"github.com/Skryensya/footprint/internal/git"
	"github.com/Skryensya/footprint/internal/repo"
)

type actionDependencies struct {
	// git
	GitAvailable func() bool
	RepoRoot     func(string) (string, error)
	OriginURL    func(string) (string, error)

	// repo
	DeriveRepoID func(string, string) (repo.RepoID, error)
	IsTracked    func(repo.RepoID) (bool, error)
	Printf       func(format string, a ...any) (n int, err error)
	Version      func() string

	// telemetry
	// InsertEvent func(string, map[string]string) error

	// misc
	Now func() time.Time
}

func defaultDeps() actionDependencies {
	return actionDependencies{
		GitAvailable: git.IsAvailable,
		RepoRoot:     git.RepoRoot,
		OriginURL:    git.OriginURL,

		DeriveRepoID: repo.DeriveID,
		IsTracked:    repo.IsTracked,

		// InsertEvent: telemetry.InsertEvent,
		Now:     time.Now,
		Printf:  fmt.Printf,
		Version: func() string { return app.Version },
	}
}
